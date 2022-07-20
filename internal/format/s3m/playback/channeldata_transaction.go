package playback

import (
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"
	"github.com/gotracker/gotracker/internal/format/s3m/layout/channel"
	"github.com/gotracker/gotracker/internal/format/s3m/playback/effect"
	"github.com/gotracker/gotracker/internal/player/intf"
	"github.com/gotracker/gotracker/internal/player/state"
	"github.com/gotracker/gotracker/internal/song"
	"github.com/gotracker/gotracker/internal/song/instrument"
	"github.com/gotracker/gotracker/internal/song/note"
)

type channelDataConverter struct{}

func (c channelDataConverter) Process(out *state.ChannelDataActions, data *channel.Data, s song.Data, cs *state.ChannelState[channel.Memory, channel.Data]) error {
	if data == nil {
		return nil
	}

	var inst *instrument.Instrument

	if data.HasNote() || data.HasInstrument() {
		instID := data.GetInstrument(cs.StoredSemitone)
		n := data.GetNote()
		var (
			wantRetrigger    bool
			wantRetriggerVol bool
		)
		if instID.IsEmpty() {
			// use current
			inst = cs.GetInstrument()
			wantRetrigger = true
		} else if !s.IsValidInstrumentID(instID) {
			out.TargetInst.Set(nil)
			n = note.InvalidNote{}
		} else {
			var str note.Semitone
			inst, str = s.GetInstrument(instID)
			n = note.CoalesceNoteSemitone(n, str)
			wantRetrigger = true
			wantRetriggerVol = true
		}

		if wantRetrigger {
			out.TargetInst.Set(inst)
			out.TargetPos.Set(sampling.Pos{})
			if inst != nil {
				if wantRetriggerVol {
					out.TargetVolume.Set(inst.GetDefaultVolume())
				}
				out.NoteAction.Set(note.ActionRetrigger)
			}
		}

		if note.IsInvalid(n) {
			out.TargetPeriod.Set(nil)
			out.NoteAction.Set(note.ActionCut)
		} else if note.IsRelease(n) {
			out.NoteAction.Set(note.ActionRelease)
		} else {
			if nn, ok := n.(note.Normal); ok {
				st := note.Semitone(nn)
				out.TargetStoredSemitone.Set(st)
				out.NoteCalcST.Set(st)
			}
		}
	}

	if data.HasVolume() {
		v := data.GetVolume()
		if v == volume.VolumeUseInstVol {
			if inst != nil {
				v = inst.GetDefaultVolume()
			}
		}
		out.TargetVolume.Set(v)
	}

	return nil
}

type channelDataTransaction struct {
	state.ChannelDataTxnHelper[channel.Memory, channel.Data, channelDataConverter]
}

func (d *channelDataTransaction) CommitPreRow(p intf.Playback, cs *state.ChannelState[channel.Memory, channel.Data], semitoneSetterFactory state.SemitoneSetterFactory[channel.Memory, channel.Data]) error {
	e := effect.Factory(cs.GetMemory(), d.Data)
	cs.SetActiveEffect(e)
	if e != nil {
		if onEff := p.GetOnEffect(); onEff != nil {
			onEff(e)
		}
		if err := intf.EffectPreStart[channel.Memory, channel.Data](e, cs, p); err != nil {
			return err
		}
	}

	return nil
}

func (d *channelDataTransaction) CommitRow(p intf.Playback, cs *state.ChannelState[channel.Memory, channel.Data], semitoneSetterFactory state.SemitoneSetterFactory[channel.Memory, channel.Data]) error {
	if pos, ok := d.TargetPos.Get(); ok {
		cs.SetTargetPos(pos)
	}

	if inst, ok := d.TargetInst.Get(); ok {
		cs.SetTargetInst(inst)
	}

	if period, ok := d.TargetPeriod.Get(); ok {
		cs.SetTargetPeriod(period)
	}

	if st, ok := d.TargetStoredSemitone.Get(); ok {
		cs.SetStoredSemitone(st)
	}

	if v, ok := d.TargetVolume.Get(); ok {
		cs.SetActiveVolume(v)
	}

	na, targetTick := d.NoteAction.Get()
	cs.UseTargetPeriod = targetTick
	cs.SetNotePlayTick(targetTick, na, 0)

	if st, ok := d.NoteCalcST.Get(); ok {
		d.AddNoteOp(semitoneSetterFactory(st, cs.SetTargetPeriod))
	}

	return nil
}

func init() {
	var _ channelDataConverter
}
