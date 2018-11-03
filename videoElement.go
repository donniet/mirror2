package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

type videoStateCommand string

const (
	Play  videoStateCommand = "play"
	Pause                   = "pause"
	Stop                    = "stop"
)

type videoCommand struct {
	Mute    *bool              `json:"mute,omitempty"`
	Volume  *int               `json:"volume,omitempty"`
	Load    string             `json:"load,omitempty"`
	Command *videoStateCommand `json:"command,omitempty"`
}

type videoElement struct {
	visible  bool
	muted    bool
	volume   int
	state    VideoState
	commands chan<- videoCommand
	receiver <-chan VideoState
	lock     *sync.Mutex
}

func newVideoElement(commands chan<- videoCommand, receiver <-chan VideoState) *videoElement {
	ret := &videoElement{
		volume:   100,
		state:    Unstarted,
		commands: commands,
		receiver: receiver,
		lock:     &sync.Mutex{},
	}
	go ret.receiverState()
	return ret
}

func (e *videoElement) MarshalJSON() ([]byte, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	m := make(map[string]interface{})
	m["visible"] = e.visible
	m["muted"] = e.muted
	m["volume"] = e.volume
	m["state"] = e.state

	return json.Marshal(&m)
}

func (e *videoElement) UnmarshalJSON(data []byte) error {
	m := make(map[string]interface{})

	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	if v, ok := m["visible"]; ok {
		vis := false
		if vis, ok = v.(bool); !ok {
			return fmt.Errorf("video visible must be a boolean")
		}

		log.Printf("video visibility: %v", vis)

		// if vis && !e.Visible() {
		// 	e.Show()
		// } else if !vis && e.Visible() {
		// 	e.Hide()
		// }
	}

	return nil
}

func (e *videoElement) ServeJSON(path []string, msg *json.RawMessage) (*json.RawMessage, error) {
	if len(path) == 0 {
		b, err := json.Marshal(e)
		return (*json.RawMessage)(&b), err
	}

	if len(path) > 1 {
		return nil, &NotFoundError{Path: path}
	}

	switch path[0] {
	case "visible":
		b, err := json.Marshal(e.visible)
		return (*json.RawMessage)(&b), err
	default:
		return nil, &NotFoundError{Path: path}
	}
}

func (v *videoElement) receiverState() {
	for s := range v.receiver {
		v.lock.Lock()
		v.state = s
		v.lock.Unlock()
	}
}

func (v *videoElement) LoadVideoByID(videoID string)   {}
func (v *videoElement) LoadVideoByURL(videoURL string) {}
func (v *videoElement) Play()                          {}
func (v *videoElement) Pause()                         {}
func (v *videoElement) Stop()                          {}
func (v *videoElement) SeekTo(seconds float32)         {}
func (v *videoElement) Mute()                          {}
func (v *videoElement) UnMute()                        {}
func (v *videoElement) IsMuted() bool {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.muted
}
func (v *videoElement) SetVolume(volume int) {}
func (v *videoElement) Volume() int {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.volume
}
func (v *videoElement) State() VideoState {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.state
}
