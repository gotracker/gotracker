package intf

import "fmt"

// Effect is an interface to command/effect
type Effect interface {
	//fmt.Stringer
}

type effectPreStartIntf[TMemory, TChannelData any] interface {
	PreStart(Channel[TMemory, TChannelData], Playback) error
}

// EffectPreStart triggers when the effect enters onto the channel state
func EffectPreStart[TMemory, TChannelData any](e Effect, cs Channel[TMemory, TChannelData], p Playback) error {
	if eff, ok := e.(effectPreStartIntf[TMemory, TChannelData]); ok {
		if err := eff.PreStart(cs, p); err != nil {
			return err
		}
	}
	return nil
}

type effectStartIntf[TMemory, TChannelData any] interface {
	Start(Channel[TMemory, TChannelData], Playback) error
}

// EffectStart triggers on the first tick, but before the Tick() function is called
func EffectStart[TMemory, TChannelData any](e Effect, cs Channel[TMemory, TChannelData], p Playback) error {
	if eff, ok := e.(effectStartIntf[TMemory, TChannelData]); ok {
		if err := eff.Start(cs, p); err != nil {
			return err
		}
	}
	return nil
}

type effectTickIntf[TMemory, TChannelData any] interface {
	Tick(Channel[TMemory, TChannelData], Playback, int) error
}

// EffectTick is called on every tick
func EffectTick[TMemory, TChannelData any](e Effect, cs Channel[TMemory, TChannelData], p Playback, currentTick int) error {
	if eff, ok := e.(effectTickIntf[TMemory, TChannelData]); ok {
		if err := eff.Tick(cs, p, currentTick); err != nil {
			return err
		}
	}
	return nil
}

type effectStopIntf[TMemory, TChannelData any] interface {
	Stop(Channel[TMemory, TChannelData], Playback, int) error
}

// EffectStop is called on the last tick of the row, but after the Tick() function is called
func EffectStop[TMemory, TChannelData any](e Effect, cs Channel[TMemory, TChannelData], p Playback, lastTick int) error {
	if eff, ok := e.(effectStopIntf[TMemory, TChannelData]); ok {
		if err := eff.Stop(cs, p, lastTick); err != nil {
			return err
		}
	}
	return nil
}

// CombinedEffect specifies multiple simultaneous effects into one
type CombinedEffect[TMemory, TChannelData any] struct {
	Effects []Effect
}

// PreStart triggers when the effect enters onto the channel state
func (e CombinedEffect[TMemory, TChannelData]) PreStart(cs Channel[TMemory, TChannelData], p Playback) error {
	for _, effect := range e.Effects {
		if err := EffectPreStart(effect, cs, p); err != nil {
			return err
		}
	}
	return nil
}

// Start triggers on the first tick, but before the Tick() function is called
func (e CombinedEffect[TMemory, TChannelData]) Start(cs Channel[TMemory, TChannelData], p Playback) error {
	for _, effect := range e.Effects {
		if err := EffectStart(effect, cs, p); err != nil {
			return err
		}
	}
	return nil
}

// Tick is called on every tick
func (e CombinedEffect[TMemory, TChannelData]) Tick(cs Channel[TMemory, TChannelData], p Playback, currentTick int) error {
	for _, effect := range e.Effects {
		if err := EffectTick(effect, cs, p, currentTick); err != nil {
			return err
		}
	}
	return nil
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e CombinedEffect[TMemory, TChannelData]) Stop(cs Channel[TMemory, TChannelData], p Playback, lastTick int) error {
	for _, effect := range e.Effects {
		if err := EffectStop(effect, cs, p, lastTick); err != nil {
			return err
		}
	}
	return nil
}

// String returns the string for the effect list
func (e CombinedEffect[TMemory, TChannelData]) String() string {
	for _, eff := range e.Effects {
		s := fmt.Sprint(eff)
		if s != "" {
			return s
		}
	}
	return ""
}

// DoEffect runs the standard tick lifetime of an effect
func DoEffect[TMemory, TChannelData any](e Effect, cs Channel[TMemory, TChannelData], p Playback, currentTick int, lastTick bool) error {
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
