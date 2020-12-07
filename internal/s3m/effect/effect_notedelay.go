package effect

import "s3mplayer/internal/player/intf"

type EffectNoteDelay uint8 // 'SDx'

func (e EffectNoteDelay) PreStart(cs intf.Channel, ss intf.Song) {
}

func (e EffectNoteDelay) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()

	cs.SetNotePlayTick(int(uint8(e) & 0x0F))
}

func (e EffectNoteDelay) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

func (e EffectNoteDelay) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}
