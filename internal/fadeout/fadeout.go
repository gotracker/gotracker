package fadeout

import (
	"github.com/gotracker/gomixing/volume"
)

// Mode is the mode used to process fade-out
type Mode uint8

const (
	// ModeDisabled is for when the fade-out is disabled (S3M/MOD)
	ModeDisabled = Mode(iota)
	// ModeAlwaysActive is for when the fade-out is always available to be used (IT-style)
	ModeAlwaysActive
	// ModeOnlyIfVolEnvActive is for when the fade-out only functions when VolEnv is enabled (XM-style)
	ModeOnlyIfVolEnvActive
)

// Settings is the settings for fade-out
type Settings struct {
	Mode   Mode
	Amount volume.Volume
}
