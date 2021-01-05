package intf

import (
	"fmt"
)

// Effect is an interface to command/effect
type Effect interface {
	fmt.Stringer
	PreStart(cs Channel, p Playback)
	Start(cs Channel, p Playback)
	Tick(cs Channel, p Playback, currentTick int)
	Stop(cs Channel, p Playback, lastTick int)
}

// CombinedEffect specifies multiple simultaneous effects into one
type CombinedEffect struct {
	Effects []Effect
}

// PreStart triggers when the effect enters onto the channel state
func (e CombinedEffect) PreStart(cs Channel, p Playback) {
	for _, effect := range e.Effects {
		effect.PreStart(cs, p)
	}
}

// Start triggers on the first tick, but before the Tick() function is called
func (e CombinedEffect) Start(cs Channel, p Playback) {
	for _, effect := range e.Effects {
		effect.Start(cs, p)
	}
}

// Tick is called on every tick
func (e CombinedEffect) Tick(cs Channel, p Playback, currentTick int) {
	for _, effect := range e.Effects {
		effect.Tick(cs, p, currentTick)
	}
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e CombinedEffect) Stop(cs Channel, p Playback, lastTick int) {
	for _, effect := range e.Effects {
		effect.Stop(cs, p, lastTick)
	}
}

// String returns the string for the effect list
func (e CombinedEffect) String() string {
	for _, eff := range e.Effects {
		s := eff.String()
		if s != "" {
			return s
		}
	}
	return ""
}
