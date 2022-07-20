package period

import (
	"github.com/gotracker/gotracker/internal/format/internal/util"
	"github.com/gotracker/gotracker/internal/song/note"
	"github.com/gotracker/voice/period"
)

type AmigaPeriodKeyboard []AmigaOctavePeriods

func (p AmigaPeriodKeyboard) GetPeriodFromNote(n note.Note) (AmigaPeriod, bool) {
	nt, ok := n.(note.Normal)
	if !ok {
		return 0, false
	}
	st := note.Semitone(nt)
	o := st.Octave()
	if int(o) >= len(p) {
		return 0, false
	}

	po := p[o]
	return po.GetPeriodFromKey(st.Key())
}

func (p AmigaPeriodKeyboard) GetPeriodFromFinetunes(ft note.Finetune) (AmigaPeriod, bool) {
	nft := ft / 64
	p0, ok := p.GetPeriodFromNote(note.Normal(nft))
	if !ok {
		return 0, false
	}
	if ft %= 64; ft == 0 {
		return p0, true
	}
	p1, ok := p.GetPeriodFromNote(note.Normal(nft + 1))
	if !ok {
		return p0, true
	}

	t := float64(ft) / 64
	return p0.Lerp(t, p1), true
}

type AmigaOctavePeriods [12]AmigaPeriod

func (o AmigaOctavePeriods) GetPeriodFromKey(key note.Key) (AmigaPeriod, bool) {
	if key.IsInvalid() {
		return 0, false
	}
	return o[int(key)], true
}

type AmigaPeriod float64

func (p AmigaPeriod) Lerp(t float64, rhs AmigaPeriod) AmigaPeriod {
	return AmigaPeriod(util.LerpFloat64(t, float64(p), float64(rhs)))
}

func (p AmigaPeriod) GetFrequency(baseClockRate period.Frequency) period.Frequency {
	if p == 0 {
		return 0
	}
	return baseClockRate / period.Frequency(p)
}
