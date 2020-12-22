package layout

import (
	"math"
	"time"

	s3mfile "github.com/heucuva/goaudiofile/music/tracked/s3m"
	"github.com/heucuva/gomixing/sampling"
	"github.com/heucuva/gomixing/volume"

	"gotracker/internal/format/s3m/playback/opl2"
	"gotracker/internal/format/s3m/playback/util"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// OPL2OperatorData is the operator data for an OPL2/Adlib instrument
type OPL2OperatorData struct {
	// KeyScaleRateSelect returns true if the modulator's envelope scales with keys
	// If enabled, the envelopes of higher notes are played more quickly than those of lower notes.
	KeyScaleRateSelect bool

	// Sustain returns true if the modulator's envelope sustain is enabled
	// If enabled, the volume envelope stays at the sustain stage and does not enter the
	// release stage of the envelope until a note-off event is encountered. Otherwise, it
	// directly advances from the decay stage to the release stage without waiting for a
	// note-off event.
	Sustain bool

	// Vibrato returns true if the modulator's vibrato is enabled
	// If enabled, adds a vibrato effect with a depth of 7 cents (0.07 semitones).
	// The rate of this vibrato is a static 6.4Hz.
	Vibrato bool

	// Tremolo returns true if the modulator's tremolo is enabled
	// If enabled, adds a tremolo effect with a depth of 1dB.
	// The rate of this tremolo is a static 3.7Hz.
	Tremolo bool

	// FrequencyMultiplier returns the modulator's frequency multiplier
	// Multiplies the frequency of the operator with a value between 0.5
	// (pitched one octave down) and 15.
	FrequencyMultiplier s3mfile.OPL2Multiple

	// KeyScaleLevel returns the key scale level
	// Attenuates the output level of the operators towards higher pitch by the given amount
	// (disabled, 1.5 dB / octave, 3 dB / octave, 6 dB / octave).
	KeyScaleLevel s3mfile.OPL2KSL

	// Volume returns the modulator's volume
	// The overall volume of the operator - if the modulator is in FM mode (i.e.: NOT in
	// additive synthesis mode), this will instead be the total pitch depth.
	Volume s3mfile.Volume

	// AttackRate returns the modulator's attack rate
	// Specifies how fast the volume envelope fades in from silence to peak volume.
	AttackRate uint8

	// DecayRate returns the modulator's decay rate
	// Specifies how fast the volume envelope reaches the sustain volume after peaking.
	DecayRate uint8

	// SustainLevel returns the modulator's sustain level
	// Specifies at which level the volume envelope is held before it is released.
	SustainLevel uint8

	// ReleaseRate returns the modulator's release rate
	// Specifies how fast the volume envelope fades out from the sustain level.
	ReleaseRate uint8

	// WaveformSelection returns the modulator's waveform selection
	WaveformSelection s3mfile.OPL2Waveform
}

// InstrumentOPL2 is an OPL2/Adlib instrument
type InstrumentOPL2 struct {
	intf.Instrument

	Modulator OPL2OperatorData
	Carrier   OPL2OperatorData

	// ModulationFeedback returns the modulation feedback
	ModulationFeedback s3mfile.OPL2Feedback

	// AdditiveSynthesis returns true if additive synthesis is enabled
	AdditiveSynthesis bool
}

type ym3812 struct {
	ch   *opl2.SingleChannel
	data []int32
}

// GetSample returns the sample at position `pos` in the instrument
func (inst *InstrumentOPL2) GetSample(ioc *InstrumentOnChannel, pos sampling.Pos) volume.Matrix {
	ym := ioc.Data.(*ym3812)

	v0 := inst.getConvertedSample(ym.data, pos.Pos)
	if pos.Frac == 0 {
		return v0
	}
	v1 := inst.getConvertedSample(ym.data, pos.Pos+1)
	for c, s := range v1 {
		v0[c] += volume.Volume(pos.Frac) * (s - v0[c])
	}
	return v0
}

func (inst *InstrumentOPL2) getConvertedSample(data []int32, pos int) volume.Matrix {
	if pos < 0 || pos >= len(data) {
		return volume.Matrix{}
	}
	o := make(volume.Matrix, 1)
	w := data[pos]
	o[0] = (volume.Volume(w)) / 65536.0
	return o
}

var ym3812Channels [32]*opl2.SingleChannel

// Initialize completes the setup of this instrument
func (inst *InstrumentOPL2) Initialize(ioc *InstrumentOnChannel) error {
	chNum := ioc.OutputChannelNum
	ch := ym3812Channels[chNum]
	if ch == nil {
		rate := opl2.OPLRATE
		ch = opl2.NewSingleChannel(uint32(rate))
		// support all waveforms
		ch.SupportAllWaveforms(true)
		ym3812Channels[chNum] = ch
	}
	ym := ym3812{
		ch: ch,
	}
	ioc.Data = &ym

	return nil
}

// SetKeyOn sets the key on flag for the instrument
func (inst *InstrumentOPL2) SetKeyOn(ioc *InstrumentOnChannel, semitone note.Semitone, on bool) {
	ym := ioc.Data.(*ym3812)
	ch := ym.ch

	// write the instrument to the channel!
	freq, block := inst.freqToFreqBlock(opl2.OPLRATE / 16)
	if !on {
		ch.WriteFNum(freq, block)
		ch.SetKeyOn(false)
		ym.data = nil
	} else {
		modReg20 := inst.getReg20(&inst.Modulator)
		modReg40 := inst.getReg40(&inst.Modulator)
		modReg60 := inst.getReg60(&inst.Modulator)
		modReg80 := inst.getReg80(&inst.Modulator)
		modRegE0 := inst.getRegE0(&inst.Modulator)

		carReg20 := inst.getReg20(&inst.Carrier)
		carReg40 := inst.getReg40(&inst.Carrier)
		carReg60 := inst.getReg60(&inst.Carrier)
		carReg80 := inst.getReg80(&inst.Carrier)
		carRegE0 := inst.getRegE0(&inst.Carrier)

		regC0 := inst.getRegC0()

		ch.WriteReg(0x20, 0, modReg20)
		ch.WriteReg(0x40, 0, modReg40)
		ch.WriteReg(0x60, 0, modReg60)
		ch.WriteReg(0x80, 0, modReg80)
		ch.WriteReg(0xE0, 0, modRegE0)

		ch.WriteFNum(freq, block)

		ch.WriteReg(0x20, 1, carReg20)
		ch.WriteReg(0x40, 1, carReg40)
		ch.WriteReg(0x60, 1, carReg60)
		ch.WriteReg(0x80, 1, carReg80)
		ch.WriteReg(0xE0, 1, carRegE0)

		ch.WriteC0(regC0)

		ch.SetKeyOn(true)
	}
}

func (inst *InstrumentOPL2) getReg20(o *OPL2OperatorData) uint8 {
	reg20 := uint8(0x00)
	if o.Tremolo {
		reg20 |= 0x80
	}
	if o.Vibrato {
		reg20 |= 0x40
	}
	if o.Sustain {
		reg20 |= 0x20
	}
	if o.KeyScaleRateSelect {
		reg20 |= 0x10
	}
	reg20 |= uint8(o.FrequencyMultiplier) & 0x0f

	return reg20
}

func (inst *InstrumentOPL2) getReg40(o *OPL2OperatorData) uint8 {
	levelScale := o.KeyScaleLevel >> 1
	levelScale |= (o.KeyScaleLevel << 1) & 2
	//levelScale := o.KeyScaleLevel
	reg40 := uint8(0x00)
	reg40 |= uint8(levelScale) << 6
	reg40 |= uint8(63-o.Volume) & 0x3f
	return reg40
}

func (inst *InstrumentOPL2) getReg60(o *OPL2OperatorData) uint8 {
	reg60 := uint8(0x00)
	reg60 |= (o.AttackRate & 0x0f) << 4
	reg60 |= o.DecayRate & 0x0f
	return reg60
}

func (inst *InstrumentOPL2) getReg80(o *OPL2OperatorData) uint8 {
	reg80 := uint8(0x00)
	reg80 |= ((15 - o.SustainLevel) & 0x0f) << 4
	reg80 |= o.ReleaseRate & 0x0f
	return reg80
}

func (inst *InstrumentOPL2) getRegC0() uint8 {
	regC0 := uint8(0x00)
	//regC0 |= 0x40 | 0x20 // channel enable: right | left
	regC0 |= (uint8(inst.ModulationFeedback) & 0x07) << 1
	if inst.AdditiveSynthesis {
		regC0 |= 0x01
	}
	return regC0
}

func (inst *InstrumentOPL2) getRegE0(o *OPL2OperatorData) uint8 {
	regE0 := uint8(0x00)
	regE0 |= uint8(o.WaveformSelection) & 0x03
	return regE0
}

// twoOperatorMelodic
var twoOperatorMelodic = [18]uint32{0, 1, 2, 6, 7, 8, 12, 13, 14, 18, 19, 20, 24, 25, 26, 30, 31, 32}

func (inst *InstrumentOPL2) getChannelIndex(channelIdx int) uint32 {
	return twoOperatorMelodic[channelIdx%18]
}

func (inst *InstrumentOPL2) semitoneToFreqBlock(semitone note.Semitone, c2spd note.C2SPD) (uint16, uint8) {
	targetFreq := float64(util.FrequencyFromSemitone(semitone, c2spd))

	return inst.freqToFreqBlock(targetFreq / 256)
}

func (inst *InstrumentOPL2) freqToFreqBlock(targetFreq float64) (uint16, uint8) {
	bestBlk := uint8(8)
	bestMatchFreqNum := uint16(0)
	bestMatchFNDelta := float64(1024)
	for blk := uint8(0); blk < 8; blk++ {
		fNum := targetFreq * float64(uint32(1<<(20-blk))) / opl2.OPLRATE
		iNum := int(fNum)
		fp := fNum - float64(iNum)
		if iNum < 1024 && iNum >= 0 && fp < bestMatchFNDelta {
			bestBlk = blk
			bestMatchFreqNum = uint16(iNum)
			bestMatchFNDelta = fp
		}
	}

	return bestMatchFreqNum, bestBlk
}

func (inst *InstrumentOPL2) freqBlockToRegA0B0(freq uint16, block uint8) (uint8, uint8) {
	regA0 := uint8(freq)
	regB0 := uint8(uint16(freq)>>8) & 0x03
	regB0 |= (block & 0x07) << 3
	return regA0, regB0
}

// GetKeyOn gets the key on flag for the instrument
func (inst *InstrumentOPL2) GetKeyOn(ioc *InstrumentOnChannel) bool {
	ym := ioc.Data.(*ym3812)
	ch := ym.ch
	return ch.GetKeyOn()
}

// Update advances time by the amount specified by `tickDuration`
func (inst *InstrumentOPL2) Update(ioc *InstrumentOnChannel, tickDuration time.Duration) {
	ym := ioc.Data.(*ym3812)
	ll := uint(math.Ceil(tickDuration.Seconds() * opl2.OPLRATE))
	gen := make([]int32, ll)
	ym.ch.GenerateBlock2(ll, gen)
	ym.data = append(ym.data, gen...)
}
