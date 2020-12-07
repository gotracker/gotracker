package effect

import (
	"gotracker/internal/player/intf"
	"gotracker/internal/player/oscillator"
	"gotracker/internal/s3m/util"
	"math"
	"math/rand"
)

func doVolSlide(cs intf.Channel, delta float32, multiplier float32) {
	v := cs.GetActiveVolume()
	vol := int16((float32(v) + delta) * multiplier)
	if vol >= 64 {
		vol = 63
	}
	if vol < 0 {
		vol = 0
	}
	cs.SetActiveVolume(uint8(vol))
}

func doPortaUp(cs intf.Channel, amount float32, multiplier float32) {
	delta := amount * multiplier
	cs.SetPeriod(cs.GetPeriod() - delta)
}

func doPortaUpToNote(cs intf.Channel, amount float32, multiplier float32, target float32) {
	delta := amount * multiplier
	newPeriod := cs.GetPeriod() - delta
	if newPeriod < target {
		newPeriod = target
	}
	cs.SetPeriod(newPeriod)
	cs.SetTargetPeriod(newPeriod)
}

func doPortaDown(cs intf.Channel, amount float32, multiplier float32) {
	delta := amount * multiplier
	cs.SetPeriod(cs.GetPeriod() + delta)
}

func doPortaDownToNote(cs intf.Channel, amount float32, multiplier float32, target float32) {
	delta := amount * multiplier
	newPeriod := cs.GetPeriod() + delta
	if newPeriod > target {
		newPeriod = target
	}
	cs.SetPeriod(newPeriod)
	cs.SetTargetPeriod(newPeriod)
}

func doVibrato(cs intf.Channel, currentTick int, speed uint8, depth uint8, multiplier float32) {
	delta := calculateWaveTable(cs, currentTick, speed, depth, multiplier, cs.GetVibratoOscillator())
	cs.SetVibratoDelta(delta)
}

func doTremor(cs intf.Channel, currentTick int, onTicks int, offTicks int) {
	if cs.GetTremorOn() {
		if cs.GetTremorTime() >= onTicks {
			cs.SetTremorOn(false)
			cs.SetTremorTime(0)
		}
	} else {
		if cs.GetTremorTime() >= offTicks {
			cs.SetTremorOn(true)
			cs.SetTremorTime(0)
		}
	}
	cs.SetTremorTime(cs.GetTremorTime() + 1)
}

func doArpeggio(cs intf.Channel, currentTick int, arpSemitoneADelta uint8, arpSemitoneBDelta uint8) {
	inst := cs.GetInstrument()
	if inst.IsInvalid() {
		return
	}
	ns := cs.GetNoteSemitone()
	var arpSemitoneTarget uint8
	switch currentTick % 3 {
	case 0:
		arpSemitoneTarget = ns
	case 1:
		arpSemitoneTarget = ns + arpSemitoneADelta
	case 2:
		arpSemitoneTarget = ns + arpSemitoneBDelta
	}
	newSemi := util.CalcSemitonePeriod(arpSemitoneTarget, inst.C2Spd())
	cs.SetTargetPeriod(newSemi)
	cs.SetTargetInst(inst)
	cs.SetTargetPos(cs.GetPos())
	cs.SetNotePlayTick(currentTick)
}

var (
	volSlideTwoThirdsTable = [...]uint8{
		0, 0, 1, 1, 2, 3, 3, 4, 5, 5, 6, 6, 7, 8, 8, 9,
		10, 10, 11, 11, 12, 13, 13, 14, 15, 15, 16, 16, 17, 18, 18, 19,
		20, 20, 21, 21, 22, 23, 23, 24, 25, 25, 26, 26, 27, 28, 28, 29,
		30, 30, 31, 31, 32, 33, 33, 34, 35, 35, 36, 36, 37, 38, 38, 39,
	}
)

func doVolSlideTwoThirds(cs intf.Channel) {
	vol := cs.GetActiveVolume()
	if vol >= 64 {
		vol = 63
	}
	cs.SetActiveVolume(volSlideTwoThirdsTable[vol])
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
