package playback

import (
	s3mfile "github.com/gotracker/goaudiofile/music/tracked/s3m"
	"github.com/gotracker/gomixing/panning"
	device "github.com/gotracker/gosound"

	"gotracker/internal/format/s3m/layout"
	effectIntf "gotracker/internal/format/s3m/playback/effect/intf"
	"gotracker/internal/format/s3m/playback/state/pattern"
	"gotracker/internal/format/s3m/playback/util"
	"gotracker/internal/player"
	"gotracker/internal/player/feature"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	playpattern "gotracker/internal/player/pattern"
	"gotracker/internal/player/state"
)

// Manager is a playback manager for S3M music
type Manager struct {
	player.Tracker
	effectIntf.S3M

	song *layout.Song

	channels []state.ChannelState
	pattern  pattern.State

	preMixRowTxn  *playpattern.RowUpdateTransaction
	postMixRowTxn *playpattern.RowUpdateTransaction

	premix *device.PremixData

	rowRenderState *rowRenderState
	OnEffect       func(intf.Effect)

	chOrder [4][]intf.Channel
}

// NewManager creates a new manager for an S3M song
func NewManager(song *layout.Song) *Manager {
	m := Manager{
		Tracker: player.Tracker{
			BaseClockRate: util.S3MBaseClock,
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
	lowpassEnabled := false
	for i, ch := range song.ChannelSettings {
		oc := m.GetOutputChannel(ch.OutputChannelNum, &m)

		cs := m.GetChannel(i).(*state.ChannelState)
		cs.SetOutputChannel(oc)
		cs.SetGlobalVolume(m.GetGlobalVolume())
		cs.SetActiveVolume(ch.InitialVolume)
		if song.Head.Stereo {
			cs.SetPanEnabled(true)
			cs.SetPan(ch.InitialPanning)
		} else {
			cs.SetPanEnabled(true)
			cs.SetPan(panning.CenterAhead)
			cs.SetPanEnabled(false)
		}
		cs.SetStoredSemitone(note.UnchangedSemitone)
		mem := &song.ChannelSettings[i].Memory
		cs.SetMemory(mem)
		if mem.LowPassFilterEnable {
			lowpassEnabled = true
		}

		// weirdly, S3M processes channels in channel category order
		// so we have to make a list with the order we're expecting
		switch s3mfile.ChannelCategory(ch.Category) {
		case s3mfile.ChannelCategoryUnknown:
			// do nothing
		default:
			cIdx := int(ch.Category) - 1
			m.chOrder[cIdx] = append(m.chOrder[cIdx], cs)
		}
	}

	if lowpassEnabled {
		m.SetFilterEnable(true)
	}

	txn := m.pattern.StartTransaction()
	defer txn.Cancel()

	txn.Ticks.Set(song.Head.InitialSpeed)
	txn.Tempo.Set(song.Head.InitialTempo)

	txn.Commit()

	return &m
}

// SetupSampler configures the internal sampler
func (m *Manager) SetupSampler(samplesPerSecond int, channels int, bitsPerSample int) error {
	if err := m.Tracker.SetupSampler(samplesPerSecond, channels, bitsPerSample); err != nil {
		return err
	}

	oplLen := len(m.chOrder[int(s3mfile.ChannelCategoryOPL2Melody)-1])
	oplLen += len(m.chOrder[int(s3mfile.ChannelCategoryOPL2Drums)-1])

	if oplLen > 0 {
		m.ensureOPL2()
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

	for ch := range m.channels {
		cs := &m.channels[ch]
		cs.ResetStates()

		cs.PortaTargetPeriod = nil
		cs.Trigger = nil
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

// SetTempo sets the desired tempo for the song
func (m *Manager) SetTempo(tempo int) {
	if m.preMixRowTxn != nil {
		m.preMixRowTxn.Tempo.Set(tempo)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.Tempo.Set(tempo)
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

// Configure sets specified features
func (m *Manager) Configure(features []feature.Feature) {
	m.Tracker.Configure(features)
	for _, feat := range features {
		switch f := feat.(type) {
		case feature.SongLoop:
			m.pattern.SongLoop = f
		}
	}
}

// CanOrderLoop returns true if the song is allowed to order loop
func (m *Manager) CanOrderLoop() bool {
	return m.pattern.SongLoop.Enabled
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
