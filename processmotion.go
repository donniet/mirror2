package main

import (
	"encoding/binary"
	"io"
	"log"
	"time"
)

type motionVector struct {
	X   int8
	Y   int8
	Sad int16
}

type MotionProcessor struct {
	mbx       int
	mby       int
	magnitude int
	total     int
	throttle  time.Duration
}

func (proc MotionProcessor) Process(reader io.Reader) <-chan time.Time {
	detected := make(chan time.Time)

	go proc.thread(reader, detected)

	return detected
}

func (proc MotionProcessor) thread(reader io.Reader, motionDetected chan<- time.Time) {
	len := (proc.mbx + 1) * proc.mby
	vect := make([]motionVector, len)

	mag2 := proc.magnitude * proc.magnitude

	last := time.Now()

	defer close(motionDetected)

	for {
		if err := binary.Read(reader, binary.LittleEndian, vect); err != nil {
			log.Println(err)
			break
		} else if time.Now().Sub(last) < proc.throttle {
			continue
		}

		last = time.Now()

		c := 0
		for _, v := range vect {
			magU := int(v.X)*int(v.X) + int(v.Y)*int(v.Y)
			if magU > mag2 {
				c++
			}
		}

		// log.Printf("total motion vectors above magnitude: %d", c)

		if c > proc.total {
			// don't get hung up here-- better to process all the vectors in this loop than wait for a full channel
			log.Printf("motion detected")
			go func() {
				motionDetected <- last
			}()
		}
	}
}
