package command

import (
	"github.com/gotracker/gotracker/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	if err := playSettings.Values.Debug.Overlay(config.StandardOverlays...).Update(playDebugCmd); err != nil {
		panic(err)
	}

	registerPlayFlags(playDebugCmd)

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
