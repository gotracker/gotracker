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

func calcLoopedSamplePos(loop LoopInfo, sustain LoopInfo, pos int, length int, keyOn bool) int {
	if keyOn {
		// sustain loop
		if enabled, newPos := calcLoopedPos(sustain, pos, length); enabled {
			return newPos
		}
	}
	// non-sustain loop
	if enabled, newPos := calcLoopedPos(loop, pos, length); enabled {
		return newPos
	}
	return calcSamplePosLoopDisabled(pos, length)
}

func calcLoopedPos(loop LoopInfo, pos int, length int) (bool, int) {
	switch loop.Mode {
	case LoopModeDisabled:
		// nothing
	case LoopModeNormalType1:
		return true, calcLoopedSamplePosMode1(pos, length, loop.Begin, loop.End)
	case LoopModeNormalType2:
		return true, calcLoopedSamplePosMode2(pos, length, loop.Begin, loop.End)
	case LoopModePingPong:
		return true, calcLoopedSamplePosPingPong(pos, length, loop.Begin, loop.End)
	default:
		panic("unhandled loop mode!")
	}
	return false, 0
}

func calcSamplePosLoopDisabled(pos int, length int) int {
	//  |start>----------------------------------------end| <= on playthrough 1, whole sample plays
	switch {
	case pos < 0:
		return 0
	case pos < length:
		return pos
	default:
		return length
	}
}

func calcLoopedSamplePosMode1(pos int, length int, loopBegin int, loopEnd int) int {
	//  |start>----------------------------------------end| <= on playthrough 1, whole sample plays
	//  |-------------|loopBegin>-----loopEnd|------------| <= only if looped and on playthrough 2+, only the part that loops plays
	//  |-------------|loopBegin>----------------------end| <= on playthrough 2+, the loop ends and playback continues to end, if !keyOn
	switch {
	case pos < 0:
		return 0
	case pos < length:
		return pos
	}

	loopLen := loopEnd - loopBegin
	if loopLen < 0 {
		return length
	} else if loopLen == 0 {
		return loopBegin
	}

	loopedPos := (pos - length) % loopLen
	return loopBegin + loopedPos
}

func calcLoopedSamplePosMode2(pos int, length int, loopBegin int, loopEnd int) int {
	//  |start>-----------------------loopEnd|------------| <= on playthrough 1, whole sample plays
	//  |-------------|loopBegin>-----loopEnd|------------| <= only if looped and on playthrough 2+, only the part that loops plays
	//  |-------------|loopBegin>----------------------end| <= on playthrough 2+, the loop ends and playback continues to end, if !keyOn
	if pos < 0 {
		return 0
	}
	if pos < loopEnd {
		return pos
	}

	loopLen := loopEnd - loopBegin
	if loopLen < 0 {
		if pos < length {
			return pos
		}
		return length
	} else if loopLen == 0 {
		return loopBegin
	}

	dist := pos - loopEnd
	loopedPos := dist % loopLen
	return loopBegin + loopedPos
}

func calcLoopedSamplePosPingPong(pos int, length int, loopBegin int, loopEnd int) int {
	//  |start>-----------------------loopEnd|------------| <= on playthrough 1, whole sample plays
	//  |-------------|loopBegin>----<loopEnd|------------| <= only if looped and on playthrough 2+, part that loops plays and ping-pongs
	//  |-------------|loopBegin>----------------------end| <= on playthrough 2+, the loop ends and playback continues to end, if !keyOn
	if pos < 0 {
		return 0
	}
	if pos < loopEnd {
		return pos
	}

	loopLen := loopEnd - loopBegin
	if loopLen < 0 {
		if pos < length {
			return pos
		}
		return length
	} else if loopLen == 0 {
		return loopBegin
	}

	dist := pos - loopEnd
	loopedPos := dist % loopLen
	if times := int(dist / loopLen); (times & 1) == 0 {
		// even loops are reversed
		return loopEnd - loopedPos - 1
	}
	// odd loops are forward... or normal loop
	return loopBegin + loopedPos
}