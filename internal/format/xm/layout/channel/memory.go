package channel

import (
	"github.com/gotracker/voice/oscillator"

	"github.com/gotracker/gotracker/internal/format/internal/effect"
	"github.com/gotracker/gotracker/internal/format/internal/memory"
	formatutil "github.com/gotracker/gotracker/internal/format/internal/util"
	oscillatorImpl "github.com/gotracker/gotracker/internal/oscillator"
)

// Memory is the storage object for custom effect/effect values
type Memory struct {
	portaToNote         memory.Value[DataEffect]
	vibrato             memory.Value[DataEffect]
	vibratoSpeed        memory.Value[DataEffect]
	sampleOffset        memory.Value[DataEffect]
	tempoDecrease       memory.Value[DataEffect]
	tempoIncrease       memory.Value[DataEffect]
	portaDown           memory.Value[DataEffect]
	portaUp             memory.Value[DataEffect]
	tremolo             memory.Value[DataEffect]
	tremor              memory.Value[DataEffect]
	volumeSlide         memory.Value[DataEffect]
	globalVolumeSlide   memory.Value[DataEffect]
	finePortaUp         memory.Value[DataEffect]
	finePortaDown       memory.Value[DataEffect]
	fineVolumeSlideUp   memory.Value[DataEffect]
	fineVolumeSlideDown memory.Value[DataEffect]
	extraFinePortaUp    memory.Value[DataEffect]
	extraFinePortaDown  memory.Value[DataEffect]

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
func (m *Memory) PortaToNote(input DataEffect) DataEffect {
	return m.portaToNote.Coalesce(input)
}

// Vibrato gets or sets the most recent non-zero value (or input) for Vibrato
func (m *Memory) Vibrato(input DataEffect) (DataEffect, DataEffect) {
	return m.vibrato.CoalesceXY(input)
}

// VibratoSpeed gets or sets the most recent non-zero value (or input) for Vibrato Speed
func (m *Memory) VibratoSpeed(input DataEffect) DataEffect {
	return m.vibratoSpeed.Coalesce(input)
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

// PortaDown gets or sets the most recent non-zero value (or input) for Portamento Down
func (m *Memory) PortaDown(input DataEffect) DataEffect {
	return m.portaDown.Coalesce(input)
}

// PortaUp gets or sets the most recent non-zero value (or input) for Portamento Up
func (m *Memory) PortaUp(input DataEffect) DataEffect {
	return m.portaUp.Coalesce(input)
}

// Tremolo gets or sets the most recent non-zero value (or input) for Tremolo
func (m *Memory) Tremolo(input DataEffect) (DataEffect, DataEffect) {
	return m.tremolo.CoalesceXY(input)
}

// Tremor gets or sets the most recent non-zero value (or input) for Tremor
func (m *Memory) Tremor(input DataEffect) (DataEffect, DataEffect) {
	return m.tremor.CoalesceXY(input)
}

// VolumeSlide gets or sets the most recent non-zero value (or input) for Volume Slide
func (m *Memory) VolumeSlide(input DataEffect) (DataEffect, DataEffect) {
	return m.volumeSlide.CoalesceXY(input)
}

// GlobalVolumeSlide gets or sets the most recent non-zero value (or input) for Global Volume Slide
func (m *Memory) GlobalVolumeSlide(input DataEffect) (DataEffect, DataEffect) {
	return m.globalVolumeSlide.CoalesceXY(input)
}

// FinePortaUp gets or sets the most recent non-zero value (or input) for Fine Portamento Up
func (m *Memory) FinePortaUp(input DataEffect) DataEffect {
	return m.finePortaUp.Coalesce(input & 0x0F)
}

// FinePortaDown gets or sets the most recent non-zero value (or input) for Fine Portamento Down
func (m *Memory) FinePortaDown(input DataEffect) DataEffect {
	return m.finePortaDown.Coalesce(input & 0x0F)
}

// FineVolumeSlideUp gets or sets the most recent non-zero value (or input) for Fine Volume Slide Up
func (m *Memory) FineVolumeSlideUp(input DataEffect) DataEffect {
	return m.fineVolumeSlideUp.Coalesce(input & 0x0F)
}

// FineVolumeSlideDown gets or sets the most recent non-zero value (or input) for Fine Volume Slide Down
func (m *Memory) FineVolumeSlideDown(input DataEffect) DataEffect {
	return m.fineVolumeSlideDown.Coalesce(input & 0x0F)
}

// ExtraFinePortaUp gets or sets the most recent non-zero value (or input) for Extra Fine Portamento Up
func (m *Memory) ExtraFinePortaUp(input DataEffect) DataEffect {
	return m.extraFinePortaUp.Coalesce(input & 0x0F)
}

// ExtraFinePortaDown gets or sets the most recent non-zero value (or input) for Extra Fine Portamento Down
func (m *Memory) ExtraFinePortaDown(input DataEffect) DataEffect {
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
		m.portaToNote.Reset()
		m.vibrato.Reset()
		m.vibratoSpeed.Reset()
		m.sampleOffset.Reset()
		m.tempoDecrease.Reset()
		m.tempoIncrease.Reset()
		m.portaDown.Reset()
		m.portaUp.Reset()
		m.tremolo.Reset()
		m.tremor.Reset()
		m.volumeSlide.Reset()
		m.globalVolumeSlide.Reset()
		m.finePortaUp.Reset()
		m.finePortaDown.Reset()
		m.fineVolumeSlideUp.Reset()
		m.fineVolumeSlideDown.Reset()
		m.extraFinePortaUp.Reset()
		m.extraFinePortaDown.Reset()
	}
}
