package voice

import (
	"github.com/gotracker/gomixing/volume"
)

// FilterApplier is an interface for applying a filter to a sample stream
type FilterApplier interface {
	ApplyFilter(dry volume.Matrix) volume.Matrix
	SetFilterEnvelopeValue(envVal float32)
}
