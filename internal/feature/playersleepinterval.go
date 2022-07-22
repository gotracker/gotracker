package feature

import "time"

// PlayerSleepInterval describes the player sleep feature
type PlayerSleepInterval struct {
	Enabled  bool
	Interval time.Duration
}
