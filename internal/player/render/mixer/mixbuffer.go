package mixer

import (
	"gotracker/internal/player/intf"
	"gotracker/internal/player/volume"
	"time"
)

// ChannelMixBuffer is a single channel's premixed volume data
type ChannelMixBuffer volume.VolumeMatrix

// SampleMixIn is the parameters for mixing in a sample into a MixBuffer
type SampleMixIn struct {
	Sample       intf.Instrument
	SamplePos    float32
	SamplePeriod float32
	StaticVol    volume.Volume
	VolMatrix    volume.VolumeMatrix
	MixPos       int
	MixLen       int
}

// MixBuffer is a buffer of premixed volume data intended to
// be eventually sent out to the sound output device after
// conversion to the output format
type MixBuffer []ChannelMixBuffer

// C returns a channel and a function that flushes any outstanding mix-ins and closes the channel
func (m *MixBuffer) C() (chan<- SampleMixIn, func()) {
	ch := make(chan SampleMixIn, 32)
	go func() {
	outerLoop:
		for {
			select {
			case d, ok := <-ch:
				if !ok {
					break outerLoop
				}
				m.mixIn(d)
			}
		}
	}()
	return ch, func() {
		for len(ch) != 0 {
			time.Sleep(1 * time.Millisecond)
		}
		close(ch)
	}
}

func (m *MixBuffer) mixIn(d SampleMixIn) {
	pos := d.MixPos
	spos := d.SamplePos
	for i := 0; i < d.MixLen; i++ {
		sdata := d.Sample.GetSample(spos)
		samp := d.StaticVol.Apply(sdata...)
		mixed := d.VolMatrix.Apply(samp...)
		for c, s := range mixed {
			(*m)[c][pos] += s
		}
		pos++
		spos += d.SamplePeriod
	}
}
