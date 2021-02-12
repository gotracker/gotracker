package pcm

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/gotracker/gomixing/volume"
)

// SampleReader is a reader interface that can return a whole multichannel sample at the current position
type SampleReader interface {
	Read() (volume.Matrix, error)
}

func readSingleChannelSample(s *SampleData, pos int, out interface{}) error {
	if pos >= len(s.data) {
		return errors.New("index out of range")
	}

	buf := bytes.NewReader(s.data[pos:])
	if err := binary.Read(buf, s.byteOrder, out); err != nil {
		return err
	}
	return nil
}

func (s *SampleData) readData(converter SampleConverter) (volume.Matrix, error) {
	bps := converter.Size()
	actualPos := s.pos * s.channels * bps
	if actualPos < 0 {
		actualPos = 0
	}
	out := make(volume.Matrix, s.channels)
	for c := range out {
		err := readSingleChannelSample(s, actualPos, converter)
		if err != nil {
			return nil, err
		}

		out[c] = converter.Volume()
		actualPos += bps
	}

	s.pos++
	return out, nil
}
