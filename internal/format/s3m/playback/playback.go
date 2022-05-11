package playback

import (
	s3mfile "github.com/gotracker/goaudiofile/music/tracked/s3m"
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"
	device "github.com/gotracker/gosound"

	"github.com/gotracker/gotracker/internal/format/s3m/layout"
	"github.com/gotracker/gotracker/internal/format/s3m/layout/channel"
	"github.com/gotracker/gotracker/internal/format/s3m/playback/state/pattern"
	"github.com/gotracker/gotracker/internal/format/s3m/playback/util"
	"github.com/gotracker/gotracker/internal/player"
	"github.com/gotracker/gotracker/internal/player/feature"
	"github.com/gotracker/gotracker/internal/player/intf"
	"github.com/gotracker/gotracker/internal/player/output"
	"github.com/gotracker/gotracker/internal/player/state"
	"github.com/gotracker/gotracker/internal/song"
	"github.com/gotracker/gotracker/internal/song/index"
	"github.com/gotracker/gotracker/internal/song/note"
	playpattern "github.com/gotracker/gotracker/internal/song/pattern"
)

// Manager is a playback manager for S3M music
type Manager struct {
	player.Tracker

	song *layout.Song

	channels []state.ChannelState[channel.Memory, channel.Data]
	pattern  pattern.State

	preMixRowTxn  *playpattern.RowUpdateTransaction
	postMixRowTxn *playpattern.RowUpdateTransaction
	premix        *device.PremixData

	rowRenderState *rowRenderState
	OnEffect       func(intf.Effect)

	chOrder [4][]*state.ChannelState[channel.Memory, channel.Data]
}

// NewManager creates a new manager for an S3M song
func NewManager(song *layout.Song) (*Manager, error) {
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
		oc := m.GetOutputChannel(ch.OutputChannelNum, m.channelInit)

		cs := m.GetChannel(i)
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
		if mem.Shared.LowPassFilterEnable {
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

	if err := txn.Commit(); err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *Manager) channelInit(ch int) *output.Channel {
	return &output.Channel{
		ChannelNum:    ch,
		Filter:        nil,
		Config:        m,
		ChannelVolume: volume.Volume(1),
	}
}

// StartPatternTransaction returns a new row update transaction for the pattern system
func (m *Manager) StartPatternTransaction() *playpattern.RowUpdateTransaction {
	return m.pattern.StartTransaction()
}

// GetNumChannels returns the number of channels
func (m *Manager) GetNumChannels() int {
	return len(m.channels)
}

// SetNumChannels updates the song to have the specified number of channels and resets their states
func (m *Manager) SetNumChannels(num int) {
	m.channels = make([]state.ChannelState[channel.Memory, channel.Data], num)

	for ch := range m.channels {
		cs := &m.channels[ch]
		cs.ResetStates()

		cs.PortaTargetPeriod.Reset()
		cs.Trigger.Reset()
		cs.RetriggerCount = 0
		cs.SetData(nil)
		ocNum := m.song.GetOutputChannel(ch)
		cs.Output = m.GetOutputChannel(ocNum, m.channelInit)
	}
}

// SetNextOrder sets the next order index
func (m *Manager) SetNextOrder(order index.Order) error {
	if m.postMixRowTxn != nil {
		m.postMixRowTxn.SetNextOrder(order)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetNextOrder(order)
		if err := rowTxn.Commit(); err != nil {
			return err
		}
	}

	return nil
}

// SetNextRow sets the next row index
func (m *Manager) SetNextRow(row index.Row) error {
	if m.postMixRowTxn != nil {
		m.postMixRowTxn.SetNextRow(row)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetNextRow(row)
		if err := rowTxn.Commit(); err != nil {
			return err
		}
	}

	return nil
}

// SetNextRowWithBacktrack will set the next row index and backtracing allowance
func (m *Manager) SetNextRowWithBacktrack(row index.Row, allowBacktrack bool) error {
	if m.postMixRowTxn != nil {
		m.postMixRowTxn.SetNextRowWithBacktrack(row, allowBacktrack)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetNextRowWithBacktrack(row, allowBacktrack)
		if err := rowTxn.Commit(); err != nil {
			return err
		}
	}

	return nil
}

// BreakOrder breaks to the next pattern in the order
func (m *Manager) BreakOrder() error {
	if m.postMixRowTxn != nil {
		m.postMixRowTxn.BreakOrder = true
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.BreakOrder = true
		if err := rowTxn.Commit(); err != nil {
			return err
		}
	}

	return nil
}

// SetTempo sets the desired tempo for the song
func (m *Manager) SetTempo(tempo int) error {
	if m.preMixRowTxn != nil {
		m.preMixRowTxn.Tempo.Set(tempo)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.Tempo.Set(tempo)
		if err := rowTxn.Commit(); err != nil {
			return err
		}
	}

	return nil
}

// DecreaseTempo reduces the tempo by the `delta` value
func (m *Manager) DecreaseTempo(delta int) error {
	if m.preMixRowTxn != nil {
		m.preMixRowTxn.AccTempoDelta(-delta)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.AccTempoDelta(-delta)
		if err := rowTxn.Commit(); err != nil {
			return err
		}
	}

	return nil
}

// IncreaseTempo increases the tempo by the `delta` value
func (m *Manager) IncreaseTempo(delta int) error {
	if m.preMixRowTxn != nil {
		m.preMixRowTxn.AccTempoDelta(delta)
	} else {
		rowTxn := m.pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.AccTempoDelta(delta)
		if err := rowTxn.Commit(); err != nil {
			return err
		}
	}

	return nil
}

// Configure sets specified features
func (m *Manager) Configure(features []feature.Feature) error {
	if err := m.Tracker.Configure(features); err != nil {
		return err
	}
	for _, feat := range features {
		switch f := feat.(type) {
		case feature.SongLoop:
			m.pattern.SongLoop = f
		case feature.PlayUntilOrderAndRow:
			m.pattern.PlayUntilOrderAndRow = f
		}
	}
	return nil
}

// CanOrderLoop returns true if the song is allowed to order loop
func (m *Manager) CanOrderLoop() bool {
	return (m.pattern.SongLoop.Count != 0)
}

// GetSongData gets the song data object
func (m *Manager) GetSongData() song.Data {
	return m.song
}

// GetChannel returns the channel interface for the specified channel number
func (m *Manager) GetChannel(ch int) *state.ChannelState[channel.Memory, channel.Data] {
	return &m.channels[ch]
}

// GetCurrentOrder returns the current order
func (m *Manager) GetCurrentOrder() index.Order {
	return m.pattern.GetCurrentOrder()
}

// GetNumOrders returns the number of orders in the song
func (m *Manager) GetNumOrders() int {
	return m.pattern.GetNumOrders()
}

// GetCurrentRow returns the current row
func (m *Manager) GetCurrentRow() index.Row {
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

func (m *Manager) SetEnvelopePosition(v int) {
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
