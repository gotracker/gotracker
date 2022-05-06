package util

import "github.com/gotracker/gotracker/internal/song/index"

// PatternLoop is a state machine for pattern loops
type PatternLoop struct {
	Enabled bool
	Start   index.Row
	End     index.Row
	Total   uint8

	Count uint8
}

// ContinueLoop returns the next expected row if a loop occurs
func (pl *PatternLoop) ContinueLoop(currentRow index.Row) (index.Row, bool) {
	if pl.Enabled {
		if currentRow == pl.End {
			pl.Count++
			if pl.Count <= pl.Total {
				return pl.Start, true
			}
			pl.Enabled = false
		}
	}
	return 0, false
}
