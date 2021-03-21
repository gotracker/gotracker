package command

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	progressBar "github.com/cheggaaa/pb"
	device "github.com/gotracker/gosound"
	"github.com/spf13/cobra"

	"gotracker/internal/format"
	itEffect "gotracker/internal/format/it/playback/effect"
	s3mEffect "gotracker/internal/format/s3m/playback/effect"
	"gotracker/internal/format/settings"
	xmEffect "gotracker/internal/format/xm/playback/effect"
	"gotracker/internal/output"
	"gotracker/internal/player"
	"gotracker/internal/player/feature"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/render"
	"gotracker/internal/song"
	"gotracker/internal/song/index"
)

// flags
var (
	outputSettings = device.Settings{
		Channels:         2,
		SamplesPerSecond: 44100,
		BitsPerSample:    16,
		Filepath:         "output.wav",
	}
	startingOrder          int  = -1
	startingRow            int  = -1
	numPremixBuffers       int  = 64
	loopSong               bool = false
	loopPlaylist           bool = false
	silent                 bool = false
	effectCoverage         bool = false
	panicOnUnhandledEffect bool = false
	disableNativeSamples   bool = false
	//disablePreconvertSamples bool = false
)

func loggingf(format string, args ...interface{}) {
	if silent {
		return
	}
	fmt.Printf(format, args...)
}

func loggingln(args ...interface{}) {
	if silent {
		return
	}
	fmt.Println(args...)
}

// func logging(args ...interface{}) {
// 	if silent {
// 		return
// 	}
// 	fmt.Print(args...)
// }

func init() {
	output.Setup()

	if persistFlags := playCmd.PersistentFlags(); persistFlags != nil {
		persistFlags.IntVarP(&outputSettings.SamplesPerSecond, "sample-rate", "s", outputSettings.SamplesPerSecond, "sample rate")
		persistFlags.IntVarP(&outputSettings.Channels, "channels", "c", outputSettings.Channels, "channels")
		persistFlags.IntVarP(&outputSettings.BitsPerSample, "bits-per-sample", "b", outputSettings.BitsPerSample, "bits per sample")
		persistFlags.IntVarP(&startingOrder, "starting-order", "o", startingOrder, "starting order")
		persistFlags.IntVarP(&startingRow, "starting-row", "r", startingRow, "starting row")
		persistFlags.IntVarP(&numPremixBuffers, "num-buffers", "B", numPremixBuffers, "number of premixed buffers")
		persistFlags.BoolVarP(&loopSong, "loop-song", "l", loopSong, "enable pattern loop (only works in single-song mode)")
		persistFlags.BoolVarP(&loopPlaylist, "loop-playlist", "L", loopPlaylist, "enable playlist loop (only useful in multi-song mode)")
		persistFlags.BoolVarP(&silent, "silent", "q", silent, "disable non-error logging")
		persistFlags.StringVarP(&outputSettings.Name, "output", "O", output.DefaultOutputDeviceName, "output device")
		persistFlags.StringVarP(&outputSettings.Filepath, "output-file", "f", outputSettings.Filepath, "output filepath")
		persistFlags.BoolVarP(&effectCoverage, "gather-effect-coverage", "E", effectCoverage, "gather and display effect coverage data")
		persistFlags.BoolVarP(&panicOnUnhandledEffect, "unhandled-effect-panic", "P", panicOnUnhandledEffect, "panic when an unhandled effect is encountered")
		persistFlags.BoolVarP(&disableNativeSamples, "disable-native-samples", "N", disableNativeSamples, "disable preconversion of samples to native sampling format")
		//persistFlags.BoolVarP(&disablePreconvertSamples, "disable-preconvert-samples", "S", disablePreconvertSamples, "disable preconversion of samples to 32-bit floats")
	}

	rootCmd.AddCommand(playCmd)
}

