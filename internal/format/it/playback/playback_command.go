package playback

import (
	"github.com/gotracker/gotracker/internal/format/internal/filter"
	"github.com/gotracker/gotracker/internal/format/it/layout/channel"
	itPeriod "github.com/gotracker/gotracker/internal/format/it/period"
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
		linearFreqSlides := cs.Memory.Shared.LinearFreqSlides
		period := itPeriod.CalcSemitonePeriod(cs.Semitone, inst.GetFinetune(), inst.GetC2Spd(), linearFreqSlides)
		o.UpdateFunc(period)
	}
	return nil
}

func (m *Manager) processEffect(ch int, cs *state.ChannelState[channel.Memory, channel.Data], currentTick int, lastTick bool) error {
	// pre-effect
	if err := cs.ProcessVolOps(m); err != nil {
		return err
	}
	if err := cs.ProcessNoteOps(m); err != nil {
		return err
	}
	if err := cs.ProcessEffects(m, currentTick, lastTick); err != nil {
		return err
	}
	// post-effect
	if err := cs.ProcessVolOps(m); err != nil {
		return err
	}
	if err := cs.ProcessNoteOps(m); err != nil {
		return err
	}
	cs.SetGlobalVolume(m.GetGlobalVolume())

	if err := m.processRowNote(ch, cs, currentTick, lastTick); err != nil {
		return err
	}

	if err := m.processVoiceUpdates(ch, cs, currentTick, lastTick); err != nil {
		return err
	}

	return nil
}

func (m *Manager) processRowNote(ch int, cs *state.ChannelState[channel.Memory, channel.Data], currentTick int, lastTick bool) error {
	targetTick, noteAction := cs.WillTriggerOn(currentTick)
	if !targetTick {
		return nil
	}

	keyOn := false
	if nc := cs.GetVoice(); nc != nil {
		keyOn = nc.IsKeyOn()
	}

	if noteAction == note.ActionRetrigger {
		cs.TransitionActiveToPastState()
	}

	wantAttack := false
	targetPeriod := cs.GetTargetPeriod()
	if targetPeriod != nil {
		targetInst := cs.GetTargetInst()
		if targetInst != nil {
			keyOn = true
			wantAttack = noteAction == note.ActionRetrigger
		}

		if cs.UseTargetPeriod {
			cs.SetPeriod(targetPeriod)
			cs.SetPortaTargetPeriod(targetPeriod)
		}

		cs.SetInstrument(targetInst)
		cs.SetPos(cs.GetTargetPos())
	}

	if nc := cs.GetVoice(); nc != nil {
		switch noteAction {
		case note.ActionRetrigger:
			if keyOn && wantAttack {
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
					o.Filter = filter.NewAmigaLPF(period.Frequency(itPeriod.DefaultC2Spd), m.GetSampleRate())
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
