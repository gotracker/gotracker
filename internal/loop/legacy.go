package loop

func calcLoopPosLegacy(pos int, length int, loopBegin int, loopEnd int) (int, bool) {
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
