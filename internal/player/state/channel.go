package state

import (
	"math"
	"time"

	"github.com/gotracker/gomixing/mixing"
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/oscillator"
)

type commandFunc func(int, *ChannelState, int, bool)

// ChannelState is the state of a single channel
type ChannelState struct {
	intf.Channel

	Instrument     intf.InstrumentOnChannel
	PrevInstrument intf.InstrumentOnChannel
	Pos            sampling.Pos
	Period         note.Period
	StoredVolume   volume.Volume
	ActiveVolume   volume.Volume
	Pan            panning.Position

	Command      commandFunc
	ActiveEffect intf.Effect

	TargetPeriod      note.Period
	TargetPos         sampling.Pos
	TargetInst        intf.Instrument
	PortaTargetPeriod note.Period
	NotePlayTick      int
	NoteSemitone      note.Semitone
	PrevNoteSemitone  note.Semitone
	DoRetriggerNote   bool
	RetriggerCount    uint8
	TremorOn          bool
	TremorTime        int
	VibratoDelta      note.Period
	Memory            intf.Memory
	effectLastNonZero uint8
	Cmd               intf.ChannelData
	freezePlayback    bool
	LastGlobalVolume  volume.Volume
	VibratoOscillator oscillator.Oscillator
	TremoloOscillator oscillator.Oscillator
	TargetC2Spd       note.C2SPD

	OutputChannelNum int
	Filter           intf.Filter
}

// ProcessRow processes a channel's row data
func (cs *ChannelState) ProcessRow(row intf.Row, channel intf.ChannelData, globalVol volume.Volume, sd intf.SongData, calcSemitonePeriod intf.CalcSemitonePeriodFunc, processCommand commandFunc) {
	cs.Command = processCommand

	cs.PrevInstrument = cs.Instrument
	cs.TargetPeriod = cs.Period
	cs.TargetPos = cs.Pos
	cs.TargetInst = nil
	if cs.Instrument != nil {
		cs.TargetInst = cs.Instrument.GetInstrument()
	}
	cs.DoRetriggerNote = true
	cs.NotePlayTick = 0
	cs.RetriggerCount = 0
	cs.TremorOn = true
	cs.TremorTime = 0
	cs.VibratoDelta = 0
	cs.Cmd = channel

	wantNoteCalc := false

	if channel.HasNote() {
		cs.VibratoOscillator.Pos = 0
		cs.TremoloOscillator.Pos = 0
		inst := channel.GetInstrument()
		if inst.IsEmpty() {
			// use current
			cs.TargetPos = sampling.Pos{}
		} else if !sd.IsValidInstrumentID(inst) {
			cs.TargetInst = nil
		} else {
			cs.TargetInst = sd.GetInstrument(inst)
			cs.TargetPos = sampling.Pos{}
			if cs.TargetInst != nil {
				vol := cs.TargetInst.GetVolume()
				cs.SetStoredVolume(vol, globalVol)
			}
		}

		n := channel.GetNote()
		if n == note.EmptyNote {
			wantNoteCalc = false
			cs.DoRetriggerNote = false
		} else if n.IsInvalid() {
			cs.TargetPeriod = 0
			wantNoteCalc = false
		} else if cs.TargetInst != nil {
			cs.PrevNoteSemitone = cs.NoteSemitone
			cs.NoteSemitone = note.Semitone(int(n.Semitone()) + int(cs.TargetInst.GetSemitoneShift()))
			cs.TargetC2Spd = cs.TargetInst.GetC2Spd()
			wantNoteCalc = true
		}
	} else {
		wantNoteCalc = false
		cs.DoRetriggerNote = false
	}

	if channel.HasVolume() {
		v := channel.GetVolume()
		if v == volume.VolumeUseInstVol {
			sample := cs.TargetInst
			if sample != nil {
				vol := sample.GetVolume()
				cs.SetStoredVolume(vol, globalVol)
			}
		} else {
			cs.SetStoredVolume(v, globalVol)
		}
	}

	if wantNoteCalc {
		cs.TargetPeriod = calcSemitonePeriod(cs.NoteSemitone, cs.TargetC2Spd)
		cs.PortaTargetPeriod = cs.TargetPeriod
	}
}

// RenderRowTick renders a channel's row data for a single tick
func (cs *ChannelState) RenderRowTick(tick int, lastTick bool, mixerData []mixing.Data, ch int, ticksThisRow int, mix *mixing.Mixer, panmixer mixing.PanMixer, samplerSpeed float32, tickSamples int, centerPanning volume.Matrix, tickDuration time.Duration) {
	if cs.Command != nil {
		cs.Command(ch, cs, tick, lastTick)
	}

	sample := cs.Instrument
	if sample != nil && cs.Period != 0 && !cs.PlaybackFrozen() {
		sample.SetVolume(cs.ActiveVolume * cs.LastGlobalVolume)
		// make a stand-alone data buffer for this channel for this tick
		data := mix.NewMixBuffer(tickSamples)

		period := cs.Period + cs.VibratoDelta
		sample.SetPeriod(period)
		samplerAdd := samplerSpeed / float32(period)
		sample.Update(tickDuration)
		mixData := mixing.SampleMixIn{
			Sample:    sampling.NewSampler(sample, cs.Pos, samplerAdd),
			StaticVol: volume.Volume(1.0),
			VolMatrix: centerPanning,
			MixPos:    0,
			MixLen:    tickSamples,
		}
		data.MixInSample(mixData)
		cs.Pos.Add(samplerAdd * float32(tickSamples))
		mixerData[tick] = mixing.Data{
			Data:       data,
			Pan:        cs.Pan,
			Volume:     volume.Volume(1.0),
			SamplesLen: tickSamples,
		}
	}
}

