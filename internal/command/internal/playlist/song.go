package playlist

import (
	"gotracker/internal/optional"
)

type Position struct {
	Order optional.Value `yaml:"order,omitempty"` //int
	Row   optional.Value `yaml:"row,omitempty"`   //int
}

type Song struct {
	Filepath string         `yaml:"file,omitempty"`
	Start    Position       `yaml:"start,omitempty"`
	End      Position       `yaml:"end,omitempty"`
	Loop     optional.Value `yaml:"loop,omitempty"` //bool
}
