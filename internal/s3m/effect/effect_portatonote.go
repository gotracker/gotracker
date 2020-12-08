package effect

import (
	"fmt"
	"gotracker/internal/player/intf"
	"gotracker/internal/s3m/channel"
)

type EffectPortaToNote uint8 // 'G'

func (e EffectPortaToNote) PreStart(cs intf.Channel, ss intf.Song) {
}

func (e EffectPortaToNote) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()

	cmd := cs.GetData().(*channel.Data)
	if cmd == nil {
		return
	}

	if cmd.What.HasNote() {
		cs.SetPortaTargetPeriod(cs.GetTargetPeriod())
	}
	cs.SetTargetPeriod(cs.GetPeriod())
	cs.SetTargetPos(cs.GetPos())
}

func (e EffectPortaToNote) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	mem := cs.GetMemory()
	xx := mem.PortaToNote(uint8(e))

	period := cs.GetPeriod()
	ptp := cs.GetPortaTargetPeriod()
	if currentTick != 0 && period != ptp {
		if period > ptp {
			doPortaUpToNote(cs, float32(xx), 4, ptp) // subtracts
		} else {
			doPortaDownToNote(cs, float32(xx), 4, ptp) // adds
		}
		cs.SetTargetPeriod(cs.GetPeriod())
	}
}

func (e EffectPortaToNote) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e EffectPortaToNote) String() string {
	return fmt.Sprintf("G%0.2x", uint8(e))
}
