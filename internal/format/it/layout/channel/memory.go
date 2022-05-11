package channel

import (
	"github.com/gotracker/voice/oscillator"

	"github.com/gotracker/gotracker/internal/format/internal/effect"
	"github.com/gotracker/gotracker/internal/format/internal/memory"
	formatutil "github.com/gotracker/gotracker/internal/format/internal/util"
	oscillatorImpl "github.com/gotracker/gotracker/internal/oscillator"
)

type SharedMemory struct {
	// LinearFreqSlides is true if linear frequency slides are enabled (false = amiga-style period-based slides)
	LinearFreqSlides bool
	// OldEffectMode performs somewhat different operations for some effects:
	// On:
	//  - Vibrato does not operate on tick 0 and has double depth
	//  - Sample Offset will ignore the command if it would exceed the length
	// Off:
	//  - Vibrato is updated every frame
	//  - Sample Offset will set the offset to the end of the sample if it would exceed the length
	OldEffectMode bool
	// EFGLinkMode will make effects Exx, Fxx, and Gxx share the same memory
	EFGLinkMode bool
	// ResetMemoryAtStartOfOrder0 if true will reset the memory registers when the first tick of the first row of the first order pattern plays
	ResetMemoryAtStartOfOrder0 bool
}

// Memory is the storage object for custom effect/effect values
type Memory struct {
	volumeSlide        memory.Value[DataEffect] `usage:"Dxy"`
	portaDown          memory.Value[DataEffect] `usage:"Exx"`
	portaUp            memory.Value[DataEffect] `usage:"Fxx"`
	portaToNote        memory.Value[DataEffect] `usage:"Gxx"`
	vibrato            memory.Value[DataEffect] `usage:"Hxy"`
	tremor             memory.Value[DataEffect] `usage:"Ixy"`
	arpeggio           memory.Value[DataEffect] `usage:"Jxy"`
	channelVolumeSlide memory.Value[DataEffect] `usage:"Nxy"`
	sampleOffset       memory.Value[DataEffect] `usage:"Oxx"`
	panningSlide       memory.Value[DataEffect] `usage:"Pxy"`
	retrigVolumeSlide  memory.Value[DataEffect] `usage:"Qxy"`
	tremolo            memory.Value[DataEffect] `usage:"Rxy"`
	tempoDecrease      memory.Value[DataEffect] `usage:"T0x"`
	tempoIncrease      memory.Value[DataEffect] `usage:"T1x"`
	globalVolumeSlide  memory.Value[DataEffect] `usage:"Wxy"`
	panbrello          memory.Value[DataEffect] `usage:"Yxy"`
	volChanVolumeSlide memory.Value[DataEffect] `usage:"vDxy"`

	tremorMem           effect.Tremor
	vibratoOscillator   oscillator.Oscillator
	tremoloOscillator   oscillator.Oscillator
	panbrelloOscillator oscillator.Oscillator
	patternLoop         formatutil.PatternLoop
	HighOffset          int

	Shared *SharedMemory
}

// ResetOscillators resets the oscillators to defaults
func (m *Memory) ResetOscillators() {
	m.vibratoOscillator = oscillatorImpl.NewImpulseTrackerOscillator(4)
	m.tremoloOscillator = oscillatorImpl.NewImpulseTrackerOscillator(4)
	m.panbrelloOscillator = oscillatorImpl.NewImpulseTrackerOscillator(1)
}

// VolumeSlide gets or sets the most recent non-zero value (or input) for Volume Slide
func (m *Memory) VolumeSlide(input DataEffect) (DataEffect, DataEffect) {
	return m.volumeSlide.CoalesceXY(input)
}

// VolChanVolumeSlide gets or sets the most recent non-zero value (or input) for Volume Slide (from the volume channel)
func (m *Memory) VolChanVolumeSlide(input DataEffect) DataEffect {
	return m.volChanVolumeSlide.Coalesce(input)
}

// PortaDown gets or sets the most recent non-zero value (or input) for Portamento Down
func (m *Memory) PortaDown(input DataEffect) DataEffect {
	if m.Shared.EFGLinkMode {
		return m.portaToNote.Coalesce(input)
	}
	return m.portaDown.Coalesce(input)
}

// PortaUp gets or sets the most recent non-zero value (or input) for Portamento Up
func (m *Memory) PortaUp(input DataEffect) DataEffect {
	if m.Shared.EFGLinkMode {
		return m.portaToNote.Coalesce(input)
	}
	return m.portaUp.Coalesce(input)
}

