package main

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"
)

var (
	pather = regexp.MustCompile(`[/]?([^/]+)`)
)

type NotFoundError struct {
	Path []string
}

func (n *NotFoundError) Error() string {
	return fmt.Sprintf("path not found: %v", n.Path)
}

func NewMirrorInterface(weatherURL string, changed chan<- socketResponse) *mirrorInterface {
	log.Printf("creating cec display interface")
	var disp Display
	var err error
	if disp, err = NewCECDisplay("", "Smart Mirror"); err != nil {
		log.Printf("error opening CEC display, using dummy: %v", err)
		disp = NewDummyDisplay()
	}

	log.Printf("creating rest of mirror interface")
	mi := &mirrorInterface{
		changed: changed,
		weather: newWeatherElement(weatherURL, make(chan bool), time.Hour),
		display: disp,
		date: &dateTimeElement{
			visible: false,
			changed: make(chan bool),
		},
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
	display       Display
	streams       []*streamElement
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
		case <-ui.display.Changed():
			ui.changed <- socketResponse{
				Request:  &socketRequest{Path: "display"},
				Response: ui.display,
			}
		case <-ui.streamChanged:
			ui.sendStreamsChanged()
		}
	}
}

func (ui *mirrorInterface) ServeJSON(path []string, msg *json.RawMessage) (ret *json.RawMessage, err error) {
	if len(path) > 0 && path[0] == "" {
		path = path[1:]
	}

	if len(path) == 0 {
		if msg != nil {
			if err = json.Unmarshal(*msg, ui); err != nil {
				return
			}
		}

		var b []byte
		b, err = json.Marshal(ui)
		ret = (*json.RawMessage)(&b)
		return
	}

	switch path[0] {
	case "streams":
		ret, err = ui.serveJSONStreams(path[1:], msg)
	case "weather":
		ret, err = ui.weather.ServeJSON(path[1:], msg)
	case "dateTime":
		ret, err = ui.date.ServeJSON(path[1:], msg)
	case "video":
		ret, err = ui.video.ServeJSON(path[1:], msg)
	case "display":
		ret, err = ui.display.ServeJSON(path[1:], msg)
	default:
		ret, err = nil, &NotFoundError{Path: path}
	}

	return
}

func (ui *mirrorInterface) serveJSONStreams(path []string, msg *json.RawMessage) (*json.RawMessage, error) {
	ss := ui.Streams()

	if len(path) == 0 {
		// if the msg is not null, consisider this a PUT
		if msg != nil {
			s := new(streamElement)
			if err := json.Unmarshal(*msg, s); err != nil {
				return nil, err
			}
			s = ui.AddStream(s.url, s.visible)
		}

		b, err := json.Marshal(ss)
		return (*json.RawMessage)(&b), err
	}

	if i, err := strconv.Atoi(path[0]); err != nil {
		return nil, err
	} else if i >= 0 || i < len(ss) {
		return ss[i].ServeJSON(path[1:], msg)
	} else if i == len(ss) && msg != nil && len(path) == 1 {
		s := new(streamElement)
		if err := json.Unmarshal(*msg, s); err != nil {
			return nil, err
		}
		s = ui.AddStream(s.url, s.visible)

		b, err := json.Marshal(s)
		return (*json.RawMessage)(&b), err
	} else {
		return nil, &NotFoundError{Path: path}
	}
}

func (ui *mirrorInterface) UnmarshalJSON(data []byte) error {
	m := make(map[string]*json.RawMessage)

	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	if w, ok := m["weather"]; ok {
		if err := json.Unmarshal(*w, ui.weather); err != nil {
			return err
		}
	}
	if dt, ok := m["dateTime"]; ok {
		if err := json.Unmarshal(*dt, ui.date); err != nil {
			return err
		}
	}
	if v, ok := m["video"]; ok {
		if err := json.Unmarshal(*v, ui.video); err != nil {
			return err
		}
	}
	if d, ok := m["display"]; ok {
		if err := json.Unmarshal(*d, ui.display); err != nil {
			return err
		}
	}
	if s, ok := m["streams"]; ok {
		var sl []*json.RawMessage

		if err := json.Unmarshal(*s, &sl); err != nil {
			return err
		}

		for i, s := range sl {
			if i >= len(ui.streams) {
				ui.streams = append(ui.streams, &streamElement{
					changed: ui.streamChanged,
				})
			}

			if err := json.Unmarshal(*s, ui.streams[i]); err != nil {
				return err
			}
		}
	}
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

func (ui *mirrorInterface) Streams() (ret []*streamElement) {
	for _, e := range ui.streams {
		ret = append(ret, e)
	}
	return
}

func (ui *mirrorInterface) Weather() *weatherElement {
	return ui.weather
}

func (ui *mirrorInterface) DateTime() *dateTimeElement {
	return ui.date
}

func (ui *mirrorInterface) Video() *videoElement {
	return ui.video
}

func (ui *mirrorInterface) Display() Display {
	return ui.display
}

func (ui *mirrorInterface) AddStream(url string, visible bool) *streamElement {
	s := &streamElement{
		url:     url,
		visible: visible,
		changed: ui.streamChanged,
	}
	ui.streams = append(ui.streams, s)
	ui.sendStreamsChanged()
	return s
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

func (ui *mirrorInterface) RemoveStream(index int) {
	if index < 0 || index >= len(ui.streams) {
		return
	}

	ui.streams = append(ui.streams[:index], ui.streams[index+1:]...)

	ui.sendStreamsChanged()
}

type streamElement struct {
	url     string
	visible bool
	changed chan<- *streamElement
}

func (e *streamElement) ServeJSON(path []string, msg *json.RawMessage) (*json.RawMessage, error) {
	if len(path) == 0 {
		b, err := json.Marshal(e)
		return (*json.RawMessage)(&b), err
	}

	if len(path) > 1 {
		return nil, &NotFoundError{Path: path}
	}

	var v interface{}

	switch path[0] {
	case "visible":
		v = e.visible
	case "url":
		v = e.url
	default:
	}

	if v == nil {
		return nil, &NotFoundError{Path: path}
	}
	b, err := json.Marshal(v)
	return (*json.RawMessage)(&b), err
}

func (e *streamElement) UnmarshalJSON(data []byte) error {
	m := make(map[string]interface{})

	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	if u, ok := m["url"]; ok {
		if e.url, ok = u.(string); !ok {
			return fmt.Errorf("stream url must be a string")
		}
	}
	if v, ok := m["visible"]; ok {
		if e.visible, ok = v.(bool); !ok {
			return fmt.Errorf("stream visible must be a bool")
		}
	}
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
