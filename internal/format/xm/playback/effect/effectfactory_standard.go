package effect

import (
	"github.com/gotracker/gotracker/internal/format/xm/layout/channel"
)

func standardEffectFactory(mem *channel.Memory, cd *channel.Data) EffectXM {
	if !cd.What.HasEffect() && !cd.What.HasEffectParameter() {
		return nil
	}

	switch cd.Effect {
	case 0x00: // Arpeggio
		return Arpeggio(cd.EffectParameter)
	case 0x01: // Porta up
		return PortaUp(cd.EffectParameter)
	case 0x02: // Porta down
		return PortaDown(cd.EffectParameter)
	case 0x03: // Tone porta
		return PortaToNote(cd.EffectParameter)
	case 0x04: // Vibrato
		return Vibrato(cd.EffectParameter)
	case 0x05: // Tone porta + Volume slide
		return NewPortaVolumeSlide(cd.EffectParameter)
	case 0x06: // Vibrato + Volume slide
		return NewVibratoVolumeSlide(cd.EffectParameter)
	case 0x07: // Tremolo
		return Tremolo(cd.EffectParameter)
	case 0x08: // Set (fine) panning
		return SetPanPosition(cd.EffectParameter)
	case 0x09: // Sample offset
		return SampleOffset(cd.EffectParameter)
	case 0x0A: // Volume slide
		return VolumeSlide(cd.EffectParameter)
	case 0x0B: // Position jump
		return OrderJump(cd.EffectParameter)
	case 0x0C: // Set volume
		return SetVolume(cd.EffectParameter)
	case 0x0D: // Pattern break
		return RowJump(cd.EffectParameter)
	case 0x0E: // extra...
		return specialEffectFactory(mem, cd.Effect, cd.EffectParameter)
	case 0x0F: // Set tempo/BPM
		if cd.EffectParameter < 0x20 {
			return SetSpeed(cd.EffectParameter)
		}
		return SetTempo(cd.EffectParameter)
	case 0x10: // Set global volume
		return SetGlobalVolume(cd.EffectParameter)
	case 0x11: // Global volume slide
		return GlobalVolumeSlide(cd.EffectParameter)

	case 0x15: // Set envelope position
		return SetEnvelopePosition(cd.EffectParameter)

	case 0x19: // Panning slide
		return PanSlide(cd.EffectParameter)

	case 0x1B: // Multi retrig note
		return RetrigVolumeSlide(cd.EffectParameter)

	case 0x1D: // Tremor
		return Tremor(cd.EffectParameter)

	case 0x21: // Extra fine porta commands
		return extraFinePortaEffectFactory(mem, cd.Effect, cd.EffectParameter)
	}
	return UnhandledCommand{Command: cd.Effect, Info: cd.EffectParameter}
}

func extraFinePortaEffectFactory(mem *channel.Memory, ce uint8, cp channel.DataEffect) EffectXM {
	switch cp >> 4 {
	case 0x0: // none
		return nil
	case 0x1: // Extra fine porta up
		return ExtraFinePortaUp(cp)
	case 0x2: // Extra fine porta down
		return ExtraFinePortaDown(cp)
	}
	return UnhandledCommand{Command: ce, Info: cp}
}

func specialEffectFactory(mem *channel.Memory, ce uint8, cp channel.DataEffect) EffectXM {
	switch cp >> 4 {
	case 0x1: // Fine porta up
		return FinePortaUp(cp)
	case 0x2: // Fine porta down
		return FinePortaDown(cp)
	//case 0x3: // Set glissando control

	case 0x4: // Set vibrato control
		return SetVibratoWaveform(cp)
	case 0x5: // Set finetune
		return SetFinetune(cp)
	case 0x6: // Set loop begin/loop
		return PatternLoop(cp)
	case 0x7: // Set tremolo control
		return SetTremoloWaveform(cp)
	case 0x8: // Set coarse panning
		return SetCoarsePanPosition(cp)
	case 0x9: // Retrig note
		return RetriggerNote(cp)
	case 0xA: // Fine volume slide up
		return FineVolumeSlideUp(cp)
	case 0xB: // Fine volume slide down
		return FineVolumeSlideDown(cp)
	case 0xC: // Note cut
		return NoteCut(cp)
	case 0xD: // Note delay
		return NoteDelay(cp)
	case 0xE: // Pattern delay
		return PatternDelay(cp)
	}
	return UnhandledCommand{Command: ce, Info: cp}
}
