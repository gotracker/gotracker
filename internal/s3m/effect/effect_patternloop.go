package effect

import (
	"fmt"
	"gotracker/internal/player/intf"
)

type EffectPatternLoop uint8 // 'SBx'

func (e EffectPatternLoop) PreStart(cs intf.Channel, ss intf.Song) {
}

func (e EffectPatternLoop) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xF

	if x == 0 {
		// set loop
		ss.SetPatternLoopStart()
	} else {
		ss.SetPatternLoopEnd(x)
	}
}

func (e EffectPatternLoop) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

func (e EffectPatternLoop) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e EffectPatternLoop) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
