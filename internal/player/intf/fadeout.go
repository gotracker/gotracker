package intf

import (
	"github.com/gotracker/gomixing/volume"
)

// FadeoutMode is the mode used to process fade-out
type FadeoutMode uint8

const (
	// FadeoutModeDisabled is for when the fade-out is disabled (S3M/MOD)
	FadeoutModeDisabled = FadeoutMode(iota)
	// FadeoutModeAlwaysActive is for when the fade-out is always available to be used (IT-style)
	FadeoutModeAlwaysActive
	// FadeoutModeOnlyIfVolEnvActive is for when the fade-out only functions when VolEnv is enabled (XM-style)
	FadeoutModeOnlyIfVolEnvActive
)

// FadeoutSettings is the settings for fade-out
type FadeoutSettings struct {
	Mode   FadeoutMode
	Amount volume.Volume
}
