package mixer

// Mixer is a manager for mixing multiple single- and multi-channel samples into a single multi-channel output stream
type Mixer struct {
}

// NewMixBuffer returns a mixer buffer with a number of channels
// of preallocated sample data
func (m *Mixer) NewMixBuffer(channels int, samples int) MixBuffer {
	mb := make(MixBuffer, channels)
	for i := range mb {
		mb[i] = make(ChannelMixBuffer, samples)
	}
	return mb
}
