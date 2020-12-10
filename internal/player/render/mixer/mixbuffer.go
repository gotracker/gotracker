package mixer

import "gotracker/internal/player/volume"

// ChannelMixBuffer is a single channel's premixed volume data
type ChannelMixBuffer []volume.Volume

// MixBuffer is a buffer of premixed volume data intended to
// be eventually sent out to the sound output device after
// conversion to the output format
type MixBuffer []ChannelMixBuffer
