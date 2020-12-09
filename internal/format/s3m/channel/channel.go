package channel

import (
	"gotracker/internal/format/s3m/util"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/volume"
)

type What uint8

func (w What) HasCommand() bool {
	return (w & 0x80) != 0
}

func (w What) HasVolume() bool {
	return (w & 0x40) != 0
}

func (w What) HasNote() bool {
	return (w & 0x20) != 0
}

func (w What) Channel() uint8 {
	return uint8(w) & 0x1F
}

type Data struct {
	intf.ChannelData
	What       What
	Note       note.Note
	Instrument uint8
	Volume     uint8
	Command    uint8
	Info       uint8
}

func (d *Data) HasNote() bool {
	return d.What.HasNote()
}

func (d *Data) GetNote() note.Note {
	return d.Note
}

func (d *Data) HasInstrument() bool {
	return d.Instrument != 0
}

func (d *Data) GetInstrument() uint8 {
	return d.Instrument
}

func (d *Data) HasVolume() bool {
	return d.What.HasVolume()
}

func (d *Data) GetVolume() volume.Volume {
	return util.VolumeFromS3M(d.Volume)
}

func (d *Data) HasCommand() bool {
	return d.What.HasCommand()
}

func (d *Data) Channel() uint8 {
	return d.What.Channel()
}
