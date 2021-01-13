package intf

import "fmt"

// Effect is an interface to command/effect
type Effect interface {
	//fmt.Stringer
}

type effectPreStartIntf interface {
	PreStart(Channel, Playback)
}

// EffectPreStart triggers when the effect enters onto the channel state
func EffectPreStart(e Effect, cs Channel, p Playback) {
	if eff, ok := e.(effectPreStartIntf); ok {
		eff.PreStart(cs, p)
	}
}

type effectStartIntf interface {
	Start(Channel, Playback)
}

// EffectStart triggers on the first tick, but before the Tick() function is called
func EffectStart(e Effect, cs Channel, p Playback) {
	if eff, ok := e.(effectStartIntf); ok {
		eff.Start(cs, p)
	}
}

type effectTickIntf interface {
	Tick(Channel, Playback, int)
}

// EffectTick is called on every tick
func EffectTick(e Effect, cs Channel, p Playback, currentTick int) {
	if eff, ok := e.(effectTickIntf); ok {
		eff.Tick(cs, p, currentTick)
	}
}

type effectStopIntf interface {
	Stop(Channel, Playback, int)
}

// EffectStop is called on the last tick of the row, but after the Tick() function is called
func EffectStop(e Effect, cs Channel, p Playback, lastTick int) {
	if eff, ok := e.(effectStopIntf); ok {
		eff.Stop(cs, p, lastTick)
	}
}

// CombinedEffect specifies multiple simultaneous effects into one
type CombinedEffect struct {
	Effects []Effect
}

// PreStart triggers when the effect enters onto the channel state
func (e CombinedEffect) PreStart(cs Channel, p Playback) {
	for _, effect := range e.Effects {
		EffectPreStart(effect, cs, p)
	}
}

// Start triggers on the first tick, but before the Tick() function is called
func (e CombinedEffect) Start(cs Channel, p Playback) {
	for _, effect := range e.Effects {
		EffectStart(effect, cs, p)
	}
}

// Tick is called on every tick
func (e CombinedEffect) Tick(cs Channel, p Playback, currentTick int) {
	for _, effect := range e.Effects {
		EffectTick(effect, cs, p, currentTick)
	}
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e CombinedEffect) Stop(cs Channel, p Playback, lastTick int) {
	for _, effect := range e.Effects {
		EffectStop(effect, cs, p, lastTick)
	}
}

// String returns the string for the effect list
func (e CombinedEffect) String() string {
	for _, eff := range e.Effects {
		s := fmt.Sprintf("%v", eff)
		if s != "" {
			return s
		}
	}
	return ""
}

// DoEffect runs the standard tick lifetime of an effect
func DoEffect(e Effect, cs Channel, p Playback, currentTick int, lastTick bool) {
	if e == nil {
		return
	}

	if currentTick == 0 {
		EffectStart(e, cs, p)
	}
	EffectTick(e, cs, p, currentTick)
	if lastTick {
		EffectStop(e, cs, p, currentTick)
	}
}
