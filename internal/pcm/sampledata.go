package pcm

import (
	"encoding/binary"
)

// Sample is the interface to a sample
type Sample interface {
	SampleReader
	Channels() int
	Length() int
	Seek(pos int)
	Tell() int
}

// SampleData is the presentation of the core data of the sample
type SampleData struct {
	pos       int // in multichannel samples
	length    int // in multichannel samples
	byteOrder binary.ByteOrder
	channels  int
	data      []byte
}

// Channels returns the channel count from the sample data
func (s *SampleData) Channels() int {
	return s.channels
}

// Length returns the sample length from the sample data
func (s *SampleData) Length() int {
	return s.length
}

// Seek sets the current position in the sample data
func (s *SampleData) Seek(pos int) {
	s.pos = pos
}

// Tell returns the current position in the sample data
func (s *SampleData) Tell() int {
	return s.pos
}

// NewSample constructs a sampler that can handle the requested sampler format
func NewSample(data []byte, length int, channels int, format SampleDataFormat) Sample {
	switch format {
	case SampleDataFormat8BitSigned:
		return &SampleReader8BitSigned{
			SampleData: SampleData{
				length:    length,
				byteOrder: binary.LittleEndian,
				channels:  channels,
				data:      data,
			},
		}
	case SampleDataFormat8BitUnsigned:
		return &SampleReader8BitUnsigned{
			SampleData: SampleData{
				length:    length,
				byteOrder: binary.LittleEndian,
				channels:  channels,
				data:      data,
			},
		}
	case SampleDataFormat16BitLESigned:
		return &SampleReader16BitSigned{
			SampleData: SampleData{
				length:    length,
				byteOrder: binary.LittleEndian,
				channels:  channels,
				data:      data,
			},
		}
	case SampleDataFormat16BitLEUnsigned:
		return &SampleReader16BitUnsigned{
			SampleData: SampleData{
				length:    length,
				byteOrder: binary.LittleEndian,
				channels:  channels,
				data:      data,
			},
		}
	case SampleDataFormat16BitBESigned:
		return &SampleReader16BitSigned{
			SampleData: SampleData{
				length:    length,
				byteOrder: binary.BigEndian,
				channels:  channels,
				data:      data,
			},
		}
	case SampleDataFormat16BitBEUnsigned:
		return &SampleReader16BitUnsigned{
			SampleData: SampleData{
				length:    length,
				byteOrder: binary.BigEndian,
				channels:  channels,
				data:      data,
			},
		}
	default:
		panic("unhandled sampler type")
	}
}
