package playback

import (
	"errors"
	"time"

	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"
	device "github.com/gotracker/gosound"
	"github.com/gotracker/opl2"

	"gotracker/internal/format/xm/layout"
	"gotracker/internal/format/xm/layout/channel"
	effectIntf "gotracker/internal/format/xm/playback/effect/intf"
	"gotracker/internal/format/xm/playback/sampler"
	"gotracker/internal/format/xm/playback/state/pattern"
	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/player/feature"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/state"
)

// Manager is a playback manager for XM music
type Manager struct {
	intf.Playback
	effectIntf.XM
	channel.OPL2Intf
	song *layout.Song

	channels     []state.ChannelState
	pattern      pattern.State
	globalVolume volume.Volume
	mixerVolume  volume.Volume

	preMixRowTxn  intf.SongPositionState
	postMixRowTxn intf.SongPositionState

	opl2 *opl2.Chip
	s    *sampler.Sampler
}

// NewManager creates a new manager for an XM song
func NewManager(song *layout.Song) *Manager {
	m := Manager{
		song: song,
	}

	m.pattern.Reset()
	m.pattern.Orders = song.OrderList
	m.pattern.Patterns = song.Patterns

	m.globalVolume = song.Head.GlobalVolume
	m.mixerVolume = song.Head.MixingVolume

	m.SetNumChannels(len(song.ChannelSettings))
	for i, ch := range song.ChannelSettings {
		cs := m.GetChannel(i)
		cs.SetOutputChannelNum(ch.OutputChannelNum)
		cs.SetStoredVolume(ch.InitialVolume, song.Head.GlobalVolume)
		cs.SetPan(ch.InitialPanning)
		cs.SetMemory(&song.ChannelSettings[i].Memory)
	}

	txn := m.pattern.StartTransaction()
	defer txn.Cancel()

	txn.SetTicks(song.Head.InitialSpeed)
	txn.SetTempo(song.Head.InitialTempo)

	txn.Commit()

	return &m
}

// Update updates the manager, producing premixed sound data
func (m *Manager) Update(deltaTime time.Duration, out chan<- *device.PremixData) error {
	premix, err := m.renderOneRow()
	if err != nil {
		return err
	}
	if premix != nil && premix.Data != nil && len(premix.Data) != 0 {
		out <- premix
	}

	return nil
}

// GetNumChannels returns the number of channels
func (m *Manager) GetNumChannels() int {
	return len(m.channels)
}

// SetNumChannels updates the song to have the specified number of channels and resets their states
func (m *Manager) SetNumChannels(num int) {
	m.channels = make([]state.ChannelState, num)

	for ch, cs := range m.channels {
		cs.Pos = sampling.Pos{}
		cs.PrevInstrument = nil
		cs.Instrument = nil
		cs.Period = 0
		cs.Command = nil

		cs.TargetPeriod = cs.Period
		cs.TargetPos = cs.Pos
		cs.TargetInst = nil
		cs.PortaTargetPeriod = cs.TargetPeriod
		cs.NotePlayTick = 0
		cs.RetriggerCount = 0
		cs.TremorOn = true
		cs.TremorTime = 0
		cs.VibratoDelta = 0
		cs.Cmd = nil
		cs.OutputChannelNum = m.song.GetOutputChannel(ch)
	}
}

// SetNextOrder sets the next order index
func (m *Manager) SetNextOrder(order intf.OrderIdx) {
	if m.postMixRowTxn != nil {
		m.postMixRowTxn.SetNextOrder(order)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetNextOrder(order)
		rowTxn.Commit()
	}
}

// SetNextRow sets the next row index
func (m *Manager) SetNextRow(row intf.RowIdx) {
	if m.postMixRowTxn != nil {
		m.postMixRowTxn.SetNextRow(row)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetNextRow(row)
		rowTxn.Commit()
	}
}

// SetTempo sets the desired tempo for the song
func (m *Manager) SetTempo(tempo int) {
	if m.preMixRowTxn != nil {
		m.preMixRowTxn.SetTempo(tempo)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetTempo(tempo)
		rowTxn.Commit()
	}
}

// DecreaseTempo reduces the tempo by the `delta` value
func (m *Manager) DecreaseTempo(delta int) {
	if m.preMixRowTxn != nil {
		m.preMixRowTxn.AccTempoDelta(-delta)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.AccTempoDelta(-delta)
		rowTxn.Commit()
	}
}

// IncreaseTempo increases the tempo by the `delta` value
func (m *Manager) IncreaseTempo(delta int) {
	if m.preMixRowTxn != nil {
		m.preMixRowTxn.AccTempoDelta(delta)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.AccTempoDelta(delta)
		rowTxn.Commit()
	}
}

// GetGlobalVolume returns the global volume value
func (m *Manager) GetGlobalVolume() volume.Volume {
	return m.globalVolume
}

// SetGlobalVolume sets the global volume to the specified `vol` value
func (m *Manager) SetGlobalVolume(vol volume.Volume) {
	m.globalVolume = vol
}

// DisableFeatures disables specified features
func (m *Manager) DisableFeatures(features []feature.Feature) {
	for _, f := range features {
		switch f {
		case feature.OrderLoop:
			m.pattern.OrderLoopEnabled = false
		}
	}
}

// CanOrderLoop returns true if the song is allowed to order loop
func (m *Manager) CanOrderLoop() bool {
	return m.pattern.OrderLoopEnabled
}

// GetSongData gets the song data object
func (m *Manager) GetSongData() intf.SongData {
	return m.song
}

// GetChannel returns the channel interface for the specified channel number
func (m *Manager) GetChannel(ch int) intf.Channel {
	return &m.channels[ch]
}

// GetCurrentOrder returns the current order
func (m *Manager) GetCurrentOrder() intf.OrderIdx {
	return m.pattern.GetCurrentOrder()
}

// GetNumOrders returns the number of orders in the song
func (m *Manager) GetNumOrders() int {
	return m.pattern.GetNumOrders()
}

// GetCurrentRow returns the current row
func (m *Manager) GetCurrentRow() intf.RowIdx {
	return m.pattern.GetCurrentRow()
}

// GetName returns the current song's name
func (m *Manager) GetName() string {
	return m.song.GetName()
}

// GetOPL2Chip returns the current song's OPL2 chip, if it's needed
func (m *Manager) GetOPL2Chip() *opl2.Chip {
	return m.opl2
}

// SetupSampler configures the internal sampler
func (m *Manager) SetupSampler(samplesPerSecond int, channels int, bitsPerSample int) error {
	m.s = sampler.NewSampler(samplesPerSecond, channels, bitsPerSample, util.XMBaseClock)
	if m.s == nil {
		return errors.New("NewSampler() returned nil")
	}

	return nil
}
