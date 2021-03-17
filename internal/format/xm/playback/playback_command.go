package playback

import (
	"github.com/gotracker/voice"

	"gotracker/internal/format/internal/filter"
	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/state"
	"gotracker/internal/song/note"
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
		linearFreqSlides := cs.Memory.(*channel.Memory).LinearFreqSlides
		period := util.CalcSemitonePeriod(cs.Semitone, inst.GetFinetune(), inst.GetC2Spd(), linearFreqSlides)
		cs.SetTargetPeriod(period)
	}
}

func (m *Manager) processEffect(ch int, cs *state.ChannelState, currentTick int, lastTick bool) error {
	// pre-effect
	m.doNoteVolCalcs(cs)
	if err := intf.DoEffect(cs.ActiveEffect, cs, m, currentTick, lastTick); err != nil {
		return err
	}
	// post-effect
	m.doNoteVolCalcs(cs)
	cs.SetGlobalVolume(m.GetGlobalVolume())

	var n note.Note = note.EmptyNote{}
	if cs.TrackData != nil {
		n = cs.TrackData.GetNote()
	}
	keyOff := false
	keyOn := false
	stop := false
	targetPeriod := cs.GetTargetPeriod()
	if targetPeriod != nil && cs.WillTriggerOn(currentTick) {
		if targetInst := cs.GetTargetInst(); targetInst != nil {
			cs.SetInstrument(targetInst)
			keyOn = true
		} else {
			cs.SetInstrument(nil)
		}
		if cs.UseTargetPeriod {
			if nc := cs.GetVoice(); nc != nil {
				nc.Release()
				if voice.IsVolumeEnvelopeEnabled(nc) {
					nc.Fadeout()
				}
			}
			cs.SetPeriod(targetPeriod)
			cs.SetPortaTargetPeriod(targetPeriod)
		}
		cs.SetPos(cs.GetTargetPos())
	}
	if inst := cs.GetInstrument(); inst != nil {
		keyOff = inst.IsReleaseNote(n)
		stop = inst.IsStopNote(n)
	}

	if nc := cs.GetVoice(); nc != nil {
		if keyOn {
			nc.Attack()
			mem := cs.GetMemory().(*channel.Memory)
			mem.Retrigger()
		} else if keyOff {
			nc.Release()
			if voice.IsVolumeEnvelopeEnabled(nc) {
				nc.Fadeout()
			}
			cs.SetPeriod(nil)
		} else if stop {
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
					o.Filter = filter.NewAmigaLPF()
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
