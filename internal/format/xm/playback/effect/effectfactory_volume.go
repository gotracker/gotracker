package effect

import (
	"github.com/gotracker/gotracker/internal/format/xm/layout/channel"
	"github.com/gotracker/gotracker/internal/format/xm/playback/util"
)

func volumeEffectFactory(mem *channel.Memory, v util.VolEffect) EffectXM {
	switch {
	case v >= 0x00 && v <= 0x0f: // nothing
		return nil
	case v >= 0x10 && v <= 0x5f: // volume set - handled elsewhere
		// really should be v >= 0x10 && v <= 0x50
		return nil
	case v >= 0x60 && v <= 0x6f: // vol slide down
		return VolumeSlide(v & 0x0f)
	case v >= 0x70 && v <= 0x7f: // vol slide up
		return VolumeSlide((v & 0x0f) << 4)
	case v >= 0x80 && v <= 0x8f: // fine volume slide down
		return FineVolumeSlideDown(v & 0x0f)
	case v >= 0x90 && v <= 0x9f: // fine volume slide up
		return FineVolumeSlideUp(v & 0x0f)
	case v >= 0xA0 && v <= 0xAf: // set vibrato speed
		mem.VibratoSpeed(channel.DataEffect(v) & 0x0f)
		return nil
	case v >= 0xB0 && v <= 0xBf: // vibrato
		vs := mem.VibratoSpeed(0x00)
		return Vibrato(vs<<4 | (channel.DataEffect(v) & 0x0f))
	case v >= 0xC0 && v <= 0xCf: // set panning
		return SetCoarsePanPosition(v & 0x0f)
	case v >= 0xD0 && v <= 0xDf: // panning slide left
		return PanSlide(v & 0x0f)
	case v >= 0xE0 && v <= 0xEf: // panning slide right
		return PanSlide((v & 0x0f) << 4)
	case v >= 0xF0 && v <= 0xFf: // tone portamento
		return PortaToNote(v & 0x0f)
	}
	return UnhandledVolCommand{Vol: v}
}
