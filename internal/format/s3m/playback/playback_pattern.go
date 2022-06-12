package playback

import (
	"errors"
	"time"

	"github.com/gotracker/gotracker/internal/format/s3m/layout/channel"
	"github.com/gotracker/gotracker/internal/format/s3m/playback/effect"
	"github.com/gotracker/gotracker/internal/player/intf"
	"github.com/gotracker/gotracker/internal/player/state"
	"github.com/gotracker/gotracker/internal/song"
	"github.com/gotracker/gotracker/internal/song/note"

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
			mem := cs.GetMemory()
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
		mem := cs.GetMemory()
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
			RenderDetails: state.RenderDetails{
				Mix:          s.Mixer(),
				SamplerSpeed: s.GetSamplerSpeed(),
				Panmixer:     panmixer,
			},
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
		cs.SetData(nil)
		if resetMemory {
			mem := cs.GetMemory()
			mem.StartOrder()
		}
	}

	// generate effects and run prestart
	channels := row.GetChannels()
	for channelNum := range channels {
		if channelNum >= m.GetNumChannels() {
			continue
		}

		cdata := &channels[channelNum]

		cs := &m.channels[channelNum]
		cs.SetData(cdata)
	}

	for ch := range m.channels {
		cs := &m.channels[ch]

		cs.SetActiveEffect(effect.Factory(cs.GetMemory(), cs.GetData()))
		if cs.GetActiveEffect() != nil {
			if m.OnEffect != nil {
				m.OnEffect(cs.GetActiveEffect())
			}
			if err := intf.EffectPreStart[channel.Memory, channel.Data](cs.GetActiveEffect(), cs, m); err != nil {
				return err
			}
		}
	}

	if err := preMixRowTxn.Commit(); err != nil {
		return err
	}

	tickDuration := tickBaseDuration / time.Duration(m.pattern.GetTempo())

	m.rowRenderState.Duration = tickDuration
	m.rowRenderState.Samples = int(tickDuration.Seconds() * float64(s.SampleRate))
	m.rowRenderState.ticksThisRow = m.pattern.GetTicksThisRow()
	m.rowRenderState.currentTick = 0

	for _, order := range m.chOrder {
		for _, cs := range order {
			if cs == nil {
				continue
			}

			cs.AdvanceRow()
			m.processRowForChannel(cs)
		}
	}

	return nil
}

func (m *Manager) processRowForChannel(cs *state.ChannelState[channel.Memory, channel.Data]) {
	mem := cs.GetMemory()
	mem.TremorMem().Reset()

	if cs.GetData() == nil {
		return
	}

	// this can probably just be assumed to be false
	targetTick, retrigger := cs.WillTriggerOn(m.rowRenderState.currentTick)

	var (
		wantVolCalc  bool
		wantNoteCalc bool
		noteCalcST   note.Semitone
	)
	if cs.GetData().HasNote() {
		cs.UseTargetPeriod = targetTick
		inst := cs.GetData().GetInstrument(cs.StoredSemitone)
		n := cs.GetData().GetNote()
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
				wantVolCalc = true
			}
		}

		if note.IsEmpty(n) {
			wantNoteCalc = false
			targetTick = cs.GetData().HasInstrument()
			if targetTick {
				cs.SetTargetPos(sampling.Pos{})
			}
		} else if note.IsInvalid(n) {
			cs.SetTargetPeriod(nil)
			wantNoteCalc = false
			targetTick = false
		} else if note.IsRelease(n) {
			cs.SetTargetPeriod(cs.GetPeriod())
			if prevInst := cs.GetPrevInst(); prevInst != nil {
				cs.SetTargetInst(prevInst)
			}
			wantNoteCalc = false
			targetTick = false
		} else if cs.GetTargetInst() != nil {
			if nn, ok := n.(note.Normal); ok {
				cs.StoredSemitone = note.Semitone(nn)
				noteCalcST = cs.StoredSemitone
				wantNoteCalc = true
			}
			targetTick = true
			retrigger = true
		}
	} else {
		wantNoteCalc = false
		wantVolCalc = false
		targetTick = false
	}

	cs.UseTargetPeriod = targetTick
	cs.SetNotePlayTick(targetTick, retrigger, m.rowRenderState.currentTick)

	if cs.GetData().HasVolume() {
		wantVolCalc = false
		v := cs.GetData().GetVolume()
		if v == volume.VolumeUseInstVol {
			if cs.GetTargetInst() != nil {
				wantVolCalc = true
			}
		} else {
			cs.SetActiveVolume(v)
		}
	}

	if wantVolCalc {
		cs.VolOps = append(cs.VolOps, doVolCalc{})
	}

	if wantNoteCalc {
		cs.NoteOps = append(cs.NoteOps, m.semitoneSetterFactory(noteCalcST, cs.SetTargetPeriod))
	}
}
