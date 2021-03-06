package effect

import (
	"fmt"

	s3mfile "github.com/gotracker/goaudiofile/music/tracked/s3m"

	"gotracker/internal/format/s3m/playback/util"
	"gotracker/internal/player/intf"
)

// SetGlobalVolume defines a set global volume effect
type SetGlobalVolume uint8 // 'V'

// PreStart triggers when the effect enters onto the channel state
func (e SetGlobalVolume) PreStart(cs intf.Channel, p intf.Playback) error {
	p.SetGlobalVolume(util.VolumeFromS3M(s3mfile.Volume(uint8(e))))
	return nil
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetGlobalVolume) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

func (e SetGlobalVolume) String() string {
	return fmt.Sprintf("V%0.2x", uint8(e))
}
