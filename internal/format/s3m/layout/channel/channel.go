package channel

import (
	s3mfile "github.com/gotracker/goaudiofile/music/tracked/s3m"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/format/s3m/playback/util"
	"gotracker/internal/song/instrument"
	"gotracker/internal/song/note"
)

// S3MInstrumentID is an instrument ID in S3M world
type S3MInstrumentID uint8

// IsEmpty returns true if the instrument ID is 'nothing'
func (s S3MInstrumentID) IsEmpty() bool {
	return s == 0
}

// Data is the data for the channel
type Data struct {
	What       s3mfile.PatternFlags
	Note       s3mfile.Note
	Instrument S3MInstrumentID
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
func (d *Data) GetInstrument(stmem note.Semitone) instrument.ID {
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
