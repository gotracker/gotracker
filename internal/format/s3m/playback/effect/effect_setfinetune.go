package effect

import (
	"fmt"

	s3mfile "github.com/gotracker/goaudiofile/music/tracked/s3m"

	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// SetFinetune defines a mod-style set finetune effect
type SetFinetune uint8 // 'S2x'

// PreStart triggers when the effect enters onto the channel state
func (e SetFinetune) PreStart(cs intf.Channel, p intf.Playback) {
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
	cs.GetTargetInst().SetC2Spd(note.C2SPD(float32(s3mfile.DefaultC2Spd) * multiplier))
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetFinetune) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e SetFinetune) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e SetFinetune) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e SetFinetune) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
