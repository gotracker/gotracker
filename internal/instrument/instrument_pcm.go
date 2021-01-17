package instrument

import (
	"math"
	"time"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/intf"
)

// PCM is a PCM-data instrument
type PCM struct {
	intf.Instrument

	Sample        []uint8
	Length        int
	LoopMode      LoopMode
	LoopBegin     int
	LoopEnd       int
	NumChannels   int
	Format        SampleDataFormat
	Panning       panning.Position
	VolumeFadeout volume.Volume
	VolEnv        InstEnv
	PanEnv        InstEnv
}

// GetSample returns the sample at position `pos` in the instrument
func (inst *PCM) GetSample(ioc intf.NoteControl, pos sampling.Pos) volume.Matrix {
	ed := ioc.GetData().(*envData)
	dry := inst.getSampleDry(pos, ed.keyOn)
	envVol := inst.getVolEnv(ed, pos)
	chVol := ioc.GetVolume()
	postVol := envVol * chVol
	wet := postVol.Apply(dry...)
	return wet
}

// GetCurrentPanning returns the panning envelope position
func (inst *PCM) GetCurrentPanning(ioc intf.NoteControl) panning.Position {
	x := inst.Panning
	if !inst.PanEnv.Enabled {
		return x
	}

	ed := ioc.GetData().(*envData)
	y := ed.panEnvValue

	// panning envelope value `y` modifies instrument panning value `x`
	// such that `x` is primary component and `y` is secondary
	// TODO: JBC - move this calculation function into gomixing lib

	xa := float64(x.Angle)
	ya := float64(y.Angle)

	const p2 = math.Pi / 2
	const p4 = math.Pi / 4
	const p8 = math.Pi / 8
	fa := xa + (ya-p8)*(p4-math.Abs(xa-p4))/p8
	if fa > p2 {
		fa = p2
	} else if fa < 0 {
		fa = 0
	}

	fd := math.Sqrt(float64(x.Distance * y.Distance))

	finalPan := panning.Position{
		Angle:    float32(fa),
		Distance: float32(fd),
	}

	return finalPan
}

// SetEnvelopePosition sets the envelope position for the note-control
func (inst *PCM) SetEnvelopePosition(ioc intf.NoteControl, ticks int) {
	ed := ioc.GetData().(*envData)
	ed.setEnvelopePosition(ticks, &ed.volEnvPos, &ed.volEnvTicksRemaining, &inst.VolEnv, ed.updateVolEnv)
	if inst.VolEnv.SustainEnabled {
		ed.setEnvelopePosition(ticks, &ed.panEnvPos, &ed.panEnvTicksRemaining, &inst.PanEnv, ed.updatePanEnv)
	}
}

func (inst *PCM) getVolEnv(ed *envData, pos sampling.Pos) volume.Volume {
	if !inst.VolEnv.Enabled {
		return volume.Volume(1.0)
	}

	fadeVol := ed.fadeoutVol
	return fadeVol * ed.volEnvValue
}

func (inst *PCM) getSampleDry(pos sampling.Pos, keyOn bool) volume.Matrix {
	v0 := inst.getConvertedSample(pos.Pos, keyOn)
	if len(v0) == 0 && inst.LoopMode != LoopModeDisabled && keyOn {
		v01 := inst.getConvertedSample(pos.Pos, keyOn)
		panic(v01)
	}
	if pos.Frac == 0 {
		return v0
	}
	v1 := inst.getConvertedSample(pos.Pos+1, keyOn)
	for c, s := range v1 {
		v0[c] += volume.Volume(pos.Frac) * (s - v0[c])
	}
	return v0
}

func (inst *PCM) getConvertedSample(pos int, keyOn bool) volume.Matrix {
	pos = calcLoopedSamplePos(inst.LoopMode, pos, inst.Length, inst.LoopBegin, inst.LoopEnd, keyOn)
	if pos < 0 || pos >= inst.Length {
		return volume.Matrix{}
	}
	return readSample(inst.Format, inst.Sample, pos, inst.NumChannels)
}

// Initialize completes the setup of this instrument
func (inst *PCM) Initialize(ioc intf.NoteControl) error {
	envData := newEnvData()
	ioc.SetData(envData)
	return nil
}

// Attack sets the key on flag for the instrument
func (inst *PCM) Attack(ioc intf.NoteControl) {
	ed := ioc.GetData().(*envData)
	ed.fadeoutVol = volume.Volume(1.0)
	ed.prevKeyOn = ed.keyOn
	ed.keyOn = true
	if inst.VolEnv.Enabled {
		ed.volEnvPos = 0
		ed.volEnvTicksRemaining = 0
		ed.updateEnv(&ed.volEnvPos, &ed.volEnvTicksRemaining, &inst.VolEnv, ed.updateVolEnv)
	}
	if inst.PanEnv.Enabled {
		ed.panEnvPos = 0
		ed.panEnvTicksRemaining = 0
		ed.updateEnv(&ed.panEnvPos, &ed.panEnvTicksRemaining, &inst.PanEnv, ed.updatePanEnv)
	}
}

// Release sets the key on flag for the instrument
func (inst *PCM) Release(ioc intf.NoteControl) {
	ed := ioc.GetData().(*envData)
	ed.prevKeyOn = ed.keyOn
	ed.keyOn = false
}

// GetKeyOn gets the key on flag for the instrument
func (inst *PCM) GetKeyOn(ioc intf.NoteControl) bool {
	ed := ioc.GetData().(*envData)
	return ed.keyOn
}

// Update advances time by the amount specified by `tickDuration`
func (inst *PCM) Update(ioc intf.NoteControl, tickDuration time.Duration) {
	ed := ioc.GetData().(*envData)

	ed.advance(&inst.VolEnv, &inst.PanEnv)

	if !ed.keyOn && inst.VolEnv.Enabled {
		ed.fadeoutVol -= inst.VolumeFadeout
		if ed.fadeoutVol < 0 {
			ed.fadeoutVol = 0
		}
	}
}

// UpdatePosition corrects the position to account for loop mode characteristics and other state parameters
func (inst *PCM) UpdatePosition(ioc intf.NoteControl, pos *sampling.Pos) {
	ed := ioc.GetData().(*envData)
	if ed.prevKeyOn != ed.keyOn && ed.prevKeyOn {
		pos.Pos = calcLoopedSamplePos(inst.LoopMode, pos.Pos, inst.Length, inst.LoopBegin, inst.LoopEnd, ed.prevKeyOn)
	}
}
