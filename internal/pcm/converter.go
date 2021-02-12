package pcm

import (
	"encoding/binary"
	"io"

	"github.com/gotracker/gomixing/volume"
)

// SampleConverter is an interface to a sample converter
type SampleConverter interface {
	Volume() volume.Volume
	Size() int
	Read(r io.Reader, b binary.ByteOrder) error
}
