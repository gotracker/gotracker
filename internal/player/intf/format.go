package intf

// Format is an interface to a music file format loader
type Format interface {
	Load(ss Song, filename string) error
	GetBaseClockRate() float32
}
