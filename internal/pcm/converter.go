package pcm

import (
	"github.com/gotracker/gomixing/volume"
)

// SampleConverter is an interface to a sample converter
type SampleConverter interface {
	Volume() volume.Volume
	Size() int
	ReadAt(s *SampleData, ofs int64) error
}
