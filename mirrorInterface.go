package main

import (
	"encoding/json"
	"log"
	"time"
)

func NewMirrorInterface(weatherURL string, changed chan<- socketResponse) *mirrorInterface {
	log.Printf("creating cec display interface")
	cec, _ := NewCECDisplay("", "Smart Mirror")

	log.Printf("creating rest of mirror interface")
	mi := &mirrorInterface{
		changed: changed,
		weather: newWeatherElement(weatherURL, make(chan bool), time.Hour),
		display: cec,
		date: &dateTimeElement{
			visible: false,
			changed: make(chan bool),
		},
		streams:       make(map[string]*streamElement),
		streamChanged: make(chan *streamElement),
	}

	log.Printf("starting changed loop")
	go mi.handleChanged()
	log.Printf("don creating mirror interface")
	return mi
}

type mirrorInterface struct {
	changed       chan<- socketResponse
	weather       *weatherElement
	date          *dateTimeElement
	display       *CECDisplay
	streams       map[string]*streamElement
	video         *videoElement
	streamChanged chan *streamElement
}

func (ui *mirrorInterface) handleChanged() {
	for {
		select {
		case <-ui.weather.changed:
			ui.changed <- socketResponse{
				Request:  &socketRequest{Path: "weather"},
				Response: ui.weather,
			}
		case <-ui.date.changed:
			ui.changed <- socketResponse{
				Request:  &socketRequest{Path: "dateTime"},
				Response: ui.date,
			}
		case <-ui.display.changed:
			ui.changed <- socketResponse{
				Request:  &socketRequest{Path: "display"},
				Response: ui.display,
			}
		case <-ui.streamChanged:
			ui.sendStreamsChanged()
		}
	}
}

func (ui *mirrorInterface) UnmarshalJSON([]byte) error {
	return nil
}
func (ui *mirrorInterface) MarshalJSON() ([]byte, error) {
	ret := make(map[string]interface{})
	ret["streams"] = ui.Streams()
	ret["weather"] = ui.Weather()
	ret["dateTime"] = ui.DateTime()
	ret["video"] = ui.Video()
	ret["display"] = ui.Display()
	return json.Marshal(ret)
}

func (ui *mirrorInterface) Streams() (ret []StreamElement) {
	for _, e := range ui.streams {
		ret = append(ret, e)
	}
	return
}

func (ui *mirrorInterface) Weather() WeatherElement {
	return ui.weather
}

func (ui *mirrorInterface) DateTime() UIElement {
	return ui.date
}

func (ui *mirrorInterface) Video() VideoService {
	return ui.video
}

func (ui *mirrorInterface) Display() Display {
	return ui.display
}

func (ui *mirrorInterface) AddStream(url string) {
	if _, ok := ui.streams[url]; !ok {
		ui.streams[url] = &streamElement{
			url:     url,
			visible: false,
			changed: ui.streamChanged,
		}
	}
	ui.sendStreamsChanged()
}

func (ui *mirrorInterface) sendStreamsChanged() {
	var ss []*streamElement
	for _, s := range ui.streams {
		ss = append(ss, s)
	}

	ui.changed <- socketResponse{
		Request:  &socketRequest{Path: "streams"},
		Response: ss,
	}
}

func (ui *mirrorInterface) RemoveStream(url string) {
	if _, ok := ui.streams[url]; ok {
		delete(ui.streams, url)

		ui.sendStreamsChanged()
	}
}

type streamElement struct {
	url     string
	visible bool
	changed chan<- *streamElement
}

func (e *streamElement) UnmarshalJSON([]byte) error {
	return nil
}

func (e *streamElement) MarshalJSON() ([]byte, error) {
	ret := make(map[string]interface{})
	ret["url"] = e.url
	ret["visible"] = e.visible
	return json.Marshal(ret)
}

func (e *streamElement) URL() string {
	return e.url
}
func (e *streamElement) Visible() bool {
	return e.visible
}
func (e *streamElement) Show() {
	if !e.visible {
		e.visible = true
		e.changed <- e
	}
}
func (e *streamElement) Hide() {
	if e.visible {
		e.visible = false
		e.changed <- e
	}
}
