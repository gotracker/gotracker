//go:build web
// +build web

package command

import (
	"github.com/gotracker/gotracker/internal/web"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(webCmd)
}

var (
	webCmd = &cobra.Command{
		Use:   "web",
		Short: "Starts a REST-based api",
		Long:  `Starts a REST-based api.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			web.Enabled = true
			if err := cmd.Parent().PersistentPreRunE(cmd, args); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			web.WaitForShutdown()
			return nil
		},
	}
)
