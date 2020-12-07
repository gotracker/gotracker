package effect

import "s3mplayer/internal/player/intf"

type EffectPortaToNote uint8 // 'G'

func (e EffectPortaToNote) PreStart(cs intf.Channel, ss intf.Song) {
}

func (e EffectPortaToNote) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()

	cmd := cs.GetData()
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
