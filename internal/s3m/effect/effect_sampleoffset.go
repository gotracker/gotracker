package effect

import (
	"fmt"
	"gotracker/internal/player/intf"
)

type EffectSampleOffset uint8 // 'O'

func (e EffectSampleOffset) PreStart(cs intf.Channel, ss intf.Song) {
}

func (e EffectSampleOffset) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
	mem := cs.GetMemory()
	xx := mem.SampleOffset(uint8(e))
	cs.SetTargetPos(float32(xx) * 0x100)
}

func (e EffectSampleOffset) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

func (e EffectSampleOffset) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e EffectSampleOffset) String() string {
	return fmt.Sprintf("O%0.2x", uint8(e))
}
