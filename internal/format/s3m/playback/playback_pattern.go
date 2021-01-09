package playback

import (
	"errors"
	"time"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/format/s3m/playback/effect"
	"gotracker/internal/player/intf"
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
		for i := range m.channels {
			mem := m.channels[i].GetMemory().(*channel.Memory)
			pl := mem.GetPatternLoop()
			pl.Count = 0
			pl.Enabled = false
		}
	}

	pat := m.song.GetPattern(patIdx)
	if pat == nil {
		return intf.ErrStopSong
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

	// generate effects and run prestart
	for channelNum, cdata := range row.GetChannels() {
		if channelNum >= m.GetNumChannels() {
			continue
		}

		cs := &m.channels[channelNum]

		cs.TrackData = cdata

		cs.ActiveEffect = effect.Factory(cs.GetMemory(), cs.TrackData)
		if cs.ActiveEffect != nil {
			if m.OnEffect != nil {
				m.OnEffect(cs.ActiveEffect)
			}
			cs.ActiveEffect.PreStart(cs, m)
		}
	}

	preMixRowTxn.Commit()

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

		m.processRowForChannel(cs)
		cs.Process(row, m.GetGlobalVolume(), m.song, m.processCommand)
	}

	return nil
}

func (m *Manager) processRowForChannel(cs intf.Channel) {
	mem := cs.GetMemory().(*channel.Memory)
	mem.TremorMem().Reset()
}
