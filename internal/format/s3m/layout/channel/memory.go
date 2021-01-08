package channel

import "gotracker/internal/player/intf"

// Memory is the storage object for custom effect/command values
type Memory struct {
	portaToNote      uint8
	vibrato          uint8
	sampleOffset     uint8
	tempoDecrease    uint8
	tempoIncrease    uint8
	lastNonZero      uint8
	patternLoopStart intf.RowIdx

	tremorMem         Tremor
	vibratoOscillator Oscillator
	tremoloOscillator Oscillator
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
func (m *Memory) TremorMem() *Tremor {
	return &m.tremorMem
}

// VibratoOscillator returns the Vibrato oscillator object
func (m *Memory) VibratoOscillator() *Oscillator {
	return &m.vibratoOscillator
}

// TremoloOscillator returns the Tremolo oscillator object
func (m *Memory) TremoloOscillator() *Oscillator {
	return &m.tremoloOscillator
}

// Retrigger runs certain operations when a note is retriggered
func (m *Memory) Retrigger() {
	m.vibratoOscillator.Pos = 0
	m.tremoloOscillator.Pos = 0
}

// SetPatternLoopStart sets the pattern loop start location in memory
func (m *Memory) SetPatternLoopStart(row intf.RowIdx) {
	m.patternLoopStart = row
}

// GetPatternLoopStart gets the pattern loop start location from memory
func (m *Memory) GetPatternLoopStart() intf.RowIdx {
	return m.patternLoopStart
}
