package state

import (
	"gotracker/internal/player/channel"
	"gotracker/internal/player/instrument"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/oscillator"
	"gotracker/internal/s3m/volume"
	"math"
)

type CommandFunc func(int, *ChannelState, int, bool)

type ChannelState struct {
	intf.Channel
	intf.SharedMemory
	Instrument   instrument.InstrumentInfo
	Pos          float32
	Period       float32
	StoredVolume uint8
	ActiveVolume uint8
	Pan          uint8

	Command      CommandFunc
	ActiveEffect intf.Effect

	DisplayNote note.Note
	DisplayInst uint8

	TargetPeriod      float32
	TargetPos         float32
	TargetInst        instrument.InstrumentInfo
	PortaTargetPeriod float32
	NotePlayTick      int
	NoteSemitone      uint8
	RetriggerCount    uint8
	TremorOn          bool
	TremorTime        int
	VibratoDelta      float32
	memory            Memory
	effectLastNonZero uint8
	Cmd               *channel.Data
	freezePlayback    bool
	LastGlobalVolume  volume.Volume
	VibratoOscillator oscillator.Oscillator
	TremoloOscillator oscillator.Oscillator
	TargetC2Spd       uint16
}

func (cs *ChannelState) SetStoredVolume(vol uint8, ss *Song) {
	if vol >= 64 {
		vol = 63
	}

	cs.StoredVolume = vol
	cs.ActiveVolume = vol
	cs.LastGlobalVolume = ss.GlobalVolume
}

func (cs *ChannelState) FreezePlayback() {
	cs.freezePlayback = true
}

func (cs *ChannelState) UnfreezePlayback() {
	cs.freezePlayback = false
}

func (cs ChannelState) PlaybackFrozen() bool {
	return cs.freezePlayback
}

func (cs *ChannelState) SetEffectSharedMemoryIfNonZero(input uint8) {
	if input != 0 {
		cs.effectLastNonZero = input
	}
}

func (cs *ChannelState) GetEffectSharedMemory(input uint8) uint8 {
	if input == 0 {
		return cs.effectLastNonZero
	}
	return input
}

func (cs *ChannelState) ResetRetriggerCount() {
	cs.RetriggerCount = 0
}

func (cs *ChannelState) GetMemory() intf.Memory {
	return &cs.memory
}

func (cs *ChannelState) GetActiveVolume() uint8 {
	return cs.ActiveVolume
}

func (cs *ChannelState) SetActiveVolume(vol uint8) {
	cs.ActiveVolume = vol
}

func (cs *ChannelState) GetData() *channel.Data {
	return cs.Cmd
}

func (cs *ChannelState) GetPortaTargetPeriod() float32 {
	return cs.PortaTargetPeriod
}

func (cs *ChannelState) SetPortaTargetPeriod(period float32) {
	cs.PortaTargetPeriod = period
}

func (cs *ChannelState) GetTargetPeriod() float32 {
	return cs.TargetPeriod
}

func (cs *ChannelState) SetTargetPeriod(period float32) {
	cs.TargetPeriod = period
}

func (cs *ChannelState) SetVibratoDelta(delta float32) {
	cs.VibratoDelta = delta
}

func (cs *ChannelState) GetVibratoOscillator() *oscillator.Oscillator {
	return &cs.VibratoOscillator
}

func (cs *ChannelState) GetTremoloOscillator() *oscillator.Oscillator {
	return &cs.TremoloOscillator
}

func (cs *ChannelState) GetTremorOn() bool {
	return cs.TremorOn
}

func (cs *ChannelState) SetTremorOn(on bool) {
	cs.TremorOn = on
}

func (cs *ChannelState) GetTremorTime() int {
	return cs.TremorTime
}

func (cs *ChannelState) SetTremorTime(time int) {
	cs.TremorTime = time
}

func (cs *ChannelState) GetInstrument() *instrument.InstrumentInfo {
	return &cs.Instrument
}

func (cs *ChannelState) GetTargetInst() *instrument.InstrumentInfo {
	return &cs.TargetInst
}

func (cs *ChannelState) SetTargetInst(inst *instrument.InstrumentInfo) {
	cs.TargetInst = *inst
}

func (cs *ChannelState) GetNoteSemitone() uint8 {
	return cs.NoteSemitone
}

func (cs *ChannelState) SetTargetPos(pos float32) {
	cs.TargetPos = pos
}

func (cs *ChannelState) GetPeriod() float32 {
	return cs.Period
}

func (cs *ChannelState) SetPeriod(period float32) {
	cs.Period = float32(math.Max(float64(period), 0))
}

func (cs *ChannelState) GetPos() float32 {
	return cs.Pos
}

func (cs *ChannelState) SetPos(pos float32) {
	cs.Pos = pos
}

func (cs *ChannelState) SetNotePlayTick(tick int) {
	cs.NotePlayTick = tick
}

func (cs *ChannelState) GetRetriggerCount() uint8 {
	return cs.RetriggerCount
}

func (cs *ChannelState) SetRetriggerCount(cnt uint8) {
	cs.RetriggerCount = cnt
}

func (cs *ChannelState) SetPan(pan uint8) {
	cs.Pan = pan
}
