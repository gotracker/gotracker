package layout

import (
	"encoding/binary"
	"time"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/player/intf"
)

// EnvPoint is a point for the envelope
type EnvPoint struct {
	Ticks int
	Y     interface{}
}

// InstEnv is an envelope for instruments
type InstEnv struct {
	Enabled        bool
	LoopEnabled    bool
	SustainEnabled bool
	LoopStart      int
	LoopEnd        int
	SustainIndex   int
	Values         []EnvPoint
}

// InstrumentPCM is a PCM-data instrument
type InstrumentPCM struct {
	intf.Instrument

	Sample        []uint8
	Length        int
	Looped        bool
	LoopBegin     int
	LoopEnd       int
	NumChannels   int
	BitsPerSample int
	VolumeFadeout volume.Volume
	VolEnv        InstEnv
	PanEnv        InstEnv
}

// GetSample returns the sample at position `pos` in the instrument
func (inst *InstrumentPCM) GetSample(ioc intf.NoteControl, pos sampling.Pos) volume.Matrix {
	ed := ioc.GetData().(*envData)
	dry := inst.getSampleDry(pos)
	volEnv := inst.getVolEnv(ed, pos)
	wet := dry
	wet = volEnv.Apply(wet...)
	wet = ed.fadeoutVol.Apply(wet...)
	return ioc.GetVolume().Apply(wet...)
}

// GetCurrentPanning returns the panning envelope position
func (inst *InstrumentPCM) GetCurrentPanning(ioc intf.NoteControl) panning.Position {
	if !inst.PanEnv.Enabled {
		return panning.CenterAhead
	}

	ed := ioc.GetData().(*envData)
	return ed.panEnvValue
}

func (inst *InstrumentPCM) getVolEnv(ed *envData, pos sampling.Pos) volume.Volume {
	if !inst.VolEnv.Enabled {
		return volume.Volume(1.0)
	}

	return ed.volEnvValue
}

func (inst *InstrumentPCM) getSampleDry(pos sampling.Pos) volume.Matrix {
	v0 := inst.getConvertedSample(pos.Pos)
	if len(v0) == 0 && inst.Looped {
		v01 := inst.getConvertedSample(pos.Pos)
		panic(v01)
	}
	if pos.Frac == 0 {
		return v0
	}
	v1 := inst.getConvertedSample(pos.Pos + 1)
	for c, s := range v1 {
		v0[c] += volume.Volume(pos.Frac) * (s - v0[c])
	}
	return v0
}

func (inst *InstrumentPCM) getConvertedSample(pos int) volume.Matrix {
	if inst.Looped {
		pos = inst.calcLoopedSamplePosMode2(pos)
	}
	bps := inst.BitsPerSample / 8
	if pos < 0 || pos >= inst.Length {
		return volume.Matrix{}
	}
	o := make(volume.Matrix, inst.NumChannels)
	actualPos := pos * inst.NumChannels * bps
	for c := 0; c < inst.NumChannels; c++ {
		switch inst.BitsPerSample {
		case 8:
			o[c] = util.VolumeFromXm8BitSample(inst.Sample[actualPos+c])
		case 16:
			s := binary.LittleEndian.Uint16(inst.Sample[actualPos+c:])
			o[c] = util.VolumeFromXm16BitSample(s)
		}
		actualPos += bps
	}
	return o
}

func (inst *InstrumentPCM) calcLoopedSamplePosMode1(pos int) int {
	// |start>-------------------------------------------------end|   <= on playthrough 1, whole sample plays
	// |----------------|loopBegin>--------loopEnd|---------------|   <= only if looped and on playthrough 2+, only the part that loops plays
	if pos < 0 {
		return 0
	}
	if pos < inst.Length {
		return pos
	}

	loopLen := inst.LoopEnd - inst.LoopBegin
	if loopLen <= 0 {
		return inst.Length
	}

	loopedPos := (pos - inst.Length) % loopLen
	return inst.LoopBegin + loopedPos
}

func (inst *InstrumentPCM) calcLoopedSamplePosMode2(pos int) int {
	// |start>-----------------------------loopEnd|>-----------end|   <= on playthrough 1, play from start to loopEnd if looped, otherwise continue to end
	// |----------------|loopBegin>--------loopEnd|---------------|   <= on playthrough 2+, only the part that loops plays
	if pos < 0 {
		return 0
	}
	if pos < inst.LoopEnd {
		return pos
	}

	loopLen := inst.LoopEnd - inst.LoopBegin
	if loopLen <= 0 {
		if pos < inst.Length {
			return pos
		}
		return inst.Length
	}

	loopedPos := (pos - inst.LoopEnd) % loopLen
	return inst.LoopBegin + loopedPos
}

// Initialize completes the setup of this instrument
func (inst *InstrumentPCM) Initialize(ioc intf.NoteControl) error {
	envData := newEnvData()
	ioc.SetData(envData)
	return nil
}

// Attack sets the key on flag for the instrument
func (inst *InstrumentPCM) Attack(ioc intf.NoteControl) {
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
func (inst *InstrumentPCM) Release(ioc intf.NoteControl) {
	ed := ioc.GetData().(*envData)
	ed.keyOn = false
}

// NoteCut cuts the current playback of the instrument
func (inst *InstrumentPCM) NoteCut(ioc intf.NoteControl) {
	ed := ioc.GetData().(*envData)
	ed.keyOn = false
	ed.fadeoutVol = volume.Volume(0.0)
}

// GetKeyOn gets the key on flag for the instrument
func (inst *InstrumentPCM) GetKeyOn(ioc intf.NoteControl) bool {
	ed := ioc.GetData().(*envData)
	return ed.keyOn
}

// Update advances time by the amount specified by `tickDuration`
func (inst *InstrumentPCM) Update(ioc intf.NoteControl, tickDuration time.Duration) {
	ed := ioc.GetData().(*envData)

	ed.advance(&inst.VolEnv, &inst.PanEnv)

	if !ed.keyOn && inst.VolumeFadeout != 0 {
		ed.fadeoutVol -= inst.VolumeFadeout
		if ed.fadeoutVol < 0 {
			ed.fadeoutVol = 0
		}
	}
}