// SetStoredVolume sets the stored volume value for the channel
// this also modifies the active volume
// and stores the active global volume value (which doesn't always get set on channels immediately)
func (cs *ChannelState) SetStoredVolume(vol volume.Volume, globalVol volume.Volume) {
	if vol != volume.VolumeUseInstVol {
		cs.StoredVolume = vol
	}
	cs.SetActiveVolume(vol)
	cs.LastGlobalVolume = globalVol
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
	return cs.Memory
}

// SetMemory sets the custom effect memory interface
func (cs *ChannelState) SetMemory(mem intf.Memory) {
	cs.Memory = mem
}

// GetActiveVolume returns the current active volume on the channel
func (cs *ChannelState) GetActiveVolume() volume.Volume {
	return cs.ActiveVolume
}

// SetActiveVolume sets the active volume on the channel
func (cs *ChannelState) SetActiveVolume(vol volume.Volume) {
	if vol != volume.VolumeUseInstVol {
		cs.ActiveVolume = vol
	}
}

// GetData returns the interface to the current channel song pattern data
func (cs *ChannelState) GetData() intf.ChannelData {
	return cs.Cmd
}

// GetPortaTargetPeriod returns the current target portamento (to note) sampler period
func (cs *ChannelState) GetPortaTargetPeriod() note.Period {
	return cs.PortaTargetPeriod
}

// SetPortaTargetPeriod sets the current target portamento (to note) sampler period
func (cs *ChannelState) SetPortaTargetPeriod(period note.Period) {
	cs.PortaTargetPeriod = period
}

// GetTargetPeriod returns the soon-to-be-committed sampler period (when the note retriggers)
func (cs *ChannelState) GetTargetPeriod() note.Period {
	return cs.TargetPeriod
}

// SetTargetPeriod sets the soon-to-be-committed sampler period (when the note retriggers)
func (cs *ChannelState) SetTargetPeriod(period note.Period) {
	cs.TargetPeriod = period
}

// SetVibratoDelta sets the vibrato (ephemeral) delta sampler period
func (cs *ChannelState) SetVibratoDelta(delta note.Period) {
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
	if cs.Instrument == nil {
		return nil
	}
	return cs.Instrument.GetInstrument()
}

// SetInstrument sets the interface to the active instrument
func (cs *ChannelState) SetInstrument(inst intf.Instrument) {
	if inst == nil {
		cs.Instrument = nil
	}
	cs.Instrument = inst.InstantiateOnChannel(int(cs.OutputChannelNum), cs.Filter)
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
func (cs *ChannelState) GetNoteSemitone() note.Semitone {
	return cs.NoteSemitone
}

// GetTargetPos returns the soon-to-be-committed sample position of the instrument
func (cs *ChannelState) GetTargetPos() sampling.Pos {
	return cs.TargetPos
}

// SetTargetPos sets the soon-to-be-committed sample position of the instrument
func (cs *ChannelState) SetTargetPos(pos sampling.Pos) {
	cs.TargetPos = pos
}

// GetPeriod returns the current sampler period of the active instrument
func (cs *ChannelState) GetPeriod() note.Period {
	return cs.Period
}

// SetPeriod sets the current sampler period of the active instrument
func (cs *ChannelState) SetPeriod(period note.Period) {
	cs.Period = note.Period(math.Max(float64(period), 0))
}

// GetPos returns the sample position of the active instrument
func (cs *ChannelState) GetPos() sampling.Pos {
	return cs.Pos
}

// SetPos sets the sample position of the active instrument
func (cs *ChannelState) SetPos(pos sampling.Pos) {
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
func (cs *ChannelState) SetPan(pan panning.Position) {
	cs.Pan = pan
}

// SetDoRetriggerNote sets the enablement flag for DoRetriggerNote
// this gets reset on every row
func (cs *ChannelState) SetDoRetriggerNote(enabled bool) {
	cs.DoRetriggerNote = enabled
}

// GetFilter returns the active filter on the channel
func (cs *ChannelState) GetFilter() intf.Filter {
	return cs.Filter
}

// SetFilter sets the active filter on the channel
func (cs *ChannelState) SetFilter(filter intf.Filter) {
	cs.Filter = filter
	if cs.Instrument != nil {
		cs.Instrument.SetFilter(filter)
	}
}

// SetOutputChannelNum sets the output channel number for the channel
func (cs *ChannelState) SetOutputChannelNum(outputChNum int) {
	cs.OutputChannelNum = outputChNum
}
