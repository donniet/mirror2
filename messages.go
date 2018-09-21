package main

type Display interface {
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
}

type FaceReceiver interface {
	FaceDetected(face string)
}

type MotionReceiver interface {
	MotionDetected()
}

type UserInterface interface {
	Streams() []UIElement
	Weather() WeatherElement
	DateTime() UIElement
	Video() VideoService
	Display() Display
}

type UIElement interface {
	Name() string
	Show()
	Hide()
	Visible() bool
	HasError() error
}

type WeatherElement interface {
	UIElement
	High() float64
	Low() float64
	Icon() string
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

type VideoService interface {
	LoadVideoByID(videoId string)
	LoadVideoByURL(videoURL string)
	Play()
	Pause()
	Stop()
	SeekTo(seconds float32)
	Mute()
	UnMute()
	IsMuted() bool
	SetVolume(volume int)
	Volume() int
	State() VideoState
}

type VideoReceiver interface {
	StateChange(state VideoState)
}
