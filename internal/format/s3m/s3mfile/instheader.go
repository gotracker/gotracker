package s3mfile

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

// SCRSFlags is a bitset for the S3M instrument/sample header definition
type SCRSFlags uint8

const (
	// SCRSFlagsLooped is looping
	SCRSFlagsLooped = SCRSFlags(0x01)
	// SCRSFlagsStereo is stereo
	SCRSFlagsStereo = SCRSFlags(0x02)
	// SCRSFlags16Bit is 16-bit
	SCRSFlags16Bit = SCRSFlags(0x04)
)

// IsLooped returns true if bit 0 is set
func (f SCRSFlags) IsLooped() bool {
	return (f & SCRSFlagsLooped) != 0
}

// IsStereo returns true if bit 1 is set
func (f SCRSFlags) IsStereo() bool {
	return (f & SCRSFlagsStereo) != 0
}

// Is16BitSample returns true if bit 2 is set
func (f SCRSFlags) Is16BitSample() bool {
	return (f & SCRSFlags16Bit) != 0
}

// HiLo32 is a 32 bit value where the first and second 16 bits are stored separately
type HiLo32 struct {
	Lo uint16
	Hi uint16
}

// SCRSType is the type of the SCRS instrument/sample
type SCRSType uint8

const (
	// SCRSTypeNone is a None type
	SCRSTypeNone = SCRSType(0 + iota)
	// SCRSTypeDigiplayer is a Digiplayer/S3M PCM sample
	SCRSTypeDigiplayer
	// SCRSTypeOPL2Melody is an Adlib/OPL2 melody instrument
	SCRSTypeOPL2Melody
	// SCRSTypeOPL2BassDrum is an Adlib/OPL2 bass drum instrument
	SCRSTypeOPL2BassDrum
	// SCRSTypeOPL2Snare is an Adlib/OPL2 snare drum instrument
	SCRSTypeOPL2Snare
	// SCRSTypeOPL2Tom is an Adlib/OPL2 tom drum instrument
	SCRSTypeOPL2Tom
	// SCRSTypeOPL2Cymbal is an Adlib/OPL2 cymbal instrument
	SCRSTypeOPL2Cymbal
	// SCRSTypeOPL2HiHat is an Adlib/OPL2 hi-hat instrument
	SCRSTypeOPL2HiHat
)

// SCRSHeader is the S3M instrument/sample header definition
type SCRSHeader struct {
	Type     SCRSType
	Filename [12]byte
}

// SCRSAncillaryHeader is the generic interface of the Type-specific header
type SCRSAncillaryHeader interface{}

// Packing is a type of sample packing format
type Packing uint8

const (
	// PackingUnpacked is an unpacked S3M PCM sample
	PackingUnpacked = Packing(iota)
	// PackingDP30ADPCM is Digiplayer/ST3 3.00 ADPCM packing
	PackingDP30ADPCM
)

// SCRSDigiplayerHeader is the remaining header for S3M PCM samples
type SCRSDigiplayerHeader struct {
	MemSeg        ParaPointer24
	Length        HiLo32
	LoopBegin     HiLo32
	LoopEnd       HiLo32
	Volume        uint8
	Reserved1D    uint8
	PackingScheme Packing
	Flags         SCRSFlags
	C2Spd         HiLo32
	Reserved24    [4]byte
	Reserved28    [2]byte
	Reserved2A    [2]byte
	Reserved2C    [4]byte
	SampleName    [28]byte
	SCRS          [4]uint8
}

// OPL2ModulatorA is a bit-field of Adlib/OPL2 modulators
type OPL2ModulatorA uint8

const (
	opl2ModulatorAScaleEnv      = OPL2ModulatorA(0x10)
	opl2ModulatorASustain       = OPL2ModulatorA(0x20)
	opl2ModulatorAPitchVibrato  = OPL2ModulatorA(0x40)
	opl2ModulatorAVolumeVibrato = OPL2ModulatorA(0x80)
)

// ScaleEnvelope returns the Scale Envelope flag
func (o OPL2ModulatorA) ScaleEnvelope() bool {
	return (o & opl2ModulatorAScaleEnv) != 0
}

// Sustain returns the Sustain flag
func (o OPL2ModulatorA) Sustain() bool {
	return (o & opl2ModulatorASustain) != 0
}

// PitchVibrato returns the Pitch Vibrato flag
func (o OPL2ModulatorA) PitchVibrato() bool {
	return (o & opl2ModulatorAPitchVibrato) != 0
}

// VolumeVibrato returns the Volume Vibrato flag
func (o OPL2ModulatorA) VolumeVibrato() bool {
	return (o & opl2ModulatorAVolumeVibrato) != 0
}

// FrequencyMultiplier returns the Frequency Multiplier
func (o OPL2ModulatorA) FrequencyMultiplier() uint8 {
	return uint8(o) & 0x0F
}

// OPL2ModulatorB is a bit-field of Adlib/OPL2 modulators
type OPL2ModulatorB uint8

