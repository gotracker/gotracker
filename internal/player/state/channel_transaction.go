package state

import (
	"github.com/gotracker/gotracker/internal/player/intf"
	"github.com/gotracker/gotracker/internal/song"
)

type ChannelDataTransaction[TMemory, TChannelData any] interface {
	GetData() *TChannelData
	SetData(data *TChannelData, s song.Data, cs *ChannelState[TMemory, TChannelData])

	CommitPreRow(p intf.Playback, cs *ChannelState[TMemory, TChannelData], semitoneSetterFactory SemitoneSetterFactory[TMemory, TChannelData]) error
	CommitRow(p intf.Playback, cs *ChannelState[TMemory, TChannelData], semitoneSetterFactory SemitoneSetterFactory[TMemory, TChannelData]) error
	CommitPostRow(p intf.Playback, cs *ChannelState[TMemory, TChannelData], semitoneSetterFactory SemitoneSetterFactory[TMemory, TChannelData]) error

	CommitPreTick(p intf.Playback, cs *ChannelState[TMemory, TChannelData], currentTick int, lastTick bool, semitoneSetterFactory SemitoneSetterFactory[TMemory, TChannelData]) error
	CommitTick(p intf.Playback, cs *ChannelState[TMemory, TChannelData], currentTick int, lastTick bool, semitoneSetterFactory SemitoneSetterFactory[TMemory, TChannelData]) error
	CommitPostTick(p intf.Playback, cs *ChannelState[TMemory, TChannelData], currentTick int, lastTick bool, semitoneSetterFactory SemitoneSetterFactory[TMemory, TChannelData]) error

	AddVolOp(op VolOp[TMemory, TChannelData])
	AddNoteOp(op NoteOp[TMemory, TChannelData])
}
