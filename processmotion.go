package main

import (
  "io"
  "time"
  "log"
  "encoding/binary"
)

type motionVector struct {
	X int8
	Y int8
	Sad int16
};

type MotionProcessor struct {
  mbx int
  mby int
  magnitude int
  total int
  throttle time.Duration
}

func (proc MotionProcessor) Process(reader io.Reader, motionDetected chan<- bool) {
  len := (proc.mbx + 1) * proc.mby
  vect := make([]motionVector, len)

  mag2 := proc.magnitude * proc.magnitude

  last := time.Now()

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

    if c > proc.total {
      // don't get hung up here-- better to process all the vectors in this loop than wait for a full channel
      go func() {
        motionDetected <- true
      }()
    }
  }
}
