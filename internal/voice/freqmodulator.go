package voice

import (
	"gotracker/internal/player/note"
)

// FreqModulator is the instrument frequency control interface
type FreqModulator interface {
	SetPeriod(period note.Period)
	GetPeriod() note.Period
	SetPeriodDelta(delta note.PeriodDelta)
	GetPeriodDelta() note.PeriodDelta
	GetFinalPeriod() note.Period
}
