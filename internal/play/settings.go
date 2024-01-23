package play

import (
	"github.com/gotracker/gotracker/internal/config"
	deviceCommon "github.com/gotracker/gotracker/internal/output/device/common"
)

type Settings struct {
	Output              *config.Config[deviceCommon.Settings]
	NumPremixBuffers    int  `pflag:"num-buffers" env:"num_buffers" usage:"number of premixed buffers"`
	ITLongChannelOutput bool `pflag:"it-long" env:"it_long" usage:"enable Impulse Tracker long channel display"`
	ITEnableNNA         bool `pflag:"it-enable-nna" env:"it_enable_nna" usage:"enable Impulse Tracker New Note Actions"`
	Debug               *config.Config[DebugSettings]
}

type DebugSettings struct {
	PanicOnUnhandledEffect bool   `flag:"unhandled-effect-panic" env:"unhandled_effect_panic" usage:"panic when an unhandled effect is encountered"`
	Tracing                bool   `flag:"tracing" env:"tracing" usage:"enable tracing"`
	TracingFile            string `flag:"tracing-file" env:"tracing_file" usage:"tracing file to output to if tracing is enabled"`
}
