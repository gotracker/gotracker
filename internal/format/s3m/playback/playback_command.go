package playback

import (
	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/format/s3m/playback/filter"
	"gotracker/internal/format/s3m/playback/util"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/state"
)

func (m *Manager) doNoteVolCalcs(cs *state.ChannelState) {
	inst := cs.TargetInst
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
		cs.TargetPeriod = period
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
	if cs.DoRetriggerNote && cs.TargetPeriod != nil && currentTick == cs.NotePlayTick {
		cs.Instrument = nil
		if cs.TargetInst != nil {
			if cs.PrevInstrument != nil && cs.PrevInstrument.GetInstrument() == cs.TargetInst {
				cs.Instrument = cs.PrevInstrument
				cs.Instrument.Release()
			} else {
				inst := cs.TargetInst.InstantiateOnChannel(cs.OutputChannelNum, cs.Filter)
				inst.SetPlayback(m)
				cs.Instrument = inst
			}
		}
		if cs.UseTargetPeriod {
			cs.Period = cs.TargetPeriod
			cs.PortaTargetPeriod = cs.TargetPeriod
		}
		cs.Pos = cs.TargetPos
		if cs.Instrument != nil {
			cs.LastGlobalVolume = m.GetGlobalVolume()
			cs.Instrument.Attack()
			keyOff = false
			mem := cs.GetMemory().(*channel.Memory)
			mem.Retrigger()
		}
	}

	if keyOff && cs.Instrument != nil && cs.Instrument.GetKeyOn() {
		if n == note.StopNote {
			cs.Instrument.Release()
			cs.Instrument = nil
			cs.Period = nil
		} else {
			cs.Instrument.Release()
		}
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

// SetPatternLoopStart sets the pattern loop start position
func (m *Manager) SetPatternLoopStart(row intf.RowIdx) {
	if m.postMixRowTxn != nil {
		m.postMixRowTxn.SetPatternLoopStart(row)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetPatternLoopStart(row)
		rowTxn.Commit()
	}
}

// SetPatternLoopEnd sets the pattern loop end position
func (m *Manager) SetPatternLoopEnd() {
	if m.postMixRowTxn != nil {
		m.postMixRowTxn.SetPatternLoopEnd()
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetPatternLoopEnd()
		rowTxn.Commit()
	}
}

// SetPatternLoopCount sets the total loops desired for the pattern loop mechanism
func (m *Manager) SetPatternLoopCount(loops int) {
	if m.postMixRowTxn != nil {
		m.postMixRowTxn.SetPatternLoopCount(loops)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetPatternLoopCount(loops)
		rowTxn.Commit()
	}
}
