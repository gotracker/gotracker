// Package s3m does a thing.
package s3m

import (
	"gotracker/internal/format/s3m/channel"
	"gotracker/internal/format/s3m/modconv"
	"gotracker/internal/format/s3m/util"
	"gotracker/internal/player/intf"

	"github.com/heucuva/gomixing/panning"
	"github.com/heucuva/gomixing/volume"
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
	OrderList       []uint8
}

// GetOrderList returns the list of all pattern orders for the song
func (s *Song) GetOrderList() []uint8 {
	return s.OrderList
}

// GetPattern returns an interface to a specific pattern indexed by `patNum`
func (s *Song) GetPattern(patNum uint8) intf.Pattern {
	if int(patNum) >= len(s.Patterns) {
		return nil
	}
	return s.Patterns[patNum]
}

// IsChannelEnabled returns true if the channel at index `channelNum` is enabled
func (s *Song) IsChannelEnabled(channelNum int) bool {
	return s.ChannelSettings[channelNum].Enabled
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

func readMOD(filename string) (*Song, error) {
	buffer, err := readFile(filename)
	if err != nil {
		return nil, err
	}

	f, err := modconv.Read(buffer)
	if err != nil {
		return nil, err
	}

	return convertS3MFileToSong(f)
}

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
