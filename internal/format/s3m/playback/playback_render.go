package playback

import (
	"time"

	"github.com/gotracker/gomixing/mixing"
	"github.com/gotracker/gomixing/panning"
	device "github.com/gotracker/gosound"

	"gotracker/internal/format/s3m/playback/effect"
	"gotracker/internal/format/s3m/playback/util"
	"gotracker/internal/player/render"
	"gotracker/internal/player/state/pattern"
)

// RenderOneRow renders the next single row from the song pattern data into a RowRender object
func (m *Manager) renderOneRow(sampler *render.Sampler) (*device.PremixData, error) {
	preMixRowTxn := m.pattern.StartTransaction()
	postMixRowTxn := m.pattern.StartTransaction()
	defer func() {
		preMixRowTxn.Cancel()
		m.preMixRowTxn = nil
		postMixRowTxn.Cancel()
		m.postMixRowTxn = nil
	}()
	m.preMixRowTxn = preMixRowTxn
	m.postMixRowTxn = postMixRowTxn

	if err := m.startNextRow(); err != nil {
		return nil, err
	}

	preMixRowTxn.Commit()

	finalData := &render.RowRender{}
	premix := &device.PremixData{
		Userdata: finalData,
	}

	m.soundRenderRow(premix, sampler)

	finalData.Order = int(m.pattern.GetCurrentOrder())
	finalData.Row = int(m.pattern.GetCurrentRow())
	finalData.RowText = m.getRowText()

	postMixRowTxn.AdvanceRow()

	postMixRowTxn.Commit()
	return premix, nil
}

func (m *Manager) startNextRow() error {
	patIdx, err := m.pattern.GetCurrentPatternIdx()
	if err != nil {
		return err
	}

	pat := m.song.GetPattern(patIdx)
	if pat == nil {
		return pattern.ErrStopSong
	}

	rows := pat.GetRows()

	myCurrentRow := m.pattern.GetCurrentRow()

	row := rows[myCurrentRow]
	for channelNum, channel := range row.GetChannels() {
		if channelNum >= m.GetNumChannels() {
			continue
		}

		cs := &m.channels[channelNum]

		cs.ProcessRow(row, channel, m.globalVolume, m.song, util.CalcSemitonePeriod, m.processCommand)

		cs.ActiveEffect = effect.Factory(cs.GetMemory(), cs.Cmd)
		if cs.ActiveEffect != nil {
			cs.ActiveEffect.PreStart(cs, m)
		}
	}

	return nil
}

func (m *Manager) soundRenderRow(premix *device.PremixData, sampler *render.Sampler) {
	mix := sampler.Mixer()

	samplerSpeed := sampler.GetSamplerSpeed()
	tickDuration := time.Duration(2500) * time.Millisecond / time.Duration(m.pattern.GetTempo())
	samplesPerTick := int(tickDuration.Seconds() * float64(sampler.SampleRate))

	ticksThisRow := m.pattern.GetTicksThisRow()

	samplesThisRow := int(ticksThisRow) * samplesPerTick

	panmixer := sampler.GetPanMixer()

	centerPanning := panmixer.GetMixingMatrix(panning.CenterAhead)

	for len(premix.Data) < len(m.channels) {
		premix.Data = append(premix.Data, nil)
	}
	premix.SamplesLen = samplesThisRow

	for ch := range m.channels {
		cs := &m.channels[ch]
		if m.song.IsChannelEnabled(ch) {
			rr := make([]mixing.Data, ticksThisRow)
			cs.RenderRow(rr, ch, ticksThisRow, mix, panmixer, samplerSpeed, samplesPerTick, centerPanning, tickDuration)

			premix.Data[ch] = rr
		}
	}
}
