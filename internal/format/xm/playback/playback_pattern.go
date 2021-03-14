package playback

import (
	"errors"
	"time"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/format/xm/playback/effect"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/state"
	"gotracker/internal/song"
	"gotracker/internal/song/note"

	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"
)

const (
	tickBaseDuration = time.Duration(2500) * time.Millisecond
)

func (m *Manager) processPatternRow() error {
	patIdx, err := m.pattern.GetCurrentPatternIdx()
	if err != nil {
		return err
	}

	if m.pattern.NeedResetPatternLoops() {
		for _, cs := range m.channels {
			mem := cs.GetMemory().(*channel.Memory)
			pl := mem.GetPatternLoop()
			pl.Count = 0
			pl.Enabled = false
		}
	}

	pat := m.song.GetPattern(patIdx)
	if pat == nil {
		return song.ErrStopSong
	}

	withinPatternLoop := false
	for _, cs := range m.channels {
		mem := cs.GetMemory().(*channel.Memory)
		pl := mem.GetPatternLoop()
		if pl.Enabled {
			withinPatternLoop = true
			break
		}
	}

	if !withinPatternLoop {
		if err := m.pattern.Observe(); err != nil {
			return err
		}
	}

	rows := pat.GetRows()

	myCurrentRow := m.pattern.GetCurrentRow()

	row := rows.GetRow(myCurrentRow)

	preMixRowTxn := m.pattern.StartTransaction()
	defer func() {
		preMixRowTxn.Cancel()
		m.preMixRowTxn = nil
	}()
	m.preMixRowTxn = preMixRowTxn

	s := m.GetSampler()
	if s == nil {
		return errors.New("sampler not configured")
	}

	if m.rowRenderState == nil {
		panmixer := s.GetPanMixer()

		m.rowRenderState = &rowRenderState{
			mix:          s.Mixer(),
			samplerSpeed: s.GetSamplerSpeed(),
			panmixer:     panmixer,
		}
	}

	var resetMemory bool
	if myCurrentRow == 0 {
		if myCurrentOrder := m.pattern.GetCurrentOrder(); myCurrentOrder == 0 {
			resetMemory = true
		}
	}

	for ch := range m.channels {
		cs := &m.channels[ch]
		cs.TrackData = nil
		if resetMemory {
			mem := cs.GetMemory()
			mem.StartOrder()
		}
	}

	// generate effects and run prestart
	for channelNum, cdata := range row.GetChannels() {
		if channelNum >= m.GetNumChannels() {
			continue
		}

		cs := &m.channels[channelNum]
		cs.TrackData = cdata
	}

	for ch := range m.channels {
		cs := &m.channels[ch]

		cs.ActiveEffect = effect.Factory(cs.GetMemory(), cs.TrackData)
		if cs.ActiveEffect != nil {
			if m.OnEffect != nil {
				m.OnEffect(cs.ActiveEffect)
			}
			if err := intf.EffectPreStart(cs.ActiveEffect, cs, m); err != nil {
				return err
			}
		}
	}

	if err := preMixRowTxn.Commit(); err != nil {
		return err
	}

	tickDuration := tickBaseDuration / time.Duration(m.pattern.GetTempo())

	m.rowRenderState.tickDuration = tickDuration
	m.rowRenderState.samplesPerTick = int(tickDuration.Seconds() * float64(s.SampleRate))
	m.rowRenderState.ticksThisRow = m.pattern.GetTicksThisRow()
	m.rowRenderState.currentTick = 0

	// run row processing, now that prestart has completed
	for channelNum := range row.GetChannels() {
		if channelNum >= m.GetNumChannels() {
			continue
		}

		cs := &m.channels[channelNum]

		cs.AdvanceRow()
		m.processRowForChannel(cs)
	}

	return nil
}

func (m *Manager) processRowForChannel(cs *state.ChannelState) {
	mem := cs.GetMemory().(*channel.Memory)
	mem.TremorMem().Reset()

	if cs.TrackData == nil {
		return
	}

	// this can probably just be assumed to be false
	willTrigger := cs.WillTriggerOn(m.rowRenderState.currentTick)

	if cs.TrackData.HasNote() {
		cs.UseTargetPeriod = true
		inst := cs.TrackData.GetInstrument(cs.StoredSemitone)
		n := cs.TrackData.GetNote()
		if inst.IsEmpty() {
			// use current
			cs.SetTargetPos(sampling.Pos{})
		} else if !m.song.IsValidInstrumentID(inst) {
			cs.SetTargetInst(nil)
		} else {
			inst, str := m.song.GetInstrument(inst)
			n = note.CoalesceNoteSemitone(n, str)
			cs.SetTargetInst(inst)
			cs.SetTargetPos(sampling.Pos{})
			if cs.GetTargetInst() != nil {
				cs.WantVolCalc = true
			}
		}

		if note.IsEmpty(n) {
			cs.WantNoteCalc = false
			willTrigger = cs.TrackData.HasInstrument()
			if willTrigger {
				cs.SetTargetPos(sampling.Pos{})
			}
		} else if note.IsInvalid(n) {
			cs.SetTargetPeriod(nil)
			cs.WantNoteCalc = false
			willTrigger = false
		} else if note.IsRelease(n) {
			cs.SetTargetPeriod(cs.GetPeriod())
			if prevInst := cs.GetPrevInst(); prevInst != nil {
				cs.SetTargetInst(prevInst)
			}
			cs.WantNoteCalc = false
			willTrigger = false
		} else if cs.GetTargetInst() != nil {
			if nn, ok := n.(note.Normal); ok {
				cs.StoredSemitone = note.Semitone(nn)
				cs.TargetSemitone = cs.StoredSemitone
				cs.WantNoteCalc = true
			}
			willTrigger = true
		}
	} else {
		cs.WantNoteCalc = false
		cs.WantVolCalc = false
		willTrigger = false
	}

	cs.UseTargetPeriod = willTrigger
	cs.SetNotePlayTick(willTrigger, m.rowRenderState.currentTick)

	if cs.TrackData.HasVolume() {
		cs.WantVolCalc = false
		v := cs.TrackData.GetVolume()
		if v == volume.VolumeUseInstVol {
			if cs.GetTargetInst() != nil {
				cs.WantVolCalc = true
			}
		} else {
			cs.SetActiveVolume(v)
		}
	}
}
