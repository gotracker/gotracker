package pcm

import (
	"github.com/gotracker/gomixing/volume"
)

const (
	cSample16BitVolumeCoeff = volume.Volume(1) / 0x8000
	cSample16BitBytes       = 2
)

// Sample16BitSigned is a signed 8-bit sample
type Sample16BitSigned int16

// Volume returns the volume value for the sample
func (s Sample16BitSigned) Volume() volume.Volume {
	return volume.Volume(s) * cSample16BitVolumeCoeff
}

// Size returns the size of the sample in bytes
func (s Sample16BitSigned) Size() int {
	return cSample16BitBytes
}

// Sample16BitUnsigned is an unsigned 8-bit sample
type Sample16BitUnsigned uint16

// Volume returns the volume value for the sample
func (s Sample16BitUnsigned) Volume() volume.Volume {
	return volume.Volume(int16(s-0x8000)) * cSample16BitVolumeCoeff
}

// Size returns the size of the sample in bytes
func (s Sample16BitUnsigned) Size() int {
	return cSample16BitBytes
}

// SampleReader16BitUnsigned is an unsigned 8-bit PCM sample reader
type SampleReader16BitUnsigned struct {
	SampleData
}

// Read returns the next multichannel sample
func (s *SampleReader16BitUnsigned) Read() (volume.Matrix, error) {
	var v Sample16BitUnsigned
	return s.readData(&v)
}

// SampleReader16BitSigned is a signed 8-bit PCM sample reader
type SampleReader16BitSigned struct {
	SampleData
}

// Read returns the next multichannel sample
func (s *SampleReader16BitSigned) Read() (volume.Matrix, error) {
	var v Sample16BitSigned
	return s.readData(&v)
}
