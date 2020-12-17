package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/util"
	"gotracker/internal/player/intf"
)

// StereoControl defines a set stereo control effect
type StereoControl uint8 // 'SAx'

// PreStart triggers when the effect enters onto the channel state
func (e StereoControl) PreStart(cs intf.Channel, ss intf.Song) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e StereoControl) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	if x > 7 {
		cs.SetPan(util.PanningFromS3M(x - 8))
	} else {
		cs.SetPan(util.PanningFromS3M(x + 8))
	}
}

// Tick is called on every tick
func (e StereoControl) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e StereoControl) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e StereoControl) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
