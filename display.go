package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/donniet/cec"
)

type DummyDisplay struct {
	powerStatus string
	err         error
	sleeping    bool
	waking      bool
	afterFunc   *time.Timer
	afterTime   time.Time
	lock        *sync.Mutex
	changed     chan bool
}

func NewDummyDisplay() *DummyDisplay {
	return &DummyDisplay{
		powerStatus: "standby",
		lock:        &sync.Mutex{},
		changed:     make(chan bool),
	}
}

func (d *DummyDisplay) Changed() <-chan bool {
	return d.changed
}

func (d *DummyDisplay) ServeJSON(path []string, msg *json.RawMessage) (*json.RawMessage, error) {
	if len(path) == 0 {
		if msg != nil {
			if err := json.Unmarshal(*msg, d); err != nil {
				return nil, err
			}
		}

		b, err := json.Marshal(d)
		return (*json.RawMessage)(&b), err
	}

	if len(path) > 1 {
		return nil, &NotFoundError{Path: path}
	}

	var v interface{}

	d.lock.Lock()
	defer d.lock.Unlock()

	switch path[0] {
	case "powerStatus":
		v = d.powerStatus
	case "vendorID":
		v = 0
	case "physicalAddress":
		v = "0:0"
	case "sleeping":
		v = d.sleeping
	case "waking":
		v = d.waking
	case "error":
		v = d.err.Error()
	default:
		return nil, &NotFoundError{Path: path}
	}

	b, err := json.Marshal(v)
	return (*json.RawMessage)(&b), err
}

func (d *DummyDisplay) UnmarshalJSON(b []byte) error {
	m := make(map[string]interface{})

	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	if p, ok := m["powerStatus"]; ok {
		var ps string
		if ps, ok = p.(string); !ok {
			return fmt.Errorf("display powerStatus must be a string")
		}

		if ps != d.PowerStatus() {
			switch ps {
			case "on":
				d.PowerOn()
			case "standby":
				d.Standby()
			default:
				return fmt.Errorf("display powerStatus must be 'on' or 'standby'")
			}
		}
	}

	if s, ok := m["sleep"]; ok {
		var ss string
		if ss, ok = s.(string); !ok {
			return fmt.Errorf("display sleep must be a string")
		}

		if err := d.Sleep(ss); err != nil {
			return err
		}
	}

	if w, ok := m["wake"]; ok {
		var ws string
		if ws, ok = w.(string); !ok {
			return fmt.Errorf("display wake must be a string")
		}

		if err := d.Wake(ws); err != nil {
			return err
		}
	}

	return nil
}
func (d *DummyDisplay) MarshalJSON() ([]byte, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	r := make(map[string]interface{})
	r["powerStatus"] = d.powerStatus
	r["vendorID"] = strconv.FormatUint(0, 16)
	r["physicalAddress"] = "0:0"
	r["sleeping"] = d.sleeping
	r["waking"] = d.waking
	if d.err != nil {
		r["error"] = d.err.Error()
	}
	return json.Marshal(r)
}

func (d *DummyDisplay) Sleep(duration string) (err error) {
	d.lock.Lock()

	var dur time.Duration
	if dur, err = time.ParseDuration(duration); err != nil {
		return
	}

	if d.afterFunc != nil {
		d.afterFunc.Stop()
		d.afterFunc = nil
	}
	d.afterFunc = time.AfterFunc(dur, func() {
		d.waking = false
		d.sleeping = false
		d.afterFunc = nil
	})
	d.afterTime = time.Now().Add(dur)
	d.waking = false
	d.sleeping = true
	d.lock.Unlock()
	d.Standby()

	d.changed <- true
	return
}
func (d *DummyDisplay) Wake(duration string) (err error) {
	d.lock.Lock()

	var dur time.Duration
	if dur, err = time.ParseDuration(duration); err != nil {
		return
	}

	if d.afterFunc != nil {
		d.afterFunc.Stop()
		d.afterFunc = nil
	}
	d.afterFunc = time.AfterFunc(dur, func() {
		d.waking = false
		d.sleeping = false
		d.afterFunc = nil
	})
	d.afterTime = time.Now().Add(dur)
	d.waking = true
	d.sleeping = false
	d.lock.Unlock()
	d.PowerOn()

	d.changed <- true
	return
}

