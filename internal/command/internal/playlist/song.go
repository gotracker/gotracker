package playlist

import (
	"gotracker/internal/optional"
)

type Position struct {
	Order optional.Value `yaml:"order,omitempty"` //int
	Row   optional.Value `yaml:"row,omitempty"`   //int
}

type Song struct {
	Filepath string   `yaml:"file,omitempty"`
	Start    Position `yaml:"start,omitempty"`
	End      Position `yaml:"end,omitempty"`
	Loop     Loop     `yaml:"loop,omitempty"`
	Fadeout  Fadeout  `yaml:"fadeout,omitempty"`
}

type Loop struct {
	Count optional.Value `yaml:"count,omitempty" default:"0"` //int  :: 0 = play 1 time / no looping; 1 = play 2 times, etc.; <0 = play indefinitely
}

func NewLoopCount(loops int) optional.Value {
	return optional.NewValue(loops)
}

func NewLoopForever() optional.Value {
	return NewLoopCount(-1)
}

type Fadeout struct {
	Length optional.Value `yaml:"length,omitempty" default:"0"` // int  :: when Song.End (and Loop.Count) is reached, this is the number of ticks to fadeout over
}
