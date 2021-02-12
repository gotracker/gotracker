package pcm

import (
	"github.com/gotracker/gomixing/volume"
)

const (
	cSample8BitVolumeCoeff = volume.Volume(1) / 0x80
	cSample8BitBytes       = 1
)

// Sample8BitSigned is a signed 8-bit sample
type Sample8BitSigned int8

// Volume returns the volume value for the sample
func (s Sample8BitSigned) Volume() volume.Volume {
	return volume.Volume(s) * cSample8BitVolumeCoeff
}

// Size returns the size of the sample in bytes
func (s Sample8BitSigned) Size() int {
	return cSample8BitBytes
}

// Sample8BitUnsigned is an unsigned 8-bit sample
type Sample8BitUnsigned uint8

// Volume returns the volume value for the sample
func (s Sample8BitUnsigned) Volume() volume.Volume {
	return volume.Volume(int8(s-0x80)) * cSample8BitVolumeCoeff
}

// Size returns the size of the sample in bytes
func (s Sample8BitUnsigned) Size() int {
	return cSample8BitBytes
}

// SampleReader8BitUnsigned is an unsigned 8-bit PCM sample reader
type SampleReader8BitUnsigned struct {
	SampleData
}

// Read returns the next multichannel sample
func (s *SampleReader8BitUnsigned) Read() (volume.Matrix, error) {
	var v Sample8BitUnsigned
	return s.readData(&v)
}

// SampleReader8BitSigned is a signed 8-bit PCM sample reader
type SampleReader8BitSigned struct {
	SampleData
}

// Read returns the next multichannel sample
func (s *SampleReader8BitSigned) Read() (volume.Matrix, error) {
	var v Sample8BitSigned
	return s.readData(&v)
}
