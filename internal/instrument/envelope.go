package instrument

import (
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"
)

// EnvPoint is a point for the envelope
type EnvPoint struct {
	Length int
	Y0     interface{}
	Y1     interface{}
}

// InstEnv is an envelope for instruments
type InstEnv struct {
	Enabled          bool
	LoopEnabled      bool
	SustainEnabled   bool
	LoopStart        int
	LoopEnd          int
	SustainLoopStart int
	SustainLoopEnd   int
	Values           []EnvPoint
	OnFinished       func(intf.NoteControl)
}

type envState struct {
	position int
	length   int
	stopped  bool
	env      *InstEnv
}

func (e *envState) Stopped() bool {
	return e.stopped
}

func (e *envState) Stop() {
	e.stopped = true
}

func (e *envState) Reset(env *InstEnv) {
	e.env = env
	if !e.env.Enabled {
		e.stopped = true
		return
	}

	e.position = 0
	pos, _ := e.calcLoopedPos(true)
	if pos < len(e.env.Values) {
		e.length = e.env.Values[pos].Length
	}
}

func (e *envState) calcLoopedPos(keyOn bool) (int, bool) {
	nPoints := len(e.env.Values)
	var (
		pos    int
		looped bool
	)
	if e.env.SustainEnabled && keyOn {
		pos, _ = calcLoopPosMode2(e.position, nPoints, e.env.SustainLoopStart, e.env.SustainLoopEnd)
		looped = true
	} else if e.env.LoopEnabled {
		pos, _ = calcLoopPosMode2(e.position, nPoints, e.env.LoopStart, e.env.LoopEnd)
		looped = true
	} else {
		pos = e.position
		looped = false
	}
	return pos, looped
}

func (e *envState) GetCurrentValue(keyOn bool) (*EnvPoint, float32) {
	if e.stopped {
		return nil, 0
	}

	pos, looped := e.calcLoopedPos(keyOn)
	if pos >= len(e.env.Values) {
		return nil, 0
	}

	cur := &e.env.Values[pos]
	t := float32(0)
	if cur.Length > 0 {
		l := float32(e.length)
		if looped {
			if e.env.SustainEnabled && keyOn && calcLoopLen(e.env.SustainLoopStart, e.env.SustainLoopEnd) == 0 {
				l = 0
			} else {
				l = float32(e.length)
			}
		}
		t = 1 - (l / float32(cur.Length))
	}
	switch {
	case t < 0:
		t = 0
	case t > 1:
		t = 1
	}
	return cur, t
}

func (e *envState) Advance(keyOn bool, prevKeyOn bool) bool {
	if e.stopped {
		return false
	}

	if e.env.SustainEnabled && keyOn {
		if calcLoopLen(e.env.SustainLoopStart, e.env.SustainLoopEnd) == 0 {
			return false
		}
	} else if e.env.LoopEnabled {
		if calcLoopLen(e.env.LoopStart, e.env.LoopEnd) == 0 {
			return false
		}
	}

	e.length--
	if e.length > 0 {
		return false
	}
	if keyOn != prevKeyOn && prevKeyOn {
		p, _ := e.calcLoopedPos(prevKeyOn)
		e.position = p
	}

	e.position++
	pos, _ := e.calcLoopedPos(keyOn)
	if pos >= len(e.env.Values) {
		e.stopped = true
		return true
	}

	e.length = e.env.Values[pos].Length
	return false
}

type pcmState struct {
	fadeoutVol    volume.Volume
	keyOn         bool
	fadingOut     bool
	volEnvState   envState
	volEnvValue   volume.Volume
	panEnvState   envState
	panEnvValue   panning.Position
	pitchEnvState envState
	pitchEnvValue note.PeriodDelta
	prevKeyOn     bool
}

func newPcmState() *pcmState {
	ed := pcmState{
		fadeoutVol:    volume.Volume(1.0),
		volEnvValue:   volume.Volume(1.0),
		panEnvValue:   panning.CenterAhead,
		pitchEnvValue: note.PeriodDelta(0),
	}
	return &ed
}

func (ed *pcmState) advance(nc intf.NoteControl, volEnv *InstEnv, panEnv *InstEnv, pitchEnv *InstEnv) {
	ed.advanceEnv(&ed.volEnvState, volEnv, nc, ed.updateVolEnv, true)
	ed.advanceEnv(&ed.panEnvState, panEnv, nc, ed.updatePanEnv, true)
	ed.advanceEnv(&ed.pitchEnvState, pitchEnv, nc, ed.updatePitchEnv, true)
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
		a = y0.(note.PeriodDelta)
	}
	if y1 != nil {
		b = y1.(note.PeriodDelta)
	}
	ed.pitchEnvValue = a + note.PeriodDelta(t)*(b-a)
}

type envUpdateFunc func(t float32, y0 interface{}, y1 interface{})

func (ed *pcmState) advanceEnv(state *envState, env *InstEnv, nc intf.NoteControl, update envUpdateFunc, runTick bool) {
	if state.Stopped() {
		return
	}

	cur, t := state.GetCurrentValue(ed.keyOn)

	var finishing bool
	if runTick {
		finishing = state.Advance(ed.keyOn, ed.prevKeyOn)
	}

	if cur != nil {
		update(t, cur.Y0, cur.Y1)
	}

	if finishing {
		env.OnFinished(nc)
	}
}

func (ed *pcmState) setEnvelopePosition(ticks int, state *envState, env *InstEnv, nc intf.NoteControl, update envUpdateFunc) {
	state.Reset(env)
	for ticks >= 0 {
		ed.advanceEnv(state, env, nc, update, true)
		ticks--
	}
}
