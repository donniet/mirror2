package visual

/*
#include <stdio.h>
#include <stdlib.h>
#cgo LDFLAGS: -lmvnc
#include <mvnc.h>
*/
import "C"

import (
	"fmt"
	"io/ioutil"
	"unsafe"
)

func ProcessImages(graphFile string, input <-chan []byte, output chan<- []byte) error {
	var deviceHandle *C.struct_ncDeviceHandle_t
	var graphHandle *C.struct_ncGraphHandle_t

	if ret := C.ncDeviceCreate(0, &deviceHandle); ret != 0 {
		return fmt.Errorf("could not create device, will continue to retry every minute")
	}
	defer C.ncDeviceDestroy(&deviceHandle)

	if ret := C.ncDeviceOpen(deviceHandle); ret != 0 {
		return fmt.Errorf("could not open device: %v", ret)
	}
	defer C.ncDeviceClose(deviceHandle)

	if ret := C.ncGraphCreate(C.CString("process"), &graphHandle); ret != 0 {
		return fmt.Errorf("could not create graph: %v", ret)
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

	errorChan := make(chan error)

	go func() {
		b := make([]byte, fifoOutputSize)
		user := unsafe.Pointer(nil)
		for {
			if ret := C.ncFifoReadElem(outputFifo, unsafe.Pointer(&b[0]), &fifoOutputSize, &user); ret != 0 {
				errorChan <- fmt.Errorf("error reading from output fifo: %v", ret)
				break
			}
			output <- b
		}
	}()

	for {
		select {
		case b, ok := <-input:
			blen := C.uint(len(b))
			if !ok {
				break
			} else if ret := C.ncFifoWriteElem(inputFifo, unsafe.Pointer(&b[0]), &blen, unsafe.Pointer(nil)); ret != 0 {
				return fmt.Errorf("error writing to input fifo: %v", ret)
			} else if ret := C.ncGraphQueueInference(graphHandle, &inputFifo, 1, &outputFifo, 1); ret != 0 {
				return fmt.Errorf("error queuing inference: %v", ret)
			}
		case err := <-errorChan:
			return err
		}
	}
	return nil
}
