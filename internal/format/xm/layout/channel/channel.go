package channel

import (
	xmfile "github.com/gotracker/goaudiofile/music/tracked/xm"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// Data is the data for the channel
type Data struct {
	intf.ChannelData
	What            xmfile.ChannelFlags
	Note            uint8
	Instrument      uint8
	Volume          uint8
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
func (d *Data) GetInstrument() uint8 {
	return d.Instrument
}

// HasVolume returns true if there exists a volume on the channel
func (d *Data) HasVolume() bool {
	if !d.What.HasVolume() {
		return false
	}

	v := d.Volume
	return v == 0x00 || v >= 0x10 && v <= 0x50
}

// GetVolume returns the volume for the channel
func (d *Data) GetVolume() volume.Volume {
	if d.Volume == 0 {
		return volume.VolumeUseInstVol
	}
	return util.VolumeFromXm(d.Volume - 0x10)
}

// HasEffect returns true if there exists a effect on the channel
func (d *Data) HasEffect() bool {
	if d.What.HasEffect() || d.What.HasEffectParameter() {
		return true
	}

	if d.What.HasVolume() {
		return d.Volume >= 0x60
	}

	return false
}

// Channel returns the channel ID for the channel
func (d *Data) Channel() uint8 {
	return 0
}
