package effect

import (
	"fmt"
	"gotracker/internal/player/intf"
)

// VolumeSlide defines a volume slide effect
type VolumeSlide uint8 // 'D'

// PreStart triggers when the effect enters onto the channel state
func (e VolumeSlide) PreStart(cs intf.Channel, ss intf.Song) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e VolumeSlide) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e VolumeSlide) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	v := cs.GetEffectSharedMemory(uint8(e))
	x := uint8(v >> 4)
	y := uint8(v & 0x0F)

	if x == 0 { // decrease every tick
		if y == 0x0F {
			doVolSlide(cs, -float32(y), 1.0)
		} else if currentTick != 0 {
			doVolSlide(cs, -float32(y), 1.0)
		}
	} else if y == 0 { // increase every tick
		if currentTick != 0 {
			doVolSlide(cs, float32(x), 1.0)
		}
	} else if x == 0x0F { // finely decrease on the first tick
		if y != 0x0F && currentTick == 0 {
			doVolSlide(cs, -float32(y), 1.0)
		}
	} else if y == 0x0F { // finely increase on the first tick
		if x != 0x0F && currentTick == 0 {
			doVolSlide(cs, float32(x), 1.0)
		}
	}
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e VolumeSlide) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e VolumeSlide) String() string {
	return fmt.Sprintf("D%0.2x", uint8(e))
}
