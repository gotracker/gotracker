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
	profiler bool
)

var rootCmd = &cobra.Command{
	Use:   "gotracker",
	Short: "Gotracker is a tracked music player",
	Long:  `Gotracker is a tracked music player written entirely in Go`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return playCmd.RunE(playCmd, args)
	},
}

func Execute() {
	if profiler {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

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
	rootCmd.PersistentFlags().BoolVarP(&profiler, "profiler", "p", false, "enable profiler (and supporting http server)")
}
