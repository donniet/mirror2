package main

import (
	"encoding/json"
)

type dateTimeElement struct {
	visible bool
	changed chan bool
}

func (e *dateTimeElement) UnmarshalJSON(b []byte) error {
	return nil
}
func (e *dateTimeElement) MarshalJSON() ([]byte, error) {
	ret := make(map[string]interface{})
	ret["visible"] = e.visible
	return json.Marshal(ret)
}

func (e *dateTimeElement) Visible() bool {
	return e.visible
}

func (e *dateTimeElement) Name() string {
	return "Date and Time"
}

func (e *dateTimeElement) Show() {
	e.visible = true
	e.changed <- true
}

func (e *dateTimeElement) Hide() {
	e.visible = false
	e.changed <- true
}
