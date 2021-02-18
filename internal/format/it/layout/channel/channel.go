package channel

import (
	itfile "github.com/gotracker/goaudiofile/music/tracked/it"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/format/it/playback/util"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// SampleID is an InstrumentID that is a combination of InstID and SampID
type SampleID struct {
	intf.InstrumentID
	InstID   uint8
	Semitone note.Semitone
}

// IsEmpty returns true if the sample ID is empty
func (s SampleID) IsEmpty() bool {
	return s.InstID == 0
}

// Data is the data for the channel
type Data struct {
	What            itfile.ChannelDataFlags
	Note            itfile.Note
	Instrument      uint8
	VolPan          uint8
	Effect          uint8
	EffectParameter uint8
}

// HasNote returns true if there exists a note on the channel
func (d *Data) HasNote() bool {
	return d.What.HasNote()
}

// GetNote returns the note for the channel
func (d *Data) GetNote() note.Note {
	return util.NoteFromItNote(d.Note)
}

// HasInstrument returns true if there exists an instrument on the channel
func (d *Data) HasInstrument() bool {
	return d.What.HasInstrument()
}

// GetInstrument returns the instrument for the channel
func (d *Data) GetInstrument(stmem note.Semitone) intf.InstrumentID {
	st := stmem
	if d.HasNote() {
		n := d.GetNote()
		if !n.IsSpecial() {
			st = n.Semitone()
		}
	}
	return SampleID{
		InstID:   d.Instrument,
		Semitone: st,
	}
}

// HasVolume returns true if there exists a volume on the channel
func (d *Data) HasVolume() bool {
	if !d.What.HasVolPan() {
		return false
	}

	v := d.VolPan
	return v <= 64
}

// GetVolume returns the volume for the channel
func (d *Data) GetVolume() volume.Volume {
	return util.VolumeFromVolPan(d.VolPan)
}

// HasCommand returns true if there exists a effect on the channel
func (d *Data) HasCommand() bool {
	if d.What.HasCommand() {
		return true
	}

	if d.What.HasVolPan() {
		return d.VolPan > 64
	}

	return false
}

// Channel returns the channel ID for the channel
func (d *Data) Channel() uint8 {
	return 0
}
