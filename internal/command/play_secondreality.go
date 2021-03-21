package command

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"gotracker/internal/command/internal/playlist"
	"gotracker/internal/optional"
)

// flags
var (
	pmPath   string = filepath.Join(".", "2nd_pm.s3m")
	skavPath string = filepath.Join(".", "2nd_skav.s3m")
)

func init() {
	if flags := secondRealityCmd.Flags(); flags != nil {
		flags.StringVar(&pmPath, "pm", pmPath, "path to 2nd_pm.s3m")
		flags.StringVar(&skavPath, "skav", skavPath, "path to 2nd_skav.s3m")
	}

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

		pl := playlist.New()
		pl.Add(playlist.Song{
			Filepath: skavPath,
			End: playlist.Position{
				Order: optional.NewValue(15),
				Row:   optional.NewValue(0),
			},
		})
		pl.Add(playlist.Song{
			Filepath: pmPath,
			End: playlist.Position{
				Order: optional.NewValue(83),
				Row:   optional.NewValue(56),
			},
		})
		pl.Add(playlist.Song{
			Filepath: skavPath,
			Start: playlist.Position{
				Order: optional.NewValue(18),
				Row:   optional.NewValue(0),
			},
			Loop: optional.NewValue(loopSong),
		})

		pl.SetLooping(loopPlaylist)
		pl.SetRandomized(false)

		playedAtLeastOne, err := playSongs(pl)
		if err != nil {
			return err
		}

		if !playedAtLeastOne {
			return cmd.Usage()
		}

		return nil
	},
}
