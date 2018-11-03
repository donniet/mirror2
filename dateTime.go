package main

import (
	"encoding/json"
	"fmt"
)

type dateTimeElement struct {
	visible bool
	changed chan bool
}

func (e *dateTimeElement) ServeJSON(path []string, msg *json.RawMessage) (*json.RawMessage, error) {
	if len(path) == 0 {
		if msg != nil {
			if err := json.Unmarshal(*msg, e); err != nil {
				return nil, err
			}
		}

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

func (e *dateTimeElement) UnmarshalJSON(b []byte) error {
	m := make(map[string]interface{})

	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	if v, ok := m["visible"]; ok {
		vis := false
		if vis, ok = v.(bool); !ok {
			return fmt.Errorf("dateTime visible element must be bool")
		}

		if vis && !e.Visible() {
			e.Show()
		} else if !vis && e.Visible() {
			e.Hide()
		}
	}
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
