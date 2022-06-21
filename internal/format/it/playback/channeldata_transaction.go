package playback

import (
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"
	"github.com/gotracker/gotracker/internal/format/it/layout/channel"
	"github.com/gotracker/gotracker/internal/optional"
	"github.com/gotracker/gotracker/internal/player/intf"
	"github.com/gotracker/gotracker/internal/player/state"
	"github.com/gotracker/gotracker/internal/song"
	"github.com/gotracker/gotracker/internal/song/instrument"
	"github.com/gotracker/gotracker/internal/song/note"
)

type noteTransaction struct {
	noteAction optional.Value[note.Action]
	noteCalcST optional.Value[note.Semitone]

	targetPos            optional.Value[sampling.Pos]
	targetInst           optional.Value[*instrument.Instrument]
	targetPeriod         optional.Value[note.Period]
	targetStoredSemitone optional.Value[note.Semitone]
	targetNewNoteAction  optional.Value[note.Action]
	targetVolume         optional.Value[volume.Volume]
}

type channelDataTransaction struct {
	data *channel.Data

	nt noteTransaction

	volOps  []state.VolOp[channel.Memory, channel.Data]
	noteOps []state.NoteOp[channel.Memory, channel.Data]
}

func (d channelDataTransaction) GetData() *channel.Data {
	return d.data
}

func (d *channelDataTransaction) SetData(cd *channel.Data, s song.Data, cs *state.ChannelState[channel.Memory, channel.Data]) {
	d.data = cd

	var n note.Note = note.EmptyNote{}
	inst := cs.GetInstrument()
	prevInst := inst

	if d.data.HasNote() || d.data.HasInstrument() {
		instID := d.data.GetInstrument(cs.StoredSemitone)
		n = d.data.GetNote()
		if instID.IsEmpty() {
			// use current
			d.nt.targetPos.Set(sampling.Pos{})
		} else if !s.IsValidInstrumentID(instID) {
			d.nt.targetInst.Set(nil)
			n = note.InvalidNote{}
		} else {
			var str note.Semitone
			inst, str = s.GetInstrument(instID)
			n = note.CoalesceNoteSemitone(n, str)
			if !note.IsEmpty(n) && inst == nil {
				inst = prevInst
			}
			d.nt.targetInst.Set(inst)
			d.nt.targetPos.Set(sampling.Pos{})
			if inst != nil {
				d.nt.targetVolume.Set(inst.GetDefaultVolume())
				d.nt.noteAction.Set(note.ActionRetrigger)
				d.nt.targetNewNoteAction.Set(inst.GetNewNoteAction())
			}
		}
	}

	if note.IsInvalid(n) {
		d.nt.targetPeriod.Set(nil)
		d.nt.noteAction.Set(note.ActionCut)
	} else if note.IsRelease(n) {
		d.nt.noteAction.Set(note.ActionRelease)
	} else if !note.IsEmpty(n) {
		if nn, ok := n.(note.Normal); ok {
			st := note.Semitone(nn)
			d.nt.targetStoredSemitone.Set(st)
			d.nt.noteCalcST.Set(st)
		} else {
			d.nt.noteAction.Set(note.ActionCut)
		}
	}

	if d.data.HasVolume() {
		v := d.data.GetVolume()
		if v == volume.VolumeUseInstVol {
			if inst != nil {
				v = inst.GetDefaultVolume()
			}
		}
		d.nt.targetVolume.Set(v)
	}
}

func (d *channelDataTransaction) CommitStartTick(cs *state.ChannelState[channel.Memory, channel.Data], currentTick int, lastTick bool, semitoneSetterFactory state.SemitoneSetterFactory[channel.Memory, channel.Data]) {
	if pos, ok := d.nt.targetPos.Get(); ok {
		cs.SetTargetPos(pos)
	}

	if inst, ok := d.nt.targetInst.Get(); ok {
		cs.SetTargetInst(inst)
	}

	if period, ok := d.nt.targetPeriod.Get(); ok {
		cs.SetTargetPeriod(period)
		cs.SetPortaTargetPeriod(period)
	}

	if st, ok := d.nt.targetStoredSemitone.Get(); ok {
		cs.SetStoredSemitone(st)
	}

	if nna, ok := d.nt.targetNewNoteAction.Get(); ok {
		cs.SetNewNoteAction(nna)
	}

	if v, ok := d.nt.targetVolume.Get(); ok {
		cs.SetActiveVolume(v)
	}

	na, targetTick := d.nt.noteAction.Get()
	cs.UseTargetPeriod = targetTick
	cs.SetNotePlayTick(targetTick, na, currentTick)

	if st, ok := d.nt.noteCalcST.Get(); ok {
		d.AddNoteOp(semitoneSetterFactory(st, cs.SetTargetPeriod))
	}
}

func (d channelDataTransaction) CommitTick(cs *state.ChannelState[channel.Memory, channel.Data], currentTick int, lastTick bool, semitoneSetterFactory state.SemitoneSetterFactory[channel.Memory, channel.Data]) {
}

func (d channelDataTransaction) CommitPostTick(cs *state.ChannelState[channel.Memory, channel.Data], currentTick int, lastTick bool, semitoneSetterFactory state.SemitoneSetterFactory[channel.Memory, channel.Data]) {
}

func (d *channelDataTransaction) AddVolOp(op state.VolOp[channel.Memory, channel.Data]) {
	d.volOps = append(d.volOps, op)
}

func (d *channelDataTransaction) ProcessVolOps(p intf.Playback, cs *state.ChannelState[channel.Memory, channel.Data]) error {
	for _, op := range d.volOps {
		if op == nil {
			continue
		}
		if err := op.Process(p, cs); err != nil {
			return err
		}
	}
	d.volOps = nil

	return nil
}

func (d *channelDataTransaction) AddNoteOp(op state.NoteOp[channel.Memory, channel.Data]) {
	d.noteOps = append(d.noteOps, op)
}

func (d *channelDataTransaction) ProcessNoteOps(p intf.Playback, cs *state.ChannelState[channel.Memory, channel.Data]) error {
	for _, op := range d.noteOps {
		if op == nil {
			continue
		}
		if err := op.Process(p, cs); err != nil {
			return err
		}
	}
	d.noteOps = nil

	return nil
}
