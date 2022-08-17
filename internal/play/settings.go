package play

import (
	deviceCommon "github.com/gotracker/gotracker/internal/output/device/common"
)

type Settings struct {
	Output                        deviceCommon.Settings
	NumPremixBuffers              int
	PanicOnUnhandledEffect        bool
	GatherEffectCoverage          bool
	ITLongChannelOutput           bool
	ITEnableNNA                   bool
	MovingAverageFilterWindowSize int
	Tracing                       bool
	TracingFile                   string
	SoloChannels                  []int
}
