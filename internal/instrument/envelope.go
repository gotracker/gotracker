package instrument

import (
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
	Enabled        bool
	LoopEnabled    bool
	SustainEnabled bool
	LoopStart      int
	LoopEnd        int
	SustainIndex   int
	Values         []EnvPoint
}

type envData struct {
	fadeoutVol           volume.Volume
	keyOn                bool
	volEnvPos            int
	volEnvTicksRemaining int
	volEnvValue          volume.Volume
	panEnvPos            int
	panEnvTicksRemaining int
	panEnvValue          panning.Position
	prevKeyOn            bool
}

func newEnvData() *envData {
	ed := envData{
		fadeoutVol:  volume.Volume(1.0),
		volEnvValue: volume.Volume(1.0),
		panEnvValue: panning.CenterAhead,
	}
	return &ed
}

func (ed *envData) advance(volEnv *InstEnv, panEnv *InstEnv) {
	ed.advanceEnv(&ed.volEnvPos, &ed.volEnvTicksRemaining, volEnv, ed.updateVolEnv)
	ed.advanceEnv(&ed.panEnvPos, &ed.panEnvTicksRemaining, panEnv, ed.updatePanEnv)
}

func (ed *envData) updateVolEnv(t float32, y0, y1 interface{}) {
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

func (ed *envData) updatePanEnv(t float32, y0, y1 interface{}) {
	a := panning.CenterAhead
	b := panning.CenterAhead
	if y0 != nil {
		a = y0.(panning.Position)
	}
	if y1 != nil {
		b = y1.(panning.Position)
	}
	ed.panEnvValue.Angle = a.Angle + t*(b.Angle-a.Angle)
	ed.panEnvValue.Distance = a.Distance + t*(b.Distance-a.Distance)
}

type envUpdateFunc func(t float32, y0 interface{}, y1 interface{})

func (ed *envData) advanceEnv(pos *int, rem *int, env *InstEnv, update envUpdateFunc) {
	if env.Enabled {
		tick := false
		if env.SustainEnabled {
			if !ed.keyOn {
				tick = true
			} else if *pos < env.SustainIndex {
				tick = true
			}
		} else {
			tick = true
		}
		if tick {
			*rem--
			if *rem <= 0 {
				*pos++
				cur, p := ed.getEnv(*pos, env)
				*pos = p
				*rem = cur.Ticks
				update(0, cur.Y, cur.Y)
			} else {
				cur, _ := ed.getEnv(*pos, env)
				next, _ := ed.getEnv(*pos+1, env)
				t := float32(0)
				if cur.Ticks > 0 {
					t = float32(*rem) / float32(cur.Ticks)
				}
				t = 1.0 - t
				update(t, cur.Y, next.Y)
			}
		}
	}
}

func (ed *envData) updateEnv(pos *int, rem *int, env *InstEnv, update envUpdateFunc) {
	if !env.Enabled {
		// not active, don't do anything
		return
	}
	cur, p := ed.getEnv(*pos, env)
	*pos = p
	*rem = cur.Ticks
	update(0, cur.Y, cur.Y)
}

func (ed *envData) getEnv(pos int, env *InstEnv) (EnvPoint, int) {
	if env.LoopEnabled {
		if pos >= env.LoopEnd {
			loopLen := env.LoopEnd - env.LoopStart
			pos = env.LoopStart + ((pos - env.LoopEnd) % loopLen)
		}
	}
	nPoints := len(env.Values)
	if pos >= nPoints {
		pos = nPoints - 1
	}
	if pos < 0 || pos >= nPoints {
		return EnvPoint{}, 0
	}

	cur := env.Values[pos]
	return cur, pos
}

func (ed *envData) setEnvelopePosition(ticks int, pos *int, rem *int, env *InstEnv, update envUpdateFunc) {
	*pos = 0
	*rem = 0
	for ticks > 0 {
		ed.updateEnv(pos, rem, env, update)
		if ticks >= *rem {
			ticks -= *rem
			*rem = 0
		} else {
			*rem -= ticks
			ticks = 0
		}
	}
	ed.updateEnv(pos, rem, env, update)
}
