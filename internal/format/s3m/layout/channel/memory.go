package channel

import (
	"github.com/gotracker/voice/oscillator"

	formatutil "gotracker/internal/format/internal/util"
	oscillatorImpl "gotracker/internal/oscillator"
)

// Memory is the storage object for custom effect/command values
type Memory struct {
	portaToNote   uint8
	vibratoSpeed  uint8
	vibratoDepth  uint8
	tremoloSpeed  uint8
	tremoloDepth  uint8
	sampleOffset  uint8
	tempoDecrease uint8
	tempoIncrease uint8
	lastNonZero   uint8

	VolSlideEveryFrame  bool
	LowPassFilterEnable bool

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
func (m *Memory) Vibrato(input uint8) (uint8, uint8) {
	// vibrato is unusual, because each nibble is treated uniquely
	vx := m.getEffectMemory(input>>4, &m.vibratoSpeed)
	vy := m.getEffectMemory(input&0x0f, &m.vibratoDepth)
	return vx, vy
}

// Tremolo gets or sets the most recent non-zero value (or input) for Vibrato
func (m *Memory) Tremolo(input uint8) (uint8, uint8) {
	// tremolo is unusual, because each nibble is treated uniquely
	vx := m.getEffectMemory(input>>4, &m.tremoloSpeed)
	vy := m.getEffectMemory(input&0x0f, &m.tremoloDepth)
	return vx, vy
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

// LastNonZero gets or sets the most recent non-zero value (or input)
func (m *Memory) LastNonZero(input uint8) uint8 {
	return m.getEffectMemory(input, &m.lastNonZero)
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
