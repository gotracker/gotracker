package play

import (
	device "github.com/gotracker/gosound"
)

type Settings struct {
	Output                 device.Settings
	NumPremixBuffers       int
	PanicOnUnhandledEffect bool
	GatherEffectCoverage   bool
}
