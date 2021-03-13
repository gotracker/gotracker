package intf

import "github.com/gotracker/voice/pcm"

// Format is an interface to a music file format loader
type Format interface {
	Load(filename string, preferredSampleFormat ...pcm.SampleDataFormat) (Playback, error)
}