// LevelScale returns the level scale
func (o OPL2ModulatorB) LevelScale() uint8 {
	v := uint8(o)
	bit0 := (v & 0x80) >> 7
	bit1 := (v & 0x40) >> 6
	return (bit1 << 1) | bit0
}

// Volume returns the volume
func (o OPL2ModulatorB) Volume() uint8 {
	v := uint8(o)
	return 63 - v
}

// OPL2ModulatorC is a bit-field of Adlib/OPL2 modulators
type OPL2ModulatorC uint8

// Attack returns the attack of the envelope
func (o OPL2ModulatorC) Attack() uint8 {
	return uint8(o) >> 4
}

// Decay returns the attack of the envelope
func (o OPL2ModulatorC) Decay() uint8 {
	return uint8(o) & 0x1F
}

// OPL2ModulatorD is a bit-field of Adlib/OPL2 modulators
type OPL2ModulatorD uint8

// Sustain returns the sustain of the envelope
func (o OPL2ModulatorD) Sustain() uint8 {
	return 15 - (uint8(o) >> 4)
}

// Release returns the release of the envelope
func (o OPL2ModulatorD) Release() uint8 {
	return uint8(o) & 0x1F
}

// OPL2ModulatorE is a bit-field of Adlib/OPL2 modulators
type OPL2ModulatorE uint8

// WaveSelect returns the wave select value
func (o OPL2ModulatorE) WaveSelect() uint8 {
	return uint8(o)
}

// OPL2ModulatorFeedbackAddSynth is a bit-field of Adlib/OPL2 values
type OPL2ModulatorFeedbackAddSynth uint8

const (
	opl2ModulatorAddSynth = OPL2ModulatorFeedbackAddSynth(0x01)
)

// ModulationFeedback returns the modulation feedback of the instrument
func (o OPL2ModulatorFeedbackAddSynth) ModulationFeedback() uint8 {
	return uint8(o) >> 1
}

// AdditiveSynth returns the additive synth flag
func (o OPL2ModulatorFeedbackAddSynth) AdditiveSynth() bool {
	return (o & opl2ModulatorAddSynth) != 0
}

// OPL2Specs is the specifiers for an OPL2/Adlib instrument
type OPL2Specs struct {
	ModulatorA    OPL2ModulatorA                // D00
	CarrierA      uint8                         // D01
	ModulatorB    OPL2ModulatorB                // D02
	CarrierB      uint8                         // D03
	ModulatorC    OPL2ModulatorC                // D04
	CarrierC      uint8                         // D05
	ModulatorD    OPL2ModulatorD                // D06
	CarrierD      uint8                         // D07
	ModulatorE    OPL2ModulatorE                // D08
	CarrierE      uint8                         // D09
	FeedbackSynth OPL2ModulatorFeedbackAddSynth // D0A
	Reserved0B    [1]byte                       // D0B
}

// SCRSAdlibHeader is the remaining header for S3M adlib instruments
type SCRSAdlibHeader struct {
	Reserved0D    [3]byte
	D             OPL2Specs
	Volume        uint8
	Dsk           uint8 // no idea what this is - maybe something to do with the Ensoniq Mirage DSK-8?
	PackingScheme uint8
	Reserved1E    [2]byte
	C2Spd         HiLo32
	Reserved24    [12]byte
	SampleName    [28]byte
	SCRI          [4]uint8
}

// SCRSNoneHeader is the remaining header for S3M none-type instrument
type SCRSNoneHeader struct {
	Reserved0D [19]byte
	Volume     uint8
	Reserved1D [3]byte
	C2Spd      HiLo32
	Reserved24 [12]byte
	SampleName [28]byte
	Reserved4C [4]uint8
}

// SCRS is a full header for an S3M instrument
type SCRS struct {
	Head      SCRSHeader
	Ancillary SCRSAncillaryHeader
}

// ReadSCRS reads an SCRS from the input stream
func ReadSCRS(r io.Reader) (*SCRS, error) {
	sh := SCRS{}
	if err := binary.Read(r, binary.LittleEndian, &sh.Head); err != nil {
		return nil, err
	}

	switch sh.Head.Type {
	case SCRSTypeNone:
		sh.Ancillary = &SCRSNoneHeader{}
	case SCRSTypeDigiplayer:
		sh.Ancillary = &SCRSDigiplayerHeader{}
	case SCRSTypeOPL2Melody, SCRSTypeOPL2BassDrum, SCRSTypeOPL2Snare, SCRSTypeOPL2Tom, SCRSTypeOPL2Cymbal, SCRSTypeOPL2HiHat:
		sh.Ancillary = &SCRSAdlibHeader{}
	default:
		return nil, errors.Errorf("unknown SCRS instrument type %0.2x", sh.Head.Type)
	}

	if sh.Ancillary != nil {
		if err := binary.Read(r, binary.LittleEndian, sh.Ancillary); err != nil {
			return nil, err
		}
	}

	return &sh, nil
}
