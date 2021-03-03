package filter

import (
	"gotracker/internal/player/intf"

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

func (e *EchoFilterFactory) Factory() intf.FilterFactory {
	return func(sampleRate float32) intf.Filter {
		echo := EchoFilter{
			EchoFilterSettings: e.EchoFilterSettings,
			sampleRate:         sampleRate,
		}
		return &echo
	}
}

//===========

type EchoFilter struct {
	EchoFilterSettings
	sampleRate float32
	delayBufL  []volume.Volume
	delayBufR  []volume.Volume
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
			e.delayBufL = append(e.delayBufL, s)
		case 1:
			e.delayBufR = append(e.delayBufR, s)
		}
	}

	for c := range wet {
		var buf []volume.Volume
		switch {
		case (c == 0) || (crossEcho && c == 1):
			if len(e.delayBufL) >= ldelay {
				pos := len(e.delayBufL) - ldelay
				e.delayBufL = e.delayBufL[pos:]
			}
			buf = e.delayBufL
		case (c == 1) || (crossEcho && c == 0):
			if len(e.delayBufR) >= rdelay {
				pos := len(e.delayBufR) - rdelay
				e.delayBufR = e.delayBufR[pos:]
			}
			buf = e.delayBufR
		}
		if buf == nil {
			continue
		}

		// Calculate the mix
		wetPre := buf[0]
		dryPre := dry[c]
		w := dryPre*dryMix + wetPre*wetMix
		wet[c] = w
		buf[len(buf)-1] += w * feedback
	}

	return wet
}

func (e *EchoFilter) UpdateEnv(val float32) {

}
