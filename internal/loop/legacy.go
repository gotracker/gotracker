package loop

// Legacy is a legacy loop based on how some old protracker players worked
//  |start>----------------------------------------end| <= on playthrough 1, whole sample plays
//  |-------------|loopBegin>-----loopEnd|------------| <= only if looped and on playthrough 2+, only the part that loops plays
//  |-------------|loopBegin>----------------------end| <= on playthrough 2+, the loop ends and playback continues to end, if !keyOn
type Legacy struct {
	Settings
}

// Enabled returns true if the loop is enabled
func (l *Legacy) Enabled() bool {
	return true
}

// Length returns the length of the loop
func (l *Legacy) Length() int {
	return calcLoopLen(l.Begin, l.End)
}

// CalcPos calculates the position based on the loop details
func (l *Legacy) CalcPos(pos int, length int) (int, bool) {
	switch {
	case pos < 0:
		return 0, false
	case pos < length:
		return pos, false
	}

	loopLen := l.Length()
	if loopLen < 0 {
		return length, false
	} else if loopLen == 0 {
		return l.Begin, true
	}

	loopedPos := (pos - length) % loopLen
	return l.Begin + loopedPos, true
}
