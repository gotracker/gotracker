package channel

import (
	xmfile "github.com/gotracker/goaudiofile/music/tracked/xm"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/song"
	"gotracker/internal/song/instrument"
	"gotracker/internal/song/note"
)

// SampleID is an InstrumentID that is a combination of InstID and SampID
type SampleID struct {
	instrument.ID
	InstID   uint8
	Semitone note.Semitone
}

// IsEmpty returns true if the sample ID is empty
func (s SampleID) IsEmpty() bool {
	return s.InstID == 0
}

// Data is the data for the channel
type Data struct {
	song.ChannelData
	What            xmfile.ChannelFlags
	Note            uint8
	Instrument      uint8
	Volume          util.VolEffect
	Effect          uint8
	EffectParameter uint8
}

// HasNote returns true if there exists a note on the channel
func (d *Data) HasNote() bool {
	return d.What.HasNote()
}

// GetNote returns the note for the channel
func (d *Data) GetNote() note.Note {
	return util.NoteFromXmNote(d.Note)
}

// HasInstrument returns true if there exists an instrument on the channel
func (d *Data) HasInstrument() bool {
	return d.What.HasInstrument()
}

// GetInstrument returns the instrument for the channel
func (d *Data) GetInstrument(stmem note.Semitone) instrument.ID {
	st := stmem
	if d.HasNote() {
		n := d.GetNote()
		if nn, ok := n.(note.Normal); ok {
			st = note.Semitone(nn)
		}
	}
	return SampleID{
		InstID:   d.Instrument,
		Semitone: st,
	}
}

// HasVolume returns true if there exists a volume on the channel
func (d *Data) HasVolume() bool {
	if !d.What.HasVolume() {
		return false
	}

	return d.Volume.IsVolume()
}

// GetVolume returns the volume for the channel
func (d *Data) GetVolume() volume.Volume {
	return d.Volume.Volume()
}

// HasEffect returns true if there exists a effect on the channel
func (d *Data) HasEffect() bool {
	if d.What.HasEffect() || d.What.HasEffectParameter() {
		return true
	}

	if d.What.HasVolume() {
		return !d.Volume.IsVolume()
	}

	return false
}

// Channel returns the channel ID for the channel
func (d *Data) Channel() uint8 {
	return 0
}
