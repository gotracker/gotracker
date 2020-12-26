package channel

import (
	"gotracker/internal/format/s3m/playback/opl2"
)

// OPL2Intf is an interface to get the active OPL2 chip
type OPL2Intf interface {
	GetOPL2Chip() *opl2.Chip
}
