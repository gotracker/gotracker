package effect

import (
	"log"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

type volEff struct {
	intf.CombinedEffect
	eff intf.Effect
}

func (e volEff) String() string {
	if e.eff == nil {
		return "..."
	}
	return e.eff.String()
}

// Factory produces an effect for the provided channel pattern data
func Factory(mi intf.Memory, data intf.ChannelData) intf.Effect {
	cd, ok := data.(*channel.Data)
	if !ok {
		return nil
	}

	if !cd.HasEffect() {
		return nil
	}

	eff := volEff{}
	if cd.What.HasVolume() {
		v := cd.Volume
		var ve intf.Effect
		switch {
		case v >= 0x60 && v <= 0x6f: // vol slide down
			ve = VolumeSlide(v & 0x0f)
		case v >= 0x70 && v <= 0x7f: // vol slide up
			ve = VolumeSlide((v & 0x0f) << 4)
		case v >= 0x80 && v <= 0x8f: // fine volume slide down
			ve = VolumeSlide(0xf0 | (v & 0x0f))
		case v >= 0x90 && v <= 0x9f: // fine volume slide up
			ve = VolumeSlide((v&0x0f)<<4 | 0x0f)
		case v >= 0xA0 && v <= 0xAf: // set vibrato speed
			mi.(*channel.Memory).VibratoSpeed(v & 0x0f)
		case v >= 0xB0 && v <= 0xBf: // vibrato
			vs := mi.(*channel.Memory).VibratoSpeed(0x00)
			ve = Vibrato(vs<<4 | (v & 0x0f))
		case v >= 0xC0 && v <= 0xCf: // set panning
			ve = SetFinePanPosition(v & 0x0f)
		case v >= 0xD0 && v <= 0xDf: // panning slide left
			// TODO
		case v >= 0xE0 && v <= 0xEf: // panning slide right
			// TODO
		case v >= 0xF0 && v <= 0xFf: // tone portamento
			ve = PortaToNote(v & 0x0f)
		}
		if ve != nil {
			eff.Effects = append(eff.Effects, ve)
		}
	}

	if e := standardEffectFactory(mi, cd); e != nil {
		eff.Effects = append(eff.Effects, e)
		eff.eff = e
	}

	if len(eff.Effects) == 0 {
		return nil
	}

	return &eff
}

func standardEffectFactory(mi intf.Memory, cd *channel.Data) intf.Effect {
	if !cd.What.HasEffect() && !cd.What.HasEffectParameter() {
		return nil
	}

	mem := mi.(*channel.Memory)
	mem.LastNonZero(cd.EffectParameter)
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
	case 0x08: // Set panning
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
		switch cd.EffectParameter >> 4 {
		case 0x1: // Fine porta up
			return FinePortaUp(cd.EffectParameter)
		case 0x2: // Fine porta down
			return FinePortaDown(cd.EffectParameter)
		case 0x3: // Set glissando control
			{
				// TODO
				log.Panicf("%0.2x%0.2x", cd.Effect, cd.EffectParameter)
			}
		case 0x4: // Set vibrato control
			return SetVibratoWaveform(cd.EffectParameter)
		case 0x5: // Set finetune
			return SetFinetune(cd.EffectParameter)
		case 0x6: // Set loop begin/loop
			return PatternLoop(cd.EffectParameter)
		case 0x7: // Set tremolo control
			return SetTremoloWaveform(cd.EffectParameter)
		case 0x8: // Set fine panning
			return SetFinePanPosition(cd.EffectParameter)
		case 0x9: // Retrig note
			return RetriggerNote(cd.EffectParameter)
		case 0xA: // Fine volume slide up
			{
				// TODO
				log.Panicf("%0.2x%0.2x", cd.Effect, cd.EffectParameter)
			}
		case 0xB: // Fine volume slide down
			{
				// TODO
				log.Panicf("%0.2x%0.2x", cd.Effect, cd.EffectParameter)
			}
		case 0xC: // Note cut
			return NoteCut(cd.EffectParameter)
		case 0xD: // Note delay
			return NoteDelay(cd.EffectParameter)
		case 0xE: // Pattern delay
			return PatternDelay(cd.EffectParameter)
		}
	case 0x0F: // Set tempo/BPM
		if cd.EffectParameter < 0x20 {
			return SetSpeed(cd.EffectParameter)
		}
		return SetTempo(cd.EffectParameter)
	case 0x10: // Set global volume
		return SetGlobalVolume(cd.EffectParameter)
	case 0x11: // Global volume slide
		{
			// TODO
			log.Panicf("%0.2x%0.2x", cd.Effect, cd.EffectParameter)
		}

	case 0x15: // Set envelope position
		{
			// TODO
			log.Panicf("%0.2x%0.2x", cd.Effect, cd.EffectParameter)
		}

	case 0x19: // Panning slide
		{
			// TODO
			log.Panicf("%0.2x%0.2x", cd.Effect, cd.EffectParameter)
		}

	case 0x1B: // Multi retrig note
		return RetrigVolumeSlide(cd.EffectParameter)

	case 0x1D: // Tremor
		return Tremor(cd.EffectParameter)

	case 0x21: // Extra fine porta up
		switch cd.EffectParameter >> 4 {
		case 0x1:
			return ExtraFinePortaUp(cd.EffectParameter)
		case 0x2: // Extra fine porta down
			return ExtraFinePortaDown(cd.EffectParameter)
		}
	}
	return nil
}
