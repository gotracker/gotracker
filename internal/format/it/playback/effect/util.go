package effect

import (
	itfile "github.com/gotracker/goaudiofile/music/tracked/it"
	"github.com/gotracker/voice/oscillator"

	"github.com/gotracker/gotracker/internal/comparison"
	"github.com/gotracker/gotracker/internal/format/it/layout/channel"
	effectIntf "github.com/gotracker/gotracker/internal/format/it/playback/effect/intf"
	itVolume "github.com/gotracker/gotracker/internal/format/it/volume"
	"github.com/gotracker/gotracker/internal/player/intf"
	"github.com/gotracker/gotracker/internal/song/note"
)

func doVolSlide(cs intf.Channel[channel.Memory, channel.Data], delta float32, multiplier float32) error {
	av := cs.GetActiveVolume()
	v := itVolume.ToItVolume(av)
	vol := int16((float32(v) + delta) * multiplier)
	if vol >= 0x40 {
		vol = 0x40
	}
	if vol < 0x00 {
		vol = 0x00
	}
	v = itfile.Volume(vol)
	nv := itVolume.FromItVolume(v)
	cs.SetActiveVolume(nv)
	return nil
}

func doGlobalVolSlide(m effectIntf.IT, delta float32, multiplier float32) error {
	gv := m.GetGlobalVolume()
	v := itVolume.ToItVolume(gv)
	vol := int16((float32(v) + delta) * multiplier)
	if vol >= 0x40 {
		vol = 0x40
	}
	if vol < 0x00 {
		vol = 0x00
	}
	v = itfile.Volume(vol)
	ngv := itVolume.FromItVolume(v)
	m.SetGlobalVolume(ngv)
	return nil
}

func doPortaByDeltaAmiga(cs intf.Channel[channel.Memory, channel.Data], delta int) error {
	period := cs.GetPeriod()
	if period == nil {
		return nil
	}

	d := note.PeriodDelta(delta)
	period = period.AddDelta(d).(note.Period)
	cs.SetPeriod(period)
	return nil
}

func doPortaByDeltaLinear(cs intf.Channel[channel.Memory, channel.Data], delta int) error {
	period := cs.GetPeriod()
	if period == nil {
		return nil
	}

	finetune := note.PeriodDelta(delta)
	period = period.AddDelta(finetune).(note.Period)
	cs.SetPeriod(period)
	return nil
}

func doPortaUp(cs intf.Channel[channel.Memory, channel.Data], amount float32, multiplier float32, linearFreqSlides bool) error {
	delta := int(amount * multiplier)
	if linearFreqSlides {
		return doPortaByDeltaLinear(cs, delta)
	}
	return doPortaByDeltaAmiga(cs, -delta)
}

func doPortaUpToNote(cs intf.Channel[channel.Memory, channel.Data], amount float32, multiplier float32, target note.Period, linearFreqSlides bool) error {
	if err := doPortaUp(cs, amount, multiplier, linearFreqSlides); err != nil {
		return err
	}
	if period := cs.GetPeriod(); note.ComparePeriods(period, target) == comparison.SpaceshipLeftGreater {
		cs.SetPeriod(target)
	}
	return nil
}

func doPortaDown(cs intf.Channel[channel.Memory, channel.Data], amount float32, multiplier float32, linearFreqSlides bool) error {
	delta := int(amount * multiplier)
	if linearFreqSlides {
		return doPortaByDeltaLinear(cs, -delta)
	}
	return doPortaByDeltaAmiga(cs, delta)
}

func doPortaDownToNote(cs intf.Channel[channel.Memory, channel.Data], amount float32, multiplier float32, target note.Period, linearFreqSlides bool) error {
	if err := doPortaDown(cs, amount, multiplier, linearFreqSlides); err != nil {
		return err
	}
	if period := cs.GetPeriod(); note.ComparePeriods(period, target) == comparison.SpaceshipRightGreater {
		cs.SetPeriod(target)
	}
	return nil
}

func doVibrato(cs intf.Channel[channel.Memory, channel.Data], currentTick int, speed channel.DataEffect, depth channel.DataEffect, multiplier float32) error {
	mem := cs.GetMemory()
	vib := calculateWaveTable(cs, currentTick, speed, depth, multiplier, mem.VibratoOscillator())
	delta := note.PeriodDelta(vib)
	cs.SetPeriodDelta(delta)
	return nil
}

func doTremor(cs intf.Channel[channel.Memory, channel.Data], currentTick int, onTicks int, offTicks int) error {
	mem := cs.GetMemory()
	tremor := mem.TremorMem()
	if tremor.IsActive() {
		if tremor.Advance() >= onTicks {
			tremor.ToggleAndReset()
		}
	} else {
		if tremor.Advance() >= offTicks {
			tremor.ToggleAndReset()
		}
	}
	cs.SetVolumeActive(tremor.IsActive())
	return nil
}

func doArpeggio(cs intf.Channel[channel.Memory, channel.Data], currentTick int, arpSemitoneADelta int8, arpSemitoneBDelta int8) error {
	ns := cs.GetNoteSemitone()
	var arpSemitoneTarget note.Semitone
	switch currentTick % 3 {
	case 0:
		arpSemitoneTarget = ns
	case 1:
		arpSemitoneTarget = note.Semitone(int8(ns) + arpSemitoneADelta)
	case 2:
		arpSemitoneTarget = note.Semitone(int8(ns) + arpSemitoneBDelta)
	}
	cs.SetOverrideSemitone(arpSemitoneTarget)
	cs.SetTargetPos(cs.GetPos())
	return nil
}

var (
	volSlideTwoThirdsTable = [...]channel.DataEffect{
		0, 0, 1, 1, 2, 3, 3, 4, 5, 5, 6, 6, 7, 8, 8, 9,
		10, 10, 11, 11, 12, 13, 13, 14, 15, 15, 16, 16, 17, 18, 18, 19,
		20, 20, 21, 21, 22, 23, 23, 24, 25, 25, 26, 26, 27, 28, 28, 29,
		30, 30, 31, 31, 32, 33, 33, 34, 35, 35, 36, 36, 37, 38, 38, 39,
	}
)

func doVolSlideTwoThirds(cs intf.Channel[channel.Memory, channel.Data]) error {
	vol := itVolume.ToItVolume(cs.GetActiveVolume())
	if vol >= 0x10 && vol <= 0x50 {
		vol -= 0x10
		if vol >= 64 {
			vol = 63
		}

		v := volSlideTwoThirdsTable[vol]
		if v >= 0x40 {
			v = 0x40
		}

		vv := itfile.Volume(v)
		cs.SetActiveVolume(itVolume.FromItVolume(vv))
	}
	return nil
}

func doTremolo(cs intf.Channel[channel.Memory, channel.Data], currentTick int, speed channel.DataEffect, depth channel.DataEffect, multiplier float32) error {
	mem := cs.GetMemory()
	delta := calculateWaveTable(cs, currentTick, speed, depth, multiplier, mem.TremoloOscillator())
	return doVolSlide(cs, delta, 1.0)
}

func calculateWaveTable(cs intf.Channel[channel.Memory, channel.Data], currentTick int, speed channel.DataEffect, depth channel.DataEffect, multiplier float32, o oscillator.Oscillator) float32 {
	delta := o.GetWave(float32(depth) * multiplier)
	o.Advance(int(speed))
	return delta
}
