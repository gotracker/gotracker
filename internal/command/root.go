package command

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/gotracker/gotracker/internal/profiling"
	"github.com/gotracker/gotracker/internal/web"
)

// flags
var (
	profiler         bool   = false
	webBindAddress   string = "localhost:6060"
	additionalRoutes []web.RouteActivator
)

var rootCmd = &cobra.Command{
	Use:   "gotracker",
	Short: "Gotracker is a tracked music player",
	Long:  `Gotracker is a tracked music player written entirely in Go`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// set up profiler
		profiling.Enabled = profiler
		// set up web (a requirement of profiler)
		web.Enabled = web.Enabled || profiler

		// listen for signals
		sigCtx, sigHandlerStop := signal.NotifyContext(context.Background(), os.Interrupt)

		// start up the web server (if enabled)
		web.Activate(sigCtx, webBindAddress, additionalRoutes...)
		go func() {
			defer sigHandlerStop()
			web.WaitForShutdown()
		}()
		return nil
	},
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Usage()
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		web.Shutdown()
		web.WaitForShutdown()
		return nil
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
		if profiling.Allowed() {
			persistFlags.BoolVar(&profiler, "profiler", profiler, "enable profiler (and supporting http server)")
		}
		if web.Allowed() || profiling.Allowed() {
			persistFlags.StringVarP(&webBindAddress, "web-bind-addr", "w", webBindAddress, "web (and/or profiler) bind address")
		}
	}
}
