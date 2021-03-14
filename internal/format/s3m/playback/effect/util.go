package effect

import (
	s3mfile "github.com/gotracker/goaudiofile/music/tracked/s3m"
	"github.com/gotracker/voice/oscillator"

	"gotracker/internal/comparison"
	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/format/s3m/playback/util"
	"gotracker/internal/player/intf"
	"gotracker/internal/song/note"
)

func doVolSlide(cs intf.Channel, delta float32, multiplier float32) error {
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
	return nil
}

func doPortaUp(cs intf.Channel, amount float32, multiplier float32) error {
	period := cs.GetPeriod()
	if period == nil {
		return nil
	}

	delta := int(amount * multiplier)
	d := note.PeriodDelta(-delta)
	period = period.AddDelta(d).(note.Period)
	cs.SetPeriod(period)
	return nil
}

func doPortaUpToNote(cs intf.Channel, amount float32, multiplier float32, target note.Period) error {
	period := cs.GetPeriod()
	if period == nil {
		return nil
	}

	delta := int(amount * multiplier)
	d := note.PeriodDelta(-delta)
	period = period.AddDelta(d).(note.Period)
	if note.ComparePeriods(period, target) == comparison.SpaceshipLeftGreater {
		period = target
	}
	cs.SetPeriod(period)
	return nil
}

func doPortaDown(cs intf.Channel, amount float32, multiplier float32) error {
	period := cs.GetPeriod()
	if period == nil {
		return nil
	}

	delta := int(amount * multiplier)
	d := note.PeriodDelta(delta)
	period = period.AddDelta(d).(note.Period)
	cs.SetPeriod(period)
	return nil
}

func doPortaDownToNote(cs intf.Channel, amount float32, multiplier float32, target note.Period) error {
	period := cs.GetPeriod()
	if period == nil {
		return nil
	}

	delta := int(amount * multiplier)
	d := note.PeriodDelta(delta)
	period = period.AddDelta(d).(note.Period)
	if note.ComparePeriods(period, target) == comparison.SpaceshipRightGreater {
		period = target
	}
	cs.SetPeriod(period)
	return nil
}

func doVibrato(cs intf.Channel, currentTick int, speed uint8, depth uint8, multiplier float32) error {
	mem := cs.GetMemory().(*channel.Memory)
	delta := calculateWaveTable(cs, currentTick, speed, depth, multiplier, mem.VibratoOscillator())
	cs.SetPeriodDelta(note.PeriodDelta(delta))
	return nil
}

func doTremor(cs intf.Channel, currentTick int, onTicks int, offTicks int) error {
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
	return nil
}

func doArpeggio(cs intf.Channel, currentTick int, arpSemitoneADelta int8, arpSemitoneBDelta int8) error {
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
	return nil
}

var (
	volSlideTwoThirdsTable = [...]s3mfile.Volume{
		0, 0, 1, 1, 2, 3, 3, 4, 5, 5, 6, 6, 7, 8, 8, 9,
		10, 10, 11, 11, 12, 13, 13, 14, 15, 15, 16, 16, 17, 18, 18, 19,
		20, 20, 21, 21, 22, 23, 23, 24, 25, 25, 26, 26, 27, 28, 28, 29,
		30, 30, 31, 31, 32, 33, 33, 34, 35, 35, 36, 36, 37, 38, 38, 39,
	}
)

func doVolSlideTwoThirds(cs intf.Channel) error {
	vol := util.VolumeToS3M(cs.GetActiveVolume())
	if vol >= 64 {
		vol = 63
	}
	cs.SetActiveVolume(util.VolumeFromS3M(volSlideTwoThirdsTable[vol]))
	return nil
}

func doTremolo(cs intf.Channel, currentTick int, speed uint8, depth uint8, multiplier float32) error {
	mem := cs.GetMemory().(*channel.Memory)
	delta := calculateWaveTable(cs, currentTick, speed, depth, multiplier, mem.TremoloOscillator())
	return doVolSlide(cs, delta, 1.0)
}

func calculateWaveTable(cs intf.Channel, currentTick int, speed uint8, depth uint8, multiplier float32, o oscillator.Oscillator) float32 {
	delta := o.GetWave(float32(depth)) * multiplier
	o.Advance(int(speed))
	return delta
}
