package mixer

// Mixer is a manager for mixing multiple single- and multi-channel samples into a single multi-channel output stream
type Mixer struct {
	Channels      int
	BitsPerSample int
}

// NewMixBuffer returns a mixer buffer with a number of channels
// of preallocated sample data
func (m *Mixer) NewMixBuffer(samples int) MixBuffer {
	mb := make(MixBuffer, m.Channels)
	for i := range mb {
		mb[i] = make(ChannelMixBuffer, samples)
	}
	return mb
}

// Flatten will to a final saturation mix of all the row's channel data into a single output buffer
func (m *Mixer) Flatten(panmixer PanMixer, samplesLen int, row []ChannelData) []byte {
	data := m.NewMixBuffer(samplesLen)
	for _, rdata := range row {
		pos := 0
		for _, cdata := range rdata {
			if cdata.Flush != nil {
				cdata.Flush()
			}
			if len(cdata.Data) > 0 {
				volMtx := cdata.Volume.Apply(panmixer.GetMixingMatrix(cdata.Pan)...)
				data.Add(pos, cdata.Data, volMtx)
			}
			pos += cdata.SamplesLen
		}
	}
	return data.ToRenderData(samplesLen, m.BitsPerSample, len(row))
}

// FlattenTo will to a final saturation mix of all the row's channel data into a single output buffer
func (m *Mixer) FlattenTo(resultBuffer []byte, panmixer PanMixer, samplesLen int, row []ChannelData) {
	data := m.NewMixBuffer(samplesLen)
	for _, rdata := range row {
		pos := 0
		for _, cdata := range rdata {
			if cdata.Flush != nil {
				cdata.Flush()
			}
			if len(cdata.Data) > 0 {
				volMtx := cdata.Volume.Apply(panmixer.GetMixingMatrix(cdata.Pan)...)
				data.Add(pos, cdata.Data, volMtx)
			}
			pos += cdata.SamplesLen
		}
	}
	data.ToRenderDataWithBuf(resultBuffer, samplesLen, m.BitsPerSample, len(row))
}
