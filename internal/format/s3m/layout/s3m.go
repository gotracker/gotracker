package layout

import (
	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/format/s3m/playback/filter"
	"gotracker/internal/player/intf"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"
)

// Header is a mildly-decoded S3M header definition
type Header struct {
	Name         string
	InitialSpeed int
	InitialTempo int
	GlobalVolume volume.Volume
	MixingVolume volume.Volume
}

// ChannelSetting is settings specific to a single channel
type ChannelSetting struct {
	Enabled          bool
	OutputChannelNum int
	InitialVolume    volume.Volume
	InitialPanning   panning.Position
	Memory           channel.Memory
}

// Song is the full definition of the song data of an Song file
type Song struct {
	intf.SongData
	Head            Header
	Instruments     []Instrument
	Patterns        []intf.Pattern
	ChannelSettings []ChannelSetting
	OrderList       []intf.PatternIdx
}

// GetOrderList returns the list of all pattern orders for the song
func (s *Song) GetOrderList() []intf.PatternIdx {
	return s.OrderList
}

// GetPattern returns an interface to a specific pattern indexed by `patNum`
func (s *Song) GetPattern(patNum intf.PatternIdx) intf.Pattern {
	if int(patNum) >= len(s.Patterns) {
		return nil
	}
	return s.Patterns[patNum]
}

// IsChannelEnabled returns true if the channel at index `channelNum` is enabled
func (s *Song) IsChannelEnabled(channelNum int) bool {
	return s.ChannelSettings[channelNum].Enabled
}

// GetOutputChannel returns the output channel for the channel at index `channelNum`
func (s *Song) GetOutputChannel(channelNum int) int {
	return s.ChannelSettings[channelNum].OutputChannelNum
}

// NumInstruments returns the number of instruments in the song
func (s *Song) NumInstruments() int {
	return len(s.Instruments)
}

// GetInstrument returns the instrument interface indexed by `instNum` (0-based)
func (s *Song) GetInstrument(instNum int) intf.Instrument {
	return &s.Instruments[instNum]
}

// GetName returns the name of the song
func (s *Song) GetName() string {
	return s.Head.Name
}

// SetFilterEnable activates or deactivates the amiga low-pass filter on the instruments
func (s *Song) SetFilterEnable(on bool, ss intf.Song) {
	for i := range s.ChannelSettings {
		c := ss.GetChannel(i)
		if on {
			if c.GetFilter() == nil {
				c.SetFilter(filter.NewAmigaLPF())
			}
		} else {
			c.SetFilter(nil)
		}
	}
}
