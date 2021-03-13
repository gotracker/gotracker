package feature

import "time"

// Feature is an interface for player features that can be optionally modified by the user and/or disabled by an output device
type Feature interface{}

// SongLoop is a setting for enabling or disabling the song looping
type SongLoop struct {
	Enabled bool
}

// PlayerSleepInterval describes the player sleep feature
type PlayerSleepInterval struct {
	Enabled  bool
	Interval time.Duration
}

// IgnoreUnknownEffect describes a bypass/ignore of unknown effects
type IgnoreUnknownEffect struct {
	Enabled bool
}

type PreConvertSamples struct {
	Enabled bool
}
