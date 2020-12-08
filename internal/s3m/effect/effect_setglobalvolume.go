package effect

import (
	"fmt"
	"gotracker/internal/player/intf"
	"gotracker/internal/s3m/util"
)

type EffectSetGlobalVolume uint8 // 'V'

func (e EffectSetGlobalVolume) PreStart(cs intf.Channel, ss intf.Song) {
	ss.SetGlobalVolume(util.VolumeFromS3M(uint8(e)))
}

func (e EffectSetGlobalVolume) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

func (e EffectSetGlobalVolume) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

func (e EffectSetGlobalVolume) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e EffectSetGlobalVolume) String() string {
	return fmt.Sprintf("V%0.2x", uint8(e))
}
