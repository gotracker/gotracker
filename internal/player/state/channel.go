package state

import (
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/oscillator"
	"gotracker/internal/player/volume"
	"math"
)

type commandFunc func(int, *ChannelState, int, bool)

// ChannelState is the state of a single channel
type ChannelState struct {
	intf.Channel
	intf.SharedMemory
	Instrument   intf.Instrument
	Pos          float32
	Period       float32
	StoredVolume volume.Volume
	ActiveVolume volume.Volume
	Pan          uint8

	Command      commandFunc
	ActiveEffect intf.Effect

	DisplayNote note.Note
	DisplayInst uint8

	TargetPeriod      float32
	TargetPos         float32
	TargetInst        intf.Instrument
	PortaTargetPeriod float32
	NotePlayTick      int
	NoteSemitone      uint8
	RetriggerCount    uint8
	TremorOn          bool
	TremorTime        int
	VibratoDelta      float32
	memory            Memory
	effectLastNonZero uint8
	Cmd               intf.ChannelData
	freezePlayback    bool
	LastGlobalVolume  volume.Volume
	VibratoOscillator oscillator.Oscillator
	TremoloOscillator oscillator.Oscillator
	TargetC2Spd       uint16
}

// SetStoredVolume sets the stored volume value for the channel
// this also modifies the active volume
// and stores the active global volume value (which doesn't always get set on channels immediately)
func (cs *ChannelState) SetStoredVolume(vol volume.Volume, ss *Song) {
	cs.StoredVolume = vol
	cs.ActiveVolume = vol
	cs.LastGlobalVolume = ss.GlobalVolume
}

// FreezePlayback suspends mixer progression on the channel
func (cs *ChannelState) FreezePlayback() {
	cs.freezePlayback = true
}

// UnfreezePlayback resumes mixer progression on the channel
func (cs *ChannelState) UnfreezePlayback() {
	cs.freezePlayback = false
}

// PlaybackFrozen returns true if the mixer progression for the channel is suspended
func (cs ChannelState) PlaybackFrozen() bool {
	return cs.freezePlayback
}

// SetEffectSharedMemoryIfNonZero stores the `input` value into memory if it is non-zero
func (cs *ChannelState) SetEffectSharedMemoryIfNonZero(input uint8) {
	if input != 0 {
		cs.effectLastNonZero = input
	}
}

// GetEffectSharedMemory returns the last non-zero value (if one exists) or the input value
func (cs *ChannelState) GetEffectSharedMemory(input uint8) uint8 {
	if input == 0 {
		return cs.effectLastNonZero
	}
	return input
}

// ResetRetriggerCount sets the retrigger count to 0
func (cs *ChannelState) ResetRetriggerCount() {
	cs.RetriggerCount = 0
}

// GetMemory returns the interface to the custom effect memory module
func (cs *ChannelState) GetMemory() intf.Memory {
	return &cs.memory
}

// GetActiveVolume returns the current active volume on the channel
func (cs *ChannelState) GetActiveVolume() volume.Volume {
	return cs.ActiveVolume
}

// SetActiveVolume sets the active volume on the channel
func (cs *ChannelState) SetActiveVolume(vol volume.Volume) {
	cs.ActiveVolume = vol
}

// GetData returns the interface to the current channel song pattern data
func (cs *ChannelState) GetData() intf.ChannelData {
	return cs.Cmd
}

// GetPortaTargetPeriod returns the current target portamento (to note) sampler period
func (cs *ChannelState) GetPortaTargetPeriod() float32 {
	return cs.PortaTargetPeriod
}

// SetPortaTargetPeriod sets the current target portamento (to note) sampler period
func (cs *ChannelState) SetPortaTargetPeriod(period float32) {
	cs.PortaTargetPeriod = period
}

// GetTargetPeriod returns the soon-to-be-committed sampler period (when the note retriggers)
func (cs *ChannelState) GetTargetPeriod() float32 {
	return cs.TargetPeriod
}

// SetTargetPeriod sets the soon-to-be-committed sampler period (when the note retriggers)
func (cs *ChannelState) SetTargetPeriod(period float32) {
	cs.TargetPeriod = period
}

// SetVibratoDelta sets the vibrato (ephemeral) delta sampler period
func (cs *ChannelState) SetVibratoDelta(delta float32) {
	cs.VibratoDelta = delta
}

// GetVibratoOscillator returns the oscillator object for the Vibrato LFO
func (cs *ChannelState) GetVibratoOscillator() *oscillator.Oscillator {
	return &cs.VibratoOscillator
}

// GetTremoloOscillator returns the oscillator object for the Tremolo LFO
func (cs *ChannelState) GetTremoloOscillator() *oscillator.Oscillator {
	return &cs.TremoloOscillator
}

// GetTremorOn returns true if the tremor setting is enabled
func (cs *ChannelState) GetTremorOn() bool {
	return cs.TremorOn
}

// SetTremorOn sets the current tremor enablement setting
func (cs *ChannelState) SetTremorOn(on bool) {
	cs.TremorOn = on
}

// GetTremorTime returns the tick the tremor should be enabled (or disabled) until
func (cs *ChannelState) GetTremorTime() int {
	return cs.TremorTime
}

// SetTremorTime sets the tick that the tremor should be enabled (or disabled) until
func (cs *ChannelState) SetTremorTime(time int) {
	cs.TremorTime = time
}

// GetInstrument returns the interface to the active instrument
func (cs *ChannelState) GetInstrument() intf.Instrument {
	return cs.Instrument
}

// GetTargetInst returns the interface to the soon-to-be-committed active instrument (when the note retriggers)
func (cs *ChannelState) GetTargetInst() intf.Instrument {
	return cs.TargetInst
}

// SetTargetInst sets the soon-to-be-committed active instrument (when the note retriggers)
func (cs *ChannelState) SetTargetInst(inst intf.Instrument) {
	cs.TargetInst = inst
}

// GetNoteSemitone returns the note semitone for the channel
func (cs *ChannelState) GetNoteSemitone() uint8 {
	return cs.NoteSemitone
}

// SetTargetPos returns the soon-to-be-committed sample position of the instrument
func (cs *ChannelState) SetTargetPos(pos float32) {
	cs.TargetPos = pos
}

// GetPeriod returns the current sampler period of the active instrument
func (cs *ChannelState) GetPeriod() float32 {
	return cs.Period
}

// SetPeriod sets the current sampler period of the active instrument
func (cs *ChannelState) SetPeriod(period float32) {
	cs.Period = float32(math.Max(float64(period), 0))
}

// GetPos returns the sample position of the active instrument
func (cs *ChannelState) GetPos() float32 {
	return cs.Pos
}

// SetPos sets the sample position of the active instrument
func (cs *ChannelState) SetPos(pos float32) {
	cs.Pos = pos
}

// SetNotePlayTick sets the tick on which the note will retrigger
func (cs *ChannelState) SetNotePlayTick(tick int) {
	cs.NotePlayTick = tick
}

// GetRetriggerCount returns the current count of the retrigger counter
func (cs *ChannelState) GetRetriggerCount() uint8 {
	return cs.RetriggerCount
}

// SetRetriggerCount sets the current count of the retrigger counter
func (cs *ChannelState) SetRetriggerCount(cnt uint8) {
	cs.RetriggerCount = cnt
}

// SetPan sets the active panning value of the channel (0 = full left, 15 = full right)
func (cs *ChannelState) SetPan(pan uint8) {
	cs.Pan = pan
}
