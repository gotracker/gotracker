package effect

import (
	"gotracker/internal/player/channel"
	"gotracker/internal/player/intf"
	"log"
)

func EffectFactory(mi intf.SharedMemory, data channel.Data) intf.Effect {
	if !data.What.HasCommand() {
		return nil
	}

	mi.SetEffectSharedMemoryIfNonZero(data.Info)
	switch data.Command + '@' {
	case 'A': // Set Speed
		return EffectSetSpeed(data.Info)
	case 'B': // Pattern Jump
		return EffectOrderJump(data.Info)
	case 'C': // Pattern Break
		return EffectRowJump(data.Info)
	case 'D': // Volume Slide / Fine Volume Slide
		return EffectVolumeSlide(data.Info)
	case 'E': // Porta Down/Fine Porta Down/Xtra Fine Porta
		return EffectPortaDown(data.Info)
	case 'F': // Porta Up/Fine Porta Up/Extra Fine Porta Down
		return EffectPortaUp(data.Info)
	case 'G': // Porta to note
		return EffectPortaToNote(data.Info)
	case 'H': // Vibrato
		return EffectVibrato(data.Info)
	case 'I': // Tremor
		return EffectTremor(data.Info)
	case 'J': // Arpeggio
		return EffectArpeggio(data.Info)
	case 'K': // Vibrato+Volume Slide
		return NewEffectVibratoVolumeSlide(data.Info)
	case 'L': // Porta+Volume Slide
		return NewEffectPortaVolumeSlide(data.Info)
	case 'M': // unused
	case 'N': // unused
	case 'O': // Sample Offset
		return EffectSampleOffset(data.Info)
	case 'P': // unused
	case 'Q': // Retrig + Volume Slide
		return EffectRetrigVolumeSlide(data.Info)
	case 'R': // Tremolo
		return EffectTremolo(data.Info)
	case 'S': // Special
		return determineSpecialActiveEffect(mi, data)
	case 'T': // Set Tempo
		return EffectSetTempo(data.Info)
	case 'U': // Fine Vibrato
		return EffectFineVibrato(data.Info)
	case 'V': // Global Volume
		return EffectSetGlobalVolume(data.Info)
	}
	return nil
}

func determineSpecialActiveEffect(mi intf.SharedMemory, data channel.Data) intf.Effect {
	var cmd = mi.GetEffectSharedMemory(data.Info)
	switch cmd >> 4 {
	case 0x0: // Set Filter on/off
		{
			// TODO
			log.Panicf("%c%0.2x", data.Command+'@', data.Info)
		}
	case 0x1: // Set Glissando on/off
		{
			// TODO
			log.Panicf("%c%0.2x", data.Command+'@', data.Info)
		}
	case 0x2: // Set FineTune
		return EffectSetFinetune(data.Info)
	case 0x3: // Set Vibrato Waveform
		return EffectSetVibratoWaveform(data.Info)
	case 0x4: // Set Tremolo Waveform
		return EffectSetTremoloWaveform(data.Info)
	case 0x5: // unused
	case 0x6: // Fine Pattern Delay
		return EffectFinePatternDelay(data.Info)
	case 0x7: // unused
	case 0x8: // Set Pan Position
		return EffectSetPanPosition(data.Info)
	case 0xA: // Stereo Control
		return EffectStereoControl(data.Info)
	case 0xB: // Pattern Loop
		return EffectPatternLoop(data.Info)
	case 0xC: // Note Cut
		return EffectNoteCut(data.Info)
	case 0xD: // Note Delay
		return EffectNoteDelay(data.Info)
	case 0xE: // Pattern Delay
		return EffectPatternDelay(data.Info)
	case 0xF: // Funk Repeat (invert loop)
		{
			// TODO
			log.Panicf("%c%0.2x", data.Command+'@', data.Info)
		}
	}
	return nil
}
