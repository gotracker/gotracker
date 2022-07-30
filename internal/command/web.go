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
		Short: "Open up a web-based user interface",
		Long:  `Opens up a web-based user interface.`,
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
