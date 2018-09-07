package main

// #include <stdio.h>
// #include <stdlib.h>
// #cgo LDFLAGS: -lmvnc -lcec
// #include <mvnc.h>
// #include <libcec/cecc.h>
import "C"

import (
	"encoding/binary"
	"flag"
	"io/ioutil"
	"log"
	"unsafe"
)

var (
	graphFile = ""
)

func init() {
	flag.StringVar(&graphFile, "graph", "", "graph file name")
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

func PowerSaver(power <-chan bool) {
	config := C.struct_libcec_configuration{
		clientVersion: C.LIBCEC_VERSION_CURRENT,
		bActivateSource: 1,
	}
	conn := C.libcec_initialise(&config)
	defer C.libcec_destroy(conn)

	log.Printf("conn: %#v", conn)
	// for {
	// 	if p, ok := <-power; !ok {
	// 		break
	// 	} else {
	// 		log.Printf("power status: %v", p)
	// 	}
	// }


}

func main() {
	flag.Parse()

	power := make(chan bool)

	PowerSaver(power)

	// images := make(chan []byte)
	// people := make(chan []int)
	//
	// ProcessImage(graphFile, images, people)

}
