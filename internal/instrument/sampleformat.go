package instrument

import (
	"encoding/binary"

	"github.com/gotracker/gomixing/volume"
)

// SampleDataFormat is the format of the sample data
type SampleDataFormat uint8

const (
	// SampleDataFormat8BitUnsigned is for unsigned 8-bit data
	SampleDataFormat8BitUnsigned = SampleDataFormat(iota)
	// SampleDataFormat8BitSigned is for signed 8-bit data
	SampleDataFormat8BitSigned
	// SampleDataFormat16BitLEUnsigned is for unsigned, little-endian, 16-bit data
	SampleDataFormat16BitLEUnsigned
	// SampleDataFormat16BitLESigned is for signed, little-endian, 16-bit data
	SampleDataFormat16BitLESigned
	// SampleDataFormat16BitBEUnsigned is for unsigned, big-endian, 16-bit data
	SampleDataFormat16BitBEUnsigned
	// SampleDataFormat16BitBESigned is for signed, big-endian, 16-bit data
	SampleDataFormat16BitBESigned
)

func getBytesPerSample(sdf SampleDataFormat) int {
	switch sdf {
	case SampleDataFormat8BitUnsigned, SampleDataFormat8BitSigned:
		return 1
	case SampleDataFormat16BitLEUnsigned, SampleDataFormat16BitLESigned:
		return 2
	}
	panic("unhandled sample data format")
}

func readSample(sdf SampleDataFormat, sample []uint8, pos int, channels int) volume.Matrix {
	o := make(volume.Matrix, channels)
	bps := getBytesPerSample(sdf)
	actualPos := pos * channels * bps
	for c := 0; c < channels; c++ {
		switch sdf {
		case SampleDataFormat8BitUnsigned:
			o[c] = volume.Volume(int8(sample[actualPos+c]-128)) / 128.0
		case SampleDataFormat8BitSigned:
			o[c] = volume.Volume(int8(sample[actualPos+c])) / 128.0
		case SampleDataFormat16BitLEUnsigned:
			s := binary.LittleEndian.Uint16(sample[actualPos+c:])
			o[c] = volume.Volume(int16(s-32768)) / 32768.0
		case SampleDataFormat16BitLESigned:
			s := binary.LittleEndian.Uint16(sample[actualPos+c:])
			o[c] = volume.Volume(int16(s)) / 32768.0
		}
		actualPos += bps
	}
	return o
}
