package common

// Kind is an enumeration of the device type
type Kind int

const (
	// KindNone is nothing!
	KindNone = Kind(iota)
	// KindFile is a file device type
	KindFile
	// KindSoundCard is an active sound playback device (e.g.: a sound card attached to speakers)
	KindSoundCard
)
