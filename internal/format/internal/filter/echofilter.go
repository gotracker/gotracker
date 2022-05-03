package filter

import (
	"gotracker/internal/filter"
	"gotracker/internal/format/internal/util"

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
	return func(sampleRate float32) filter.Filter {
		echo := EchoFilter{
			EchoFilterSettings: e.EchoFilterSettings,
			sampleRate:         sampleRate,
		}
		ldelay := int(e.LeftDelay * echo.sampleRate)
		rdelay := int(e.RightDelay * echo.sampleRate)
		echo.delayBufL = util.NewRingBuffer[volume.Volume](ldelay * 3)
		echo.delayBufR = util.NewRingBuffer[volume.Volume](rdelay * 3)
		return &echo
	}
}

//===========

type EchoFilter struct {
	EchoFilterSettings
	sampleRate float32
	delayBufL  util.RingBuffer[volume.Volume]
	delayBufR  util.RingBuffer[volume.Volume]
}

func (e *EchoFilter) Filter(dry volume.Matrix) volume.Matrix {
	wetMix := volume.Volume(e.WetDryMix)
	dryMix := 1 - wetMix
	wet := dry

	ldelay := int(e.LeftDelay * e.sampleRate)
	rdelay := int(e.RightDelay * e.sampleRate)

	feedback := volume.Volume(e.Feedback)

	crossEcho := e.PanDelay >= 0.5

	for c, s := range dry {
		switch c {
		case 0:
			e.delayBufL.Write([]volume.Volume{s})
		case 1:
			e.delayBufR.Write([]volume.Volume{s})
		}
	}

	for c := range wet {
		var (
			buf   *util.RingBuffer[volume.Volume]
			delay int
		)

		switch {
		case (c == 0) || (crossEcho && c == 1):
			buf = &e.delayBufL
			delay = ldelay
		case (c == 1) || (crossEcho && c == 0):
			buf = &e.delayBufR
			delay = rdelay
		}
		if buf == nil {
			continue
		}

		// Calculate the mix
		var wetPre [1]volume.Volume
		buf.ReadFrom(delay, wetPre[:])
		dryPre := dry[c]
		w := dryPre*dryMix + wetPre[0]*wetMix
		wet[c] = w
		buf.Accumulate(w * feedback)
	}

	return wet
}

func (e *EchoFilter) UpdateEnv(val float32) {

}
