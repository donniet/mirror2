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
}

func NewCECDisplay(name string, deviceName string) (ret *CECDisplay, err error) {
	ret = new(CECDisplay)
	ret.address = 0
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

func (d *CECDisplay) HasError() error {
	return d.err
}

func (d *CECDisplay) PowerOn() {
	d.connection.PowerOn(d.address)
}

func (d *CECDisplay) Standby() {
	d.connection.Standby(d.address)
}

func (d *CECDisplay) VolumeUp() {
	d.connection.VolumeUp()
}

func (d *CECDisplay) VolumeDown() {
	d.connection.VolumeDown()
}

func (d *CECDisplay) Mute() {
	d.connection.Mute()
}

func (d *CECDisplay) KeyPress(key int) {
	d.connection.KeyPress(d.address, key)
}

func (d *CECDisplay) KeyRelease() {
	d.connection.KeyRelease(d.address)
}

func (d *CECDisplay) Key(key int) {
	d.connection.Key(d.address, key)
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
