package playback

import (
	"github.com/gotracker/gotracker/internal/format/internal/filter"
	"github.com/gotracker/gotracker/internal/format/it/layout/channel"
	"github.com/gotracker/gotracker/internal/format/it/playback/util"
	"github.com/gotracker/gotracker/internal/player/state"
	"github.com/gotracker/gotracker/internal/song/note"
	"github.com/gotracker/voice/period"
)

func (m *Manager) doNoteVolCalcs(cs *state.ChannelState[channel.Memory, channel.Data]) {
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
		linearFreqSlides := cs.Memory.Shared.LinearFreqSlides
		period := util.CalcSemitonePeriod(cs.Semitone, inst.GetFinetune(), inst.GetC2Spd(), linearFreqSlides)
		cs.SetTargetPeriod(period)
	}
}

func (m *Manager) processEffect(ch int, cs *state.ChannelState[channel.Memory, channel.Data], currentTick int, lastTick bool) error {
	// pre-effect
	m.doNoteVolCalcs(cs)
	if err := cs.ProcessEffects(m, currentTick, lastTick); err != nil {
		return err
	}
	// post-effect
	m.doNoteVolCalcs(cs)
	cs.SetGlobalVolume(m.GetGlobalVolume())

	var n note.Note = note.EmptyNote{}
	if cs.GetData() != nil {
		n = cs.GetData().GetNote()
	}
	nna := note.ActionContinue
	keyOn := false
	newNote := false
	targetPeriod := cs.GetTargetPeriod()
	if targetPeriod != nil && cs.WillTriggerOn(currentTick) {
		targetInst := cs.GetTargetInst()
		if targetInst != nil {
			newNote = true
			keyOn = true
		}
		if cs.UseTargetPeriod {
			newNote = true
		}

		if newNote {
			cs.TransitionActiveToPastState()
		}

		cs.SetInstrument(targetInst)

		if cs.UseTargetPeriod {
			cs.SetPeriod(targetPeriod)
			cs.SetPortaTargetPeriod(targetPeriod)
		}
		cs.SetPos(cs.GetTargetPos())
	}
	if inst := cs.GetInstrument(); inst != nil {
		if inst.IsReleaseNote(n) {
			nna = note.ActionRelease
		}
		if inst.IsStopNote(n) {
			nna = note.ActionCut
		}
	}

	if nc := cs.GetVoice(); nc != nil {
		switch nna {
		case note.ActionContinue:
			if keyOn {
				nc.Attack()
				mem := cs.GetMemory()
				mem.Retrigger()
			}
		case note.ActionRelease:
			nc.Release()
		case note.ActionCut:
			cs.SetInstrument(nil)
			cs.SetPeriod(nil)
		}
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
