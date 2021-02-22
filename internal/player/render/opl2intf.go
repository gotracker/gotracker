package render

import (
	"github.com/gotracker/opl2"
	"github.com/gotracker/voice/render"
)

// NewOPL2Chip generates a new OPL2 chip
func NewOPL2Chip(rate uint32) render.OPL2Chip {
	return opl2.NewChip(rate, false)
}
