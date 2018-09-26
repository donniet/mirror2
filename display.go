package main

import (
	"encoding/json"
	"strconv"

	"github.com/chbmuc/cec"
)

type CECDisplay struct {
	connection      *cec.Connection
	address         int
	powerStatus     string
	vendorID        uint64
	physicalAddress string
	err             error
	changed         chan bool
}

func NewCECDisplay(name string, deviceName string) (ret *CECDisplay, err error) {
	ret = new(CECDisplay)
	ret.address = 0
	ret.changed = make(chan bool)
	ret.connection, ret.err = cec.Open(name, deviceName)
	if ret.err == nil {
		ret.powerStatus = ret.PowerStatus()
		ret.vendorID = ret.VendorID()
		ret.physicalAddress = ret.PhysicalAddress()
	}
	return ret, ret.err
}

func (d *CECDisplay) UnmarshalJSON(b []byte) error {
	return nil
}
func (d *CECDisplay) MarshalJSON() ([]byte, error) {
	r := make(map[string]interface{})
	r["powerStatus"] = d.powerStatus
	r["vendorID"] = strconv.FormatUint(d.vendorID, 16)
	r["physicalAddress"] = d.physicalAddress
	if d.err != nil {
		r["error"] = d.err.Error()
	}
	return json.Marshal(r)
}

func (d *CECDisplay) PowerOn() {
	d.connection.PowerOn(d.address)
	d.changed <- true
}

func (d *CECDisplay) Standby() {
	d.connection.Standby(d.address)
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
	return d.connection.GetDevicePowerStatus(d.address)
}
