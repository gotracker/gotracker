package channel

import (
	"github.com/gotracker/voice/oscillator"

	"gotracker/internal/format/internal/effect"
	"gotracker/internal/format/internal/memory"
	formatutil "gotracker/internal/format/internal/util"
	oscillatorImpl "gotracker/internal/oscillator"
)

// Memory is the storage object for custom effect/effect values
type Memory struct {
	portaToNote         memory.UInt8
	vibrato             memory.UInt8
	vibratoSpeed        memory.UInt8
	sampleOffset        memory.UInt8
	tempoDecrease       memory.UInt8
	tempoIncrease       memory.UInt8
	portaDown           memory.UInt8
	portaUp             memory.UInt8
	tremolo             memory.UInt8
	tremor              memory.UInt8
	volumeSlide         memory.UInt8
	globalVolumeSlide   memory.UInt8
	finePortaUp         memory.UInt8
	finePortaDown       memory.UInt8
	fineVolumeSlideUp   memory.UInt8
	fineVolumeSlideDown memory.UInt8
	extraFinePortaUp    memory.UInt8
	extraFinePortaDown  memory.UInt8

	// LinearFreqSlides is true if linear frequency slides are enabled (false = amiga-style period-based slides)
	LinearFreqSlides bool
	// ResetMemoryAtStartOfOrder0 if true will reset the memory registers when the first tick of the first row of the first order pattern plays
	ResetMemoryAtStartOfOrder0 bool

	tremorMem         effect.Tremor
	vibratoOscillator oscillator.Oscillator
	tremoloOscillator oscillator.Oscillator
	patternLoop       formatutil.PatternLoop
}

// ResetOscillators resets the oscillators to defaults
func (m *Memory) ResetOscillators() {
	m.vibratoOscillator = oscillatorImpl.NewProtrackerOscillator()
	m.tremoloOscillator = oscillatorImpl.NewProtrackerOscillator()
}

// PortaToNote gets or sets the most recent non-zero value (or input) for Portamento-to-note
func (m *Memory) PortaToNote(input uint8) uint8 {
	return m.portaToNote.Coalesce(input)
}

// Vibrato gets or sets the most recent non-zero value (or input) for Vibrato
func (m *Memory) Vibrato(input uint8) (uint8, uint8) {
	return m.vibrato.CoalesceXY(input)
}

// VibratoSpeed gets or sets the most recent non-zero value (or input) for Vibrato Speed
func (m *Memory) VibratoSpeed(input uint8) uint8 {
	return m.vibratoSpeed.Coalesce(input)
}

// SampleOffset gets or sets the most recent non-zero value (or input) for Sample Offset
func (m *Memory) SampleOffset(input uint8) uint8 {
	return m.sampleOffset.Coalesce(input)
}

// TempoDecrease gets or sets the most recent non-zero value (or input) for Tempo Decrease
func (m *Memory) TempoDecrease(input uint8) uint8 {
	return m.tempoDecrease.Coalesce(input)
}

// TempoIncrease gets or sets the most recent non-zero value (or input) for Tempo Increase
func (m *Memory) TempoIncrease(input uint8) uint8 {
	return m.tempoIncrease.Coalesce(input)
}

// PortaDown gets or sets the most recent non-zero value (or input) for Portamento Down
func (m *Memory) PortaDown(input uint8) uint8 {
	return m.portaDown.Coalesce(input)
}

// PortaUp gets or sets the most recent non-zero value (or input) for Portamento Up
func (m *Memory) PortaUp(input uint8) uint8 {
	return m.portaUp.Coalesce(input)
}

// Tremolo gets or sets the most recent non-zero value (or input) for Tremolo
func (m *Memory) Tremolo(input uint8) (uint8, uint8) {
	return m.tremolo.CoalesceXY(input)
}

// Tremor gets or sets the most recent non-zero value (or input) for Tremor
func (m *Memory) Tremor(input uint8) (uint8, uint8) {
	return m.tremor.CoalesceXY(input)
}

// VolumeSlide gets or sets the most recent non-zero value (or input) for Volume Slide
func (m *Memory) VolumeSlide(input uint8) (uint8, uint8) {
	return m.volumeSlide.CoalesceXY(input)
}

// GlobalVolumeSlide gets or sets the most recent non-zero value (or input) for Global Volume Slide
func (m *Memory) GlobalVolumeSlide(input uint8) (uint8, uint8) {
	return m.globalVolumeSlide.CoalesceXY(input)
}

// FinePortaUp gets or sets the most recent non-zero value (or input) for Fine Portamento Up
func (m *Memory) FinePortaUp(input uint8) uint8 {
	return m.finePortaUp.Coalesce(input & 0x0F)
}

// FinePortaDown gets or sets the most recent non-zero value (or input) for Fine Portamento Down
func (m *Memory) FinePortaDown(input uint8) uint8 {
	return m.finePortaDown.Coalesce(input & 0x0F)
}

// FineVolumeSlideUp gets or sets the most recent non-zero value (or input) for Fine Volume Slide Up
func (m *Memory) FineVolumeSlideUp(input uint8) uint8 {
	return m.fineVolumeSlideUp.Coalesce(input & 0x0F)
}

// FineVolumeSlideDown gets or sets the most recent non-zero value (or input) for Fine Volume Slide Down
func (m *Memory) FineVolumeSlideDown(input uint8) uint8 {
	return m.fineVolumeSlideDown.Coalesce(input & 0x0F)
}

// ExtraFinePortaUp gets or sets the most recent non-zero value (or input) for Extra Fine Portamento Up
func (m *Memory) ExtraFinePortaUp(input uint8) uint8 {
	return m.extraFinePortaUp.Coalesce(input & 0x0F)
}

// ExtraFinePortaDown gets or sets the most recent non-zero value (or input) for Extra Fine Portamento Down
func (m *Memory) ExtraFinePortaDown(input uint8) uint8 {
	return m.extraFinePortaDown.Coalesce(input & 0x0F)
}

// TremorMem returns the Tremor object
func (m *Memory) TremorMem() *effect.Tremor {
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

// StartOrder is called when the first order's row at tick 0 is started
func (m *Memory) StartOrder() {
	if m.ResetMemoryAtStartOfOrder0 {
		m.portaToNote = 0
		m.vibrato = 0
		m.vibratoSpeed = 0
		m.sampleOffset = 0
		m.tempoDecrease = 0
		m.tempoIncrease = 0
		m.portaDown = 0
		m.portaUp = 0
		m.tremolo = 0
		m.tremor = 0
		m.volumeSlide = 0
		m.globalVolumeSlide = 0
		m.finePortaUp = 0
		m.finePortaDown = 0
		m.fineVolumeSlideUp = 0
		m.fineVolumeSlideDown = 0
		m.extraFinePortaUp = 0
		m.extraFinePortaDown = 0
	}
}
