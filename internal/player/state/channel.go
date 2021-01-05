package state

import (
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

	Instrument     intf.NoteControl
	PrevInstrument intf.NoteControl
	Pos            sampling.Pos
	Period         note.Period
	StoredVolume   volume.Volume
	ActiveVolume   volume.Volume
	Pan            panning.Position

	Command      commandFunc
	ActiveEffect intf.Effect

	TargetPeriod   note.Period
	TargetPos      sampling.Pos
	TargetInst     intf.Instrument
	TargetSemitone note.Semitone // from pattern, modified

	StoredSemitone     note.Semitone // from pattern, unmodified, current note
	PrevStoredSemitone note.Semitone // from pattern, unmodified, previous note
	DoRetriggerNote    bool
	PortaTargetPeriod  note.Period
	NotePlayTick       int
	RetriggerCount     uint8
	VibratoDelta       note.Period
	Memory             intf.Memory
	effectLastNonZero  uint8
	Cmd                intf.ChannelData
	freezePlayback     bool
	LastGlobalVolume   volume.Volume
	VibratoOscillator  oscillator.Oscillator
	TremoloOscillator  oscillator.Oscillator
	TargetC2Spd        note.C2SPD
	Semitone           note.Semitone // from TargetSemitone, modified further, used in period calculations
	WantNoteCalc       bool
	WantVolCalc        bool
	UseTargetPeriod    bool
	volumeActive       bool

	OutputChannelNum int
	Filter           intf.Filter
}

// Process processes a channel's row data
func (cs *ChannelState) Process(row intf.Row, globalVol volume.Volume, sd intf.SongData, processCommand commandFunc) {
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
	cs.VibratoDelta = nil

	cs.WantNoteCalc = false
	cs.WantVolCalc = false
	cs.UseTargetPeriod = false

	if cs.Cmd == nil {
		return
	}

	if cs.Cmd.HasNote() {
		cs.UseTargetPeriod = true
		cs.VibratoOscillator.Pos = 0
		cs.TremoloOscillator.Pos = 0
		inst := cs.Cmd.GetInstrument()
		if inst.IsEmpty() {
			// use current
			cs.TargetPos = sampling.Pos{}
		} else if !sd.IsValidInstrumentID(inst) {
			cs.TargetInst = nil
		} else {
			cs.TargetInst = sd.GetInstrument(inst)
			cs.TargetPos = sampling.Pos{}
			if cs.TargetInst != nil {
				cs.WantVolCalc = true
			}
		}

		n := cs.Cmd.GetNote()
		if n == note.EmptyNote {
			cs.WantNoteCalc = false
			cs.DoRetriggerNote = false
		} else if n.IsInvalid() {
			cs.TargetPeriod = nil
			cs.WantNoteCalc = false
		} else if cs.TargetInst != nil {
			cs.PrevStoredSemitone = cs.StoredSemitone
			cs.StoredSemitone = n.Semitone()
			cs.TargetSemitone = cs.StoredSemitone
			cs.WantNoteCalc = true
		}
	} else {
		cs.WantNoteCalc = false
		cs.WantVolCalc = false
		cs.DoRetriggerNote = false
	}

	if cs.Cmd.HasVolume() {
		cs.WantVolCalc = false
		v := cs.Cmd.GetVolume()
		if v == volume.VolumeUseInstVol {
			if cs.TargetInst != nil {
				cs.WantVolCalc = true
			}
		} else {
			cs.SetStoredVolume(v, globalVol)
		}
	}
}

// RenderRowTick renders a channel's row data for a single tick
func (cs *ChannelState) RenderRowTick(tick int, lastTick bool, ch int, ticksThisRow int, mix *mixing.Mixer, panmixer mixing.PanMixer, samplerSpeed float32, tickSamples int, centerPanning volume.Matrix, tickDuration time.Duration) (*mixing.Data, error) {
	if cs.Command != nil {
		cs.Command(ch, cs, tick, lastTick)
	}

	sample := cs.Instrument
	if sample != nil && cs.Period != nil && !cs.PlaybackFrozen() {
		sample.SetVolume(cs.ActiveVolume * cs.LastGlobalVolume)
		period := cs.Period.Add(cs.VibratoDelta)
		sample.SetPeriod(period)

		samplerAdd := float32(period.GetSamplerAdd(float64(samplerSpeed)))

		sample.Update(tickDuration)

		// make a stand-alone data buffer for this channel for this tick
		var data mixing.MixBuffer
		if cs.volumeActive {
			data = mix.NewMixBuffer(tickSamples)
			mixData := mixing.SampleMixIn{
				Sample:    sampling.NewSampler(sample, cs.Pos, samplerAdd),
				StaticVol: volume.Volume(1.0),
				VolMatrix: centerPanning,
				MixPos:    0,
				MixLen:    tickSamples,
			}
			data.MixInSample(mixData)
		}

		cs.Pos.Add(samplerAdd * float32(tickSamples))

		return &mixing.Data{
			Data:       data,
			Pan:        cs.Pan,
			Volume:     volume.Volume(1.0),
			SamplesLen: tickSamples,
		}, nil
	}
	return nil, nil
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

// SetVolumeActive enables or disables the sample of the instrument
func (cs *ChannelState) SetVolumeActive(on bool) {
	cs.volumeActive = on
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
	return cs.StoredSemitone
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
	cs.Period = period
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
	cs.UseTargetPeriod = enabled
	if enabled {
		cs.WantNoteCalc = true
	}
}

// SetSemitone sets the target semitone for the channel
func (cs *ChannelState) SetSemitone(st note.Semitone) {
	cs.TargetSemitone = st
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
