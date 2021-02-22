package note

import (
	"github.com/gotracker/voice/period"
)

// C2SPD defines the C-2 (or in some players cases C-4) note sampling rate
type C2SPD float32

func (c C2SPD) ToFrequency() period.Frequency {
	return period.Frequency(c)
}

// Finetune is a 1/64th of a Semitone
type Finetune int16
