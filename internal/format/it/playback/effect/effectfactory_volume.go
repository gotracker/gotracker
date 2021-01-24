package effect

import (
	"gotracker/internal/player/intf"
)

func volPanEffectFactory(mi intf.Memory, v uint8) intf.Effect {
	switch {
	case v >= 0x00 && v <= 40: // volume set - handled elsewhere
		return nil
	case v >= 0x41 && v <= 0x4a: // fine volume slide up
		return FineVolumeSlideUp(v - 0x41)
	case v >= 0x4b && v <= 0x54: // fine volume slide down
		return FineVolumeSlideDown(v - 0x4b)
	case v >= 0x55 && v <= 0x5e: // volume slide up
		return VolumeSlideUp(v - 0x55)
	case v >= 0x5f && v <= 0x68: // volume slide down
		return VolumeSlideDown(v - 0x5f)
	case v >= 0x69 && v <= 0x72: // portamento down
		return PortaDown(v - 0x69)
	case v >= 0x73 && v <= 0x7c: // portamento up
		return PortaUp(v - 0x73)
	case v >= 0x80 && v <= 0xc0: // set panning
		return SetPanPosition(v - 0x80)
	case v >= 0xc1 && v <= 0xca: // portamento to note
		return PortaToNote(v - 0xc1)
	case v >= 0xcb && v <= 0xd4: // vibrato
		return Vibrato(v - 0xcb)
	}
	return UnhandledVolCommand{Vol: v}
}
