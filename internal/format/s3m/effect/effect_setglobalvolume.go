package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/util"
	"gotracker/internal/module/player/intf"
)

// SetGlobalVolume defines a set global volume effect
type SetGlobalVolume uint8 // 'V'

// PreStart triggers when the effect enters onto the channel state
func (e SetGlobalVolume) PreStart(cs intf.Channel, ss intf.Song) {
	ss.SetGlobalVolume(util.VolumeFromS3M(uint8(e)))
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetGlobalVolume) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e SetGlobalVolume) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e SetGlobalVolume) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e SetGlobalVolume) String() string {
	return fmt.Sprintf("V%0.2x", uint8(e))
}
