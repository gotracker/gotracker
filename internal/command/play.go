package command

import (
	device "github.com/gotracker/gosound"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"gotracker/internal/command/internal/logging"
	"gotracker/internal/command/internal/play"
	"gotracker/internal/command/internal/playlist"
	"gotracker/internal/format/settings"
	"gotracker/internal/optional"
	"gotracker/internal/output"
)

// persistent flags
var (
	playSettings = play.Settings{
		Output: device.Settings{
			Channels:         2,
			SamplesPerSecond: 44100,
			BitsPerSample:    16,
			Filepath:         "output.wav",
		},
		NumPremixBuffers:       64,
		PanicOnUnhandledEffect: false,
		GatherEffectCoverage:   false,
	}
	loopPlaylist         bool = false
	logger               logging.Squelchable
	disableNativeSamples bool = false
	//disablePreconvertSamples bool = false
)

// flags
var (
	loopSong      bool = false
	startingOrder int  = -1
	startingRow   int  = -1
	randomized    bool = false
)

func init() {
	output.Setup()

	if persistFlags := playCmd.PersistentFlags(); persistFlags != nil {
		persistFlags.IntVarP(&playSettings.Output.SamplesPerSecond, "sample-rate", "s", playSettings.Output.SamplesPerSecond, "sample rate")
		persistFlags.IntVarP(&playSettings.Output.Channels, "channels", "c", playSettings.Output.Channels, "channels")
		persistFlags.IntVarP(&playSettings.Output.BitsPerSample, "bits-per-sample", "b", playSettings.Output.BitsPerSample, "bits per sample")
		persistFlags.IntVar(&playSettings.NumPremixBuffers, "num-buffers", playSettings.NumPremixBuffers, "number of premixed buffers")
		persistFlags.BoolVarP(&loopPlaylist, "loop-playlist", "L", loopPlaylist, "enable playlist loop (only useful in multi-song mode)")
		persistFlags.BoolVarP(&logger.Squelch, "silent", "q", logger.Squelch, "disable non-error logging")
		persistFlags.StringVarP(&playSettings.Output.Name, "output", "O", output.DefaultOutputDeviceName, "output device")
		persistFlags.StringVarP(&playSettings.Output.Filepath, "output-file", "f", playSettings.Output.Filepath, "output filepath")
		persistFlags.BoolVar(&disableNativeSamples, "disable-native-samples", disableNativeSamples, "disable preconversion of samples to native sampling format")
		//persistFlags.BoolVar(&disablePreconvertSamples, "disable-preconvert-samples", disablePreconvertSamples, "disable preconversion of samples to 32-bit floats")
	}

	registerPlayFlags(playCmd.Flags())

	rootCmd.AddCommand(playCmd)
}

func registerPlayFlags(flags *pflag.FlagSet) {
	if flags == nil {
		return
	}
	flags.IntVarP(&startingOrder, "starting-order", "o", startingOrder, "starting order")
	flags.IntVarP(&startingRow, "starting-row", "r", startingRow, "starting row")
	flags.BoolVarP(&loopSong, "loop-song", "l", loopSong, "enable pattern loop (only works in single-song mode)")
	flags.BoolVarP(&randomized, "random", "R", randomized, "randomize the playlist")
}

var playCmd = &cobra.Command{
	Use:   "play [flags] <file(s)>",
	Short: "Play a tracked music file using Gotracker",
	Long:  "Play one or more tracked music file(s) using Gotracker.",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		pl := playlist.New()
		for _, fn := range args {
			song := playlist.Song{
				Filepath: fn,
				Start: playlist.Position{
					Order: optional.NewValue(startingOrder),
					Row:   optional.NewValue(startingRow),
				},
			}
			if len(args) == 1 {
				song.Loop.Set(loopSong)
			}
			pl.Add(song)
		}

		pl.SetLooping(loopPlaylist)
		pl.SetRandomized(randomized)

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

func playSongs(pl *playlist.Playlist) (bool, error) {
	var options []settings.OptionFunc
	// NOTE: JBC - disabled because Native Samples are working now :)
	// leaving this code here so down-rezing of samples can be added later.
	//if !disablePreconvertSamples {
	//	var preferredSampleFormat pcm.SampleDataFormat = pcm.SampleDataFormat32BitLEFloat
	//	// HACK: I wish we had access to the `sys.BigEndian` bool
	//	if (*(*[2]uint8)(unsafe.Pointer(&[]uint16{1}[0])))[0] == 0 {
	//		preferredSampleFormat = pcm.SampleDataFormat32BitBEFloat
	//	}
	//	options = append(options, settings.PreferredSampleFormat(preferredSampleFormat))
	//}
	if !disableNativeSamples {
		options = append(options, settings.UseNativeSampleFormat())
	}

	return play.Playlist(pl, options, &playSettings, &logger)
}
