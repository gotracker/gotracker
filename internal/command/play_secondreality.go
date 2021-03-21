package command

import (
	"gotracker/internal/command/internal/playlist"

	"github.com/spf13/cobra"
)

// flags
var (
	pmPath   string = "2nd_pm.s3m"
	skavPath string = "2nd_skav.s3m"
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

		songs := playlist.New()
		songs.Add(playlist.Song{
			Filepath: skavPath,
			Start: playlist.Position{
				Order: 0,
				Row:   0,
			},
			End: playlist.Position{
				Order: 15,
				Row:   0,
			},
		})
		songs.Add(playlist.Song{
			Filepath: pmPath,
			Start: playlist.Position{
				Order: -1,
				Row:   -1,
			},
			End: playlist.Position{
				Order: 83,
				Row:   56,
			},
		})
		songs.Add(playlist.Song{
			Filepath: skavPath,
			Start: playlist.Position{
				Order: 18,
				Row:   0,
			},
			End: playlist.Position{
				Order: -1,
				Row:   -1,
			},
			Loop: loopSong,
		})

		randomized = false

		playedAtLeastOne, err := playSongs(songs, loopPlaylist)
		if err != nil {
			return err
		}

		if !playedAtLeastOne {
			return cmd.Usage()
		}

		return nil
	},
}
