package command

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deviceCmd)
}

var (
	deviceCmd = &cobra.Command{
		Use:   "device",
		Short: "Perform an operation regarding output devices",
		Long:  `Perform an operation regarding output devices.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
)