func (d *DummyDisplay) MotionActivated() bool {
	d.lock.Lock()
	defer d.lock.Unlock()

	return !d.sleeping && !d.waking
}

func (d *DummyDisplay) Sleeping() bool {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.sleeping
}
func (d *DummyDisplay) Waking() bool {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.waking
}
func (d *DummyDisplay) SleepingUntil() time.Time {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.afterTime
}
func (d *DummyDisplay) WakingUntil() time.Time {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.afterTime
}

func (d *DummyDisplay) PowerOn() {
	d.lock.Lock()
	d.powerStatus = "on"
	d.lock.Unlock()
	d.changed <- true
}

func (d *DummyDisplay) Standby() {
	d.lock.Lock()
	d.powerStatus = "standby"
	d.lock.Unlock()
	d.changed <- true
}

func (d *DummyDisplay) VolumeUp() {
	d.changed <- true
}

func (d *DummyDisplay) VolumeDown() {
	d.changed <- true
}

func (d *DummyDisplay) Mute() {
	d.changed <- true
}

func (d *DummyDisplay) KeyPress(key int) {
	d.changed <- true
}

func (d *DummyDisplay) KeyRelease() {
	d.changed <- true
}

func (d *DummyDisplay) Key(key int) {
	d.changed <- true
}

func (d *DummyDisplay) OSDName() string {
	return "Dummy Display"
}

func (d *DummyDisplay) IsActiveSource() bool {
	return true
}

func (d *DummyDisplay) VendorID() uint64 {
	return 0
}

func (d *DummyDisplay) PhysicalAddress() string {
	return "0:0"
}

func (d *DummyDisplay) PowerStatus() string {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.powerStatus
}

type CECDisplay struct {
	connection      *cec.Connection
	address         int
	powerStatus     string
	vendorID        uint64
	physicalAddress string
	afterFunc       *time.Timer
	afterTime       time.Time
	sleeping        bool
	waking          bool
	err             error
	changed         chan bool
	lock            *sync.Mutex
	commands        chan *cec.Command
}

func NewCECDisplay(name string, deviceName string) (ret *CECDisplay, err error) {
	ret = new(CECDisplay)
	ret.address = 0
	ret.commands = make(chan *cec.Command)
	ret.lock = &sync.Mutex{}
	ret.changed = make(chan bool)
	ret.connection, ret.err = cec.Open(name, deviceName)
	if ret.err == nil {
		ret.connection.Commands = ret.commands
		ret.powerStatus = ret.PowerStatus()
		ret.vendorID = ret.VendorID()
		ret.physicalAddress = ret.PhysicalAddress()
	}
	go ret.handleCommands()
	return ret, ret.err
}

func (d *CECDisplay) Destroy() {
	d.connection.Destroy()
}

func (d *CECDisplay) handleCommands() {
	for c := range d.commands {
		switch c.Operation {
		case "STANDBY":
			d.lock.Lock()
			d.powerStatus = "standby"
			d.lock.Unlock()
		case "ROUTING_CHANGE":
			d.lock.Lock()
			d.powerStatus = "on"
			d.lock.Unlock()
		case "REPORT_POWER_STATUS":
			log.Printf("power status change: %#v", c)
		}
	}
}

func (d *CECDisplay) Changed() <-chan bool {
	return d.changed
}

func (d *CECDisplay) ServeJSON(path []string, msg *json.RawMessage) (*json.RawMessage, error) {
	if len(path) == 0 {
		b, err := json.Marshal(d)
		return (*json.RawMessage)(&b), err
	}

	if len(path) > 1 {
		return nil, &NotFoundError{Path: path}
	}

	var v interface{}

	d.lock.Lock()
	defer d.lock.Unlock()

	switch path[0] {
	case "powerStatus":
		v = d.powerStatus
	case "vendorID":
		v = d.vendorID
	case "physicalAddress":
		v = d.physicalAddress
	case "sleeping":
		v = d.sleeping
	case "waking":
		v = d.waking
	case "error":
		v = d.err.Error()
	default:
		return nil, &NotFoundError{Path: path}
	}

	b, err := json.Marshal(v)
	return (*json.RawMessage)(&b), err
}