var playCmd = &cobra.Command{
	Use:   "play [flags] <file(s)>",
	Short: "Play a tracked music file using Gotracker",
	Long:  "Play one or more tracked music file(s) using Gotracker.",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var songs []songDetails
		for _, fn := range args {
			songs = append(songs, songDetails{
				fn: fn,
				start: orderDetails{
					order: startingOrder,
					row:   startingRow,
				},
				end: orderDetails{
					order: -1,
					row:   -1,
				},
				loopEnabled: loopSong,
			})
		}

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

type orderDetails struct {
	order int
	row   int
}

type songDetails struct {
	fn          string
	start       orderDetails
	end         orderDetails
	loopEnabled bool
}

func playSongs(songs []songDetails, loopListDesired bool) (bool, error) {
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

	var (
		playback  intf.Playback
		progress  *progressBar.ProgressBar
		lastOrder int
	)

	outputSettings.OnRowOutput = func(deviceKind device.Kind, premix *device.PremixData) {
		row := premix.Userdata.(*render.RowRender)
		switch deviceKind {
		case device.KindSoundCard:
			if row.RowText != nil {
				loggingf("[%0.3d:%0.3d] %s\n", row.Order, row.Row, row.RowText.String())
			}
		case device.KindFile:
			if progress == nil {
				progress = progressBar.StartNew(playback.GetNumOrders())
				lastOrder = row.Order
			}
			if lastOrder != row.Order {
				progress.Increment()
				lastOrder = row.Order
			}
		}
	}

	waveOut, configuration, err := output.CreateOutputDevice(outputSettings)
	if err != nil {
		return false, err
	}
	defer waveOut.Close()

	outBufs := make(chan *device.PremixData, numPremixBuffers)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := waveOut.Play(outBufs); err != nil {
			switch {
			case errors.Is(err, song.ErrStopSong):
			case errors.Is(err, context.Canceled):

			default:
				log.Fatalln(err)
			}
		}
	}()

	configuration = append(configuration, feature.IgnoreUnknownEffect{Enabled: !panicOnUnhandledEffect})

	loggingf("Output device: %s\n", waveOut.Name())

	playedAtLeastOne, err := renderSongs(songs, outBufs, options, configuration, loopListDesired, func(pb intf.Playback, tickInterval time.Duration) error {
		playback = pb
		defer func() {
			if progress != nil {
				progress.Set64(progress.Total)
				progress.Finish()
			}
		}()

		var effectMap map[string]int
		if effectCoverage {
			effectMap = make(map[string]int)
			playback.SetOnEffect(func(e intf.Effect) {
				var name string
				switch t := e.(type) {
				case *xmEffect.VolEff:
					for _, eff := range t.Effects {
						typ := reflect.TypeOf(eff)
						name = typ.Name()
						effectMap[name]++
					}
				case *itEffect.VolEff:
					for _, eff := range t.Effects {
						typ := reflect.TypeOf(eff)
						name = typ.Name()
						effectMap[name]++
					}
				case *s3mEffect.UnhandledCommand:
					name = fmt.Sprintf("UnhandledCommand(%c)", t.Command+'@')
					effectMap[name]++
				default:
					typ := reflect.TypeOf(t)
					name = typ.Name()
					effectMap[name]++
				}
			})
		}

		loggingf("Order Looping Enabled: %v\n", playback.CanOrderLoop())
		loggingf("Song: %s\n", playback.GetName())

		p, err := player.NewPlayer(context.TODO(), outBufs, tickInterval)
		if err != nil {
			return err
		}

		if err := p.Play(playback); err != nil {
			return err
		}

		if err := p.WaitUntilDone(); err != nil {
			switch {
			case errors.Is(err, song.ErrStopSong):
			case errors.Is(err, context.Canceled):

			default:
				return err
			}
		}

		return nil
	})
	if !playedAtLeastOne || err != nil {
		return playedAtLeastOne, err
	}

	wg.Wait()

	loggingln()
	loggingln("done!")

	return true, nil
}

func findFeatureByName(configuration []feature.Feature, name string) (feature.Feature, bool) {
	for _, feature := range configuration {
		tf := reflect.TypeOf(feature)
		if tf.Name() == name {
			return feature, true
		}
	}
	return nil, false
}

func renderSongs(songs []songDetails, outBufs chan<- *device.PremixData, options []settings.OptionFunc, configuration []feature.Feature, loopListDesired bool, startPlayingCB func(pb intf.Playback, tickInterval time.Duration) error) (bool, error) {
	defer close(outBufs)

	tickInterval := time.Duration(5) * time.Millisecond
	if feat, found := findFeatureByName(configuration, "PlayerSleepInterval"); found {
		if f, ok := feat.(feature.PlayerSleepInterval); ok {
			if f.Enabled {
				tickInterval = f.Interval
			} else {
				tickInterval = time.Duration(0)
			}
		}
	}

	canPossiblyLoop := true
	if feat, found := findFeatureByName(configuration, "SongLoop"); found {
		if f, ok := feat.(feature.SongLoop); ok {
			canPossiblyLoop = f.Enabled
		}
	}

	var playedAtLeastOne bool
playlistLoop:
	for _, song := range songs {
		playback, songFmt, err := format.Load(song.fn, options...)
		if err != nil {
			return playedAtLeastOne, fmt.Errorf("Could not create song state! err[%v]", err)
		} else if songFmt != nil {
			if err := playback.SetupSampler(outputSettings.SamplesPerSecond, outputSettings.Channels, outputSettings.BitsPerSample); err != nil {
				return playedAtLeastOne, fmt.Errorf("Could not setup playback sampler! err[%v]", err)
			}
		}
		if song.start.order != -1 || song.start.row != -1 {
			txn := playback.StartPatternTransaction()
			defer txn.Cancel()
			if song.start.order != -1 {
				txn.SetNextOrder(index.Order(song.start.order))
			}
			if song.start.row != -1 {
				txn.SetNextRow(index.Row(song.start.row))
			}
			if err := txn.Commit(); err != nil {
				return playedAtLeastOne, err
			}
		}

		cfg := append([]feature.Feature{}, configuration...)
		if song.end.order != -1 && song.end.row != -1 {
			cfg = append(cfg, feature.PlayUntilOrderAndRow{
				Order: song.end.order,
				Row:   song.end.row,
			})
		}
		if canPossiblyLoop {
			cfg = append(cfg, feature.SongLoop{Enabled: song.loopEnabled})
		}

		playback.Configure(cfg)

		if err = startPlayingCB(playback, tickInterval); err != nil {
			continue
		}

		playedAtLeastOne = true
	}

	if loopPlaylist && loopListDesired {
		goto playlistLoop
	}

	return playedAtLeastOne, nil
}
