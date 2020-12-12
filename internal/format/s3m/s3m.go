// Package s3m does a thing.
package s3m

import (
	"gotracker/internal/format/s3m/channel"
	"gotracker/internal/format/s3m/util"
	"gotracker/internal/player/intf"
)

// ChannelID is the S3M value for a channel specification (found within the ChanenlSetting header block)
type ChannelID uint8

const (
	// ChannelIDL1 is the Left Channel 1
	ChannelIDL1 = ChannelID(0 + iota)
	// ChannelIDL2 is the Left Channel 2
	ChannelIDL2
	// ChannelIDL3 is the Left Channel 3
	ChannelIDL3
	// ChannelIDL4 is the Left Channel 4
	ChannelIDL4
	// ChannelIDL5 is the Left Channel 5
	ChannelIDL5
	// ChannelIDL6 is the Left Channel 6
	ChannelIDL6
	// ChannelIDL7 is the Left Channel 7
	ChannelIDL7
	// ChannelIDL8 is the Left Channel 8
	ChannelIDL8
	// ChannelIDR1 is the Right Channel 1
	ChannelIDR1
	// ChannelIDR2 is the Right Channel 2
	ChannelIDR2
	// ChannelIDR3 is the Right Channel 3
	ChannelIDR3
	// ChannelIDR4 is the Right Channel 4
	ChannelIDR4
	// ChannelIDR5 is the Right Channel 5
	ChannelIDR5
	// ChannelIDR6 is the Right Channel 6
	ChannelIDR6
	// ChannelIDR7 is the Right Channel 7
	ChannelIDR7
	// ChannelIDR8 is the Right Channel 8
	ChannelIDR8
	// ChannelIDOPL2Melody1 is the Adlib/OPL2 Melody Channel 1
	ChannelIDOPL2Melody1
	// ChannelIDOPL2Melody2 is the Adlib/OPL2 Melody Channel 2
	ChannelIDOPL2Melody2
	// ChannelIDOPL2Melody3 is the Adlib/OPL2 Melody Channel 3
	ChannelIDOPL2Melody3
	// ChannelIDOPL2Melody4 is the Adlib/OPL2 Melody Channel 4
	ChannelIDOPL2Melody4
	// ChannelIDOPL2Melody5 is the Adlib/OPL2 Melody Channel 5
	ChannelIDOPL2Melody5
	// ChannelIDOPL2Melody6 is the Adlib/OPL2 Melody Channel 6
	ChannelIDOPL2Melody6
	// ChannelIDOPL2Melody7 is the Adlib/OPL2 Melody Channel 7
	ChannelIDOPL2Melody7
	// ChannelIDOPL2Melody8 is the Adlib/OPL2 Melody Channel 8
	ChannelIDOPL2Melody8
	// ChannelIDOPL2Drums1 is the Adlib/OPL2 Drums Channel 1
	ChannelIDOPL2Drums1
	// ChannelIDOPL2Drums2 is the Adlib/OPL2 Drums Channel 2
	ChannelIDOPL2Drums2
	// ChannelIDOPL2Drums3 is the Adlib/OPL2 Drums Channel 3
	ChannelIDOPL2Drums3
	// ChannelIDOPL2Drums4 is the Adlib/OPL2 Drums Channel 4
	ChannelIDOPL2Drums4
	// ChannelIDOPL2Drums5 is the Adlib/OPL2 Drums Channel 5
	ChannelIDOPL2Drums5
	// ChannelIDOPL2Drums6 is the Adlib/OPL2 Drums Channel 6
	ChannelIDOPL2Drums6
	// ChannelIDOPL2Drums7 is the Adlib/OPL2 Drums Channel 7
	ChannelIDOPL2Drums7
	// ChannelIDOPL2Drums8 is the Adlib/OPL2 Drums Channel 8
	ChannelIDOPL2Drums8
)

// ParaPointer is a pointer offset within the S3M file format
type ParaPointer uint16

// ChannelSetting is a full channel setting (with flags) definition from the S3M header
type ChannelSetting uint8

// IsEnabled returns the enabled flag (bit 7 is unset)
func (cs ChannelSetting) IsEnabled() bool {
	return (uint8(cs) & 0x80) == 0
}

// GetChannel returns the ChannelID for the channel
func (cs ChannelSetting) GetChannel() ChannelID {
	ch := uint8(cs) & 0x7F
	return ChannelID(ch)
}

// IsPCM returns true if the channel is one of the left or right channels (non-Adlib/OPL2)
func (cs ChannelSetting) IsPCM() bool {
	ch := uint8(cs) & 0x7F
	return (ch < 16)
}

// IsOPL2 returns true if the channel is an Adlib/OPL2 channel (non-PCM)
func (cs ChannelSetting) IsOPL2() bool {
	ch := uint8(cs) & 0x7F
	return (ch >= 16)
}

// PanningFlags is a flagset and panning value for the panning system
type PanningFlags uint8

const (
	// PanningFlagValid is the flag used to determine that the panning value is valid
	PanningFlagValid = PanningFlags(0x20)
)

// IsValid returns true if bit 5 is set
func (pf PanningFlags) IsValid() bool {
	return uint8(pf&PanningFlagValid) != 0
}

// Value returns the panning position (0=full left, 15=full right)
func (pf PanningFlags) Value() uint8 {
	return uint8(pf) & 0x0F
}

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

// RowData is the data for each row
type RowData struct {
	intf.Row
	Channels [32]channel.Data
}

// GetChannels returns an interface to all the channels in the row
func (r RowData) GetChannels() []intf.ChannelData {
	c := make([]intf.ChannelData, len(r.Channels))
	for i := range r.Channels {
		c[i] = &r.Channels[i]
	}

	return c
}

// Pattern is the data for each pattern
type Pattern struct {
	intf.Pattern
	Packed PackedPattern
	Rows   [64]RowData
}

// GetRow returns the interface to the row at index `row`
func (p Pattern) GetRow(row uint8) intf.Row {
	return &p.Rows[row]
}

// GetRows returns the interfaces to all the rows in the pattern
func (p Pattern) GetRows() []intf.Row {
	rows := make([]intf.Row, len(p.Rows))
	for i, pr := range p.Rows {
		rows[i] = pr
	}
	return rows
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
