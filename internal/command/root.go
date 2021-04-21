package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gotracker/internal/command/internal/profiling"
)

// flags
var (
	profiler            bool   = false
	profilerBindAddress string = "localhost:6060"
)

var rootCmd = &cobra.Command{
	Use:   "gotracker",
	Short: "Gotracker is a tracked music player",
	Long:  `Gotracker is a tracked music player written entirely in Go`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if profiler {
			profiling.Activate(profilerBindAddress)
		}
		return nil
	},
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Usage()
	},
}

func Execute() {
	args := os.Args[1:]
	cmd, _, err := rootCmd.Find(args)
	cmdExec := rootCmd
	if err != nil || (cmd == rootCmd && len(os.Args) > 1) {
		// assume play command if command argument not provided
		cmdExec = playCmd
		os.Args = append([]string{os.Args[0], "play"}, args...)
	}

	if err := cmdExec.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	if persistFlags := rootCmd.PersistentFlags(); persistFlags != nil {
		if profiling.Allowed {
			persistFlags.BoolVar(&profiler, "profiler", profiler, "enable profiler (and supporting http server)")
			persistFlags.StringVar(&profilerBindAddress, "profiler-bind-addr", profilerBindAddress, "profiler bind address (if enabled)")
		}
	}
}
