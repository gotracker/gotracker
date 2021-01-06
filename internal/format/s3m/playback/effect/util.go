package effect

import (
	"fmt"
	"math"
	"math/rand"

	s3mfile "github.com/gotracker/goaudiofile/music/tracked/s3m"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/format/s3m/playback/util"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/oscillator"
)

// UnhandledCommand is an unhandled command
type UnhandledCommand struct {
	intf.Effect
	Command uint8
	Info    uint8
}

// PreStart triggers when the effect enters onto the channel state
func (e UnhandledCommand) PreStart(cs intf.Channel, p intf.Playback) {
	panic("unhandled command")
}

// Start triggers on the first tick, but before the Tick() function is called
func (e UnhandledCommand) Start(cs intf.Channel, p intf.Playback) {
}

// Tick is called on every tick
func (e UnhandledCommand) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e UnhandledCommand) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e UnhandledCommand) String() string {
	return fmt.Sprintf("%c%0.2x", e.Command+'@', e.Info)
}

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
	d := util.AmigaPeriod(-delta)
	period = period.Add(&d)
	cs.SetPeriod(period)
}

func doPortaUpToNote(cs intf.Channel, amount float32, multiplier float32, target note.Period) {
	period := cs.GetPeriod()
	if period == nil {
		return
	}

	delta := int(amount * multiplier)
	d := util.AmigaPeriod(-delta)
	period = period.Add(&d)
	if period.Compare(target) == 1 {
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
	d := util.AmigaPeriod(delta)
	period = period.Add(&d)
	cs.SetPeriod(period)
}

func doPortaDownToNote(cs intf.Channel, amount float32, multiplier float32, target note.Period) {
	period := cs.GetPeriod()
	if period == nil {
		return
	}

	delta := int(amount * multiplier)
	d := util.AmigaPeriod(delta)
	period = period.Add(&d)
	if period.Compare(target) == -1 {
		period = target
	}
	cs.SetPeriod(period)
}

func doVibrato(cs intf.Channel, currentTick int, speed uint8, depth uint8, multiplier float32) {
	delta := util.AmigaPeriod(calculateWaveTable(cs, currentTick, speed, depth, multiplier, cs.GetVibratoOscillator()))
	cs.SetVibratoDelta(&delta)
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
	cs.SetSemitone(arpSemitoneTarget)
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
	delta := calculateWaveTable(cs, currentTick, speed, depth, multiplier, cs.GetTremoloOscillator())
	doVolSlide(cs, delta, 1.0)
}

func calculateWaveTable(cs intf.Channel, currentTick int, speed uint8, depth uint8, multiplier float32, o *oscillator.Oscillator) float32 {
	var vib float32
	switch o.Table {
	case oscillator.WaveTableSelectSine:
		vib = float32(math.Sin(float64(o.Pos) * math.Pi / 32.0))
	case oscillator.WaveTableSelectSawtooth:
		vib = (32.0 - float32(o.Pos&64)) / 32.0
	case oscillator.WaveTableSelectSquare:
		v := float32(math.Sin(float64(o.Pos) * math.Pi / 32.0))
		if v > 0 {
			vib = 1.0
		} else {
			vib = -1.0
		}
	case oscillator.WaveTableSelectRandom:
		vib = rand.Float32()*2.0 - 1.0
	}
	delta := float32(vib) * float32(depth) * multiplier
	o.Pos += int8(speed)
	if o.Pos > 31 {
		o.Pos -= 64
	}
	return delta
}
