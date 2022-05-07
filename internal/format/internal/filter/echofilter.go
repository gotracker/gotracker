package filter

import (
	"github.com/gotracker/gotracker/internal/filter"
	"github.com/gotracker/gotracker/internal/format/internal/util"

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
	return func(sampleRate int) filter.Filter {
		echo := EchoFilter{
			EchoFilterSettings: e.EchoFilterSettings,
			sampleRate:         float32(sampleRate),
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
	if dry.Channels == 0 {
		return volume.Matrix{}
	}
	wetMix := volume.Volume(e.WetDryMix)
	dryMix := 1 - wetMix
	wet := dry

	ldelay := int(e.LeftDelay * e.sampleRate)
	rdelay := int(e.RightDelay * e.sampleRate)

	feedback := volume.Volume(e.Feedback)

	crossEcho := e.PanDelay >= 0.5

	for c := 0; c < dry.Channels; c++ {
		s := dry.StaticMatrix[c]
		switch c {
		case 0:
			e.delayBufL.Write(s)
		case 1:
			e.delayBufR.Write(s)
		}
	}

	type delayInfo struct {
		buf   *util.RingBuffer[volume.Volume]
		delay int
	}

	var delayBuf [2]delayInfo

	lbuf := 0
	rbuf := 1
	if crossEcho {
		lbuf = 1
		rbuf = 0
	}

	delayBuf[lbuf] = delayInfo{
		buf:   &e.delayBufL,
		delay: ldelay,
	}
	delayBuf[rbuf] = delayInfo{
		buf:   &e.delayBufR,
		delay: rdelay,
	}

	for c := 0; c < dry.Channels; c++ {
		dryPre := dry.StaticMatrix[c]
		d := delayBuf[c]

		if d.buf == nil {
			continue
		}

		// Calculate the mix
		var wetPre [1]volume.Volume
		d.buf.ReadFrom(d.delay, wetPre[:])
		w := dryPre*dryMix + wetPre[0]*wetMix
		wet.StaticMatrix[c] = w
		d.buf.Accumulate(w * feedback)
	}

	return wet
}

func (e *EchoFilter) UpdateEnv(val int8) {

}
