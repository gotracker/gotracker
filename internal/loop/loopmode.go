package loop

// Mode is the mode of operation for the looping ranges of "things" (samples, envelope points, etc)
type Mode uint8

const (
	// ModeDisabled is for disabled looping
	//  |start>----------------------------------------end| <= on playthrough 1, whole sample plays
	ModeDisabled = Mode(iota)
	// ModeLegacy is for legacy looping: (old MOD players)
	//  |start>----------------------------------------end| <= on playthrough 1, whole sample plays
	//  |-------------|loopBegin>-----loopEnd|------------| <= only if looped and on playthrough 2+, only the part that loops plays
	//  |-------------|loopBegin>----------------------end| <= on playthrough 2+, the loop ends and playback continues to end, if !keyOn
	ModeLegacy
	// ModeNormal is for normal looping: (S3M/XM/IT)
	//  |start>-----------------------loopEnd|------------| <= on playthrough 1, whole sample plays
	//  |-------------|loopBegin>-----loopEnd|------------| <= only if looped and on playthrough 2+, only the part that loops plays
	//  |-------------|loopBegin>----------------------end| <= on playthrough 2+, the loop ends and playback continues to end, if !keyOn
	ModeNormal
	// ModePingPong is for ping-pong looping:
	//  |start>-----------------------loopEnd|------------| <= on playthrough 1, whole sample plays
	//  |-------------|loopBegin>----<loopEnd|------------| <= only if looped and on playthrough 2+, part that loops plays and ping-pongs
	//  |-------------|loopBegin>----------------------end| <= on playthrough 2+, the loop ends and playback continues to end, if !keyOn
	ModePingPong
)

func internalCalcLoopPos(loop *Loop, pos int, length int) (bool, int, bool) {
	switch loop.Mode {
	case ModeDisabled:
		// nothing
	case ModeLegacy:
		newPos, looped := calcLoopPosMode1(pos, length, loop.Begin, loop.End)
		return true, newPos, looped
	case ModeNormal:
		newPos, looped := calcLoopPosMode2(pos, length, loop.Begin, loop.End)
		return true, newPos, looped
	case ModePingPong:
		newPos, looped := calcLoopPosPingPong(pos, length, loop.Begin, loop.End)
		return true, newPos, looped
	default:
		panic("unhandled loop mode!")
	}
	return false, pos, false
}

func calcLoopPosDisabled(pos int, length int) (int, bool) {
	//  |start>----------------------------------------end| <= on playthrough 1, whole sample plays
	switch {
	case pos < 0:
		return 0, false
	case pos < length:
		return pos, false
	default:
		return length, false
	}
}

// simple helper to consolidate loop length calculations
// (yeah, it could just be the math in place, but whatever)
func calcLoopLen(loopBegin int, loopEnd int) int {
	return loopEnd - loopBegin
}
