package playlist

import "gotracker/internal/optional"

type Position struct {
	Order optional.Value //int
	Row   optional.Value //int
}

type Song struct {
	Filepath string
	Start    Position
	End      Position
	Loop     optional.Value //bool
}
