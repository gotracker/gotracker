package channel

import (
	"github.com/gotracker/voice/oscillator"

	formatutil "gotracker/internal/format/internal/util"
	oscillatorImpl "gotracker/internal/oscillator"
)

// Memory is the storage object for custom effect/effect values
type Memory struct {
	portaToNote         uint8
	vibrato             uint8
	vibratoSpeed        uint8
	sampleOffset        uint8
	tempoDecrease       uint8
	tempoIncrease       uint8
	portaDown           uint8
	portaUp             uint8
	tremolo             uint8
	tremor              uint8
	volumeSlide         uint8
	globalVolumeSlide   uint8
	finePortaUp         uint8
	finePortaDown       uint8
	fineVolumeSlideUp   uint8
	fineVolumeSlideDown uint8
	extraFinePortaUp    uint8
	extraFinePortaDown  uint8

	// LinearFreqSlides is true if linear frequency slides are enabled (false = amiga-style period-based slides)
	LinearFreqSlides bool

	tremorMem         formatutil.Tremor
	vibratoOscillator oscillator.Oscillator
	tremoloOscillator oscillator.Oscillator
	patternLoop       formatutil.PatternLoop
}

// ResetOscillators resets the oscillators to defaults
func (m *Memory) ResetOscillators() {
	m.vibratoOscillator = oscillatorImpl.NewProtrackerOscillator()
	m.tremoloOscillator = oscillatorImpl.NewProtrackerOscillator()
}

func (m *Memory) getEffectMemory(input uint8, reg *uint8) uint8 {
	if input == 0 {
		return *reg
	}
	if input != 0 {
		*reg = input
	}
	return input
}

// PortaToNote gets or sets the most recent non-zero value (or input) for Portamento-to-note
func (m *Memory) PortaToNote(input uint8) uint8 {
	return m.getEffectMemory(input, &m.portaToNote)
}

// Vibrato gets or sets the most recent non-zero value (or input) for Vibrato
func (m *Memory) Vibrato(input uint8) uint8 {
	return m.getEffectMemory(input, &m.vibrato)
}

// VibratoSpeed gets or sets the most recent non-zero value (or input) for Vibrato Speed
func (m *Memory) VibratoSpeed(input uint8) uint8 {
	return m.getEffectMemory(input, &m.vibratoSpeed)
}

// SampleOffset gets or sets the most recent non-zero value (or input) for Sample Offset
func (m *Memory) SampleOffset(input uint8) uint8 {
	return m.getEffectMemory(input, &m.sampleOffset)
}

// TempoDecrease gets or sets the most recent non-zero value (or input) for Tempo Decrease
func (m *Memory) TempoDecrease(input uint8) uint8 {
	return m.getEffectMemory(input, &m.tempoDecrease)
}

// TempoIncrease gets or sets the most recent non-zero value (or input) for Tempo Increase
func (m *Memory) TempoIncrease(input uint8) uint8 {
	return m.getEffectMemory(input, &m.tempoIncrease)
}

// PortaDown gets or sets the most recent non-zero value (or input) for Portamento Down
func (m *Memory) PortaDown(input uint8) uint8 {
	return m.getEffectMemory(input, &m.portaDown)
}

// PortaUp gets or sets the most recent non-zero value (or input) for Portamento Up
func (m *Memory) PortaUp(input uint8) uint8 {
	return m.getEffectMemory(input, &m.portaUp)
}

// Tremolo gets or sets the most recent non-zero value (or input) for Tremolo
func (m *Memory) Tremolo(input uint8) uint8 {
	return m.getEffectMemory(input, &m.tremolo)
}

// Tremor gets or sets the most recent non-zero value (or input) for Tremor
func (m *Memory) Tremor(input uint8) uint8 {
	return m.getEffectMemory(input, &m.tremor)
}

// VolumeSlide gets or sets the most recent non-zero value (or input) for Volume Slide
func (m *Memory) VolumeSlide(input uint8) uint8 {
	return m.getEffectMemory(input, &m.volumeSlide)
}

// GlobalVolumeSlide gets or sets the most recent non-zero value (or input) for Global Volume Slide
func (m *Memory) GlobalVolumeSlide(input uint8) uint8 {
	return m.getEffectMemory(input, &m.globalVolumeSlide)
}

// FinePortaUp gets or sets the most recent non-zero value (or input) for Fine Portamento Up
func (m *Memory) FinePortaUp(input uint8) uint8 {
	return m.getEffectMemory(input&0x0F, &m.finePortaUp)
}

// FinePortaDown gets or sets the most recent non-zero value (or input) for Fine Portamento Down
func (m *Memory) FinePortaDown(input uint8) uint8 {
	return m.getEffectMemory(input&0x0F, &m.finePortaDown)
}

// FineVolumeSlideUp gets or sets the most recent non-zero value (or input) for Fine Volume Slide Up
func (m *Memory) FineVolumeSlideUp(input uint8) uint8 {
	return m.getEffectMemory(input&0x0F, &m.fineVolumeSlideUp)
}

// FineVolumeSlideDown gets or sets the most recent non-zero value (or input) for Fine Volume Slide Down
func (m *Memory) FineVolumeSlideDown(input uint8) uint8 {
	return m.getEffectMemory(input&0x0F, &m.fineVolumeSlideDown)
}

// ExtraFinePortaUp gets or sets the most recent non-zero value (or input) for Extra Fine Portamento Up
func (m *Memory) ExtraFinePortaUp(input uint8) uint8 {
	return m.getEffectMemory(input&0x0F, &m.extraFinePortaUp)
}

// ExtraFinePortaDown gets or sets the most recent non-zero value (or input) for Extra Fine Portamento Down
func (m *Memory) ExtraFinePortaDown(input uint8) uint8 {
	return m.getEffectMemory(input&0x0F, &m.extraFinePortaDown)
}

// TremorMem returns the Tremor object
func (m *Memory) TremorMem() *formatutil.Tremor {
	return &m.tremorMem
}

// VibratoOscillator returns the Vibrato oscillator object
func (m *Memory) VibratoOscillator() oscillator.Oscillator {
	return m.vibratoOscillator
}

// TremoloOscillator returns the Tremolo oscillator object
func (m *Memory) TremoloOscillator() oscillator.Oscillator {
	return m.tremoloOscillator
}

// Retrigger runs certain operations when a note is retriggered
func (m *Memory) Retrigger() {
	for _, osc := range []oscillator.Oscillator{m.VibratoOscillator(), m.TremoloOscillator()} {
		osc.Reset()
	}
}

// GetPatternLoop returns the pattern loop object from the memory
func (m *Memory) GetPatternLoop() *formatutil.PatternLoop {
	return &m.patternLoop
}
