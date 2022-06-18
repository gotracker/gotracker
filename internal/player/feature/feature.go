package feature

import "time"

// Feature is an interface for player features that can be optionally modified by the user and/or disabled by an output device
type Feature any

// SongLoop is a setting for enabling or disabling the song looping
type SongLoop struct {
	Count int
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

type EnableTracing struct {
	Filename string
}

type PreConvertSamples struct {
	Enabled bool
}

type PlayUntilOrderAndRow struct {
	Order int
	Row   int
}

type ITLongChannelOutput struct {
	Enabled bool
}

type ITNewNoteActions struct {
	Enabled bool
}

type SetDefaultTempo struct {
	Tempo int
}

type SetDefaultBPM struct {
	BPM int
}
