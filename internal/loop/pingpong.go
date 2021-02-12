package loop

// PingPong is a loop that bounces forward and backward between loopBegin and loopEnd
//  |start>-----------------------loopEnd|------------| <= on playthrough 1, whole sample plays
//  |-------------|loopBegin>----<loopEnd|------------| <= only if looped and on playthrough 2+, part that loops plays and ping-pongs
//  |-------------|loopBegin>----------------------end| <= on playthrough 2+, the loop ends and playback continues to end, if !keyOn
type PingPong struct {
	Settings
}

// Enabled returns true if the loop is enabled
func (l *PingPong) Enabled() bool {
	return true
}

// Length returns the length of the loop
func (l *PingPong) Length() int {
	return calcLoopLen(l.Begin, l.End)
}

// CalcPos calculates the position based on the loop details
func (l *PingPong) CalcPos(pos int, length int) (int, bool) {
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
	if times := int(dist / loopLen); (times & 1) == 0 {
		// even loops are reversed
		return l.End - loopedPos - 1, true
	}
	// odd loops are forward... or normal loop
	return l.Begin + loopedPos, true
}
