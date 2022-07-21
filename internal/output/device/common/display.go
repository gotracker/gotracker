package common

import "github.com/gotracker/playback/output"

// WrittenCallback defines the callback for when a premix buffer is mixed/rendered and output on the device
type WrittenCallback func(deviceCommon Kind, premix *output.PremixData)
