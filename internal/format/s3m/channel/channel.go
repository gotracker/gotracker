package channel

import (
	"github.com/heucuva/gomixing/volume"

	"gotracker/internal/format/s3m/util"
	"gotracker/internal/player/note"
)

// What is a flagset (and channel id) for data in the channel
type What uint8

const (
	// WhatCommand is the flag that denotes existence of a command on the channel
	WhatCommand = What(0x80)
	// WhatVolume is the flag that denotes existence of a volume on the channel
	WhatVolume = What(0x40)
	// WhatNote is the flag that denotes existence of a note on the channel
	WhatNote = What(0x20)
)

// HasCommand returns true if there exists a command on the channel
func (w What) HasCommand() bool {
	return (w & WhatCommand) != 0
}

// HasVolume returns true if there exists a volume on the channel
func (w What) HasVolume() bool {
	return (w & WhatVolume) != 0
}

// HasNote returns true if there exists a note on the channel
func (w What) HasNote() bool {
	return (w & WhatNote) != 0
}

// Channel returns the channel ID for this channel
func (w What) Channel() uint8 {
	return uint8(w) & 0x1F
}

// Data is the data for the channel
type Data struct {
	What       What
	Note       note.Note
	Instrument uint8
	Volume     uint8
	Command    uint8
	Info       uint8
}

// HasNote returns true if there exists a note on the channel
func (d *Data) HasNote() bool {
	return d.What.HasNote()
}

// GetNote returns the note for the channel
func (d *Data) GetNote() note.Note {
	return d.Note
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
