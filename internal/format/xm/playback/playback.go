package playback

import (
	device "github.com/gotracker/gosound"

	"gotracker/internal/format/xm/layout"
	effectIntf "gotracker/internal/format/xm/playback/effect/intf"
	"gotracker/internal/format/xm/playback/state/pattern"
	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/player"
	"gotracker/internal/player/feature"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/state"
)

// Manager is a playback manager for XM music
type Manager struct {
	player.Tracker
	effectIntf.XM

	song *layout.Song

	channels []state.ChannelState
	pattern  pattern.State

	preMixRowTxn  intf.SongPositionState
	postMixRowTxn intf.SongPositionState
	premix        *device.PremixData

	rowRenderState *rowRenderState
	OnEffect       func(intf.Effect)
}

// NewManager creates a new manager for an XM song
func NewManager(song *layout.Song) *Manager {
	m := Manager{
		Tracker: player.Tracker{
			BaseClockRate: util.XMBaseClock,
		},
		song: song,
	}

	m.Tracker.Tickable = &m
	m.Tracker.Premixable = &m

	m.pattern.Reset()
	m.pattern.Orders = song.OrderList
	m.pattern.Patterns = song.Patterns

	m.SetGlobalVolume(song.Head.GlobalVolume)
	m.SetMixerVolume(song.Head.MixingVolume)

	m.SetNumChannels(len(song.ChannelSettings))
	for i, ch := range song.ChannelSettings {
		oc := m.GetOutputChannel(ch.OutputChannelNum, &m)

		cs := m.GetChannel(i)
		cs.SetOutputChannel(oc)
		cs.SetGlobalVolume(m.GetGlobalVolume())
		cs.SetActiveVolume(ch.InitialVolume)
		cs.SetPanEnabled(true)
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

// GetNumChannels returns the number of channels
func (m *Manager) GetNumChannels() int {
	return len(m.channels)
}

// SetNumChannels updates the song to have the specified number of channels and resets their states
func (m *Manager) SetNumChannels(num int) {
	m.channels = make([]state.ChannelState, num)

	for ch := range m.channels {
		cs := &m.channels[ch]
		cs.ResetStates()

		cs.PortaTargetPeriod = nil
		cs.NotePlayTick = 0
		cs.RetriggerCount = 0
		cs.TrackData = nil
		ocNum := m.song.GetOutputChannel(ch)
		cs.Output = m.GetOutputChannel(ocNum, m)
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
func (m *Manager) SetNextRow(row intf.RowIdx, opts ...bool) {
	if m.postMixRowTxn != nil {
		m.postMixRowTxn.SetNextRow(row, opts...)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetNextRow(row, opts...)
		rowTxn.Commit()
	}
}

// BreakOrder breaks to the next pattern in the order
func (m *Manager) BreakOrder() {
	if m.postMixRowTxn != nil {
		m.postMixRowTxn.BreakOrder()
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.BreakOrder()
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

// DisableFeatures disables specified features
func (m *Manager) DisableFeatures(features []feature.Feature) {
	m.Tracker.DisableFeatures(features)
	for _, f := range features {
		switch f {
		case feature.OrderLoop:
			m.pattern.SongLoopEnabled = false
		}
	}
}

// CanOrderLoop returns true if the song is allowed to order loop
func (m *Manager) CanOrderLoop() bool {
	return m.pattern.SongLoopEnabled
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

// SetOnEffect sets the callback for an effect being generated for a channel
func (m *Manager) SetOnEffect(fn func(intf.Effect)) {
	m.OnEffect = fn
}
