package loop

// Normal is a normal loop
//  |start>-----------------------loopEnd|------------| <= on playthrough 1, whole sample plays
//  |-------------|loopBegin>-----loopEnd|------------| <= only if looped and on playthrough 2+, only the part that loops plays
//  |-------------|loopBegin>----------------------end| <= on playthrough 2+, the loop ends and playback continues to end, if !keyOn
type Normal struct {
	Settings
}

// Enabled returns true if the loop is enabled
func (l *Normal) Enabled() bool {
	return true
}

// Length returns the length of the loop
func (l *Normal) Length() int {
	return calcLoopLen(l.Begin, l.End)
}

// CalcPos calculates the position based on the loop details
func (l *Normal) CalcPos(pos int, length int) (int, bool) {
	if pos < 0 {
		return 0, false
	}
	if pos < l.End {
		return pos, false
	}

	loopLen := l.Length()
	if loopLen < 0 {
		if pos < length {
			return pos, false
		}
		return length, false
	} else if loopLen == 0 {
		return l.Begin, true
	}

	dist := pos - l.End
	loopedPos := dist % loopLen
	return l.Begin + loopedPos, true
}
