// Package s3m does a thing.
package s3m

import (
	"gotracker/internal/format/s3m/util"
	"gotracker/internal/player/intf"
)

// ParaPointer is a pointer offset within the S3M file format
type ParaPointer uint16

// Header is a mildly-decoded S3M header definition
type Header struct {
	Name               string
	Info               ModuleHeader
	ChannelSettings    [32]ChannelSetting
	OrderList          []uint8
	InstrumentPointers []ParaPointer
	PatternPointers    []ParaPointer
	Panning            [32]PanningFlags
}

// Song is the full definition of the song data of an Song file
type Song struct {
	intf.SongData
	Head        Header
	Instruments []Instrument
	Patterns    []Pattern
}

// GetOrderList returns the list of all pattern orders for the song
func (s *Song) GetOrderList() []uint8 {
	return s.Head.OrderList
}

// GetPatternsInterface returns an interface to all the patterns
func (s *Song) GetPatternsInterface() []intf.Pattern {
	p := make([]intf.Pattern, len(s.Patterns))
	for i, sp := range s.Patterns {
		p[i] = sp
	}
	return p
}

// GetPattern returns an interface to a specific pattern indexed by `patNum`
func (s *Song) GetPattern(patNum uint8) intf.Pattern {
	if int(patNum) >= len(s.Patterns) {
		return nil
	}
	return &s.Patterns[patNum]
}

// IsChannelEnabled returns true if the channel at index `channelNum` is enabled
func (s *Song) IsChannelEnabled(channelNum int) bool {
	return s.Head.ChannelSettings[channelNum].IsEnabled()
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

type format struct {
	intf.Format
}

var (
	// S3M is the exported interface to the S3M file loader
	S3M = format{}
)

// LoadMOD loads a MOD file and upgrades it into an S3M file internally
func LoadMOD(s intf.Song, filename string) error {
	return load(s, filename, readMOD)
}

type readerFunc func(filename string) (*Song, error)

// GetBaseClockRate returns the base clock rate for the S3M player
func (f format) GetBaseClockRate() float32 {
	return util.S3MBaseClock
}

// Load loads an S3M file into the song state `s`
func (f format) Load(s intf.Song, filename string) error {
	return load(s, filename, readS3M)
}
