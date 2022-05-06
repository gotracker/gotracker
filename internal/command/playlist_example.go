package command

import (
	"io"
	"os"
	"path/filepath"

	"github.com/gotracker/gotracker/internal/command/internal/playlist"
	"github.com/gotracker/gotracker/internal/optional"

	"github.com/spf13/cobra"
)

var (
	pmPath   string = filepath.Join(".", "2ND_PM.S3M")
	skavPath string = filepath.Join(".", "2ND_SKAV.S3M")

	playlistExampleOutputFilepath string
)

func init() {
	if flags := playlistExampleCmd.Flags(); flags != nil {
		flags.StringVarP(&playlistExampleOutputFilepath, "output", "o", playlistExampleOutputFilepath, "output path [blank for stdout]")
	}

	playlistCmd.AddCommand(playlistExampleCmd)
}

var (
	playlistExampleCmd = &cobra.Command{
		Use:   "example",
		Short: "Create an example playlist file",
		Long:  `Create an example playlist file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			pl := playlist.New()
			pl.Add(playlist.Song{
				Filepath: skavPath,
				End: playlist.Position{
					Order: optional.NewValue(15),
					//Row:   optional.NewValue(0),
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
					//Row:   optional.NewValue(0),
				},
				Loop: playlist.Loop{
					Count: playlist.NewLoopForever(),
				},
			})

			var ow io.Writer = os.Stdout
			if playlistExampleOutputFilepath != "" {
				basePath := filepath.Dir(playlistExampleOutputFilepath)
				if basePath != "" && basePath != "." {
					if err := os.MkdirAll(basePath, 0755); err != nil {
						return err
					}
				}
				var err error
				ow, err = os.Create(playlistExampleOutputFilepath)
				if err != nil {
					return err
				}
			}

			return pl.WriteYAML(ow)
		},
	}
)
