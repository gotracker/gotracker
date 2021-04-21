package command

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(playlistCmd)
}

var (
	playlistCmd = &cobra.Command{
		Use:   "playlist",
		Short: "Perform playlist operations",
		Long:  `Perform playlist-related operations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
)
