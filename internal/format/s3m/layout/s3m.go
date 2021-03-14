package layout

import (
	s3mfile "github.com/gotracker/goaudiofile/music/tracked/s3m"
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/index"
	"gotracker/internal/instrument"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/pattern"
)

// Header is a mildly-decoded S3M header definition
type Header struct {
	Name         string
	InitialSpeed int
	InitialTempo int
	GlobalVolume volume.Volume
	MixingVolume volume.Volume
	Stereo       bool
}

// ChannelSetting is settings specific to a single channel
type ChannelSetting struct {
	Enabled          bool
	OutputChannelNum int
	Category         s3mfile.ChannelCategory
	InitialVolume    volume.Volume
	InitialPanning   panning.Position
	Memory           channel.Memory
}

// Song is the full definition of the song data of an Song file
type Song struct {
	intf.SongData
	Head            Header
	Instruments     []instrument.Instrument
	Patterns        []pattern.Pattern
	ChannelSettings []ChannelSetting
	OrderList       []index.Pattern
}

// GetOrderList returns the list of all pattern orders for the song
func (s *Song) GetOrderList() []index.Pattern {
	return s.OrderList
}

// GetPattern returns an interface to a specific pattern indexed by `patNum`
func (s *Song) GetPattern(patNum index.Pattern) intf.Pattern {
	if int(patNum) >= len(s.Patterns) {
		return nil
	}
	return &s.Patterns[patNum]
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

// IsValidInstrumentID returns true if the instrument exists
func (s *Song) IsValidInstrumentID(instNum intf.InstrumentID) bool {
	if instNum.IsEmpty() {
		return false
	}
	switch id := instNum.(type) {
	case channel.S3MInstrumentID:
		iid := int(id)
		return iid > 0 && iid <= len(s.Instruments)
	}
	return false
}

// GetInstrument returns the instrument interface indexed by `instNum` (0-based)
func (s *Song) GetInstrument(instID intf.InstrumentID) (intf.Instrument, note.Semitone) {
	if instID.IsEmpty() {
		return nil, note.UnchangedSemitone
	}
	switch id := instID.(type) {
	case channel.S3MInstrumentID:
		return &s.Instruments[int(id)-1], note.UnchangedSemitone
	}

	return nil, note.UnchangedSemitone
}

// GetName returns the name of the song
func (s *Song) GetName() string {
	return s.Head.Name
}
