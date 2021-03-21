package command

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/spf13/cobra"
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
			go func() {
				log.Println(http.ListenAndServe(profilerBindAddress, nil))
			}()
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return playCmd.RunE(playCmd, args)
	},
}

func Execute() {
	args := os.Args[1:]
	cmd, _, err := rootCmd.Find(args)
	if err != nil || cmd == rootCmd {
		// assume play command if args not provided
		os.Args = append([]string{os.Args[0], "play"}, args...)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	if persistFlags := rootCmd.PersistentFlags(); persistFlags != nil {
		persistFlags.BoolVar(&profiler, "profiler", profiler, "enable profiler (and supporting http server)")
		persistFlags.StringVar(&profilerBindAddress, "profiler-bind-addr", profilerBindAddress, "profiler bind address (if enabled)")
	}
}
