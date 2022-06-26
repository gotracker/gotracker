package playback

import (
	"github.com/gotracker/gotracker/internal/format/internal/filter"
	"github.com/gotracker/gotracker/internal/format/s3m/layout/channel"
	"github.com/gotracker/gotracker/internal/format/s3m/playback/util"
	"github.com/gotracker/gotracker/internal/player/intf"
	"github.com/gotracker/gotracker/internal/player/state"
	"github.com/gotracker/gotracker/internal/song/note"
	"github.com/gotracker/voice/period"
)

type doNoteCalc struct {
	Semitone   note.Semitone
	UpdateFunc state.PeriodUpdateFunc
}

func (o doNoteCalc) Process(p intf.Playback, cs *state.ChannelState[channel.Memory, channel.Data]) error {
	if o.UpdateFunc == nil {
		return nil
	}

	if inst := cs.GetTargetInst(); inst != nil {
		cs.Semitone = note.Semitone(int(o.Semitone) + int(inst.GetSemitoneShift()))
		period := util.CalcSemitonePeriod(cs.Semitone, inst.GetFinetune(), inst.GetC2Spd())
		o.UpdateFunc(period)
	}
	return nil
}

func (m *Manager) processEffect(ch int, cs *state.ChannelState[channel.Memory, channel.Data], currentTick int, lastTick bool) error {
	if txn := cs.GetTxn(); txn != nil {
		if err := txn.CommitPreTick(m, cs, currentTick, lastTick, cs.SemitoneSetterFactory); err != nil {
			return err
		}
		if err := txn.CommitTick(m, cs, currentTick, lastTick, cs.SemitoneSetterFactory); err != nil {
			return err
		}
		if err := txn.CommitPostTick(m, cs, currentTick, lastTick, cs.SemitoneSetterFactory); err != nil {
			return err
		}
	}

	if err := m.processRowNote(ch, cs, currentTick, lastTick); err != nil {
		return err
	}

	if err := m.processVoiceUpdates(ch, cs, currentTick, lastTick); err != nil {
		return err
	}

	return nil
}

func (m *Manager) processRowNote(ch int, cs *state.ChannelState[channel.Memory, channel.Data], currentTick int, lastTick bool) error {
	triggerTick, noteAction := cs.WillTriggerOn(currentTick)
	if !triggerTick {
		return nil
	}
	var n note.Note = note.EmptyNote{}
	if cs.GetData() != nil {
		n = cs.GetData().GetNote()
	}
	keyOn := false
	keyOff := false
	stop := false

	if targetInst := cs.GetTargetInst(); targetInst != nil {
		cs.SetInstrument(targetInst)
		keyOn = true
	} else {
		cs.SetInstrument(nil)
	}

	if cs.UseTargetPeriod {
		if nc := cs.GetVoice(); nc != nil {
			nc.Release()
			nc.Fadeout()
		}
		targetPeriod := cs.GetTargetPeriod()
		cs.SetPeriod(targetPeriod)
		cs.SetPortaTargetPeriod(targetPeriod)
	}
	cs.SetPos(cs.GetTargetPos())

	if inst := cs.GetInstrument(); inst != nil {
		keyOff = inst.IsReleaseNote(n)
		stop = inst.IsStopNote(n)
	}

	if nc := cs.GetVoice(); nc != nil {
		if keyOn && noteAction == note.ActionRetrigger {
			// S3M is weird and only sets the global volume on the channel when a KeyOn happens
			cs.SetGlobalVolume(m.GetGlobalVolume())
			nc.Attack()
			mem := cs.GetMemory()
			mem.Retrigger()
		} else if keyOff {
			nc.Release()
			nc.Fadeout()
			cs.SetPeriod(nil)
		} else if stop {
			cs.SetInstrument(nil)
			cs.SetPeriod(nil)
		}
	}
	return nil
}

func (m *Manager) processVoiceUpdates(ch int, cs *state.ChannelState[channel.Memory, channel.Data], currentTick int, lastTick bool) error {
	if cs.UsePeriodOverride {
		cs.UsePeriodOverride = false
		arpeggioPeriod := cs.GetPeriodOverride()
		cs.SetPeriod(arpeggioPeriod)
	}
	return nil
}

// SetFilterEnable activates or deactivates the amiga low-pass filter on the instruments
func (m *Manager) SetFilterEnable(on bool) {
	for i := range m.song.ChannelSettings {
		c := m.GetChannel(i)
		if o := c.GetOutputChannel(); o != nil {
			if on {
				if o.Filter == nil {
					o.Filter = filter.NewAmigaLPF(period.Frequency(util.DefaultC2Spd), m.GetSampleRate())
				}
			} else {
				o.Filter = nil
			}
		}
	}
}

// SetTicks sets the number of ticks the row expects to play for
func (m *Manager) SetTicks(ticks int) error {
	if m.preMixRowTxn != nil {
		m.preMixRowTxn.Ticks.Set(ticks)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.Ticks.Set(ticks)
		if err := rowTxn.Commit(); err != nil {
			return err
		}
	}

	return nil
}

// AddRowTicks increases the number of ticks the row expects to play for
func (m *Manager) AddRowTicks(ticks int) error {
	if m.preMixRowTxn != nil {
		m.preMixRowTxn.FinePatternDelay.Set(ticks)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.FinePatternDelay.Set(ticks)
		if err := rowTxn.Commit(); err != nil {
			return err
		}
	}

	return nil
}

// SetPatternDelay sets the repeat number for the row to `rept`
// NOTE: this may be set 1 time (first in wins) and will be reset only by the next row being read in
func (m *Manager) SetPatternDelay(rept int) error {
	if m.preMixRowTxn != nil {
		m.preMixRowTxn.SetPatternDelay(rept)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetPatternDelay(rept)
		if err := rowTxn.Commit(); err != nil {
			return err
		}
	}

	return nil
}
