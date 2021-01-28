package instrument

// LoopMode is the mode of operation for the looping instrument sample
type LoopMode uint8

const (
	// LoopModeDisabled is for disabled looping
	//  |start>----------------------------------------end| <= on playthrough 1, whole sample plays
	LoopModeDisabled = LoopMode(iota)
	// LoopModeNormalType1 is for normal looping, type 1: (S3M)
	//  |start>----------------------------------------end| <= on playthrough 1, whole sample plays
	//  |-------------|loopBegin>-----loopEnd|------------| <= only if looped and on playthrough 2+, only the part that loops plays
	//  |-------------|loopBegin>----------------------end| <= on playthrough 2+, the loop ends and playback continues to end, if !keyOn
	LoopModeNormalType1
	// LoopModeNormalType2 is for normal looping, type 2: (XM)
	//  |start>-----------------------loopEnd|------------| <= on playthrough 1, whole sample plays
	//  |-------------|loopBegin>-----loopEnd|------------| <= only if looped and on playthrough 2+, only the part that loops plays
	//  |-------------|loopBegin>----------------------end| <= on playthrough 2+, the loop ends and playback continues to end, if !keyOn
	LoopModeNormalType2
	// LoopModePingPong is for ping-pong looping:
	//  |start>-----------------------loopEnd|------------| <= on playthrough 1, whole sample plays
	//  |-------------|loopBegin>----<loopEnd|------------| <= only if looped and on playthrough 2+, part that loops plays and ping-pongs
	//  |-------------|loopBegin>----------------------end| <= on playthrough 2+, the loop ends and playback continues to end, if !keyOn
	LoopModePingPong
)

// LoopInfo is details about a loop
type LoopInfo struct {
	Mode  LoopMode
	Begin int
	End   int
}

func calcLoopPos(loop LoopInfo, sustain LoopInfo, pos int, length int, keyOn bool) (int, bool) {
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

func internalCalcLoopPos(loop LoopInfo, pos int, length int) (bool, int, bool) {
	switch loop.Mode {
	case LoopModeDisabled:
		// nothing
	case LoopModeNormalType1:
		newPos, looped := calcLoopPosMode1(pos, length, loop.Begin, loop.End)
		return true, newPos, looped
	case LoopModeNormalType2:
		newPos, looped := calcLoopPosMode2(pos, length, loop.Begin, loop.End)
		return true, newPos, looped
	case LoopModePingPong:
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

func calcLoopPosMode1(pos int, length int, loopBegin int, loopEnd int) (int, bool) {
	//  |start>----------------------------------------end| <= on playthrough 1, whole sample plays
	//  |-------------|loopBegin>-----loopEnd|------------| <= only if looped and on playthrough 2+, only the part that loops plays
	//  |-------------|loopBegin>----------------------end| <= on playthrough 2+, the loop ends and playback continues to end, if !keyOn
	switch {
	case pos < 0:
		return 0, false
	case pos < length:
		return pos, false
	}

	loopLen := calcLoopLen(loopBegin, loopEnd)
	if loopLen < 0 {
		return length, false
	} else if loopLen == 0 {
		return loopBegin, true
	}

	loopedPos := (pos - length) % loopLen
	return loopBegin + loopedPos, true
}

func calcLoopPosMode2(pos int, length int, loopBegin int, loopEnd int) (int, bool) {
	//  |start>-----------------------loopEnd|------------| <= on playthrough 1, whole sample plays
	//  |-------------|loopBegin>-----loopEnd|------------| <= only if looped and on playthrough 2+, only the part that loops plays
	//  |-------------|loopBegin>----------------------end| <= on playthrough 2+, the loop ends and playback continues to end, if !keyOn
	if pos < 0 {
		return 0, false
	}
	if pos < loopEnd {
		return pos, false
	}

	loopLen := calcLoopLen(loopBegin, loopEnd)
	if loopLen < 0 {
		if pos < length {
			return pos, false
		}
		return length, false
	} else if loopLen == 0 {
		return loopBegin, true
	}

	dist := pos - loopEnd
	loopedPos := dist % loopLen
	return loopBegin + loopedPos, true
}

func calcLoopPosPingPong(pos int, length int, loopBegin int, loopEnd int) (int, bool) {
	//  |start>-----------------------loopEnd|------------| <= on playthrough 1, whole sample plays
	//  |-------------|loopBegin>----<loopEnd|------------| <= only if looped and on playthrough 2+, part that loops plays and ping-pongs
	//  |-------------|loopBegin>----------------------end| <= on playthrough 2+, the loop ends and playback continues to end, if !keyOn
	if pos < 0 {
		return 0, false
	}
	if pos < loopEnd {
		return pos, false
	}

	loopLen := calcLoopLen(loopBegin, loopEnd)
	if loopLen < 0 {
		if pos < length {
			return pos, false
		}
		return length, false
	} else if loopLen == 0 {
		return loopBegin, true
	}

	dist := pos - loopEnd
	loopedPos := dist % loopLen
	if times := int(dist / loopLen); (times & 1) == 0 {
		// even loops are reversed
		return loopEnd - loopedPos - 1, true
	}
	// odd loops are forward... or normal loop
	return loopBegin + loopedPos, true
}
