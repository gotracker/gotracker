package effect

import (
	"fmt"

	"github.com/gotracker/voice/oscillator"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// SetVibratoWaveform defines a set vibrato waveform effect
type SetVibratoWaveform uint8 // 'S3x'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetVibratoWaveform) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	mem := cs.GetMemory()
	vib := mem.VibratoOscillator()
	vib.SetWaveform(oscillator.WaveTableSelect(x))
	return nil
}

func (e SetVibratoWaveform) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
