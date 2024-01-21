package play

import (
	deviceCommon "github.com/gotracker/gotracker/internal/output/device/common"
)

type Settings struct {
	Output                 deviceCommon.Settings
	NumPremixBuffers       int
	PanicOnUnhandledEffect bool
	ITLongChannelOutput    bool
	ITEnableNNA            bool
	Tracing                bool
	TracingFile            string
}