// PortaToNote gets or sets the most recent non-zero value (or input) for Portamento-to-note
func (m *Memory) PortaToNote(input DataEffect) DataEffect {
	return m.portaToNote.Coalesce(input)
}

// Vibrato gets or sets the most recent non-zero value (or input) for Vibrato
func (m *Memory) Vibrato(input DataEffect) (DataEffect, DataEffect) {
	return m.vibrato.CoalesceXY(input)
}

// Tremor gets or sets the most recent non-zero value (or input) for Tremor
func (m *Memory) Tremor(input DataEffect) (DataEffect, DataEffect) {
	return m.tremor.CoalesceXY(input)
}

// Arpeggio gets or sets the most recent non-zero value (or input) for Arpeggio
func (m *Memory) Arpeggio(input DataEffect) (DataEffect, DataEffect) {
	return m.arpeggio.CoalesceXY(input)
}

// ChannelVolumeSlide gets or sets the most recent non-zero value (or input) for Channel Volume Slide
func (m *Memory) ChannelVolumeSlide(input DataEffect) (DataEffect, DataEffect) {
	return m.channelVolumeSlide.CoalesceXY(input)
}

// SampleOffset gets or sets the most recent non-zero value (or input) for Sample Offset
func (m *Memory) SampleOffset(input DataEffect) DataEffect {
	return m.sampleOffset.Coalesce(input)
}

// PanningSlide gets or sets the most recent non-zero value (or input) for Panning Slide
func (m *Memory) PanningSlide(input DataEffect) DataEffect {
	return m.panningSlide.Coalesce(input)
}

// RetrigVolumeSlide gets or sets the most recent non-zero value (or input) for Retrigger+VolumeSlide
func (m *Memory) RetrigVolumeSlide(input DataEffect) (DataEffect, DataEffect) {
	return m.retrigVolumeSlide.CoalesceXY(input)
}

// Tremolo gets or sets the most recent non-zero value (or input) for Tremolo
func (m *Memory) Tremolo(input DataEffect) (DataEffect, DataEffect) {
	return m.tremolo.CoalesceXY(input)
}

// TempoDecrease gets or sets the most recent non-zero value (or input) for Tempo Decrease
func (m *Memory) TempoDecrease(input DataEffect) DataEffect {
	return m.tempoDecrease.Coalesce(input)
}

// TempoIncrease gets or sets the most recent non-zero value (or input) for Tempo Increase
func (m *Memory) TempoIncrease(input DataEffect) DataEffect {
	return m.tempoIncrease.Coalesce(input)
}

// GlobalVolumeSlide gets or sets the most recent non-zero value (or input) for Global Volume Slide
func (m *Memory) GlobalVolumeSlide(input DataEffect) (DataEffect, DataEffect) {
	return m.globalVolumeSlide.CoalesceXY(input)
}

// Panbrello gets or sets the most recent non-zero value (or input) for Panbrello
func (m *Memory) Panbrello(input DataEffect) DataEffect {
	return m.panbrello.Coalesce(input)
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

// PanbrelloOscillator returns the Panbrello oscillator object
func (m *Memory) PanbrelloOscillator() oscillator.Oscillator {
	return m.panbrelloOscillator
}

// Retrigger runs certain operations when a note is retriggered
func (m *Memory) Retrigger() {
	for _, osc := range []oscillator.Oscillator{m.VibratoOscillator(), m.TremoloOscillator(), m.PanbrelloOscillator()} {
		osc.Reset()
	}
}

// GetPatternLoop returns the pattern loop object from the memory
func (m *Memory) GetPatternLoop() *formatutil.PatternLoop {
	return &m.patternLoop
}

// StartOrder is called when the first order's row at tick 0 is started
func (m *Memory) StartOrder() {
	if m.Shared.ResetMemoryAtStartOfOrder0 {
		m.volumeSlide.Reset()
		m.portaDown.Reset()
		m.portaUp.Reset()
		m.portaToNote.Reset()
		m.vibrato.Reset()
		m.tremor.Reset()
		m.arpeggio.Reset()
		m.channelVolumeSlide.Reset()
		m.sampleOffset.Reset()
		m.panningSlide.Reset()
		m.retrigVolumeSlide.Reset()
		m.tremolo.Reset()
		m.tempoDecrease.Reset()
		m.tempoIncrease.Reset()
		m.globalVolumeSlide.Reset()
		m.panbrello.Reset()
		m.volChanVolumeSlide.Reset()
	}
}
