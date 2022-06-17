package playback

import (
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"
	"github.com/gotracker/gotracker/internal/format/xm/layout"
	"github.com/gotracker/gotracker/internal/format/xm/layout/channel"
	"github.com/gotracker/gotracker/internal/optional"
	"github.com/gotracker/gotracker/internal/player/state"
	"github.com/gotracker/gotracker/internal/song/instrument"
	"github.com/gotracker/gotracker/internal/song/note"
	"github.com/gotracker/voice"
)

type channelDataTransaction struct {
	noteAction optional.Value[note.Action]
	noteCalcST optional.Value[note.Semitone]

	targetPos            optional.Value[sampling.Pos]
	targetInst           optional.Value[*instrument.Instrument]
	targetPeriod         optional.Value[note.Period]
	targetStoredSemitone optional.Value[note.Semitone]
	targetVolume         optional.Value[volume.Volume]
}

func (d *channelDataTransaction) Calculate(data *channel.Data, song *layout.Song, cs *state.ChannelState[channel.Memory, channel.Data]) {
	var inst *instrument.Instrument

	if data.HasNote() || data.HasInstrument() {
		instID := data.GetInstrument(cs.StoredSemitone)
		n := data.GetNote()
		if instID.IsEmpty() {
			// use current
			d.targetPos.Set(sampling.Pos{})
		} else if !song.IsValidInstrumentID(instID) {
			d.targetInst.Set(nil)
			n = note.InvalidNote{}
		} else {
			var str note.Semitone
			inst, str = song.GetInstrument(instID)
			n = note.CoalesceNoteSemitone(n, str)
			d.targetInst.Set(inst)
			d.targetPos.Set(sampling.Pos{})
			if inst != nil {
				d.targetVolume.Set(inst.GetDefaultVolume())
				d.noteAction.Set(note.ActionRetrigger)
			}
		}

		if note.IsInvalid(n) {
			d.targetPeriod.Set(nil)
			d.noteAction.Set(note.ActionCut)
		} else if note.IsRelease(n) {
			d.noteAction.Set(note.ActionRelease)
		} else {
			if nn, ok := n.(note.Normal); ok {
				st := note.Semitone(nn)
				d.targetStoredSemitone.Set(st)
				d.noteCalcST.Set(st)
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
		d.targetVolume.Set(v)
	}
}

func (d channelDataTransaction) Commit(cs *state.ChannelState[channel.Memory, channel.Data], currentTick int, semitoneSetterFactory state.SemitoneSetterFactory[channel.Memory, channel.Data]) {
	if pos, ok := d.targetPos.Get(); ok {
		cs.SetTargetPos(pos)
	}

	if inst, ok := d.targetInst.Get(); ok {
		if nc := cs.GetVoice(); nc != nil {
			nc.Release()
			if voice.IsVolumeEnvelopeEnabled(nc) {
				nc.Fadeout()
			}
		}
		cs.SetTargetInst(inst)
	}

	if period, ok := d.targetPeriod.Get(); ok {
		cs.SetTargetPeriod(period)
		cs.SetPortaTargetPeriod(period)
	}

	if st, ok := d.targetStoredSemitone.Get(); ok {
		cs.SetStoredSemitone(st)
	}

	if v, ok := d.targetVolume.Get(); ok {
		cs.SetActiveVolume(v)
	}

	na, targetTick := d.noteAction.Get()
	cs.UseTargetPeriod = targetTick
	cs.SetNotePlayTick(targetTick, na, currentTick)

	if st, ok := d.noteCalcST.Get(); ok {
		cs.NoteOps = append(cs.NoteOps, semitoneSetterFactory(st, cs.SetTargetPeriod))
	}
}
