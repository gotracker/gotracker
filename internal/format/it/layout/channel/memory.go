package channel

import (
	formatutil "gotracker/internal/format/internal/util"
	"gotracker/internal/oscillator"
)

// Memory is the storage object for custom effect/effect values
type Memory struct {
	volumeSlide        uint8 `usage:"Dxy"`
	portaDown          uint8 `usage:"Exx"`
	portaUp            uint8 `usage:"Fxx"`
	portaToNote        uint8 `usage:"Gxx"`
	vibrato            uint8 `usage:"Hxy"`
	tremor             uint8 `usage:"Ixy"`
	arpeggio           uint8 `usage:"Jxy"`
	channelVolumeSlide uint8 `usage:"Nxy"`
	sampleOffset       uint8 `usage:"Oxx"`
	panningSlide       uint8 `usage:"Pxy"`
	retrigVolumeSlide  uint8 `usage:"Qxy"`
	tremolo            uint8 `usage:"Rxy"`
	tempoDecrease      uint8 `usage:"T0x"`
	tempoIncrease      uint8 `usage:"T1x"`
	globalVolumeSlide  uint8 `usage:"Wxy"`
	panbrello          uint8 `usage:"Yxy"`
	volChanVolumeSlide uint8 `usage:"vDxy"`

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

	tremorMem           formatutil.Tremor
	vibratoOscillator   oscillator.Oscillator
	tremoloOscillator   oscillator.Oscillator
	panbrelloOscillator oscillator.Oscillator
	patternLoop         formatutil.PatternLoop
	HighOffset          int
}

// ResetOscillators resets the oscillators to defaults
func (m *Memory) ResetOscillators() {
	m.vibratoOscillator = oscillator.NewImpulseTrackerOscillator(4)
	m.tremoloOscillator = oscillator.NewImpulseTrackerOscillator(4)
	m.panbrelloOscillator = oscillator.NewImpulseTrackerOscillator(1)
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

// VolumeSlide gets or sets the most recent non-zero value (or input) for Volume Slide
func (m *Memory) VolumeSlide(input uint8) (uint8, uint8) {
	xy := m.getEffectMemory(input, &m.volumeSlide)
	return xy >> 4, xy & 0x0f
}

// VolChanVolumeSlide gets or sets the most recent non-zero value (or input) for Volume Slide (from the volume channel)
func (m *Memory) VolChanVolumeSlide(input uint8) uint8 {
	return m.getEffectMemory(input, &m.volChanVolumeSlide)
}

// PortaDown gets or sets the most recent non-zero value (or input) for Portamento Down
func (m *Memory) PortaDown(input uint8) uint8 {
	if m.EFGLinkMode {
		return m.getEffectMemory(input, &m.portaToNote)
	}
	return m.getEffectMemory(input, &m.portaDown)
}

// PortaUp gets or sets the most recent non-zero value (or input) for Portamento Up
func (m *Memory) PortaUp(input uint8) uint8 {
	if m.EFGLinkMode {
		return m.getEffectMemory(input, &m.portaToNote)
	}
	return m.getEffectMemory(input, &m.portaUp)
}

// PortaToNote gets or sets the most recent non-zero value (or input) for Portamento-to-note
func (m *Memory) PortaToNote(input uint8) uint8 {
	return m.getEffectMemory(input, &m.portaToNote)
}

// Vibrato gets or sets the most recent non-zero value (or input) for Vibrato
func (m *Memory) Vibrato(input uint8) (uint8, uint8) {
	xy := m.getEffectMemory(input, &m.vibrato)
	return xy >> 4, xy & 0x0f
}

// Tremor gets or sets the most recent non-zero value (or input) for Tremor
func (m *Memory) Tremor(input uint8) (uint8, uint8) {
	xy := m.getEffectMemory(input, &m.tremor)
	return xy >> 4, xy & 0x0f
}

// Arpeggio gets or sets the most recent non-zero value (or input) for Arpeggio
func (m *Memory) Arpeggio(input uint8) uint8 {
	return m.getEffectMemory(input, &m.arpeggio)
}

// ChannelVolumeSlide gets or sets the most recent non-zero value (or input) for Channel Volume Slide
func (m *Memory) ChannelVolumeSlide(input uint8) (uint8, uint8) {
	xy := m.getEffectMemory(input, &m.channelVolumeSlide)
	return xy >> 4, xy & 0x0f
}

// SampleOffset gets or sets the most recent non-zero value (or input) for Sample Offset
func (m *Memory) SampleOffset(input uint8) uint8 {
	return m.getEffectMemory(input, &m.sampleOffset)
}

// PanningSlide gets or sets the most recent non-zero value (or input) for Panning Slide
func (m *Memory) PanningSlide(input uint8) uint8 {
	return m.getEffectMemory(input, &m.panningSlide)
}

// RetrigVolumeSlide gets or sets the most recent non-zero value (or input) for Retrigger+VolumeSlide
func (m *Memory) RetrigVolumeSlide(input uint8) (uint8, uint8) {
	xy := m.getEffectMemory(input, &m.retrigVolumeSlide)
	return xy >> 4, xy & 0x0f
}

// Tremolo gets or sets the most recent non-zero value (or input) for Tremolo
func (m *Memory) Tremolo(input uint8) (uint8, uint8) {
	xy := m.getEffectMemory(input, &m.tremolo)
	return xy >> 4, xy & 0x0f
}

// TempoDecrease gets or sets the most recent non-zero value (or input) for Tempo Decrease
func (m *Memory) TempoDecrease(input uint8) uint8 {
	return m.getEffectMemory(input, &m.tempoDecrease)
}

// TempoIncrease gets or sets the most recent non-zero value (or input) for Tempo Increase
func (m *Memory) TempoIncrease(input uint8) uint8 {
	return m.getEffectMemory(input, &m.tempoIncrease)
}

// GlobalVolumeSlide gets or sets the most recent non-zero value (or input) for Global Volume Slide
func (m *Memory) GlobalVolumeSlide(input uint8) uint8 {
	return m.getEffectMemory(input, &m.globalVolumeSlide)
}

// Panbrello gets or sets the most recent non-zero value (or input) for Panbrello
func (m *Memory) Panbrello(input uint8) uint8 {
	return m.getEffectMemory(input, &m.panbrello)
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
