package intf

// Format is an interface to a music file format loader
type Format interface {
	Load(filename string) (Playback, error)
	GetBaseClockRate() float32
}
