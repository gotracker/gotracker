package channel

import (
	"fmt"
	"strings"

	itfile "github.com/gotracker/goaudiofile/music/tracked/it"
	"github.com/gotracker/gomixing/volume"

	"github.com/gotracker/gotracker/internal/format/it/playback/util"
	"github.com/gotracker/gotracker/internal/song/instrument"
	"github.com/gotracker/gotracker/internal/song/note"
)

const MaxTotalChannels = 32

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

func (d Data) String() string {
	return DataToString(&d, true)
}

func DataToString(data *Data, longChannelOutput bool) string {
	n := "..."
	i := ".."
	v := ".."
	e := "..."

	if data != nil {
		if data.HasNote() {
			nt := data.GetNote()
			switch note.Type(nt) {
			case note.SpecialTypeRelease:
				n = "==="
			case note.SpecialTypeStop:
				n = "^^^"
			case note.SpecialTypeNormal:
				n = nt.String()
			default:
				n = "???"
			}
		}

		if longChannelOutput {
			if data.HasInstrument() {
				if inst := data.Instrument; inst != 0 {
					i = fmt.Sprintf("%0.2X", inst)
				}
			}

			if data.HasVolume() {
				vol := data.VolPan
				v = fmt.Sprintf("%0.2X", vol)
			}

			if data.HasCommand() && data.Effect != 0 {
				var c uint8
				switch {
				case data.Effect <= 26:
					c = '@' + data.Effect
				default:
					panic("effect out of range")
				}
				e = fmt.Sprintf("%c%0.2X", c, data.EffectParameter)
			}
		}
	}

	if longChannelOutput {
		return strings.Join([]string{n, i, v, e}, " ")

	}

	return n
}
