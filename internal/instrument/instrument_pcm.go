package instrument

import (
	"math"
	"time"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// FadeoutMode is the mode used to process fade-out
type FadeoutMode uint8

const (
	// FadeoutModeDisabled is for when the fade-out is disabled (S3M/MOD)
	FadeoutModeDisabled = FadeoutMode(iota)
	// FadeoutModeAlwaysActive is for when the fade-out is always available to be used (IT-style)
	FadeoutModeAlwaysActive
	// FadeoutModeOnlyIfVolEnvActive is for when the fade-out only functions when VolEnv is enabled (XM-style)
	FadeoutModeOnlyIfVolEnvActive
)

// FadeoutSettings is the settings for fade-out
type FadeoutSettings struct {
	Mode   FadeoutMode
	Amount volume.Volume
}

// PCM is a PCM-data instrument
type PCM struct {
	Sample       []uint8
	Length       int
	Loop         LoopInfo
	SustainLoop  LoopInfo
	NumChannels  int
	Format       SampleDataFormat
	Panning      panning.Position
	MixingVolume volume.Volume
	FadeOut      FadeoutSettings
	VolEnv       InstEnv
	PanEnv       InstEnv
	PitchEnv     InstEnv
}

// GetSample returns the sample at position `pos` in the instrument
func (inst *PCM) GetSample(ioc intf.NoteControl, pos sampling.Pos) volume.Matrix {
	ncs := ioc.GetPlaybackState()
	if ncs == nil {
		panic("no playback state on note-control interface")
	}
	ed := ioc.GetData().(*pcmState)
	envVol := inst.getVolEnv(ed)
	if envVol <= 0 {
		return volume.Matrix{}
	}

	dry := inst.getSampleDry(pos, ed.keyOn)
	chVol := ncs.Volume
	postVol := envVol * chVol * inst.MixingVolume
	wet := postVol.Apply(dry...)
	return wet
}

// IsLooped returns true if the instrument is looping
func (inst *PCM) IsLooped() bool {
	if inst.SustainLoop.Mode != LoopModeDisabled {
		return true
	}
	return inst.Loop.Mode != LoopModeDisabled
}

// GetCurrentPeriodDelta returns the current pitch envelope value
func (inst *PCM) GetCurrentPeriodDelta(ioc intf.NoteControl) note.PeriodDelta {
	if !inst.PitchEnv.Enabled {
		return note.PeriodDelta(0)
	}

	ed := ioc.GetData().(*pcmState)
	return ed.pitchEnvValue
}

// GetCurrentPanning returns the panning envelope position
func (inst *PCM) GetCurrentPanning(ioc intf.NoteControl) panning.Position {
	x := inst.Panning
	if !inst.PanEnv.Enabled {
		return x
	}

	ed := ioc.GetData().(*pcmState)
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
	ed := ioc.GetData().(*pcmState)
	ed.setEnvelopePosition(ticks, &ed.volEnvState, &inst.VolEnv, ioc, ed.updateVolEnv)
	if inst.VolEnv.SustainEnabled {
		ed.setEnvelopePosition(ticks, &ed.panEnvState, &inst.PanEnv, ioc, ed.updatePanEnv)
	}
}

func (inst *PCM) getVolEnv(ed *pcmState) volume.Volume {
	switch inst.FadeOut.Mode {
	case FadeoutModeDisabled:
		if !inst.VolEnv.Enabled {
			return volume.Volume(1)
		}
		return ed.volEnvValue
	case FadeoutModeAlwaysActive:
		if !inst.VolEnv.Enabled {
			return ed.fadeoutVol
		}
	case FadeoutModeOnlyIfVolEnvActive:
		if !inst.VolEnv.Enabled {
			return volume.Volume(1)
		}
	default:
		panic("unhandled method")
	}

	fadeVol := ed.fadeoutVol
	return fadeVol * ed.volEnvValue
}

func (inst *PCM) getSampleDry(pos sampling.Pos, keyOn bool) volume.Matrix {
	v0 := inst.getConvertedSample(pos.Pos, keyOn)
	if len(v0) == 0 && ((keyOn && inst.SustainLoop.Mode != LoopModeDisabled) || inst.Loop.Mode != LoopModeDisabled) {
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
	pos = calcLoopedSamplePos(inst.Loop, inst.SustainLoop, pos, inst.Length, keyOn)
	if pos < 0 || pos >= inst.Length {
		return volume.Matrix{}
	}
	return readSample(inst.Format, inst.Sample, pos, inst.NumChannels)
}

// Initialize completes the setup of this instrument
func (inst *PCM) Initialize(ioc intf.NoteControl) error {
	pcmState := newPcmState()
	ioc.SetData(pcmState)
	return nil
}

// Attack sets the key on flag for the instrument
func (inst *PCM) Attack(ioc intf.NoteControl) {
	ed := ioc.GetData().(*pcmState)
	ed.fadeoutVol = volume.Volume(1.0)
	ed.prevKeyOn = ed.keyOn
	ed.keyOn = true
	ed.fadingOut = false
	ed.setEnvelopePosition(0, &ed.volEnvState, &inst.VolEnv, ioc, ed.updateVolEnv)
	ed.setEnvelopePosition(0, &ed.panEnvState, &inst.PanEnv, ioc, ed.updatePanEnv)
	ed.setEnvelopePosition(0, &ed.pitchEnvState, &inst.PitchEnv, ioc, ed.updatePitchEnv)
}

// Release sets the key on flag for the instrument
func (inst *PCM) Release(ioc intf.NoteControl) {
	ed := ioc.GetData().(*pcmState)
	ed.prevKeyOn = ed.keyOn
	ed.keyOn = false
}

// Fadeout sets the instrument to fading-out mode (if able)
func (inst *PCM) Fadeout(ioc intf.NoteControl) {
	if inst.FadeOut.Mode == FadeoutModeDisabled {
		return
	}

	ed := ioc.GetData().(*pcmState)
	ed.fadingOut = true
}

// GetKeyOn gets the key on flag for the instrument
func (inst *PCM) GetKeyOn(ioc intf.NoteControl) bool {
	ed := ioc.GetData().(*pcmState)
	return ed.keyOn
}

// Update advances time by the amount specified by `tickDuration`
func (inst *PCM) Update(ioc intf.NoteControl, tickDuration time.Duration) {
	ed := ioc.GetData().(*pcmState)

	if ed.prevKeyOn != ed.keyOn && ed.prevKeyOn {
		ncs := ioc.GetPlaybackState()
		if ncs != nil {
			ncs.Pos.Pos = calcLoopedSamplePos(inst.Loop, inst.SustainLoop, ncs.Pos.Pos, inst.Length, ed.prevKeyOn)
		}
	}

	ed.advance(ioc, &inst.VolEnv, &inst.PanEnv, &inst.PitchEnv)

	if ed.fadingOut {
		performFade := false
		switch inst.FadeOut.Mode {
		case FadeoutModeDisabled:
			// nothing
		case FadeoutModeOnlyIfVolEnvActive, FadeoutModeAlwaysActive:
			performFade = true
		}
		if performFade {
			ed.fadeoutVol -= inst.FadeOut.Amount
			if ed.fadeoutVol < 0 {
				ed.fadeoutVol = 0
			}
		}
	}
}

// IsVolumeEnvelopeEnabled returns true if the volume envelope is enabled
func (inst *PCM) IsVolumeEnvelopeEnabled() bool {
	return inst.VolEnv.Enabled
}

// GetKind returns the kind of the instrument
func (inst *PCM) GetKind() note.InstrumentKind {
	return note.InstrumentKindPCM
}

// CloneData clones the data associated to the note-control interface
func (inst *PCM) CloneData(ioc intf.NoteControl) interface{} {
	ed := *ioc.GetData().(*pcmState)
	return &ed
}

// IsDone returns true if the instrument has stopped
func (inst *PCM) IsDone(ioc intf.NoteControl) bool {
	if inst.FadeOut.Mode == FadeoutModeDisabled {
		return false
	}
	ed := ioc.GetData().(*pcmState)
	return ed.fadeoutVol <= 0
}
