package effect

import (
	"fmt"
	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// VolEff is a combined effect that includes a volume effect and a standard effect
type VolEff struct {
	intf.CombinedEffect[channel.Memory, channel.Data]
	eff intf.Effect
}

func (e VolEff) String() string {
	if e.eff == nil {
		return "..."
	}
	return fmt.Sprint(e.eff)
}

// Factory produces an effect for the provided channel pattern data
func Factory(mem *channel.Memory, data *channel.Data) intf.Effect {
	if data == nil {
		return nil
	}

	if !data.What.HasCommand() && !data.What.HasVolPan() {
		return nil
	}

	eff := VolEff{}
	if data.What.HasVolPan() {
		ve := volPanEffectFactory(mem, data.VolPan)
		if ve != nil {
			eff.Effects = append(eff.Effects, ve)
		}
	}

	if e := standardEffectFactory(mem, data); e != nil {
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

func standardEffectFactory(mem *channel.Memory, data *channel.Data) intf.Effect {
	switch data.Effect + '@' {
	case '@': // unused
		return nil
	case 'A': // Set Speed
		return SetSpeed(data.EffectParameter)
	case 'B': // Pattern Jump
		return OrderJump(data.EffectParameter)
	case 'C': // Pattern Break
		return RowJump(data.EffectParameter)
	case 'D': // Volume Slide / Fine Volume Slide
		return volumeSlideFactory(mem, data.Effect, data.EffectParameter)
	case 'E': // Porta Down/Fine Porta Down/Xtra Fine Porta
		xx := mem.PortaDown(channel.DataEffect(data.EffectParameter))
		x := xx >> 4
		if x == 0x0F {
			return FinePortaDown(xx)
		} else if x == 0x0E {
			return ExtraFinePortaDown(xx)
		}
		return PortaDown(data.EffectParameter)
	case 'F': // Porta Up/Fine Porta Up/Extra Fine Porta Down
		xx := mem.PortaUp(channel.DataEffect(data.EffectParameter))
		x := xx >> 4
		if x == 0x0F {
			return FinePortaUp(xx)
		} else if x == 0x0E {
			return ExtraFinePortaUp(xx)
		}
		return PortaUp(data.EffectParameter)
	case 'G': // Porta to note
		return PortaToNote(data.EffectParameter)
	case 'H': // Vibrato
		return Vibrato(data.EffectParameter)
	case 'I': // Tremor
		return Tremor(data.EffectParameter)
	case 'J': // Arpeggio
		return Arpeggio(data.EffectParameter)
	case 'K': // Vibrato+Volume Slide
		return NewVibratoVolumeSlide(mem, data.Effect, data.EffectParameter)
	case 'L': // Porta+Volume Slide
		return NewPortaVolumeSlide(mem, data.Effect, data.EffectParameter)
	case 'M': // Set Channel Volume
		return SetChannelVolume(data.EffectParameter)
	case 'N': // Channel Volume Slide
		return ChannelVolumeSlide(data.EffectParameter)
	case 'O': // Sample Offset
		return SampleOffset(data.EffectParameter)
	case 'P': // Panning Slide
		//return panningSlideFactory(mem, data.Effect, data.EffectParameter)
	case 'Q': // Retrig + Volume Slide
		return RetrigVolumeSlide(data.EffectParameter)
	case 'R': // Tremolo
		return Tremolo(data.EffectParameter)
	case 'S': // Special
		return specialEffect(data)
	case 'T': // Set Tempo
		return SetTempo(data.EffectParameter)
	case 'U': // Fine Vibrato
		return FineVibrato(data.EffectParameter)
	case 'V': // Global Volume
		return SetGlobalVolume(data.EffectParameter)
	case 'W': // Global Volume Slide
		return GlobalVolumeSlide(data.EffectParameter)
	case 'X': // Set Pan Position
		return SetPanPosition(data.EffectParameter)
	case 'Y': // Panbrello
		//return Panbrello(data.EffectParameter)
	case 'Z': // MIDI Macro
		return nil // TODO: MIDIMacro
	default:
	}
	return UnhandledCommand{Command: data.Effect, Info: data.EffectParameter}
}

func specialEffect(data *channel.Data) intf.Effect {
	switch data.EffectParameter >> 4 {
	case 0x0: // unused
		return nil
	//case 0x1: // Set Glissando on/off

	case 0x2: // Set FineTune
		return SetFinetune(data.EffectParameter)
	case 0x3: // Set Vibrato Waveform
		return SetVibratoWaveform(data.EffectParameter)
	case 0x4: // Set Tremolo Waveform
		return SetTremoloWaveform(data.EffectParameter)
	case 0x5: // Set Panbrello Waveform
		return SetPanbrelloWaveform(data.EffectParameter)
	case 0x6: // Fine Pattern Delay
		return FinePatternDelay(data.EffectParameter)
	case 0x7: // special note operations
		return specialNoteEffects(data)
	case 0x8: // Set Coarse Pan Position
		return SetCoarsePanPosition(data.EffectParameter)
	case 0x9: // Sound Control
		return soundControlEffect(data)
	case 0xA: // High Offset
		return HighOffset(data.EffectParameter)
	case 0xB: // Pattern Loop
		return PatternLoop(data.EffectParameter)
	case 0xC: // Note Cut
		return NoteCut(data.EffectParameter)
	case 0xD: // Note Delay
		return NoteDelay(data.EffectParameter)
	case 0xE: // Pattern Delay
		return PatternDelay(data.EffectParameter)
	case 0xF: // Set Active Macro
		return nil // TODO: SetActiveMacro
	default:
	}
	return UnhandledCommand{Command: data.Effect, Info: data.EffectParameter}
}

func specialNoteEffects(data *channel.Data) intf.Effect {
	switch data.EffectParameter & 0xf {
	case 0x0: // Past Note Cut
		return PastNoteCut(data.EffectParameter)
	case 0x1: // Past Note Off
		return PastNoteOff(data.EffectParameter)
	case 0x2: // Past Note Fade
		return PastNoteFade(data.EffectParameter)
	case 0x3: // New Note Action: Note Cut
		return NewNoteActionNoteCut(data.EffectParameter)
	case 0x4: // New Note Action: Note Continue
		return NewNoteActionNoteContinue(data.EffectParameter)
	case 0x5: // New Note Action: Note Off
		return NewNoteActionNoteOff(data.EffectParameter)
	case 0x6: // New Note Action: Note Fade
		return NewNoteActionNoteFade(data.EffectParameter)
	case 0x7: // Volume Envelope Off
		return VolumeEnvelopeOff(data.EffectParameter)
	case 0x8: // Volume Envelope On
		return VolumeEnvelopeOn(data.EffectParameter)
	case 0x9: // Panning Envelope Off
		return PanningEnvelopeOff(data.EffectParameter)
	case 0xA: // Panning Envelope On
		return PanningEnvelopeOn(data.EffectParameter)
	case 0xB: // Pitch Envelope Off
		return PitchEnvelopeOff(data.EffectParameter)
	case 0xC: // Pitch Envelope On
		return PitchEnvelopeOn(data.EffectParameter)
	case 0xD, 0xE, 0xF: // unused
		return nil
	}
	return UnhandledCommand{Command: data.Effect, Info: data.EffectParameter}
}

func volumeSlideFactory(mem *channel.Memory, cd uint8, ce channel.DataEffect) intf.Effect {
	x, y := mem.VolumeSlide(channel.DataEffect(ce))
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

func soundControlEffect(data *channel.Data) intf.Effect {
	switch data.EffectParameter & 0xF {
	case 0x0: // Surround Off
	case 0x1: // Surround On
		// only S91 is supported directly by IT
		return nil // TODO: SurroundOn
	case 0x8: // Reverb Off
	case 0x9: // Reverb On
	case 0xA: // Center Surround
	case 0xB: // Quad Surround
	case 0xC: // Global Filters
	case 0xD: // Local Filters
	case 0xE: // Play Forward
	case 0xF: // Play Backward
	}
	return UnhandledCommand{Command: data.Effect, Info: data.EffectParameter}
}
