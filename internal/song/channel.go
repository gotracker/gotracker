package song

import (
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/song/instrument"
	"gotracker/internal/song/note"
)

// ChannelData is an interface for channel data
type ChannelData interface {
	HasNote() bool
	GetNote() note.Note

	HasInstrument() bool
	GetInstrument(note.Semitone) instrument.ID

	HasVolume() bool
	GetVolume() volume.Volume

	HasCommand() bool

	Channel() uint8
}
