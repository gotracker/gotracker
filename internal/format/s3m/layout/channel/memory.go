package channel

import (
	"github.com/gotracker/voice/oscillator"

	"gotracker/internal/format/internal/memory"
	formatutil "gotracker/internal/format/internal/util"
	oscillatorImpl "gotracker/internal/oscillator"
)

// Memory is the storage object for custom effect/command values
type Memory struct {
	portaToNote   memory.UInt8
	vibratoSpeed  memory.UInt8
	vibratoDepth  memory.UInt8
	tremoloSpeed  memory.UInt8
	tremoloDepth  memory.UInt8
	sampleOffset  memory.UInt8
	tempoDecrease memory.UInt8
	tempoIncrease memory.UInt8
	lastNonZero   memory.UInt8

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

// PortaToNote gets or sets the most recent non-zero value (or input) for Portamento-to-note
func (m *Memory) PortaToNote(input uint8) uint8 {
	return m.portaToNote.Coalesce(input)
}

// Vibrato gets or sets the most recent non-zero value (or input) for Vibrato
func (m *Memory) Vibrato(input uint8) (uint8, uint8) {
	// vibrato is unusual, because each nibble is treated uniquely
	vx := m.vibratoSpeed.Coalesce(input >> 4)
	vy := m.vibratoDepth.Coalesce(input & 0x0f)
	return vx, vy
}

// Tremolo gets or sets the most recent non-zero value (or input) for Vibrato
func (m *Memory) Tremolo(input uint8) (uint8, uint8) {
	// tremolo is unusual, because each nibble is treated uniquely
	vx := m.tremoloSpeed.Coalesce(input >> 4)
	vy := m.tremoloDepth.Coalesce(input & 0x0f)
	return vx, vy
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

// LastNonZero gets or sets the most recent non-zero value (or input)
func (m *Memory) LastNonZero(input uint8) uint8 {
	return m.lastNonZero.Coalesce(input)
}

// LastNonZero gets or sets the most recent non-zero value (or input)
func (m *Memory) LastNonZeroXY(input uint8) (uint8, uint8) {
	xy := m.LastNonZero(input)
	return xy >> 4, xy & 0x0f
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
