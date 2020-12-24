package intf

import "github.com/gotracker/gomixing/volume"

// Filter is an interface to a filter
type Filter interface {
	Filter(volume.Matrix) volume.Matrix
}
