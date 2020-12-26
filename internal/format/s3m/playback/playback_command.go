package playback

import (
	"gotracker/internal/format/s3m/layout"
	"gotracker/internal/format/s3m/playback/filter"
	"gotracker/internal/player/note"
	"gotracker/internal/player/state"
)

func (m *Manager) processCommand(ch int, cs *state.ChannelState, currentTick int, lastTick bool) {
	if cs.ActiveEffect != nil {
		if currentTick == 0 {
			cs.ActiveEffect.Start(cs, m)
		}
		cs.ActiveEffect.Tick(cs, m, currentTick)
		if lastTick {
			cs.ActiveEffect.Stop(cs, m, currentTick)
		}
	}

	if cs.TargetPeriod == 0 && cs.Instrument != nil && cs.Instrument.GetKeyOn() {
		if cs.Cmd.GetNote() == note.StopNote {
			cs.Instrument.SetVolume(0)
		}
		cs.Instrument.SetKeyOn(cs.Period, false)
	} else if cs.DoRetriggerNote && currentTick == cs.NotePlayTick {
		cs.Instrument = nil
		if cs.TargetInst != nil {
			if cs.PrevInstrument != nil && cs.PrevInstrument.GetInstrument() == cs.TargetInst {
				cs.Instrument = cs.PrevInstrument
				cs.Instrument.SetKeyOn(cs.Period, false)
			} else {
				inst := cs.TargetInst.InstantiateOnChannel(cs.OutputChannelNum, cs.Filter)
				if opl, ok := inst.(*layout.InstrumentOnChannel); ok {
					opl.Playback = m
				}
				cs.Instrument = inst
			}
		}
		cs.Period = cs.TargetPeriod
		cs.Pos = cs.TargetPos
		if cs.Period != 0 && cs.Instrument != nil {
			cs.Instrument.SetKeyOn(cs.Period, true)
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
func (m *Manager) SetPatternLoopStart() {
	if m.preMixRowTxn != nil {
		m.preMixRowTxn.SetPatternLoopStart()
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetPatternLoopStart()
		rowTxn.Commit()
	}
}

// SetPatternLoopEnd sets the pattern loop end position
func (m *Manager) SetPatternLoopEnd() {
	if m.preMixRowTxn != nil {
		m.preMixRowTxn.SetPatternLoopEnd()
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetPatternLoopEnd()
		rowTxn.Commit()
	}
}

// SetPatternLoopCount sets the total loops desired for the pattern loop mechanism
func (m *Manager) SetPatternLoopCount(loops int) {
	if m.preMixRowTxn != nil {
		m.preMixRowTxn.SetPatternLoopCount(loops)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetPatternLoopCount(loops)
		rowTxn.Commit()
	}
}
