package main

// #include <stdio.h>
// #include <stdlib.h>
// #cgo LDFLAGS: -lmvnc
// #include <mvnc.h>
import "C"

import (
	"encoding/binary"
	"flag"
	"github.com/chbmuc/cec"
	"io/ioutil"
	"log"
	"unsafe"
	"fmt"
)

var (
	graphFile = ""
	deviceName = "Smart Mirror"
)

func init() {
	flag.StringVar(&graphFile, "graph", "", "graph file name")
	flag.StringVar(&deviceName, "deviceName", "Smart Mirror", "CEC Device Name")
}

func ProcessImage(graphFile string, input <-chan []byte, output chan<- []int) {
	var deviceHandle *C.struct_ncDeviceHandle_t
	var graphHandle *C.struct_ncGraphHandle_t

	if ret := C.ncDeviceCreate(0, &deviceHandle); ret != 0 {
		log.Fatalf("could not get device name, error code: %v", ret)
	}
	defer C.ncDeviceDestroy(&deviceHandle)

	if ret := C.ncDeviceOpen(deviceHandle); ret != 0 {
		log.Fatalf("could not open device: %v", ret)
	}
	defer C.ncDeviceClose(deviceHandle)

	if ret := C.ncGraphCreate(C.CString("faces"), &graphHandle); ret != 0 {
		log.Fatalf("could not create graph, %v", ret)
	}

	var inputFifo, outputFifo *C.struct_ncFifoHandle_t

	if b, err := ioutil.ReadFile(graphFile); err != nil {
		log.Fatal(err)
	} else if ret := C.ncGraphAllocateWithFifos(deviceHandle, graphHandle, unsafe.Pointer(&b[0]), C.uint(len(b)), &inputFifo, &outputFifo); ret != 0 {
		log.Fatalf("error allocating graph: %v", ret)
	}

	defer C.ncFifoDestroy(&inputFifo)
	defer C.ncFifoDestroy(&outputFifo)
	defer C.ncGraphDestroy(&graphHandle)

	fifoOutputSize := C.uint(0)
	optionDataLen := C.uint(4)

	C.ncFifoGetOption(outputFifo, C.NC_RO_FIFO_ELEMENT_DATA_SIZE, unsafe.Pointer(&fifoOutputSize), &optionDataLen)

	log.Printf("fifo output size: %d", fifoOutputSize)

	go func() {
		b := make([]byte, fifoOutputSize)
		user := unsafe.Pointer(nil)
		for {
			if ret := C.ncFifoReadElem(outputFifo, unsafe.Pointer(&b[0]), &fifoOutputSize, &user); ret != 0 {
				log.Printf("error reading from output fifo: %v", ret)
				break
			}
			var dat []int

			for start := 0; start < int(fifoOutputSize); start += 4 {
				d, _ := binary.Varint(b[start : start+4])
				dat = append(dat, int(d))
			}
			output <- dat
		}
	}()

	for {
		select {
		case b, ok := <-input:
			blen := C.uint(len(b))
			if !ok {
				break
			} else if ret := C.ncFifoWriteElem(inputFifo, unsafe.Pointer(&b[0]), &blen, unsafe.Pointer(nil)); ret != 0 {
				log.Printf("error writing fifo, %v", ret)
			} else if ret := C.ncGraphQueueInference(graphHandle, &inputFifo, 1, &outputFifo, 1); ret != 0 {
				log.Printf("error queuing inference, %v", ret)
			}
		}
	}
}


func main() {
	flag.Parse()

	c, err := cec.Open("", deviceName)
	if err != nil {
		fmt.Println(err)
	}
	//c.PowerOn(0)
	c.Standby(0)

	// images := make(chan []byte)
	// people := make(chan []int)
	//
	// ProcessImage(graphFile, images, people)

}
