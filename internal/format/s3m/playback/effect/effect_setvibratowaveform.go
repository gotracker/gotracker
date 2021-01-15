package effect

import (
	"fmt"

	formatutil "gotracker/internal/format/internal/util"
	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// SetVibratoWaveform defines a set vibrato waveform effect
type SetVibratoWaveform uint8 // 'S3x'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetVibratoWaveform) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	mem := cs.GetMemory().(*channel.Memory)
	vib := mem.VibratoOscillator()
	vib.Table = formatutil.WaveTableSelect(x)
}

func (e SetVibratoWaveform) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
