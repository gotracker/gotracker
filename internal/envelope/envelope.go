package envelope

import (
	"gotracker/internal/loop"
)

// State is the state information about an envelope
type State struct {
	position int
	length   int
	stopped  bool
	env      *Envelope
}

// Stopped returns true if the envelope state is stopped
func (e *State) Stopped() bool {
	return e.stopped
}

// Stop stops the envelope state
func (e *State) Stop() {
	e.stopped = true
}

// Reset resets the envelope
func (e *State) Reset(env *Envelope) {
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

func (e *State) calcLoopedPos(keyOn bool) (int, bool) {
	nPoints := len(e.env.Values)
	var looped bool
	pos, _ := loop.CalcLoopPos(&e.env.Loop, &e.env.Sustain, e.position, nPoints, keyOn)
	if (keyOn && e.env.Sustain.Enabled()) || e.env.Loop.Enabled() {
		looped = true
	}
	return pos, looped
}

// GetCurrentValue returns the current value
func (e *State) GetCurrentValue(keyOn bool) (*EnvPoint, float32) {
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
			if e.env.Sustain.Enabled() && keyOn && e.env.Sustain.Length() == 0 {
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

// Advance advances the state by 1 tick
func (e *State) Advance(keyOn bool, prevKeyOn bool) bool {
	if e.stopped {
		return false
	}

	if e.env.Sustain.Enabled() && keyOn {
		if e.env.Sustain.Length() == 0 {
			return false
		}
	} else if e.env.Loop.Enabled() {
		if e.env.Loop.Length() == 0 {
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
