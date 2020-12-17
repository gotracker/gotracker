package intf

import (
	"fmt"
)

// Effect is an interface to command/effect
type Effect interface {
	fmt.Stringer
	PreStart(cs Channel, ss Song)
	Start(cs Channel, ss Song)
	Tick(cs Channel, ss Song, currentTick int)
	Stop(cs Channel, ss Song, lastTick int)
}

// CombinedEffect specifies multiple simultaneous effects into one
type CombinedEffect struct {
	Effects []Effect
}

// PreStart triggers when the effect enters onto the channel state
func (e CombinedEffect) PreStart(cs Channel, ss Song) {
	for _, effect := range e.Effects {
		effect.PreStart(cs, ss)
	}
}

// Start triggers on the first tick, but before the Tick() function is called
func (e CombinedEffect) Start(cs Channel, ss Song) {
	for _, effect := range e.Effects {
		effect.Start(cs, ss)
	}
}

// Tick is called on every tick
func (e CombinedEffect) Tick(cs Channel, ss Song, currentTick int) {
	for _, effect := range e.Effects {
		effect.Tick(cs, ss, currentTick)
	}
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e CombinedEffect) Stop(cs Channel, ss Song, lastTick int) {
	for _, effect := range e.Effects {
		effect.Stop(cs, ss, lastTick)
	}
}
