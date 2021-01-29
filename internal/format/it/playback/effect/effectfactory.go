package effect

import (
	"fmt"
	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// VolEff is a combined effect that includes a volume effect and a standard effect
type VolEff struct {
	intf.CombinedEffect
	eff intf.Effect
}

func (e VolEff) String() string {
	if e.eff == nil {
		return "..."
	}
	return fmt.Sprintf("%v", e.eff)
}

// Factory produces an effect for the provided channel pattern data
func Factory(mi intf.Memory, data intf.ChannelData) intf.Effect {
	cd, ok := data.(*channel.Data)
	if !ok {
		return nil
	}

	if !cd.What.HasCommand() && !cd.What.HasVolPan() {
		return nil
	}

	eff := VolEff{}
	if cd.What.HasVolPan() {
		ve := volPanEffectFactory(mi, cd.VolPan)
		if ve != nil {
			eff.Effects = append(eff.Effects, ve)
		}
	}

	if e := standardEffectFactory(mi, cd); e != nil {
		eff.Effects = append(eff.Effects, e)
		eff.eff = e
	}

	switch len(eff.Effects) {
	case 0:
		return nil
	case 1:
		return eff.Effects[0]
	default:
		return &eff
	}
}

func standardEffectFactory(mi intf.Memory, cd *channel.Data) intf.Effect {
	mem := mi.(*channel.Memory)
	switch cd.Effect + '@' {
	case '@': // unused
		return nil
	case 'A': // Set Speed
		return SetSpeed(cd.EffectParameter)
	case 'B': // Pattern Jump
		return OrderJump(cd.EffectParameter)
	case 'C': // Pattern Break
		return RowJump(cd.EffectParameter)
	case 'D': // Volume Slide / Fine Volume Slide
		return volumeSlideFactory(mem, cd.Effect, cd.EffectParameter)
	case 'E': // Porta Down/Fine Porta Down/Xtra Fine Porta
		xx := mem.PortaDown(uint8(cd.EffectParameter))
		x := xx >> 4
		if x == 0x0F {
			return FinePortaDown(xx)
		} else if x == 0x0E {
			return ExtraFinePortaDown(xx)
		}
		return PortaDown(cd.EffectParameter)
	case 'F': // Porta Up/Fine Porta Up/Extra Fine Porta Down
		xx := mem.PortaUp(uint8(cd.EffectParameter))
		x := xx >> 4
		if x == 0x0F {
			return FinePortaUp(xx)
		} else if x == 0x0E {
			return ExtraFinePortaUp(xx)
		}
		return PortaUp(cd.EffectParameter)
	case 'G': // Porta to note
		return PortaToNote(cd.EffectParameter)
	case 'H': // Vibrato
		return Vibrato(cd.EffectParameter)
	case 'I': // Tremor
		return Tremor(cd.EffectParameter)
	case 'J': // Arpeggio
		return Arpeggio(cd.EffectParameter)
	case 'K': // Vibrato+Volume Slide
		return NewVibratoVolumeSlide(mem, cd.Effect, cd.EffectParameter)
	case 'L': // Porta+Volume Slide
		return NewPortaVolumeSlide(mem, cd.Effect, cd.EffectParameter)
	case 'M': // Set Channel Volume
		return SetChannelVolume(cd.EffectParameter)
	case 'N': // Channel Volume Slide
		return ChannelVolumeSlide(cd.EffectParameter)
	case 'O': // Sample Offset
		return SampleOffset(cd.EffectParameter)
	case 'P': // Panning Slide
		//return panningSlideFactory(mem, cd.Effect, cd.EffectParameter)
	case 'Q': // Retrig + Volume Slide
		return RetrigVolumeSlide(cd.EffectParameter)
	case 'R': // Tremolo
		return Tremolo(cd.EffectParameter)
	case 'S': // Special
		return specialEffect(cd)
	case 'T': // Set Tempo
		return SetTempo(cd.EffectParameter)
	case 'U': // Fine Vibrato
		return FineVibrato(cd.EffectParameter)
	case 'V': // Global Volume
		return SetGlobalVolume(cd.EffectParameter)
	case 'W': // Global Volume Slide
		return GlobalVolumeSlide(cd.EffectParameter)
	case 'X': // Set Pan Position
		return SetPanPosition(cd.EffectParameter)
	case 'Y': // Panbrello
		//return Panbrello(cd.EffectParameter)
	case 'Z': // MIDI Macro
		return nil // TODO: MIDIMacro
	default:
	}
	return UnhandledCommand{Command: cd.Effect, Info: cd.EffectParameter}
}

func specialEffect(cd *channel.Data) intf.Effect {
	switch cd.EffectParameter >> 4 {
	case 0x0: // unused
		return nil
	//case 0x1: // Set Glissando on/off

	case 0x2: // Set FineTune
		return SetFinetune(cd.EffectParameter)
	case 0x3: // Set Vibrato Waveform
		return SetVibratoWaveform(cd.EffectParameter)
	case 0x4: // Set Tremolo Waveform
		return SetTremoloWaveform(cd.EffectParameter)
	case 0x5: // Set Panbrello Waveform
		return SetPanbrelloWaveform(cd.EffectParameter)
	case 0x6: // Fine Pattern Delay
		return FinePatternDelay(cd.EffectParameter)
	case 0x7: // special note operations
		return specialNoteEffects(cd)
	case 0x8: // Set Coarse Pan Position
		return SetCoarsePanPosition(cd.EffectParameter)
	case 0x9: // Sound Control
		if cd.EffectParameter&0xF == 1 {
			return nil // TODO: SoundControl
		}
	case 0xA: // High Offset
		return HighOffset(cd.EffectParameter)
	case 0xB: // Pattern Loop
		return PatternLoop(cd.EffectParameter)
	case 0xC: // Note Cut
		return NoteCut(cd.EffectParameter)
	case 0xD: // Note Delay
		return NoteDelay(cd.EffectParameter)
	case 0xE: // Pattern Delay
		return PatternDelay(cd.EffectParameter)
	case 0xF: // Set Active Macro
		return nil // TODO: SetActiveMacro
	default:
	}
	return UnhandledCommand{Command: cd.Effect, Info: cd.EffectParameter}
}

func specialNoteEffects(cd *channel.Data) intf.Effect {
	switch cd.EffectParameter & 0xf {
	case 0x0: // Past Note Cut
		return PastNoteCut(cd.EffectParameter)
	case 0x1: // Past Note Off
		return PastNoteOff(cd.EffectParameter)
	case 0x2: // Past Note Fade
		return PastNoteFade(cd.EffectParameter)
	case 0x3: // New Note Action: Note Cut
		return NewNoteActionNoteCut(cd.EffectParameter)
	case 0x4: // New Note Action: Note Continue
		return NewNoteActionNoteContinue(cd.EffectParameter)
	case 0x5: // New Note Action: Note Off
		return NewNoteActionNoteOff(cd.EffectParameter)
	case 0x6: // New Note Action: Note Fade
		return NewNoteActionNoteFade(cd.EffectParameter)
	case 0x7: // Volume Envelope Off
	case 0x8: // Volume Envelope On
	case 0x9: // Panning Envelope Off
	case 0xA: // Panning Envelope On
	case 0xB: // Pitch Envelope Off
	case 0xC: // Pitch Envelope On
	case 0xD, 0xE, 0xF: // unused
		return nil
	}
	return UnhandledCommand{Command: cd.Effect, Info: cd.EffectParameter}
}

func volumeSlideFactory(mem *channel.Memory, cd uint8, ce uint8) intf.Effect {
	x, y := mem.VolumeSlide(uint8(ce))
	switch {
	case x == 0:
		return VolumeSlideDown(ce)
	case y == 0:
		return VolumeSlideUp(ce)
	case x == 0x0f:
		return FineVolumeSlideDown(ce)
	case y == 0x0f:
		return FineVolumeSlideUp(ce)
	}
	// There is a chance that a volume slide command is set with an invalid
	// value or is 00, in which case the memory might have the invalid value,
	// so we need to handle it by deferring to using a no-op instead of a
	// VolumeSlideDown
	return nil
}
