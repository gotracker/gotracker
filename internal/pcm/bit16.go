package pcm

import (
	"encoding/binary"
	"io"

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

// ReadAt reads a value from the reader provided in the byte order provided
func (s *Sample16BitSigned) ReadAt(r io.ReaderAt, ofs int64, b binary.ByteOrder) error {
	var in [2]byte
	if _, err := r.ReadAt(in[:], ofs); err != nil {
		return err
	}
	*s = Sample16BitSigned(b.Uint16(in[:]))
	return nil
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

// ReadAt reads a value from the reader provided in the byte order provided
func (s *Sample16BitUnsigned) ReadAt(r io.ReaderAt, ofs int64, b binary.ByteOrder) error {
	var in [2]byte
	if _, err := r.ReadAt(in[:], ofs); err != nil {
		return err
	}
	*s = Sample16BitUnsigned(b.Uint16(in[:]))
	return nil
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
