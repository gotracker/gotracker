package effect

import (
	"log"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// Factory produces an effect for the provided channel pattern data
func Factory(mi intf.Memory, data intf.ChannelData) intf.Effect {
	cd, ok := data.(*channel.Data)
	if !ok {
		return nil
	}

	if !cd.What.HasCommand() {
		return nil
	}

	mem := mi.(*channel.Memory)
	mem.LastNonZero(cd.Info)
	switch cd.Command + '@' {
	case 'A': // Set Speed
		return SetSpeed(cd.Info)
	case 'B': // Pattern Jump
		return OrderJump(cd.Info)
	case 'C': // Pattern Break
		return RowJump(cd.Info)
	case 'D': // Volume Slide / Fine Volume Slide
		return VolumeSlide(cd.Info)
	case 'E': // Porta Down/Fine Porta Down/Xtra Fine Porta
		xx := mem.LastNonZero(uint8(cd.Info))
		x := xx >> 4
		if x == 0x0F {
			return FinePortaDown(cd.Info)
		} else if x == 0x0E {
			return ExtraFinePortaDown(cd.Info)
		}
		return PortaDown(cd.Info)
	case 'F': // Porta Up/Fine Porta Up/Extra Fine Porta Down
		xx := mem.LastNonZero(uint8(cd.Info))
		x := xx >> 4
		if x == 0x0F {
			return FinePortaUp(cd.Info)
		} else if x == 0x0E {
			return ExtraFinePortaUp(cd.Info)
		}
		return PortaUp(cd.Info)
	case 'G': // Porta to note
		return PortaToNote(cd.Info)
	case 'H': // Vibrato
		return Vibrato(cd.Info)
	case 'I': // Tremor
		return Tremor(cd.Info)
	case 'J': // Arpeggio
		return Arpeggio(cd.Info)
	case 'K': // Vibrato+Volume Slide
		return NewVibratoVolumeSlide(cd.Info)
	case 'L': // Porta+Volume Slide
		return NewPortaVolumeSlide(cd.Info)
	case 'M': // unused
	case 'N': // unused
	case 'O': // Sample Offset
		return SampleOffset(cd.Info)
	case 'P': // unused
	case 'Q': // Retrig + Volume Slide
		return RetrigVolumeSlide(cd.Info)
	case 'R': // Tremolo
		return Tremolo(cd.Info)
	case 'S': // Special
		return specialEffect(mem, cd)
	case 'T': // Set Tempo
		return SetTempo(cd.Info)
	case 'U': // Fine Vibrato
		return FineVibrato(cd.Info)
	case 'V': // Global Volume
		return SetGlobalVolume(cd.Info)
	}
	return nil
}

func specialEffect(mem *channel.Memory, cd *channel.Data) intf.Effect {
	var cmd = mem.LastNonZero(cd.Info)
	switch cmd >> 4 {
	case 0x0: // Set Filter on/off
		{
			// TODO
			log.Panicf("%c%0.2x", cd.Command+'@', cd.Info)
		}
	case 0x1: // Set Glissando on/off
		{
			// TODO
			log.Panicf("%c%0.2x", cd.Command+'@', cd.Info)
		}
	case 0x2: // Set FineTune
		return SetFinetune(cd.Info)
	case 0x3: // Set Vibrato Waveform
		return SetVibratoWaveform(cd.Info)
	case 0x4: // Set Tremolo Waveform
		return SetTremoloWaveform(cd.Info)
	case 0x5: // unused
	case 0x6: // Fine Pattern Delay
		return FinePatternDelay(cd.Info)
	case 0x7: // unused
	case 0x8: // Set Pan Position
		return SetPanPosition(cd.Info)
	case 0xA: // Stereo Control
		return StereoControl(cd.Info)
	case 0xB: // Pattern Loop
		return PatternLoop(cd.Info)
	case 0xC: // Note Cut
		return NoteCut(cd.Info)
	case 0xD: // Note Delay
		return NoteDelay(cd.Info)
	case 0xE: // Pattern Delay
		return PatternDelay(cd.Info)
	case 0xF: // Funk Repeat (invert loop)
		{
			// TODO
			log.Panicf("%c%0.2x", cd.Command+'@', cd.Info)
		}
	}
	return nil
}
