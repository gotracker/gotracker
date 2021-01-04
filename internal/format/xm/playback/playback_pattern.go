package playback

import (
	"errors"
	"time"

	"github.com/gotracker/gomixing/panning"

	"gotracker/internal/format/xm/playback/effect"
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

		fullVolume := panmixer.GetMixingMatrix(panning.CenterAhead)
		// we don't want to apply a double panning, so make sure it's full volume
		for i := range fullVolume {
			fullVolume[i] = 1.0
		}

		m.rowRenderState = &rowRenderState{
			mix:           s.Mixer(),
			samplerSpeed:  s.GetSamplerSpeed(),
			panmixer:      panmixer,
			centerPanning: fullVolume,
		}
	}

	// generate effects and run prestart
	for channelNum, cdata := range row.GetChannels() {
		if channelNum >= m.GetNumChannels() {
			continue
		}

		cs := &m.channels[channelNum]

		cs.Cmd = cdata

		cs.ActiveEffect = effect.Factory(cs.GetMemory(), cs.Cmd)
		if cs.ActiveEffect != nil {
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

		cs.Process(row, m.GetGlobalVolume(), m.song, m.processEffect)
	}

	return nil
}
