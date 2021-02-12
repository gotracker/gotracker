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

// simple helper to consolidate loop length calculations
// (yeah, it could just be the math in place, but whatever)
func calcLoopLen(loopBegin int, loopEnd int) int {
	return loopEnd - loopBegin
}

// NewLoop creates a loop based on the specified mode and settings
func NewLoop(mode Mode, settings Settings) Loop {
	switch mode {
	case ModeDisabled:
		return &Disabled{}
	case ModeLegacy:
		return &Legacy{
			Settings: settings,
		}
	case ModeNormal:
		return &Normal{
			Settings: settings,
		}
	case ModePingPong:
		return &PingPong{
			Settings: settings,
		}
	default:
		panic("unhandled loop mode")
	}
}
