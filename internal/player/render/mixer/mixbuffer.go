package mixer

import (
	"bytes"
	"encoding/binary"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/sample"
	"gotracker/internal/player/volume"
	"time"
)

// ChannelMixBuffer is a single channel's premixed volume data
type ChannelMixBuffer volume.VolumeMatrix

// SampleMixIn is the parameters for mixing in a sample into a MixBuffer
type SampleMixIn struct {
	Sample       intf.Instrument
	SamplePos    sample.Pos
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
		spos.Add(d.SamplePeriod)
	}
}

// Add will mix in another MixBuffer's data
func (m *MixBuffer) Add(pos int, rhs MixBuffer, volMtx volume.VolumeMatrix) {
	sdata := make(volume.VolumeMatrix, len(*m))
	for i := 0; i < len(rhs[0]); i++ {
		for c := 0; c < len(rhs); c++ {
			sdata[c] = rhs[c][i]
		}
		sd := volMtx.Apply(sdata...)
		for c, s := range sd {
			(*m)[c][pos+i] += s
		}
	}
}

// ToRenderData converts a mixbuffer into a byte stream intended to be
// output to the output sound device
func (m *MixBuffer) ToRenderData(samples int, bitsPerSample int, mixedChannels int) []byte {
	writer := &bytes.Buffer{}
	samplePostMultiply := 1.0 / volume.Volume(mixedChannels)
	for i := 0; i < samples; i++ {
		for _, buf := range *m {
			v := buf[i] * samplePostMultiply
			val := v.ToSample(bitsPerSample)
			binary.Write(writer, binary.LittleEndian, val)
		}
	}
	return writer.Bytes()
}

// ToRenderDataWithBuf converts a mixbuffer into a byte stream intended to be
// output to the output sound device
func (m *MixBuffer) ToRenderDataWithBuf(out []byte, samples int, bitsPerSample int, mixedChannels int) {
	pos := 0
	samplePostMultiply := 1.0 / volume.Volume(mixedChannels)
	for i := 0; i < samples; i++ {
		for _, buf := range *m {
			v := buf[i] * samplePostMultiply
			val := v.ToSample(bitsPerSample)
			switch d := val.(type) {
			case uint8:
				out[pos] = d
				pos++
			case uint16:
				binary.LittleEndian.PutUint16(out[pos:], d)
				pos += 2
			case uint32:
				binary.LittleEndian.PutUint32(out[pos:], d)
				pos += 4
			default:
				writer := &bytes.Buffer{}
				binary.Write(writer, binary.LittleEndian, val)
				pos += copy(out[pos:], writer.Bytes())
			}
		}
	}
}
