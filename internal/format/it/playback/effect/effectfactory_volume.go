package effect

import (
	"gotracker/internal/player/intf"
)

func volPanEffectFactory(mi intf.Memory, v uint8) intf.Effect {
	switch {
	case v >= 0x00 && v <= 0x40: // volume set - handled elsewhere
		return nil
	case v >= 0x41 && v <= 0x4a: // fine volume slide up
		return VolChanFineVolumeSlideUp(v - 0x41)
	case v >= 0x4b && v <= 0x54: // fine volume slide down
		return VolChanFineVolumeSlideDown(v - 0x4b)
	case v >= 0x55 && v <= 0x5e: // volume slide up
		return VolChanVolumeSlideUp(v - 0x55)
	case v >= 0x5f && v <= 0x68: // volume slide down
		return VolChanVolumeSlideDown(v - 0x5f)
	case v >= 0x69 && v <= 0x72: // portamento down
		return volPortaDown(v - 0x69)
	case v >= 0x73 && v <= 0x7c: // portamento up
		return volPortaUp(v - 0x73)
	case v >= 0x80 && v <= 0xc0: // set panning
		return SetPanPosition(v - 0x80)
	case v >= 0xc1 && v <= 0xca: // portamento to note
		return volPortaToNote(v - 0xc1)
	case v >= 0xcb && v <= 0xd4: // vibrato
		return Vibrato(v - 0xcb)
	}
	return UnhandledVolCommand{Vol: v}
}

func volPortaDown(v uint8) intf.Effect {
	return PortaDown(v * 4)
}
func volPortaUp(v uint8) intf.Effect {
	return PortaUp(v * 4)
}

func volPortaToNote(v uint8) intf.Effect {
	switch v {
	case 0:
		return PortaToNote(0x00)
	case 1:
		return PortaToNote(0x01)
	case 2:
		return PortaToNote(0x04)
	case 3:
		return PortaToNote(0x08)
	case 4:
		return PortaToNote(0x10)
	case 5:
		return PortaToNote(0x20)
	case 6:
		return PortaToNote(0x40)
	case 7:
		return PortaToNote(0x60)
	case 8:
		return PortaToNote(0x80)
	case 9:
		return PortaToNote(0xFF)
	}
	// impossible, but hey...
	return UnhandledVolCommand{Vol: v + 0xc1}
}
