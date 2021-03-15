package layout

import (
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/song"
	"gotracker/internal/song/index"
	"gotracker/internal/song/instrument"
	"gotracker/internal/song/note"
	"gotracker/internal/song/pattern"
)

// Header is a mildly-decoded XM header definition
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
	song.Data
	Head              Header
	Instruments       map[uint8]*instrument.Instrument
	InstrumentNoteMap map[uint8]map[note.Semitone]*instrument.Instrument
	Patterns          []pattern.Pattern
	ChannelSettings   []ChannelSetting
	OrderList         []index.Pattern
}

// GetOrderList returns the list of all pattern orders for the song
func (s *Song) GetOrderList() []index.Pattern {
	return s.OrderList
}

// GetPattern returns an interface to a specific pattern indexed by `patNum`
func (s *Song) GetPattern(patNum index.Pattern) song.Pattern {
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
func (s *Song) IsValidInstrumentID(instNum instrument.InstrumentID) bool {
	if instNum.IsEmpty() {
		return false
	}
	switch id := instNum.(type) {
	case channel.SampleID:
		_, ok := s.Instruments[id.InstID]
		return ok
	}
	return false
}

// GetInstrument returns the instrument interface indexed by `instNum` (0-based)
func (s *Song) GetInstrument(instNum instrument.InstrumentID) (*instrument.Instrument, note.Semitone) {
	if instNum.IsEmpty() {
		return nil, note.UnchangedSemitone
	}
	switch id := instNum.(type) {
	case channel.SampleID:
		if nm, ok1 := s.InstrumentNoteMap[id.InstID]; ok1 {
			if sm, ok2 := nm[id.Semitone]; ok2 {
				return sm, note.UnchangedSemitone
			}
		}
	}
	return nil, note.UnchangedSemitone
}

// GetName returns the name of the song
func (s *Song) GetName() string {
	return s.Head.Name
}
