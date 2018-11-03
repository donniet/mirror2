package main

import (
	"encoding/json"
	"time"
)

type Server interface {
	ServeJSON(paths []string, msg *json.RawMessage) (*json.RawMessage, error)
}

type Display interface {
	Server
	PowerOn()
	Standby()
	VolumeUp()
	VolumeDown()
	Mute()
	KeyPress(key int)
	KeyRelease()
	Key(key int)
	OSDName() string
	IsActiveSource() bool
	VendorID() uint64
	PhysicalAddress() string
	PowerStatus() string
	Sleep(duration string) error /* puts the screen in standby mode and ignores motion for the duration */
	Wake(duration string) error  /* ensures the screen stays awake for the duration regardless of motion */
	MotionActivated() bool
	SleepingUntil() time.Time
	WakingUntil() time.Time
	Sleeping() bool
	Waking() bool
	Changed() <-chan bool
}

type socketRequest struct {
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type socketResponse struct {
	Request  *socketRequest `json:"request,omitempty"`
	Response interface{}    `json:"response,omitempty"`
	Error    string         `json:"error,omitempty"`
}

type VideoState int

const (
	Unstarted VideoState = -1
	Ended     VideoState = 0
	Playing   VideoState = 1
	Paused    VideoState = 2
	Buffering VideoState = 3
	Cued      VideoState = 5
)
