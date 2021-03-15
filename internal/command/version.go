package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var (
	// Version is the version string
	Version string = "No Version Provided"
	// GitHash is the hash for the current git change
	GitHash string = "HEAD"

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Gotracker",
		Long:  `All software has versions. This is Gotracker's`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Gotracker %s -- %s\n", Version, GitHash)
		},
	}
)
