package channel

import (
	s3mfile "github.com/heucuva/goaudiofile/music/tracked/s3m"
	"github.com/heucuva/gomixing/volume"

	"gotracker/internal/format/s3m/playback/util"
	"gotracker/internal/player/note"
)

// Data is the data for the channel
type Data struct {
	What       s3mfile.PatternFlags
	Note       s3mfile.Note
	Instrument uint8
	Volume     s3mfile.Volume
	Command    uint8
	Info       uint8
}

// HasNote returns true if there exists a note on the channel
func (d *Data) HasNote() bool {
	return d.What.HasNote()
}

// GetNote returns the note for the channel
func (d *Data) GetNote() note.Note {
	return util.NoteFromS3MNote(d.Note)
}

// HasInstrument returns true if there exists an instrument on the channel
func (d *Data) HasInstrument() bool {
	return d.Instrument != 0
}

// GetInstrument returns the instrument for the channel
func (d *Data) GetInstrument() uint8 {
	return d.Instrument
}

// HasVolume returns true if there exists a volume on the channel
func (d *Data) HasVolume() bool {
	return d.What.HasVolume()
}

// GetVolume returns the volume for the channel
func (d *Data) GetVolume() volume.Volume {
	return util.VolumeFromS3M(d.Volume)
}

// HasCommand returns true if there exists a command on the channel
func (d *Data) HasCommand() bool {
	return d.What.HasCommand()
}

// Channel returns the channel ID for the channel
func (d *Data) Channel() uint8 {
	return d.What.Channel()
}
