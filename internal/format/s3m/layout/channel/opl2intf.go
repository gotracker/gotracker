package channel

import (
	"github.com/gotracker/opl2"
)

// OPL2Intf is an interface to get the active OPL2 chip
type OPL2Intf interface {
	GetOPL2Chip() *opl2.Chip
}
