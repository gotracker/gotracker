package intf

import "fmt"

// Effect is an interface to command/effect
type Effect interface {
	//fmt.Stringer
}

type effectPreStartIntf interface {
	PreStart(Channel, Playback) error
}

// EffectPreStart triggers when the effect enters onto the channel state
func EffectPreStart(e Effect, cs Channel, p Playback) error {
	if eff, ok := e.(effectPreStartIntf); ok {
		if err := eff.PreStart(cs, p); err != nil {
			return err
		}
	}
	return nil
}

type effectStartIntf interface {
	Start(Channel, Playback) error
}

// EffectStart triggers on the first tick, but before the Tick() function is called
func EffectStart(e Effect, cs Channel, p Playback) error {
	if eff, ok := e.(effectStartIntf); ok {
		if err := eff.Start(cs, p); err != nil {
			return err
		}
	}
	return nil
}

type effectTickIntf interface {
	Tick(Channel, Playback, int) error
}

// EffectTick is called on every tick
func EffectTick(e Effect, cs Channel, p Playback, currentTick int) error {
	if eff, ok := e.(effectTickIntf); ok {
		if err := eff.Tick(cs, p, currentTick); err != nil {
			return err
		}
	}
	return nil
}

type effectStopIntf interface {
	Stop(Channel, Playback, int) error
}

// EffectStop is called on the last tick of the row, but after the Tick() function is called
func EffectStop(e Effect, cs Channel, p Playback, lastTick int) error {
	if eff, ok := e.(effectStopIntf); ok {
		if err := eff.Stop(cs, p, lastTick); err != nil {
			return err
		}
	}
	return nil
}

// CombinedEffect specifies multiple simultaneous effects into one
type CombinedEffect struct {
	Effects []Effect
}

// PreStart triggers when the effect enters onto the channel state
func (e CombinedEffect) PreStart(cs Channel, p Playback) error {
	for _, effect := range e.Effects {
		if err := EffectPreStart(effect, cs, p); err != nil {
			return err
		}
	}
	return nil
}

// Start triggers on the first tick, but before the Tick() function is called
func (e CombinedEffect) Start(cs Channel, p Playback) error {
	for _, effect := range e.Effects {
		if err := EffectStart(effect, cs, p); err != nil {
			return err
		}
	}
	return nil
}

// Tick is called on every tick
func (e CombinedEffect) Tick(cs Channel, p Playback, currentTick int) error {
	for _, effect := range e.Effects {
		if err := EffectTick(effect, cs, p, currentTick); err != nil {
			return err
		}
	}
	return nil
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e CombinedEffect) Stop(cs Channel, p Playback, lastTick int) error {
	for _, effect := range e.Effects {
		if err := EffectStop(effect, cs, p, lastTick); err != nil {
			return err
		}
	}
	return nil
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
func DoEffect(e Effect, cs Channel, p Playback, currentTick int, lastTick bool) error {
	if e == nil {
		return nil
	}

	if currentTick == 0 {
		if err := EffectStart(e, cs, p); err != nil {
			return err
		}
	}
	if err := EffectTick(e, cs, p, currentTick); err != nil {
		return err
	}
	if lastTick {
		if err := EffectStop(e, cs, p, currentTick); err != nil {
			return err
		}
	}
	return nil
}
