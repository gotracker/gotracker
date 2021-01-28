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
	Stopped  bool
}

func (e *envState) Reset() {
	e.position = -1
	e.length = 0
	e.Stopped = false
}

func (e *envState) Advance() bool {
	e.length--
	return e.length <= 0
}

func (e *envState) Pos() int {
	p := e.position
	if e.length <= 0 {
		p++
	}
	return p
}

func (e *envState) SetPos(newPos int, length int) {
	e.position = newPos
	e.length = length
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
	a := volume.Volume(1)
	b := volume.Volume(0)
	if y0 != nil {
		a = y0.(volume.Volume)
	}
	if y1 != nil {
		b = y1.(volume.Volume)
	}
	ed.volEnvValue = a + volume.Volume(t)*(b-a)
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
	if !env.Enabled || state.Stopped {
		return
	}

	var updateState bool

	if runTick {
		updateState = state.Advance()
	}

	var finishing bool
	cur, p := ed.getEnv(state.Pos(), env)
	if updateState {
		state.SetPos(p, cur.Length)
		looping := env.LoopEnabled || (env.SustainEnabled && ed.keyOn)
		if env.OnFinished != nil && !looping {
			if state.position >= len(env.Values)-1 {
				finishing = true
			}
		}
	}

	t := float32(0)
	if cur.Length > 0 {
		t = float32(state.length) / float32(cur.Length)
	}
	t = 1.0 - t
	update(t, cur.Y0, cur.Y1)

	if finishing {
		env.OnFinished(nc)
		state.Stopped = true
	}
}

func (ed *pcmState) calcEnvPos(env *InstEnv, pos int) int {
	nPoints := len(env.Values)
	if env.SustainEnabled && ed.keyOn {
		pos = calcLoopedSamplePosMode2(pos, nPoints, env.SustainLoopStart, env.SustainLoopEnd)
	} else if env.LoopEnabled {
		pos = calcLoopedSamplePosMode2(pos, nPoints, env.LoopStart, env.LoopEnd)
	}
	if pos >= nPoints {
		pos = nPoints - 1
	}
	return pos
}

func (ed *pcmState) getEnv(pos int, env *InstEnv) (EnvPoint, int) {
	pos = ed.calcEnvPos(env, pos)
	nPoints := len(env.Values)
	if pos < 0 || pos >= nPoints {
		return EnvPoint{}, 0
	}

	cur := env.Values[pos]
	return cur, pos
}

func (ed *pcmState) setEnvelopePosition(ticks int, state *envState, env *InstEnv, nc intf.NoteControl, update envUpdateFunc) {
	state.Reset()
	for ticks >= 0 {
		ed.advanceEnv(state, env, nc, update, true)
		ticks--
	}
}
