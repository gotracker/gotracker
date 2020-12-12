package mixer

import "gotracker/internal/player/volume"

// ChannelMixBuffer is a single channel's premixed volume data
type ChannelMixBuffer volume.VolumeMatrix

// MixBuffer is a buffer of premixed volume data intended to
// be eventually sent out to the sound output device after
// conversion to the output format
type MixBuffer []ChannelMixBuffer

// MixInAt mixes in samples into the mixbuffer at a particular position `pos`
func (m *MixBuffer) MixInAt(pos int, samp volume.VolumeMatrix) {
	for c, s := range samp {
		(*m)[c][pos] += s
	}
}
