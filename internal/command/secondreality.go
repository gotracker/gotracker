package command

import (
	"github.com/spf13/cobra"
)

// flags
var (
	pmPath   string = "2nd_pm.s3m"
	skavPath string = "2nd_skav.s3m"
)

func init() {
	secondRealityCmd.Flags().StringVar(&pmPath, "pm", pmPath, "path to 2nd_pm.s3m")
	secondRealityCmd.Flags().StringVar(&skavPath, "skav", skavPath, "path to 2nd_skav.s3m")

	playCmd.AddCommand(secondRealityCmd)
}

var secondRealityCmd = &cobra.Command{
	Use:   "second-reality [flags] --pm <path to 2nd_pm.s3m> --skav <path to 2nd_skav.s3m>",
	Short: "Stitch together the Second Reality music files and play them using Gotracker",
	Long:  "Stitch together the Second Reality music files and play them using Gotracker.",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if pmPath == "" || skavPath == "" {
			return cmd.Usage()
		}

		songs := []songDetails{
			{
				fn: skavPath,
				start: orderDetails{
					order: 0,
					row:   0,
				},
				end: orderDetails{
					order: 15,
					row:   0,
				},
			},
			{
				fn: pmPath,
				start: orderDetails{
					order: -1,
					row:   -1,
				},
				end: orderDetails{
					order: -1,
					row:   -1,
				},
			},
			{
				fn: skavPath,
				start: orderDetails{
					order: 18,
					row:   0,
				},
				end: orderDetails{
					order: -1,
					row:   -1,
				},
			},
		}

		playedAtLeastOne, err := playSongs(songs)
		if err != nil {
			return err
		}

		if !playedAtLeastOne {
			return cmd.Usage()
		}

		return nil
	},
}
