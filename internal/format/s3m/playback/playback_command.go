package playback

import (
	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/format/s3m/playback/filter"
	"gotracker/internal/format/s3m/playback/util"
	"gotracker/internal/player/note"
	"gotracker/internal/player/state"
)

func (m *Manager) doNoteVolCalcs(cs *state.ChannelState) {
	inst := cs.GetTargetInst()
	if inst == nil {
		return
	}

	if cs.WantVolCalc {
		cs.WantVolCalc = false
		cs.SetActiveVolume(inst.GetDefaultVolume())
	}
	if cs.WantNoteCalc {
		cs.WantNoteCalc = false
		cs.Semitone = note.Semitone(int(cs.TargetSemitone) + int(inst.GetSemitoneShift()))
		period := util.CalcSemitonePeriod(cs.Semitone, inst.GetFinetune(), inst.GetC2Spd())
		cs.SetTargetPeriod(period)
	}
}

func (m *Manager) processCommand(ch int, cs *state.ChannelState, currentTick int, lastTick bool) {
	// pre-effect
	m.doNoteVolCalcs(cs)
	if cs.ActiveEffect != nil {
		if currentTick == 0 {
			cs.ActiveEffect.Start(cs, m)
		}
		cs.ActiveEffect.Tick(cs, m, currentTick)
		if lastTick {
			cs.ActiveEffect.Stop(cs, m, currentTick)
		}
	}
	// post-effect
	m.doNoteVolCalcs(cs)

	n := note.EmptyNote
	if cs.TrackData != nil {
		n = cs.TrackData.GetNote()
	}
	keyOff := n.IsEmpty() || n.IsStop()
	targetPeriod := cs.GetTargetPeriod()
	if cs.DoRetriggerNote && targetPeriod != nil && currentTick == cs.NotePlayTick {
		if targetInst := cs.GetTargetInst(); targetInst != nil {
			if targetInst != cs.GetInstrument() {
				cs.SetInstrument(targetInst, m)
			}
		} else {
			cs.SetInstrument(nil, nil)
		}
		if cs.UseTargetPeriod {
			cs.SetPeriod(targetPeriod)
			cs.PortaTargetPeriod = targetPeriod
		}
		cs.SetPos(cs.GetTargetPos())
		if nc := cs.GetNoteControl(); nc != nil {
			cs.LastGlobalVolume = m.GetGlobalVolume()
			if nc.GetKeyOn() {
				nc.Release()
			}
			nc.Attack()
			keyOff = false
			mem := cs.GetMemory().(*channel.Memory)
			mem.Retrigger()
		}
	}

	if keyOff {
		if nc := cs.GetNoteControl(); nc != nil && nc.GetKeyOn() {
			nc.Release()
		}
		cs.SetPeriod(nil)
	}
}

// SetFilterEnable activates or deactivates the amiga low-pass filter on the instruments
func (m *Manager) SetFilterEnable(on bool) {
	for i := range m.song.ChannelSettings {
		c := m.GetChannel(i)
		if on {
			if c.GetFilter() == nil {
				c.SetFilter(filter.NewAmigaLPF())
			}
		} else {
			c.SetFilter(nil)
		}
	}
}

// SetTicks sets the number of ticks the row expects to play for
func (m *Manager) SetTicks(ticks int) {
	if m.preMixRowTxn != nil {
		m.preMixRowTxn.SetTicks(ticks)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetTicks(ticks)
		rowTxn.Commit()
	}
}

// AddRowTicks increases the number of ticks the row expects to play for
func (m *Manager) AddRowTicks(ticks int) {
	if m.preMixRowTxn != nil {
		m.preMixRowTxn.SetFinePatternDelay(ticks)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetFinePatternDelay(ticks)
		rowTxn.Commit()
	}
}

// SetPatternDelay sets the repeat number for the row to `rept`
// NOTE: this may be set 1 time (first in wins) and will be reset only by the next row being read in
func (m *Manager) SetPatternDelay(rept int) {
	if m.preMixRowTxn != nil {
		m.preMixRowTxn.SetPatternDelay(rept)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetPatternDelay(rept)
		rowTxn.Commit()
	}
}
