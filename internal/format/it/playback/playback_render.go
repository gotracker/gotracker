package playback

import (
	"time"

	"github.com/gotracker/gomixing/mixing"
	device "github.com/gotracker/gosound"

	"gotracker/internal/player/render"
)

// OnTick runs the IT tick processing
func (m *Manager) OnTick() error {
	m.premix = nil
	premix, err := m.renderTick()
	if err != nil {
		return err
	}

	m.premix = premix
	return nil
}

// GetPremixData gets the current premix data from the manager
func (m *Manager) GetPremixData() (*device.PremixData, error) {
	return m.premix, nil
}

// RenderOneRow renders the next single row from the song pattern data into a RowRender object
func (m *Manager) renderTick() (*device.PremixData, error) {
	postMixRowTxn := m.pattern.StartTransaction()
	defer func() {
		postMixRowTxn.Cancel()
		m.postMixRowTxn = nil
	}()
	m.postMixRowTxn = postMixRowTxn

	if m.rowRenderState == nil || m.rowRenderState.currentTick >= m.rowRenderState.ticksThisRow {
		if err := m.processPatternRow(); err != nil {
			return nil, err
		}
	}

	finalData := &render.RowRender{}
	premix := &device.PremixData{
		Userdata:   finalData,
		SamplesLen: m.rowRenderState.samplesPerTick,
	}

	if err := m.soundRenderTick(premix); err != nil {
		return nil, err
	}

	finalData.Order = int(m.pattern.GetCurrentOrder())
	finalData.Row = int(m.pattern.GetCurrentRow())
	finalData.Tick = m.rowRenderState.currentTick
	if m.rowRenderState.currentTick == 0 {
		finalData.RowText = m.getRowText()
	}

	m.rowRenderState.currentTick++
	if m.rowRenderState.currentTick >= m.rowRenderState.ticksThisRow {
		postMixRowTxn.AdvanceRow = true
	}

	postMixRowTxn.Commit()
	return premix, nil
}

type rowRenderState struct {
	mix            *mixing.Mixer
	samplerSpeed   float32
	tickDuration   time.Duration
	samplesPerTick int
	ticksThisRow   int
	panmixer       mixing.PanMixer

	currentTick int
}

func (m *Manager) soundRenderTick(premix *device.PremixData) error {
	tick := m.rowRenderState.currentTick
	var lastTick = (tick+1 == m.rowRenderState.ticksThisRow)

	for ch := range m.channels {
		cs := &m.channels[ch]
		if m.song.IsChannelEnabled(ch) {

			m.processEffect(ch, cs, tick, lastTick)

			rr, err := cs.RenderRowTick(m.rowRenderState.mix,
				m.rowRenderState.panmixer,
				m.rowRenderState.samplerSpeed,
				m.rowRenderState.samplesPerTick,
				m.rowRenderState.tickDuration)
			if err != nil {
				return err
			}
			if rr != nil {
				premix.Data = append(premix.Data, rr)
			}
		}
	}

	premix.MixerVolume = m.GetMixerVolume()
	return nil
}

/** unused in IT, so far
func (m *Manager) ensureOPL2() {
	if opl2 := m.GetOPL2Chip(); opl2 == nil {
		if s := m.GetSampler(); s != nil {
			opl2 = render.NewOPL2Chip(uint32(s.SampleRate))
			opl2.WriteReg(0x01, 0x20) // enable all waveforms
			opl2.WriteReg(0x04, 0x00) // clear timer flags
			opl2.WriteReg(0x08, 0x40) // clear CSW and set NOTE-SEL
			opl2.WriteReg(0xBD, 0x00) // set default notes
			m.SetOPL2Chip(opl2)
		}
	}
}
*/
