package channel

import (
	"fmt"
	"strings"

	itfile "github.com/gotracker/goaudiofile/music/tracked/it"
	"github.com/gotracker/gomixing/volume"

	itNote "github.com/gotracker/gotracker/internal/format/it/conversion/note"
	itVolume "github.com/gotracker/gotracker/internal/format/it/conversion/volume"
	"github.com/gotracker/gotracker/internal/song/instrument"
	"github.com/gotracker/gotracker/internal/song/note"
)

const MaxTotalChannels = 64

// DataEffect is the type of a channel's EffectParameter value
type DataEffect uint8

// SampleID is an InstrumentID that is a combination of InstID and SampID
type SampleID struct {
	InstID   uint8
	Semitone note.Semitone
}

// IsEmpty returns true if the sample ID is empty
func (s SampleID) IsEmpty() bool {
	return s.InstID == 0
}

func (s SampleID) String() string {
	return fmt.Sprint(s.InstID)
}

// Data is the data for the channel
type Data struct {
	What            itfile.ChannelDataFlags
	Note            itfile.Note
	Instrument      uint8
	VolPan          uint8
	Effect          uint8
	EffectParameter DataEffect
}

// HasNote returns true if there exists a note on the channel
func (d Data) HasNote() bool {
	return d.What.HasNote()
}

// GetNote returns the note for the channel
func (d Data) GetNote() note.Note {
	return itNote.FromItNote(d.Note)
}

// HasInstrument returns true if there exists an instrument on the channel
func (d Data) HasInstrument() bool {
	return d.What.HasInstrument()
}

// GetInstrument returns the instrument for the channel
func (d Data) GetInstrument(stmem note.Semitone) instrument.ID {
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
func (d Data) HasVolume() bool {
	if !d.What.HasVolPan() {
		return false
	}

	v := d.VolPan
	return v <= 64
}

// GetVolume returns the volume for the channel
func (d Data) GetVolume() volume.Volume {
	return itVolume.FromVolPan(d.VolPan)
}

// HasCommand returns true if there exists a effect on the channel
func (d Data) HasCommand() bool {
	if d.What.HasCommand() {
		return true
	}

	if d.What.HasVolPan() {
		return d.VolPan > 64
	}

	return false
}

// Channel returns the channel ID for the channel
func (d Data) Channel() uint8 {
	return 0
}

func (Data) getNoteString(n note.Note) string {
	switch note.Type(n) {
	case note.SpecialTypeRelease:
		return "==="
	case note.SpecialTypeStop:
		return "^^^"
	case note.SpecialTypeNormal:
		return n.String()
	default:
		return "???"
	}
}

func (Data) getCommandString(cmd uint8) rune {
	switch {
	case cmd > 0 && cmd <= 26:
		return '@' + rune(cmd)
	default:
		panic("effect out of range")
	}
}

func (d Data) String() string {
	pieces := []string{
		"...", // note
		"..",  // inst
		"..",  // vol
		"...", // eff
	}
	if d.HasNote() {
		pieces[0] = d.getNoteString(d.GetNote())
	}
	if d.HasInstrument() {
		pieces[1] = fmt.Sprintf("%02X", d.Instrument)
	}
	if d.HasVolume() {
		pieces[2] = fmt.Sprintf("%02X", d.VolPan)
	}
	if d.HasCommand() && d.Effect != 0 {
		pieces[3] = fmt.Sprintf("%c%02X", d.getCommandString(d.Effect), d.EffectParameter)
	}
	return strings.Join(pieces, " ")
}

func (d Data) ShortString() string {
	if d.HasNote() {
		return d.GetNote().String()
	}
	return "..."
}
