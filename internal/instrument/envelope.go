package instrument

import (
	"gotracker/internal/player/note"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"
)

// EnvPoint is a point for the envelope
type EnvPoint struct {
	Ticks int
	Y     interface{}
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
}

type envState struct {
	Pos            int
	TicksRemaining int
	Value          interface{}
}

type pcmState struct {
	fadeoutVol    volume.Volume
	keyOn         bool
	fadingOut     bool
	volEnvState   envState
	panEnvState   envState
	pitchEnvState envState
	prevKeyOn     bool
}

func newPcmState() *pcmState {
	ed := pcmState{
		fadeoutVol: volume.Volume(1.0),
		volEnvState: envState{
			Value: volume.Volume(1.0),
		},
		panEnvState: envState{
			Value: panning.CenterAhead,
		},
		pitchEnvState: envState{
			Value: note.PeriodDelta(0),
		},
	}
	return &ed
}

func (ed *pcmState) advance(volEnv *InstEnv, panEnv *InstEnv, pitchEnv *InstEnv) {
	ed.advanceEnv(&ed.volEnvState, volEnv, ed.updateVolEnv)
	ed.advanceEnv(&ed.panEnvState, panEnv, ed.updatePanEnv)
	ed.advanceEnv(&ed.pitchEnvState, pitchEnv, ed.updatePitchEnv)
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
	ed.volEnvState.Value = a + volume.Volume(t)*(b-a)
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
	ed.panEnvState.Value = panning.Position{
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
	ed.pitchEnvState.Value = a + note.PeriodDelta(t)*(b-a)
}

type envUpdateFunc func(t float32, y0 interface{}, y1 interface{})

func (ed *pcmState) advanceEnv(state *envState, env *InstEnv, update envUpdateFunc) {
	if env.Enabled {
		state.TicksRemaining--
		if state.TicksRemaining <= 0 {
			state.Pos++
			cur, p := ed.getEnv(state.Pos, env)
			state.Pos = p
			state.TicksRemaining = cur.Ticks
			update(0, cur.Y, cur.Y)
		} else {
			cur, _ := ed.getEnv(state.Pos, env)
			next, _ := ed.getEnv(state.Pos+1, env)
			t := float32(0)
			if cur.Ticks > 0 {
				t = float32(state.TicksRemaining) / float32(cur.Ticks)
			}
			t = 1.0 - t
			update(t, cur.Y, next.Y)
		}
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

func (ed *pcmState) updateEnv(state *envState, env *InstEnv, update envUpdateFunc) {
	if !env.Enabled {
		// not active, don't do anything
		return
	}
	cur, p := ed.getEnv(state.Pos, env)
	state.Pos = p
	state.TicksRemaining = cur.Ticks
	update(0, cur.Y, cur.Y)
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

func (ed *pcmState) setEnvelopePosition(ticks int, state *envState, env *InstEnv, update envUpdateFunc) {
	state.Pos = 0
	state.TicksRemaining = 0
	for ticks > 0 {
		ed.updateEnv(state, env, update)
		if ticks >= state.TicksRemaining {
			ticks -= state.TicksRemaining
			state.TicksRemaining = 0
		} else {
			state.TicksRemaining -= ticks
			ticks = 0
		}
	}
	ed.updateEnv(state, env, update)
}
