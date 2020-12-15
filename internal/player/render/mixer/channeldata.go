package mixer

import (
	"gotracker/internal/player/panning"
	"gotracker/internal/player/volume"
)

// Data is a single buffer of data at a specific panning position
type Data struct {
	Data       MixBuffer
	Pan        panning.Position
	Volume     volume.Volume
	SamplesLen int
	Flush      func()
}

// ChannelData is a single channel's data
type ChannelData []Data
