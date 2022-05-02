package effect

import (
	"fmt"

	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// ChannelVolumeSlide defines a set channel volume effect
type ChannelVolumeSlide uint8 // 'Nxy'

// Start triggers on the first tick, but before the Tick() function is called
func (e ChannelVolumeSlide) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	mem := cs.GetMemory()
	x, y := mem.ChannelVolumeSlide(uint8(e))

	switch {
	case y == 0x0 && x != 0xF:
	case y != 0xF && x == 0x0:
	case y == 0xF:
		vol := cs.GetChannelVolume() + (volume.Volume(x) / 64)
		if vol > 1 {
			vol = 1
		}
		cs.SetChannelVolume(vol)
	case x == 0xF:
		vol := cs.GetChannelVolume() - (volume.Volume(x) / 64)
		if vol < 0 {
			vol = 0
		}
		cs.SetChannelVolume(vol)
	}
	return nil
}

// Tick is called on every tick
func (e ChannelVolumeSlide) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	mem := cs.GetMemory()
	x, y := mem.ChannelVolumeSlide(uint8(e))

	switch {
	case y == 0x0 && x != 0xF:
		vol := cs.GetChannelVolume() + (volume.Volume(x) / 64)
		if vol > 1 {
			vol = 1
		}
		cs.SetChannelVolume(vol)
	case y != 0xF && x == 0x0:
		vol := cs.GetChannelVolume() - (volume.Volume(x) / 64)
		if vol < 0 {
			vol = 0
		}
		cs.SetChannelVolume(vol)

	case y == 0xF, x == 0xF:
		// nothing
	}
	return nil
}

func (e ChannelVolumeSlide) String() string {
	return fmt.Sprintf("N%0.2x", uint8(e))
}
