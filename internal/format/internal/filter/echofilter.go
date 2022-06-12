package filter

import (
	"math"

	"github.com/gotracker/gotracker/internal/filter"
	"github.com/gotracker/voice/period"

	"github.com/gotracker/gomixing/volume"
)

type EchoFilterSettings struct {
	WetDryMix  float32
	Feedback   float32
	LeftDelay  float32
	RightDelay float32
	PanDelay   float32
}

type EchoFilterFactory struct {
	Reserved00 [4]byte
	EchoFilterSettings
}

func (e *EchoFilterFactory) Factory() filter.Factory {
	return func(instrument, playback period.Frequency) filter.Filter {
		echo := EchoFilter{
			EchoFilterSettings: e.EchoFilterSettings,
			sampleRate:         playback,
		}
		echo.recalculate()
		return &echo
	}
}

type delayInfo struct {
	buf   []volume.Volume
	delay int
}

//===========

type EchoFilter struct {
	EchoFilterSettings
	sampleRate      period.Frequency
	initialFeedback volume.Volume
	writePos        int
	delay           [2]delayInfo // L,R
}

func (e *EchoFilter) Clone() filter.Filter {
	clone := EchoFilter{
		EchoFilterSettings: e.EchoFilterSettings,
		sampleRate:         e.sampleRate,
		writePos:           e.writePos,
	}
	clone.recalculate()
	for i := range clone.delay {
		copy(clone.delay[i].buf, e.delay[i].buf)
	}
	return &clone
}

func (e *EchoFilter) Filter(dry volume.Matrix) volume.Matrix {
	if dry.Channels == 0 {
		return volume.Matrix{}
	}
	wetMix := volume.Volume(e.WetDryMix)
	dryMix := 1 - wetMix
	wet := dry.Apply(dryMix)

	feedback := volume.Volume(e.Feedback)

	crossEcho := e.PanDelay >= 0.5

	bufferLen := len(e.delay[0].buf)

	for e.writePos >= bufferLen {
		e.writePos -= bufferLen
	}
	for e.writePos < 0 {
		e.writePos += bufferLen
	}

	for c := 0; c < dry.Channels; c++ {
		readChannel := c
		if crossEcho {
			readChannel = 1 - c
		}
		read := &e.delay[readChannel]
		write := &e.delay[c]

		readPos := e.writePos - read.delay
		for readPos < 0 {
			readPos += bufferLen
		}
		for readPos >= bufferLen {
			readPos -= bufferLen
		}

		chnInput := dry.StaticMatrix[c]
		chnDelay := read.buf[readPos]

		chnOutput := chnInput * e.initialFeedback
		chnOutput += chnDelay * feedback

		write.buf[e.writePos] = chnOutput

		wet.StaticMatrix[c] += chnDelay * wetMix
	}

	e.writePos++

	return wet
}

func (e *EchoFilter) recalculate() {
	e.initialFeedback = volume.Volume(math.Sqrt(float64(1.0 - (e.Feedback * e.Feedback))))

	playbackRate := float32(e.sampleRate)
	bufferSize := int(playbackRate * 2)

	for c, delayMs := range [2]float32{e.LeftDelay, e.RightDelay} {
		delay := int(delayMs * 2.0 * playbackRate)
		e.delay[c].delay = delay
		e.delay[c].buf = make([]volume.Volume, bufferSize)
	}
}

func (e *EchoFilter) UpdateEnv(val int8) {

}
