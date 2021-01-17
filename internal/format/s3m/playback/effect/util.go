package effect

import (
	s3mfile "github.com/gotracker/goaudiofile/music/tracked/s3m"

	formatutil "gotracker/internal/format/internal/util"
	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/format/s3m/playback/util"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

func doVolSlide(cs intf.Channel, delta float32, multiplier float32) {
	av := cs.GetActiveVolume()
	v := util.VolumeToS3M(av)
	vol := int16((float32(v) + delta) * multiplier)
	if vol >= 64 {
		vol = 63
	}
	if vol < 0 {
		vol = 0
	}
	sv := s3mfile.Volume(uint8(vol))
	nv := util.VolumeFromS3M(sv)
	cs.SetActiveVolume(nv)
}

func doPortaUp(cs intf.Channel, amount float32, multiplier float32) {
	period := cs.GetPeriod()
	if period == nil {
		return
	}

	delta := int(amount * multiplier)
	d := note.PeriodDelta(-delta)
	period = period.Add(d)
	cs.SetPeriod(period)
}

func doPortaUpToNote(cs intf.Channel, amount float32, multiplier float32, target note.Period) {
	period := cs.GetPeriod()
	if period == nil {
		return
	}

	delta := int(amount * multiplier)
	d := note.PeriodDelta(-delta)
	period = period.Add(d)
	if note.ComparePeriods(period, target) == note.CompareLeftHigher {
		period = target
	}
	cs.SetPeriod(period)
}

func doPortaDown(cs intf.Channel, amount float32, multiplier float32) {
	period := cs.GetPeriod()
	if period == nil {
		return
	}

	delta := int(amount * multiplier)
	d := note.PeriodDelta(delta)
	period = period.Add(d)
	cs.SetPeriod(period)
}

func doPortaDownToNote(cs intf.Channel, amount float32, multiplier float32, target note.Period) {
	period := cs.GetPeriod()
	if period == nil {
		return
	}

	delta := int(amount * multiplier)
	d := note.PeriodDelta(delta)
	period = period.Add(d)
	if note.ComparePeriods(period, target) == note.CompareRightHigher {
		period = target
	}
	cs.SetPeriod(period)
}

func doVibrato(cs intf.Channel, currentTick int, speed uint8, depth uint8, multiplier float32) {
	mem := cs.GetMemory().(*channel.Memory)
	delta := calculateWaveTable(cs, currentTick, speed, depth, multiplier, mem.VibratoOscillator())
	cs.SetPeriodDelta(note.PeriodDelta(delta))
}

func doTremor(cs intf.Channel, currentTick int, onTicks int, offTicks int) {
	mem := cs.GetMemory().(*channel.Memory)
	tremor := mem.TremorMem()
	if tremor.IsActive() {
		if tremor.Advance() > onTicks {
			tremor.ToggleAndReset()
		}
	} else {
		if tremor.Advance() > offTicks {
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
	cs.SetNotePlayTick(currentTick)
	cs.SetDoRetriggerNote(true)
}

var (
	volSlideTwoThirdsTable = [...]s3mfile.Volume{
		0, 0, 1, 1, 2, 3, 3, 4, 5, 5, 6, 6, 7, 8, 8, 9,
		10, 10, 11, 11, 12, 13, 13, 14, 15, 15, 16, 16, 17, 18, 18, 19,
		20, 20, 21, 21, 22, 23, 23, 24, 25, 25, 26, 26, 27, 28, 28, 29,
		30, 30, 31, 31, 32, 33, 33, 34, 35, 35, 36, 36, 37, 38, 38, 39,
	}
)

func doVolSlideTwoThirds(cs intf.Channel) {
	vol := util.VolumeToS3M(cs.GetActiveVolume())
	if vol >= 64 {
		vol = 63
	}
	cs.SetActiveVolume(util.VolumeFromS3M(volSlideTwoThirdsTable[vol]))
}

func doTremolo(cs intf.Channel, currentTick int, speed uint8, depth uint8, multiplier float32) {
	mem := cs.GetMemory().(*channel.Memory)
	delta := calculateWaveTable(cs, currentTick, speed, depth, multiplier, mem.TremoloOscillator())
	doVolSlide(cs, delta, 1.0)
}

func calculateWaveTable(cs intf.Channel, currentTick int, speed uint8, depth uint8, multiplier float32, o *formatutil.Oscillator) float32 {
	delta := o.GetWave(float32(depth)) * multiplier
	o.Advance(int(speed))
	return delta
}
