package effect

import (
	"github.com/gotracker/voice/oscillator"

	"gotracker/internal/comparison"
	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

func doVolSlide(cs intf.Channel, delta float32, multiplier float32) {
	av := cs.GetActiveVolume()
	v := util.ToVolumeXM(av)
	vol := int16((float32(v) + delta) * multiplier)
	if vol >= 0x40 {
		vol = 0x40
	}
	if vol < 0x00 {
		vol = 0x00
	}
	v = util.VolumeXM(uint8(vol))
	cs.SetActiveVolume(v.Volume())
}

func doGlobalVolSlide(p intf.Playback, delta float32, multiplier float32) {
	gv := p.GetGlobalVolume()
	v := util.ToVolumeXM(gv)
	vol := int16((float32(v) + delta) * multiplier)
	if vol >= 0x40 {
		vol = 0x40
	}
	if vol < 0x00 {
		vol = 0x00
	}
	v = util.VolumeXM(uint8(vol))
	p.SetGlobalVolume(v.Volume())
}

func doPortaByDeltaAmiga(cs intf.Channel, delta int) {
	period := cs.GetPeriod()
	if period == nil {
		return
	}

	d := note.PeriodDelta(delta)
	period = period.AddDelta(d).(note.Period)
	cs.SetPeriod(period)
}

func doPortaByDeltaLinear(cs intf.Channel, delta int) {
	period := cs.GetPeriod()
	if period == nil {
		return
	}

	finetune := note.PeriodDelta(delta)
	period = period.AddDelta(finetune).(note.Period)
	cs.SetPeriod(period)
}

func doPortaUp(cs intf.Channel, amount float32, multiplier float32, linearFreqSlides bool) {
	delta := int(amount * multiplier)
	if linearFreqSlides {
		doPortaByDeltaLinear(cs, delta)
	} else {
		doPortaByDeltaAmiga(cs, -delta)
	}
}

func doPortaUpToNote(cs intf.Channel, amount float32, multiplier float32, target note.Period, linearFreqSlides bool) {
	doPortaUp(cs, amount, multiplier, linearFreqSlides)
	if period := cs.GetPeriod(); note.ComparePeriods(period, target) == comparison.SpaceshipLeftGreater {
		cs.SetPeriod(target)
	}
}

func doPortaDown(cs intf.Channel, amount float32, multiplier float32, linearFreqSlides bool) {
	delta := int(amount * multiplier)
	if linearFreqSlides {
		doPortaByDeltaLinear(cs, -delta)
	} else {
		doPortaByDeltaAmiga(cs, delta)
	}
}

func doPortaDownToNote(cs intf.Channel, amount float32, multiplier float32, target note.Period, linearFreqSlides bool) {
	doPortaDown(cs, amount, multiplier, linearFreqSlides)
	if period := cs.GetPeriod(); note.ComparePeriods(period, target) == comparison.SpaceshipRightGreater {
		cs.SetPeriod(target)
	}
}

func doVibrato(cs intf.Channel, currentTick int, speed uint8, depth uint8, multiplier float32) {
	mem := cs.GetMemory().(*channel.Memory)
	vib := calculateWaveTable(cs, currentTick, speed, depth, multiplier, mem.VibratoOscillator())
	delta := note.PeriodDelta(vib)
	cs.SetPeriodDelta(delta)
}

func doTremor(cs intf.Channel, currentTick int, onTicks int, offTicks int) {
	mem := cs.GetMemory().(*channel.Memory)
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
}

func doArpeggio(cs intf.Channel, currentTick int, arpSemitoneADelta int8, arpSemitoneBDelta int8) {
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
	cs.SetTargetSemitone(arpSemitoneTarget)
	cs.SetTargetPos(cs.GetPos())
	cs.SetNotePlayTick(true, currentTick)
}

var (
	volSlideTwoThirdsTable = [...]util.VolumeXM{
		0, 0, 1, 1, 2, 3, 3, 4, 5, 5, 6, 6, 7, 8, 8, 9,
		10, 10, 11, 11, 12, 13, 13, 14, 15, 15, 16, 16, 17, 18, 18, 19,
		20, 20, 21, 21, 22, 23, 23, 24, 25, 25, 26, 26, 27, 28, 28, 29,
		30, 30, 31, 31, 32, 33, 33, 34, 35, 35, 36, 36, 37, 38, 38, 39,
	}
)

func doVolSlideTwoThirds(cs intf.Channel) {
	vol := util.ToVolumeXM(cs.GetActiveVolume())
	if vol >= 64 {
		vol = 63
	}

	v := volSlideTwoThirdsTable[vol]
	if v >= 0x40 {
		v = 0x40
	}

	cs.SetActiveVolume(v.Volume())
}

func doTremolo(cs intf.Channel, currentTick int, speed uint8, depth uint8, multiplier float32) {
	mem := cs.GetMemory().(*channel.Memory)
	delta := calculateWaveTable(cs, currentTick, speed, depth, multiplier, mem.TremoloOscillator())
	doVolSlide(cs, delta, 1.0)
}

func calculateWaveTable(cs intf.Channel, currentTick int, speed uint8, depth uint8, multiplier float32, o oscillator.Oscillator) float32 {
	delta := o.GetWave(float32(depth) * multiplier)
	o.Advance(int(speed))
	return delta
}
