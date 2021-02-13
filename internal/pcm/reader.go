package pcm

import (
	"bytes"

	"github.com/gotracker/gomixing/volume"
)

// SampleReader is a reader interface that can return a whole multichannel sample at the current position
type SampleReader interface {
	Read() (volume.Matrix, error)
}

func (s *SampleData) readData(converter SampleConverter) (volume.Matrix, error) {
	bps := converter.Size()
	actualPos := int64(s.pos * s.channels * bps)
	if actualPos < 0 {
		actualPos = 0
	}
	if s.reader == nil {
		s.reader = bytes.NewReader(s.data)
	}

	out := make(volume.Matrix, s.channels)
	for c := range out {
		if err := converter.ReadAt(s.reader, actualPos, s.byteOrder); err != nil {
			return nil, err
		}

		out[c] = converter.Volume()
		actualPos += int64(bps)
	}

	s.pos++
	return out, nil
}
