package filter

import (
	"github.com/gotracker/gomixing/volume"
)

// Filter is an interface to a filter
type Filter interface {
	Filter(volume.Matrix) volume.Matrix
	UpdateEnv(int8)
}

// Factory is a function type that builds a filter with an input parameter taking a value between 0 and 1
type Factory func(sampleRate int) Filter
