package intf

type Effect interface {
	PreStart(cs Channel, ss Song)
	Start(cs Channel, ss Song)
	Tick(cs Channel, ss Song, currentTick int)
	Stop(cs Channel, ss Song, lastTick int)
}

type CombinedEffect struct {
	Effects []Effect
}

func (e CombinedEffect) PreStart(cs Channel, ss Song) {
	for _, effect := range e.Effects {
		effect.PreStart(cs, ss)
	}
}

func (e CombinedEffect) Start(cs Channel, ss Song) {
	for _, effect := range e.Effects {
		effect.Start(cs, ss)
	}
}

func (e CombinedEffect) Tick(cs Channel, ss Song, currentTick int) {
	for _, effect := range e.Effects {
		effect.Tick(cs, ss, currentTick)
	}
}

func (e CombinedEffect) Stop(cs Channel, ss Song, lastTick int) {
	for _, effect := range e.Effects {
		effect.Stop(cs, ss, lastTick)
	}
}
