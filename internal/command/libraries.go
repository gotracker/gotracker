package command

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(librariesCmd)
}

var (
	librariesCmd = &cobra.Command{
		Use:   "libraries",
		Short: "List information about embedded libraries",
		Long:  `List information about embedded libraries.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
)
