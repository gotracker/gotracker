package channel

import "gotracker/internal/player/note"

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
	What       What
	Note       note.Note
	Instrument uint8
	Volume     uint8
	Command    uint8
	Info       uint8
}
