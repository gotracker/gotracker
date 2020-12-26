package playback

import (
	"time"

	s3mfile "github.com/gotracker/goaudiofile/music/tracked/s3m"
	"github.com/gotracker/gomixing/mixing"
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"
	device "github.com/gotracker/gosound"

	"gotracker/internal/format/s3m/playback/effect"
	"gotracker/internal/format/s3m/playback/opl2"
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

	premix.SamplesLen = samplesThisRow

	chRrs := make([][]mixing.Data, len(m.channels))
	for ch := range m.channels {
		chRrs[ch] = make([]mixing.Data, ticksThisRow)
	}

	firstOplCh := -1
	for tick := 0; tick < ticksThisRow; tick++ {
		var lastTick = (tick+1 == ticksThisRow)

		for ch := range m.channels {
			cs := &m.channels[ch]
			if m.song.IsChannelEnabled(ch) {
				chCat := m.song.ChannelSettings[ch].Category
				switch chCat {
				case s3mfile.ChannelCategoryOPL2Melody, s3mfile.ChannelCategoryOPL2Drums:
					if m.opl2 == nil {
						m.setOPL2Chip(uint32(sampler.SampleRate))
					}
					if firstOplCh < 0 {
						firstOplCh = ch
					}
				}

				rr := chRrs[ch]
				cs.RenderRowTick(tick, lastTick, rr, ch, ticksThisRow, mix, panmixer, samplerSpeed, samplesPerTick, centerPanning, tickDuration)

				switch chCat {
				case s3mfile.ChannelCategoryPCMLeft, s3mfile.ChannelCategoryPCMRight:
					for len(premix.Data) <= ch {
						premix.Data = append(premix.Data, nil)
					}
					premix.Data[ch] = rr
				}
			}
		}
		if m.opl2 != nil {
			ch := firstOplCh
			for len(premix.Data) <= ch {
				premix.Data = append(premix.Data, nil)
			}
			rr := chRrs[ch]
			m.renderOPL2RowTick(tick, rr, ticksThisRow, mix, panmixer, samplerSpeed, samplesPerTick, centerPanning, tickDuration)
			premix.Data[ch] = rr
		}
	}
}

func (m *Manager) renderOPL2RowTick(tick int, mixerData []mixing.Data, ticksThisRow int, mix *mixing.Mixer, panmixer mixing.PanMixer, samplerSpeed float32, tickSamples int, centerPanning volume.Matrix, tickDuration time.Duration) {
	// make a stand-alone data buffer for this channel for this tick
	data := mix.NewMixBuffer(tickSamples)

	opl2data := make([]int32, tickSamples)

	m.opl2.GenerateBlock2(uint(tickSamples), opl2data)

	for i, s := range opl2data {
		sv := volume.Volume(s) / 32768.0
		for c := range data {
			data[c][i] = sv
		}
	}
	mixerData[tick] = mixing.Data{
		Data:       data,
		Pan:        panning.CenterAhead,
		Volume:     util.DefaultVolume * m.globalVolume,
		SamplesLen: tickSamples,
	}
}

func (m *Manager) setOPL2Chip(rate uint32) {
	m.opl2 = opl2.NewChip(rate, false)
	m.opl2.WriteReg(0x01, 0x20) // enable all waveforms
	m.opl2.WriteReg(0x08, 0x00) // clear CSW and NOTE-SEL
	m.opl2.WriteReg(0xBD, 0x00) // set default notes
}
