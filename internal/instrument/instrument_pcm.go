package instrument

import (
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
	VolumeFadeout volume.Volume
	VolEnv        InstEnv
	PanEnv        InstEnv
}

// GetSample returns the sample at position `pos` in the instrument
func (inst *PCM) GetSample(ioc intf.NoteControl, pos sampling.Pos) volume.Matrix {
	ed := ioc.GetData().(*envData)
	dry := inst.getSampleDry(pos, ed.keyOn)
	envVol := inst.getVolEnv(ed, pos)
	fadeVol := ed.fadeoutVol
	chVol := ioc.GetVolume()
	postVol := fadeVol * envVol * chVol
	wet := postVol.Apply(dry...)
	return wet
}

// GetCurrentPanning returns the panning envelope position
func (inst *PCM) GetCurrentPanning(ioc intf.NoteControl) panning.Position {
	if !inst.PanEnv.Enabled {
		return panning.CenterAhead
	}

	ed := ioc.GetData().(*envData)
	return ed.panEnvValue
}

func (inst *PCM) getVolEnv(ed *envData, pos sampling.Pos) volume.Volume {
	if !inst.VolEnv.Enabled {
		return volume.Volume(1.0)
	}

	return ed.volEnvValue
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
	ed.keyOn = false
}

// NoteCut cuts the current playback of the instrument
func (inst *PCM) NoteCut(ioc intf.NoteControl) {
	ed := ioc.GetData().(*envData)
	ed.keyOn = false
	ed.fadeoutVol = volume.Volume(0.0)
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

	if !ed.keyOn && inst.VolumeFadeout != 0 {
		ed.fadeoutVol -= inst.VolumeFadeout
		if ed.fadeoutVol < 0 {
			ed.fadeoutVol = 0
		}
	}
}
