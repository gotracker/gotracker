package instrument

import (
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"
	"github.com/gotracker/voice/envelope"
	"github.com/gotracker/voice/fadeout"
	"github.com/gotracker/voice/loop"
	"github.com/gotracker/voice/pcm"
)

// PCM is a PCM-data instrument
type PCM struct {
	Sample        pcm.Sample
	Loop          loop.Loop
	SustainLoop   loop.Loop
	Panning       panning.Position
	MixingVolume  volume.Volume
	FadeOut       fadeout.Settings
	VolEnv        envelope.Envelope
	PanEnv        envelope.Envelope
	PitchFiltMode bool              // true = filter, false = pitch
	PitchFiltEnv  envelope.Envelope // this is either pitch or filter
}
