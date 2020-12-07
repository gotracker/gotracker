package effect

import "gotracker/internal/player/intf"

type EffectSetTempo uint8 // 'T'

func (e EffectSetTempo) PreStart(cs intf.Channel, ss intf.Song) {
	if e > 0x20 {
		ss.SetTempo(int(e))
	}
}

func (e EffectSetTempo) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

func (e EffectSetTempo) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	switch uint8(e >> 4) {
	case 0: // decrease tempo
		if currentTick != 0 {
			mem := cs.GetMemory()
			val := int(mem.TempoDecrease(uint8(e & 0x0F)))
			ss.DecreaseTempo(val)
		}
	case 1: // increase tempo
		if currentTick != 0 {
			mem := cs.GetMemory()
			val := int(mem.TempoIncrease(uint8(e & 0x0F)))
			ss.IncreaseTempo(val)
		}
	default:
	}
}

func (e EffectSetTempo) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}
