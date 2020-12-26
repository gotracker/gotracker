package layout

import (
	"encoding/binary"
	"time"

	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/format/s3m/playback/util"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

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
}

// GetSample returns the sample at position `pos` in the instrument
func (inst *InstrumentPCM) GetSample(ioc *InstrumentOnChannel, pos sampling.Pos) volume.Matrix {
	dry := inst.getSampleDry(ioc, pos)
	return ioc.Volume.Apply(dry...)
}

func (inst *InstrumentPCM) getSampleDry(ioc *InstrumentOnChannel, pos sampling.Pos) volume.Matrix {
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
		pos = inst.calcLoopedSamplePosMode1(pos)
	}
	if pos < 0 || pos >= inst.Length {
		return volume.Matrix{}
	}
	o := make(volume.Matrix, inst.NumChannels)
	for c := 0; c < inst.NumChannels; c++ {
		switch inst.BitsPerSample {
		case 8:
			o[c] = util.VolumeFromS3M8BitSample(inst.Sample[pos+c])
		case 16:
			s := binary.LittleEndian.Uint16(inst.Sample[pos+c:])
			o[c] = util.VolumeFromS3M16BitSample(s)
		}
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
		return inst.Length
	}

	loopedPos := (pos - inst.LoopEnd) % loopLen
	return inst.LoopBegin + loopedPos
}

// Initialize completes the setup of this instrument
func (inst *InstrumentPCM) Initialize(ioc *InstrumentOnChannel) error {
	return nil
}

// SetKeyOn sets the key on flag for the instrument
func (inst *InstrumentPCM) SetKeyOn(ioc *InstrumentOnChannel, period note.Period, on bool) {
}

// GetKeyOn gets the key on flag for the instrument
func (inst *InstrumentPCM) GetKeyOn(ioc *InstrumentOnChannel) bool {
	return false
}

// Update advances time by the amount specified by `tickDuration`
func (inst *InstrumentPCM) Update(ioc *InstrumentOnChannel, tickDuration time.Duration) {
}
