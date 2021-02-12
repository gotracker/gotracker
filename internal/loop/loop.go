package loop

// Loop is a loop interface
type Loop interface {
	Enabled() bool
	Length() int
	CalcPos(pos int, length int) (int, bool)
}

// Settings is details about a loop
type Settings struct {
	Begin int
	End   int
}

// CalcLoopPos returns the new location and looped flag within a pair of loops (normal and sustain)
func CalcLoopPos(loop Loop, sustain Loop, pos int, length int, keyOn bool) (int, bool) {
	if keyOn && sustain.Enabled() {
		// sustain loop
		return sustain.CalcPos(pos, length)
	}
	// non-sustain loop
	return loop.CalcPos(pos, length)
}
