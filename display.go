package main

import (
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/donniet/cec"
)

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
	ret.connection.Commands = ret.commands
	if ret.err == nil {
		ret.powerStatus = ret.PowerStatus()
		ret.vendorID = ret.VendorID()
		ret.physicalAddress = ret.PhysicalAddress()
	}
	go ret.handleCommands()
	return ret, ret.err
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
