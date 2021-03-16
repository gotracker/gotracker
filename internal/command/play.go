package command

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sort"
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
	outputSettings         device.Settings
	startingOrder          int
	startingRow            int
	canLoop                bool
	effectCoverage         bool
	panicOnUnhandledEffect bool
	disableNativeSamples   bool
	//disablePreconvertSamples bool
)

func init() {
	output.Setup()

	playCmd.PersistentFlags().IntVarP(&outputSettings.SamplesPerSecond, "sample-rate", "s", 44100, "sample rate")
	playCmd.PersistentFlags().IntVarP(&outputSettings.Channels, "channels", "c", 2, "channels")
	playCmd.PersistentFlags().IntVarP(&outputSettings.BitsPerSample, "bits-per-sample", "b", 16, "bits per sample")
	playCmd.PersistentFlags().IntVarP(&startingOrder, "starting-order", "o", -1, "starting order")
	playCmd.PersistentFlags().IntVarP(&startingRow, "starting-row", "r", -1, "starting row")
	playCmd.PersistentFlags().BoolVarP(&canLoop, "can-loop", "l", false, "enable pattern loop (only works in single-song mode)")
	playCmd.PersistentFlags().StringVarP(&outputSettings.Name, "output", "O", output.DefaultOutputDeviceName, "output device")
	playCmd.PersistentFlags().StringVarP(&outputSettings.Filepath, "output-file", "f", "output.wav", "output filepath")
	playCmd.PersistentFlags().BoolVarP(&effectCoverage, "gather-effect-coverage", "E", false, "gather and display effect coverage data")
	playCmd.PersistentFlags().BoolVarP(&panicOnUnhandledEffect, "unhandled-effect-panic", "P", false, "panic when an unhandled effect is encountered")
	playCmd.PersistentFlags().BoolVarP(&disableNativeSamples, "disable-native-samples", "N", false, "disable preconversion of samples to native sampling format")
	//playCmd.PersistentFlags().BoolVarP(&disablePreconvertSamples, "disable-preconvert-samples", "S", false, "disable preconversion of samples to 32-bit floats")

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
			})
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

type orderDetails struct {
	order int
	row   int
}

type songDetails struct {
	fn    string
	start orderDetails
	end   orderDetails
}

func playSongs(songs []songDetails) (bool, error) {
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
				fmt.Printf("[%0.3d:%0.3d] %s\n", row.Order, row.Row, row.RowText.String())
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

	outBufs := make(chan *device.PremixData, 64)

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

	if len(songs) != 1 {
		canLoop = false
	}
	configuration = append(configuration, feature.SongLoop{Enabled: canLoop})
	configuration = append(configuration, feature.IgnoreUnknownEffect{Enabled: !panicOnUnhandledEffect})

	fmt.Printf("Output device: %s\n", waveOut.Name())

	playedAtLeastOne, err := renderSongs(songs, outBufs, options, configuration, func(pb intf.Playback, tickInterval time.Duration) error {
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

		fmt.Printf("Order Looping Enabled: %v\n", playback.CanOrderLoop())
		fmt.Printf("Song: %s\n", playback.GetName())

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

	fmt.Println()
	fmt.Println("done!")

	return true, nil
}

func renderSongs(songs []songDetails, outBufs chan<- *device.PremixData, options []settings.OptionFunc, configuration []feature.Feature, startPlayingCB func(pb intf.Playback, tickInterval time.Duration) error) (bool, error) {
	defer close(outBufs)

	tickInterval := time.Duration(5) * time.Millisecond
	disableSleepIdx := sort.Search(len(configuration), func(i int) bool {
		switch configuration[i].(type) {
		case feature.PlayerSleepInterval:
			return true
		}
		return false
	})
	if disableSleepIdx < len(configuration) {
		feat := configuration[disableSleepIdx]
		switch f := feat.(type) {
		case feature.PlayerSleepInterval:
			if f.Enabled {
				tickInterval = f.Interval
			} else {
				tickInterval = time.Duration(0)
			}
		}
	}

	var playedAtLeastOne bool
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

		playback.Configure(cfg)

		if err = startPlayingCB(playback, tickInterval); err != nil {
			continue
		}

		playedAtLeastOne = true
	}

	return playedAtLeastOne, nil
}