func (d *CECDisplay) UnmarshalJSON(b []byte) error {
	return nil
}
func (d *CECDisplay) MarshalJSON() ([]byte, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	r := make(map[string]interface{})
	r["powerStatus"] = d.powerStatus
	r["vendorID"] = strconv.FormatUint(d.vendorID, 16)
	r["physicalAddress"] = d.physicalAddress
	r["sleeping"] = d.sleeping
	r["waking"] = d.waking
	if d.err != nil {
		r["error"] = d.err.Error()
	}
	return json.Marshal(r)
}

func (d *CECDisplay) Sleep(duration string) (err error) {
	d.lock.Lock()

	var dur time.Duration
	if dur, err = time.ParseDuration(duration); err != nil {
		return
	}

	if d.afterFunc != nil {
		d.afterFunc.Stop()
		d.afterFunc = nil
	}
	d.afterFunc = time.AfterFunc(dur, func() {
		d.waking = false
		d.sleeping = false
		d.afterFunc = nil
	})
	d.afterTime = time.Now().Add(dur)
	d.waking = false
	d.sleeping = true
	d.lock.Unlock()
	d.Standby()

	d.changed <- true
	return
}
func (d *CECDisplay) Wake(duration string) (err error) {
	d.lock.Lock()

	var dur time.Duration
	if dur, err = time.ParseDuration(duration); err != nil {
		return
	}

	if d.afterFunc != nil {
		d.afterFunc.Stop()
		d.afterFunc = nil
	}
	d.afterFunc = time.AfterFunc(dur, func() {
		d.waking = false
		d.sleeping = false
		d.afterFunc = nil
	})
	d.afterTime = time.Now().Add(dur)
	d.waking = true
	d.sleeping = false
	d.lock.Unlock()
	d.PowerOn()

	d.changed <- true
	return
}

func (d *CECDisplay) MotionActivated() bool {
	d.lock.Lock()
	defer d.lock.Unlock()

	return !d.sleeping && !d.waking
}

func (d *CECDisplay) Sleeping() bool {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.sleeping
}
func (d *CECDisplay) Waking() bool {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.waking
}
func (d *CECDisplay) SleepingUntil() time.Time {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.afterTime
}
func (d *CECDisplay) WakingUntil() time.Time {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.afterTime
}

func (d *CECDisplay) PowerOn() {
	d.lock.Lock()
	d.connection.PowerOn(d.address)
	d.powerStatus = "on"
	d.lock.Unlock()
	d.changed <- true
}

func (d *CECDisplay) Standby() {
	d.lock.Lock()
	d.connection.Standby(d.address)
	d.powerStatus = "standby"
	d.lock.Unlock()
	d.changed <- true
}

func (d *CECDisplay) VolumeUp() {
	d.connection.VolumeUp()
	d.changed <- true
}

func (d *CECDisplay) VolumeDown() {
	d.connection.VolumeDown()
	d.changed <- true
}

func (d *CECDisplay) Mute() {
	d.connection.Mute()
	d.changed <- true
}

func (d *CECDisplay) KeyPress(key int) {
	d.connection.KeyPress(d.address, key)
	d.changed <- true
}

func (d *CECDisplay) KeyRelease() {
	d.connection.KeyRelease(d.address)
	d.changed <- true
}

func (d *CECDisplay) Key(key int) {
	d.connection.Key(d.address, key)
	d.changed <- true
}

func (d *CECDisplay) OSDName() string {
	return d.connection.GetDeviceOSDName(d.address)
}

func (d *CECDisplay) IsActiveSource() bool {
	return d.connection.IsActiveSource(d.address)
}

func (d *CECDisplay) VendorID() uint64 {
	return d.connection.GetDeviceVendorID(d.address)
}

func (d *CECDisplay) PhysicalAddress() string {
	return d.connection.GetDevicePhysicalAddress(d.address)
}

func (d *CECDisplay) PowerStatus() string {
	d.lock.Lock()

	p := d.connection.GetDevicePowerStatus(d.address)
	if p != d.powerStatus {
		d.powerStatus = p
		d.lock.Unlock()
		d.changed <- true
	} else {
		d.lock.Unlock()
	}
	return p
}
