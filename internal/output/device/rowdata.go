package device

import "gotracker/internal/audio/mixing"

// PremixData is a structure containing the audio pre-mix data for a specific row or buffer
type PremixData struct {
	SamplesLen int
	Data       []mixing.ChannelData
	Userdata   interface{}
}
