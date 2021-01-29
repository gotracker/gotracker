package loop

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
