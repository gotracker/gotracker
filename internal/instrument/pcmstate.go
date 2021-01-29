package instrument

import (
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/envelope"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

type pcmState struct {
	fadeoutVol          volume.Volume
	keyOn               bool
	fadingOut           bool
	volEnvEnabled       bool
	volEnvState         envelope.State
	volEnvValue         volume.Volume
	panEnvEnabled       bool
	panEnvState         envelope.State
	panEnvValue         panning.Position
	pitchFiltEnvEnabled bool
	pitchFiltEnvState   envelope.State
	pitchFiltEnvMode    bool
	pitchEnvValue       note.PeriodDelta
	filtEnvValue        float32
	prevKeyOn           bool
}

func newPcmState() *pcmState {
	ed := pcmState{
		fadeoutVol:    volume.Volume(1.0),
		volEnvValue:   volume.Volume(1.0),
		panEnvValue:   panning.CenterAhead,
		pitchEnvValue: note.PeriodDelta(0),
		filtEnvValue:  1,
	}
	return &ed
}

func (ed *pcmState) advance(nc intf.NoteControl, volEnv *envelope.Envelope, panEnv *envelope.Envelope, pitchEnv *envelope.Envelope) {
	if ed.volEnvEnabled {
		ed.advanceEnv(&ed.volEnvState, volEnv, nc, ed.updateVolEnv, true)
	}
	if ed.panEnvEnabled {
		ed.advanceEnv(&ed.panEnvState, panEnv, nc, ed.updatePanEnv, true)
	}
	if ed.pitchFiltEnvEnabled {
		var pitchFiltEnvFunc envUpdateFunc = ed.updatePitchEnv
		if ed.pitchFiltEnvMode {
			pitchFiltEnvFunc = ed.updateFiltEnv
		}
		ed.advanceEnv(&ed.pitchFiltEnvState, pitchEnv, nc, pitchFiltEnvFunc, true)
	}
}

func (ed *pcmState) updateVolEnv(t float32, y0, y1 interface{}) {
	switch {
	case t < 0:
		t = 0
	case t > 1:
		t = 1
	}
	a := volume.Volume(1)
	b := volume.Volume(0)
	if y0 != nil {
		a = y0.(volume.Volume)
	}
	if y1 != nil {
		b = y1.(volume.Volume)
	}
	v := a + volume.Volume(t)*(b-a)
	switch {
	case v < 0:
		v = 0
	case v > 1:
		v = 1
	}
	ed.volEnvValue = v
}

func (ed *pcmState) updatePanEnv(t float32, y0, y1 interface{}) {
	a := panning.CenterAhead
	b := panning.CenterAhead
	if y0 != nil {
		a = y0.(panning.Position)
	}
	if y1 != nil {
		b = y1.(panning.Position)
	}
	ed.panEnvValue = panning.Position{
		Angle:    a.Angle + t*(b.Angle-a.Angle),
		Distance: a.Distance + t*(b.Distance-a.Distance),
	}
}

func (ed *pcmState) updatePitchEnv(t float32, y0, y1 interface{}) {
	a := note.PeriodDelta(0)
	b := note.PeriodDelta(0)
	if y0 != nil {
		a = note.PeriodDelta(int8(uint8(y0.(float32) * 128)))
	}
	if y1 != nil {
		a = note.PeriodDelta(int8(uint8(y1.(float32) * 128)))
	}
	ed.pitchEnvValue = a + note.PeriodDelta(t)*(b-a)
}

func (ed *pcmState) updateFiltEnv(t float32, y0, y1 interface{}) {
	a := float32(0)
	b := float32(0)
	if y0 != nil {
		a = y0.(float32)
	}
	if y1 != nil {
		b = y1.(float32)
	}
	lerp := t * (b - a)
	v := a + lerp
	ed.filtEnvValue = v / 255
}

type envUpdateFunc func(t float32, y0 interface{}, y1 interface{})

func (ed *pcmState) advanceEnv(state *envelope.State, env *envelope.Envelope, nc intf.NoteControl, update envUpdateFunc, runTick bool) {
	if state.Stopped() {
		return
	}

	cur, next, t := state.GetCurrentValue(ed.keyOn)

	var finishing bool
	if runTick {
		finishing = state.Advance(ed.keyOn, ed.prevKeyOn)
	}

	if cur != nil {
		update(t, cur.Y0, next.Y0)
	}

	if finishing {
		env.OnFinished(nc)
	}
}

func (ed *pcmState) setEnvelopePosition(ticks int, state *envelope.State, env *envelope.Envelope, nc intf.NoteControl, update envUpdateFunc) {
	state.Reset(env)
	for ticks >= 0 {
		ed.advanceEnv(state, env, nc, update, true)
		ticks--
	}
}
