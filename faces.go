package main

// #include <stdio.h>
// #include <stdlib.h>
// #cgo LDFLAGS: -lmvnc
// #include <mvnc.h>
import "C"

import (
	"io"
	"io/ioutil"
	"log"
	"time"
	"unsafe"
)

type Faces struct {
	graphFile string
	names     []string
	width     int
	height    int
	channels  int
	threshold float32
	throttle  time.Duration
}

func (f Faces) Process(reader io.Reader, detected chan<- string) {
	l := f.width * f.height * f.channels
	bb := make([]byte, l)

	last := time.Now()

	defer close(detected)

	var deviceHandle *C.struct_ncDeviceHandle_t
	var graphHandle *C.struct_ncGraphHandle_t

	if ret := C.ncDeviceCreate(0, &deviceHandle); ret != 0 {
		log.Printf("could not get device name, error code: %v", ret)
		return
	}
	defer C.ncDeviceDestroy(&deviceHandle)

	if ret := C.ncDeviceOpen(deviceHandle); ret != 0 {
		log.Printf("could not open device: %v", ret)
		return
	}
	defer C.ncDeviceClose(deviceHandle)

	if ret := C.ncGraphCreate(C.CString("faces"), &graphHandle); ret != 0 {
		log.Printf("could not create graph, %v", ret)
		return
	}
	defer C.ncGraphDestroy(&graphHandle)

	var inputFifo, outputFifo *C.struct_ncFifoHandle_t

	if b, err := ioutil.ReadFile(graphFile); err != nil {
		log.Println(err)
		return
	} else if ret := C.ncGraphAllocateWithFifos(deviceHandle, graphHandle, unsafe.Pointer(&b[0]), C.uint(len(b)), &inputFifo, &outputFifo); ret != 0 {
		log.Printf("error allocating graph: %v", ret)
		return
	}

	defer C.ncFifoDestroy(&inputFifo)
	defer C.ncFifoDestroy(&outputFifo)

	fifoOutputSize := C.uint(0)
	fifoInputSize := C.uint(0)
	optionDataLen := C.uint(4)

	C.ncFifoGetOption(outputFifo, C.NC_RO_FIFO_ELEMENT_DATA_SIZE, unsafe.Pointer(&fifoOutputSize), &optionDataLen)
	C.ncFifoGetOption(inputFifo, C.NC_RO_FIFO_ELEMENT_DATA_SIZE, unsafe.Pointer(&fifoInputSize), &optionDataLen)

	log.Printf("fifo input/output sizes: %d/%d", fifoInputSize, fifoOutputSize)

	for {
		if n, err := reader.Read(bb); err != nil {
			log.Println(err)
			break
		} else if n < l {
			log.Printf("not enough bytes from reader, got %d expected %d", n, l)
			break
		}

		fifoWriteFillLevel := C.int(0)
		fifoWriteFillLevelSize := C.uint(4)

		if now := time.Now(); now.Sub(last) < f.throttle {
			continue
		} else if ret := C.ncFifoGetOption(inputFifo, C.NC_RO_FIFO_WRITE_FILL_LEVEL, unsafe.Pointer(&fifoWriteFillLevel), &fifoWriteFillLevelSize); ret != C.NC_OK {
			log.Printf("error getting fifo fill level %v", ret)
			return
		} else if fifoWriteFillLevel > 0 {
			log.Println("fifo has elements, skipping this frame")
			continue
		} else {
			last = now
		}

		if int(fifoOutputSize)/4 > len(f.names) {
			log.Printf("outputsize %d greater than names %d", fifoOutputSize/4, len(f.names))
		}

		bout := make([]float32, fifoOutputSize/4)
		user := unsafe.Pointer(nil)

		blen := C.uint(l)

		if ret := C.ncFifoWriteElem(inputFifo, unsafe.Pointer(&bb[0]), &blen, unsafe.Pointer(nil)); ret != 0 {
			log.Printf("error writing fifo, %v", ret)
			return
		} else if ret := C.ncGraphQueueInference(graphHandle, &inputFifo, 1, &outputFifo, 1); ret != 0 {
			log.Printf("error queuing inference, %v", ret)
			return
		} else if ret := C.ncFifoReadElem(outputFifo, unsafe.Pointer(&bout[0]), &fifoOutputSize, &user); ret != 0 {
			log.Printf("error reading output of inference, %v", ret)
			return
		}

		for i, r := range bout {
			if r > f.threshold {
				if i < len(f.names) {
					detected <- f.names[i]
				}
			}
		}
	}
}
