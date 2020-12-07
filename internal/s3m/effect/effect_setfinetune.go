package effect

import (
	"s3mplayer/internal/player/intf"
	"s3mplayer/internal/s3m/util"
)

type EffectSetFinetune uint8 // 'S2x'

func (e EffectSetFinetune) PreStart(cs intf.Channel, ss intf.Song) {
	x := uint8(e) & 0xf

	var multiplier float32
	switch x {
	case 0:
		multiplier = 1.0
	case 1:
		multiplier = 1.0059787157718522061461198134641
	case 2:
		multiplier = 1.0119574315437044122922396269281
	case 3:
		multiplier = 1.0198493363625493244051177807007
	case 4:
		multiplier = 1.0260672007652756187970823867033
	case 5:
		multiplier = 1.034437402845868707401650125553
	case 6:
		multiplier = 1.0430467535573358842520626569413
	case 7:
		multiplier = 1.0471122802821953844314241300969
	case 8:
		multiplier = 0.94403922037546335047231854597632
	case 9:
		multiplier = 0.94953963888556738012674877436327
	case 10:
		multiplier = 0.95480090876479732153533421021165
	case 11:
		multiplier = 0.96209494200645701303360038263781
	case 12:
		multiplier = 0.96938897524811670453186655506397
	case 13:
		multiplier = 0.97680258280521344015305512375942
	case 14:
		multiplier = 0.98433576467774721989716608872414
	case 15:
		multiplier = 0.99007533181872533779744110964965
	default:
		multiplier = 1.0
	}
	cs.GetTargetInst().SetC2Spd(uint16(float32(util.DefaultC2Spd) * multiplier))
}

func (e EffectSetFinetune) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

func (e EffectSetFinetune) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

func (e EffectSetFinetune) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}
