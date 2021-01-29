package loop

// Loop is details about a loop
type Loop struct {
	Mode  Mode
	Begin int
	End   int
}

// Enabled returns true if the loop is enabled
func (l *Loop) Enabled() bool {
	return l.Mode != ModeDisabled
}

// Length returns the length of the loop
func (l *Loop) Length() int {
	return calcLoopLen(l.Begin, l.End)
}

// CalcLoopPos returns the new location and looped flag within a pair of loops (normal and sustain)
func CalcLoopPos(loop *Loop, sustain *Loop, pos int, length int, keyOn bool) (int, bool) {
	if keyOn {
		// sustain loop
		if enabled, newPos, looped := internalCalcLoopPos(sustain, pos, length); enabled {
			return newPos, looped
		}
	}
	// non-sustain loop
	if enabled, newPos, looped := internalCalcLoopPos(loop, pos, length); enabled {
		return newPos, looped
	}
	return calcLoopPosDisabled(pos, length)
}
