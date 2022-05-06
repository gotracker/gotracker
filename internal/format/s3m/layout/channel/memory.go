package channel

import (
	"github.com/gotracker/voice/oscillator"

	"github.com/gotracker/gotracker/internal/format/internal/effect"
	"github.com/gotracker/gotracker/internal/format/internal/memory"
	formatutil "github.com/gotracker/gotracker/internal/format/internal/util"
	oscillatorImpl "github.com/gotracker/gotracker/internal/oscillator"
)

// Memory is the storage object for custom effect/command values
type Memory struct {
	portaToNote   memory.Value[DataEffect]
	vibratoSpeed  memory.Value[DataEffect]
	vibratoDepth  memory.Value[DataEffect]
	tremoloSpeed  memory.Value[DataEffect]
	tremoloDepth  memory.Value[DataEffect]
	sampleOffset  memory.Value[DataEffect]
	tempoDecrease memory.Value[DataEffect]
	tempoIncrease memory.Value[DataEffect]
	lastNonZero   memory.Value[DataEffect]

	VolSlideEveryFrame  bool
	LowPassFilterEnable bool
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
func (m *Memory) PortaToNote(input DataEffect) DataEffect {
	return m.portaToNote.Coalesce(input)
}

// Vibrato gets or sets the most recent non-zero value (or input) for Vibrato
func (m *Memory) Vibrato(input DataEffect) (DataEffect, DataEffect) {
	// vibrato is unusual, because each nibble is treated uniquely
	vx := m.vibratoSpeed.Coalesce(input >> 4)
	vy := m.vibratoDepth.Coalesce(input & 0x0f)
	return vx, vy
}

// Tremolo gets or sets the most recent non-zero value (or input) for Vibrato
func (m *Memory) Tremolo(input DataEffect) (DataEffect, DataEffect) {
	// tremolo is unusual, because each nibble is treated uniquely
	vx := m.tremoloSpeed.Coalesce(input >> 4)
	vy := m.tremoloDepth.Coalesce(input & 0x0f)
	return vx, vy
}

// SampleOffset gets or sets the most recent non-zero value (or input) for Sample Offset
func (m *Memory) SampleOffset(input DataEffect) DataEffect {
	return m.sampleOffset.Coalesce(input)
}

// TempoDecrease gets or sets the most recent non-zero value (or input) for Tempo Decrease
func (m *Memory) TempoDecrease(input DataEffect) DataEffect {
	return m.tempoDecrease.Coalesce(input)
}

// TempoIncrease gets or sets the most recent non-zero value (or input) for Tempo Increase
func (m *Memory) TempoIncrease(input DataEffect) DataEffect {
	return m.tempoIncrease.Coalesce(input)
}

// LastNonZero gets or sets the most recent non-zero value (or input)
func (m *Memory) LastNonZero(input DataEffect) DataEffect {
	return m.lastNonZero.Coalesce(input)
}

// LastNonZero gets or sets the most recent non-zero value (or input)
func (m *Memory) LastNonZeroXY(input DataEffect) (DataEffect, DataEffect) {
	xy := m.LastNonZero(input)
	return xy >> 4, xy & 0x0f
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
		m.portaToNote.Reset()
		m.vibratoSpeed.Reset()
		m.vibratoDepth.Reset()
		m.tremoloSpeed.Reset()
		m.tremoloDepth.Reset()
		m.sampleOffset.Reset()
		m.tempoDecrease.Reset()
		m.tempoIncrease.Reset()
		m.lastNonZero.Reset()
	}
}
