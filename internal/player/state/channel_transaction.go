package state

import (
	"github.com/gotracker/gotracker/internal/player/intf"
	"github.com/gotracker/gotracker/internal/song"
)

type ChannelDataTransaction[TMemory, TChannelData any] interface {
	GetData() *TChannelData
	SetData(data *TChannelData, s song.Data, cs *ChannelState[TMemory, TChannelData])

	Commit(cs *ChannelState[TMemory, TChannelData], currentTick int, semitoneSetterFactory SemitoneSetterFactory[TMemory, TChannelData])

	AddVolOp(op VolOp[TMemory, TChannelData])
	ProcessVolOps(p intf.Playback, cs *ChannelState[TMemory, TChannelData]) error

	AddNoteOp(op NoteOp[TMemory, TChannelData])
	ProcessNoteOps(p intf.Playback, cs *ChannelState[TMemory, TChannelData]) error
}
