package command

import (
	"github.com/spf13/cobra"
)

func init() {
	if flags := playDebugCmd.Flags(); flags != nil {
		flags.BoolVar(&playSettings.GatherEffectCoverage, "gather-effect-coverage", playSettings.GatherEffectCoverage, "gather and display effect coverage data")
		flags.BoolVar(&playSettings.PanicOnUnhandledEffect, "unhandled-effect-panic", playSettings.PanicOnUnhandledEffect, "panic when an unhandled effect is encountered")

		registerPlayFlags(flags)
	}

	playCmd.AddCommand(playDebugCmd)
}

var playDebugCmd = &cobra.Command{
	Use:   "debug [flags] <file(s)>",
	Short: "Play a tracked music file using Gotracker (with added debugging)",
	Long:  "Play one or more tracked music file(s) using Gotracker (with added debugging).",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		parent := cmd.Parent()
		return parent.RunE(cmd, args)
	},
}
