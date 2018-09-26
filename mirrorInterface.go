package main

import (
	"time"
	"encoding/json"
)

func NewMirrorInterface(weatherURL string, changed chan<- socketResponse) *mirrorInterface {
	cec, _ := NewCECDisplay("", "Smart Mirror")
	mi := &mirrorInterface{
		changed: changed,
		weather: newWeatherElement(weatherURL, make(chan bool), time.Hour),
		display: cec,
		date: &dateTimeElement{
			visible: false,
			changed: make(chan bool),
		},
		streams: make(map[string]*streamElement),
	}
	go mi.handleChanged()
	return mi
}

type mirrorInterface struct {
	changed chan<- socketResponse
	weather *weatherElement
	date    *dateTimeElement
	display *CECDisplay
	streams map[string]*streamElement
	video   *videoElement
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

func (ui *mirrorInterface) AddStream(name string, url string) {
	if s, ok := ui.streams[url]; ok {
		s.name = name
	} else {
		ui.streams[url] = &streamElement{
			name:    name,
			url:     url,
			visible: false,
		}
	}
}

func (ui *mirrorInterface) RemoveStream(url string) {
	delete(ui.streams, url)
}
