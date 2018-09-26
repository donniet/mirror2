package main

import (
	"encoding/json"
)

type streamElement struct {
	url     string
	name    string
	visible bool
	changed chan<- *streamElement
}

func (e *streamElement) UnmarshalJSON([]byte) error {
	return nil
}

func (e *streamElement) MarshalJSON() ([]byte, error) {
	ret := make(map[string]interface{})
	ret["url"] = e.url
	ret["name"] = e.name
	ret["visible"] = e.visible
	return json.Marshal(ret)
}

func (e *streamElement) URL() string {
	return e.url
}
func (e *streamElement) Visible() bool {
	return e.visible
}
func (e *streamElement) Name() string {
	return e.name
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
