package effect

import (
	"fmt"

	effectIntf "gotracker/internal/format/s3m/playback/effect/intf"
	"gotracker/internal/player/intf"
)

// EnableFilter defines a set filter enable effect
type EnableFilter uint8 // 'S0x'

// Start triggers on the first tick, but before the Tick() function is called
func (e EnableFilter) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf
	on := x != 0

	pb := p.(effectIntf.S3M)
	pb.SetFilterEnable(on)
	return nil
}

func (e EnableFilter) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
