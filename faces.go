package main

// #include <stdio.h>
// #include <stdlib.h>
// #cgo LDFLAGS: -lmvnc
// #include <mvnc.h>
import "C"

import (
  "fmt"
  "io/ioutil"
  "unsafe"
  "log"
)

type ImageRequest struct {
	Image []byte
	Response chan<- []byte
}

func ProcessImage(graphFile string, input <-chan ImageRequest) error {
	var deviceHandle *C.struct_ncDeviceHandle_t
	var graphHandle *C.struct_ncGraphHandle_t

	if ret := C.ncDeviceCreate(0, &deviceHandle); ret != 0 {
		return fmt.Errorf("could not get device name, error code: %v", ret)
	}
	defer C.ncDeviceDestroy(&deviceHandle)

	if ret := C.ncDeviceOpen(deviceHandle); ret != 0 {
		return fmt.Errorf("could not open device: %v", ret)
	}
	defer C.ncDeviceClose(deviceHandle)

	if ret := C.ncGraphCreate(C.CString("faces"), &graphHandle); ret != 0 {
		return fmt.Errorf("could not create graph, %v", ret)
	}
	defer C.ncGraphDestroy(&graphHandle)

	var inputFifo, outputFifo *C.struct_ncFifoHandle_t

	if b, err := ioutil.ReadFile(graphFile); err != nil {
		return err
	} else if ret := C.ncGraphAllocateWithFifos(deviceHandle, graphHandle, unsafe.Pointer(&b[0]), C.uint(len(b)), &inputFifo, &outputFifo); ret != 0 {
		return fmt.Errorf("error allocating graph: %v", ret)
	}

	defer C.ncFifoDestroy(&inputFifo)
	defer C.ncFifoDestroy(&outputFifo)

	fifoOutputSize := C.uint(0)
	optionDataLen := C.uint(4)

	C.ncFifoGetOption(outputFifo, C.NC_RO_FIFO_ELEMENT_DATA_SIZE, unsafe.Pointer(&fifoOutputSize), &optionDataLen)

	log.Printf("fifo output size: %d", fifoOutputSize)

	for {
		bout := make([]byte, fifoOutputSize)
		user := unsafe.Pointer(nil)
		select {
		case req, ok := <-input:
			blen := C.uint(len(req.Image))
			if !ok {
				log.Printf("error reading from input channel")
				break
			} else if len(req.Image) == 0 {
				log.Printf("empty imput image")
				continue
			} else if ret := C.ncFifoWriteElem(inputFifo, unsafe.Pointer(&req.Image[0]), &blen, unsafe.Pointer(nil)); ret != 0 {
				return fmt.Errorf("error writing fifo, %v", ret)
			} else if ret := C.ncGraphQueueInference(graphHandle, &inputFifo, 1, &outputFifo, 1); ret != 0 {
				return fmt.Errorf("error queuing inference, %v", ret)
			} else if ret := C.ncFifoReadElem(outputFifo, unsafe.Pointer(&bout[0]), &fifoOutputSize, &user); ret != 0 {
				return fmt.Errorf("error reading output of inference, %v", ret)
			} else {
				req.Response <- bout
			}
		}
	}
}
