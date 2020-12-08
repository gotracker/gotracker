package effect

import (
	"gotracker/internal/player/intf"
	"gotracker/internal/s3m/channel"
	"log"
)

func EffectFactory(mi intf.SharedMemory, data intf.ChannelData) intf.Effect {
	cd, ok := data.(*channel.Data)
	if !ok {
		return nil
	}

	if !cd.What.HasCommand() {
		return nil
	}

	mi.SetEffectSharedMemoryIfNonZero(cd.Info)
	switch cd.Command + '@' {
	case 'A': // Set Speed
		return EffectSetSpeed(cd.Info)
	case 'B': // Pattern Jump
		return EffectOrderJump(cd.Info)
	case 'C': // Pattern Break
		return EffectRowJump(cd.Info)
	case 'D': // Volume Slide / Fine Volume Slide
		return EffectVolumeSlide(cd.Info)
	case 'E': // Porta Down/Fine Porta Down/Xtra Fine Porta
		return EffectPortaDown(cd.Info)
	case 'F': // Porta Up/Fine Porta Up/Extra Fine Porta Down
		return EffectPortaUp(cd.Info)
	case 'G': // Porta to note
		return EffectPortaToNote(cd.Info)
	case 'H': // Vibrato
		return EffectVibrato(cd.Info)
	case 'I': // Tremor
		return EffectTremor(cd.Info)
	case 'J': // Arpeggio
		return EffectArpeggio(cd.Info)
	case 'K': // Vibrato+Volume Slide
		return NewEffectVibratoVolumeSlide(cd.Info)
	case 'L': // Porta+Volume Slide
		return NewEffectPortaVolumeSlide(cd.Info)
	case 'M': // unused
	case 'N': // unused
	case 'O': // Sample Offset
		return EffectSampleOffset(cd.Info)
	case 'P': // unused
	case 'Q': // Retrig + Volume Slide
		return EffectRetrigVolumeSlide(cd.Info)
	case 'R': // Tremolo
		return EffectTremolo(cd.Info)
	case 'S': // Special
		return determineSpecialActiveEffect(mi, cd)
	case 'T': // Set Tempo
		return EffectSetTempo(cd.Info)
	case 'U': // Fine Vibrato
		return EffectFineVibrato(cd.Info)
	case 'V': // Global Volume
		return EffectSetGlobalVolume(cd.Info)
	}
	return nil
}

func determineSpecialActiveEffect(mi intf.SharedMemory, cd *channel.Data) intf.Effect {
	var cmd = mi.GetEffectSharedMemory(cd.Info)
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
		return EffectSetFinetune(cd.Info)
	case 0x3: // Set Vibrato Waveform
		return EffectSetVibratoWaveform(cd.Info)
	case 0x4: // Set Tremolo Waveform
		return EffectSetTremoloWaveform(cd.Info)
	case 0x5: // unused
	case 0x6: // Fine Pattern Delay
		return EffectFinePatternDelay(cd.Info)
	case 0x7: // unused
	case 0x8: // Set Pan Position
		return EffectSetPanPosition(cd.Info)
	case 0xA: // Stereo Control
		return EffectStereoControl(cd.Info)
	case 0xB: // Pattern Loop
		return EffectPatternLoop(cd.Info)
	case 0xC: // Note Cut
		return EffectNoteCut(cd.Info)
	case 0xD: // Note Delay
		return EffectNoteDelay(cd.Info)
	case 0xE: // Pattern Delay
		return EffectPatternDelay(cd.Info)
	case 0xF: // Funk Repeat (invert loop)
		{
			// TODO
			log.Panicf("%c%0.2x", cd.Command+'@', cd.Info)
		}
	}
	return nil
}
