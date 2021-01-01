package playback

import (
	"time"

	s3mfile "github.com/gotracker/goaudiofile/music/tracked/s3m"
	"github.com/gotracker/gomixing/mixing"
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"
	device "github.com/gotracker/gosound"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/format/s3m/playback/effect"
	"gotracker/internal/format/s3m/playback/util"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/render"
)

const (
	tickBaseDuration = time.Duration(2500) * time.Millisecond
)

// RenderOneRow renders the next single row from the song pattern data into a RowRender object
func (m *Manager) renderTick() (*device.PremixData, error) {
	postMixRowTxn := m.pattern.StartTransaction()
	defer func() {
		postMixRowTxn.Cancel()
		m.postMixRowTxn = nil
	}()
	m.postMixRowTxn = postMixRowTxn

	if m.rowRenderState == nil || m.rowRenderState.currentTick >= m.rowRenderState.ticksThisRow {
		if err := m.startNextRow(); err != nil {
			return nil, err
		}
	}

	//for ch := range m.channels {
	//	m.rowRenderState.chRrs[ch] = make(mixing.ChannelData, 1)
	//}

	finalData := &render.RowRender{}
	premix := &device.PremixData{
		Userdata:   finalData,
		SamplesLen: m.rowRenderState.samplesPerTick,
	}

	m.soundRenderTick(premix)

	finalData.Order = int(m.pattern.GetCurrentOrder())
	finalData.Row = int(m.pattern.GetCurrentRow())
	finalData.Tick = m.rowRenderState.currentTick
	if m.rowRenderState.currentTick == 0 {
		finalData.RowText = m.getRowText()
	}
	m.rowRenderState.currentTick++
	if m.rowRenderState.currentTick >= m.rowRenderState.ticksThisRow {
		postMixRowTxn.AdvanceRow()
	}

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

	if m.rowRenderState == nil {
		panmixer := m.s.GetPanMixer()
		m.rowRenderState = &rowRenderState{
			mix:           m.s.Mixer(),
			samplerSpeed:  m.s.GetSamplerSpeed(),
			panmixer:      panmixer,
			centerPanning: panmixer.GetMixingMatrix(panning.CenterAhead),
		}
	}

	tickDuration := tickBaseDuration / time.Duration(m.pattern.GetTempo())

	m.rowRenderState.tickDuration = tickDuration
	m.rowRenderState.samplesPerTick = int(tickDuration.Seconds() * float64(m.s.SampleRate))
	m.rowRenderState.ticksThisRow = m.pattern.GetTicksThisRow()
	m.rowRenderState.firstOplCh = -1
	m.rowRenderState.currentTick = 0

	for channelNum, cdata := range row.GetChannels() {
		if channelNum >= m.GetNumChannels() {
			continue
		}

		cs := &m.channels[channelNum]

		cs.ProcessRow(row, cdata, m.globalVolume, m.song, util.CalcSemitonePeriod, m.processCommand)

		cs.ActiveEffect = effect.Factory(cs.GetMemory(), cs.Cmd)
		if cs.ActiveEffect != nil {
			cs.ActiveEffect.PreStart(cs, m)
		}
	}

	preMixRowTxn.Commit()
	return nil
}

type rowRenderState struct {
	mix            *mixing.Mixer
	samplerSpeed   float32
	tickDuration   time.Duration
	samplesPerTick int
	ticksThisRow   int
	panmixer       mixing.PanMixer
	centerPanning  volume.Matrix
	firstOplCh     int

	currentTick int
}

func (m *Manager) soundRenderTick(premix *device.PremixData) {
	tick := m.rowRenderState.currentTick
	var lastTick = (tick+1 == m.rowRenderState.ticksThisRow)

	for ch := range m.channels {
		cs := &m.channels[ch]
		if m.song.IsChannelEnabled(ch) {
			chCat := m.song.ChannelSettings[ch].Category
			switch chCat {
			case s3mfile.ChannelCategoryOPL2Melody, s3mfile.ChannelCategoryOPL2Drums:
				if m.opl2 == nil {
					m.setOPL2Chip(uint32(m.s.SampleRate))
				}
				if m.rowRenderState.firstOplCh < 0 {
					m.rowRenderState.firstOplCh = ch
				}
			}

			rr := [1]mixing.Data{}
			cs.RenderRowTick(tick, lastTick, &rr[0], ch,
				m.rowRenderState.ticksThisRow,
				m.rowRenderState.mix,
				m.rowRenderState.panmixer,
				m.rowRenderState.samplerSpeed,
				m.rowRenderState.samplesPerTick,
				m.rowRenderState.centerPanning,
				m.rowRenderState.tickDuration)
			premix.Data = append(premix.Data, rr[:])
		}
	}
	if m.opl2 != nil {
		rr := [1]mixing.Data{}
		m.renderOPL2RowTick(tick, &rr[0],
			m.rowRenderState.ticksThisRow,
			m.rowRenderState.mix,
			m.rowRenderState.panmixer,
			m.rowRenderState.samplerSpeed,
			m.rowRenderState.samplesPerTick,
			m.rowRenderState.centerPanning,
			m.rowRenderState.tickDuration)
		premix.Data = append(premix.Data, rr[:])
	}

	premix.MixerVolume = m.mixerVolume
	if m.opl2 != nil {
		// make room in the mixer for the OPL2 data
		// effectively, we can do this by calculating the new number (+1) of channels from the mixer volume (channels = reciprocal of mixer volume):
		//   numChannels = (1/mv) + 1
		// then by taking the reciprocal of it:
		//   1 / numChannels
		// but that ends up being simplified to:
		//   mv / (mv + 1)
		// and we get protection from div/0 in the process - provided, of course, that the mixerVolume is not exactly -1...
		premix.MixerVolume = m.mixerVolume / (m.mixerVolume + 1)
	}
}

func (m *Manager) renderOPL2RowTick(tick int, mixerData *mixing.Data, ticksThisRow int, mix *mixing.Mixer, panmixer mixing.PanMixer, samplerSpeed float32, tickSamples int, centerPanning volume.Matrix, tickDuration time.Duration) {
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
	*mixerData = mixing.Data{
		Data:       data,
		Pan:        panning.CenterAhead,
		Volume:     util.DefaultVolume * m.globalVolume,
		SamplesLen: tickSamples,
	}
}

func (m *Manager) setOPL2Chip(rate uint32) {
	m.opl2 = channel.NewOPL2Chip(rate)
	m.opl2.WriteReg(0x01, 0x20) // enable all waveforms
	m.opl2.WriteReg(0x04, 0x00) // clear timer flags
	m.opl2.WriteReg(0x08, 0x40) // clear CSW and set NOTE-SEL
	m.opl2.WriteReg(0xBD, 0x00) // set default notes
}
