package main

import (
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
