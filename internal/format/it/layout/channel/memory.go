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
	volumeSlide        memory.UInt8 `usage:"Dxy"`
	portaDown          memory.UInt8 `usage:"Exx"`
	portaUp            memory.UInt8 `usage:"Fxx"`
	portaToNote        memory.UInt8 `usage:"Gxx"`
	vibrato            memory.UInt8 `usage:"Hxy"`
	tremor             memory.UInt8 `usage:"Ixy"`
	arpeggio           memory.UInt8 `usage:"Jxy"`
	channelVolumeSlide memory.UInt8 `usage:"Nxy"`
	sampleOffset       memory.UInt8 `usage:"Oxx"`
	panningSlide       memory.UInt8 `usage:"Pxy"`
	retrigVolumeSlide  memory.UInt8 `usage:"Qxy"`
	tremolo            memory.UInt8 `usage:"Rxy"`
	tempoDecrease      memory.UInt8 `usage:"T0x"`
	tempoIncrease      memory.UInt8 `usage:"T1x"`
	globalVolumeSlide  memory.UInt8 `usage:"Wxy"`
	panbrello          memory.UInt8 `usage:"Yxy"`
	volChanVolumeSlide memory.UInt8 `usage:"vDxy"`

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

	tremorMem           effect.Tremor
	vibratoOscillator   oscillator.Oscillator
	tremoloOscillator   oscillator.Oscillator
	panbrelloOscillator oscillator.Oscillator
	patternLoop         formatutil.PatternLoop
	HighOffset          int
}

// ResetOscillators resets the oscillators to defaults
func (m *Memory) ResetOscillators() {
	m.vibratoOscillator = oscillatorImpl.NewImpulseTrackerOscillator(4)
	m.tremoloOscillator = oscillatorImpl.NewImpulseTrackerOscillator(4)
	m.panbrelloOscillator = oscillatorImpl.NewImpulseTrackerOscillator(1)
}

// VolumeSlide gets or sets the most recent non-zero value (or input) for Volume Slide
func (m *Memory) VolumeSlide(input uint8) (uint8, uint8) {
	return m.volumeSlide.CoalesceXY(input)
}

// VolChanVolumeSlide gets or sets the most recent non-zero value (or input) for Volume Slide (from the volume channel)
func (m *Memory) VolChanVolumeSlide(input uint8) uint8 {
	return m.volChanVolumeSlide.Coalesce(input)
}

// PortaDown gets or sets the most recent non-zero value (or input) for Portamento Down
func (m *Memory) PortaDown(input uint8) uint8 {
	if m.EFGLinkMode {
		return m.portaToNote.Coalesce(input)
	}
	return m.portaDown.Coalesce(input)
}

// PortaUp gets or sets the most recent non-zero value (or input) for Portamento Up
func (m *Memory) PortaUp(input uint8) uint8 {
	if m.EFGLinkMode {
		return m.portaToNote.Coalesce(input)
	}
	return m.portaUp.Coalesce(input)
}

// PortaToNote gets or sets the most recent non-zero value (or input) for Portamento-to-note
func (m *Memory) PortaToNote(input uint8) uint8 {
	return m.portaToNote.Coalesce(input)
}

// Vibrato gets or sets the most recent non-zero value (or input) for Vibrato
func (m *Memory) Vibrato(input uint8) (uint8, uint8) {
	return m.vibrato.CoalesceXY(input)
}

// Tremor gets or sets the most recent non-zero value (or input) for Tremor
func (m *Memory) Tremor(input uint8) (uint8, uint8) {
	return m.tremor.CoalesceXY(input)
}

// Arpeggio gets or sets the most recent non-zero value (or input) for Arpeggio
func (m *Memory) Arpeggio(input uint8) (uint8, uint8) {
	return m.arpeggio.CoalesceXY(input)
}

// ChannelVolumeSlide gets or sets the most recent non-zero value (or input) for Channel Volume Slide
func (m *Memory) ChannelVolumeSlide(input uint8) (uint8, uint8) {
	return m.channelVolumeSlide.CoalesceXY(input)
}

// SampleOffset gets or sets the most recent non-zero value (or input) for Sample Offset
func (m *Memory) SampleOffset(input uint8) uint8 {
	return m.sampleOffset.Coalesce(input)
}

// PanningSlide gets or sets the most recent non-zero value (or input) for Panning Slide
func (m *Memory) PanningSlide(input uint8) uint8 {
	return m.panningSlide.Coalesce(input)
}

// RetrigVolumeSlide gets or sets the most recent non-zero value (or input) for Retrigger+VolumeSlide
func (m *Memory) RetrigVolumeSlide(input uint8) (uint8, uint8) {
	return m.retrigVolumeSlide.CoalesceXY(input)
}

// Tremolo gets or sets the most recent non-zero value (or input) for Tremolo
func (m *Memory) Tremolo(input uint8) (uint8, uint8) {
	return m.tremolo.CoalesceXY(input)
}

// TempoDecrease gets or sets the most recent non-zero value (or input) for Tempo Decrease
func (m *Memory) TempoDecrease(input uint8) uint8 {
	return m.tempoDecrease.Coalesce(input)
}

// TempoIncrease gets or sets the most recent non-zero value (or input) for Tempo Increase
func (m *Memory) TempoIncrease(input uint8) uint8 {
	return m.tempoIncrease.Coalesce(input)
}

// GlobalVolumeSlide gets or sets the most recent non-zero value (or input) for Global Volume Slide
func (m *Memory) GlobalVolumeSlide(input uint8) (uint8, uint8) {
	return m.globalVolumeSlide.CoalesceXY(input)
}

// Panbrello gets or sets the most recent non-zero value (or input) for Panbrello
func (m *Memory) Panbrello(input uint8) uint8 {
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
