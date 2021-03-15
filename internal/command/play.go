package command

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sort"
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

	playCmd.Flags().IntVarP(&outputSettings.SamplesPerSecond, "sample-rate", "s", 44100, "sample rate")
	playCmd.Flags().IntVarP(&outputSettings.Channels, "channels", "c", 2, "channels")
	playCmd.Flags().IntVarP(&outputSettings.BitsPerSample, "bits-per-sample", "b", 16, "bits per sample")
	playCmd.Flags().IntVarP(&startingOrder, "starting-order", "o", -1, "starting order")
	playCmd.Flags().IntVarP(&startingRow, "starting-row", "r", -1, "starting row")
	playCmd.Flags().BoolVarP(&canLoop, "can-loop", "l", false, "enable pattern loop")
	playCmd.Flags().StringVarP(&outputSettings.Name, "output", "O", output.DefaultOutputDeviceName, "output device")
	playCmd.Flags().StringVarP(&outputSettings.Filepath, "output-file", "f", "output.wav", "output filepath")
	playCmd.Flags().BoolVarP(&effectCoverage, "gather-effect-coverage", "E", false, "gather and display effect coverage data")
	playCmd.Flags().BoolVarP(&panicOnUnhandledEffect, "unhandled-effect-panic", "P", false, "panic when an unhandled effect is encountered")
	playCmd.Flags().BoolVarP(&disableNativeSamples, "disable-native-samples", "N", false, "disable preconversion of samples to native sampling format")
	//playCmd.Flags().BoolVarP(&disablePreconvertSamples, "disable-preconvert-samples", "S", false, "disable preconversion of samples to 32-bit floats")

	rootCmd.AddCommand(playCmd)
}

var playCmd = &cobra.Command{
	Use:   "play [flags] <file>",
	Short: "Play a tracked music file using Gotracker",
	Long:  "Play a tracked music file using Gotracker.",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Usage()
		}

		fn := args[0]
		if fn == "" {
			return cmd.Usage()
		}

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

		playback, songFmt, err := format.Load(fn, options...)
		if err != nil {
			return fmt.Errorf("Could not create song state! err[%v]", err)
		} else if songFmt != nil {
			if err := playback.SetupSampler(outputSettings.SamplesPerSecond, outputSettings.Channels, outputSettings.BitsPerSample); err != nil {
				return fmt.Errorf("Could not setup playback sampler! err[%v]", err)
			}
		}
		if startingOrder != -1 {
			if err := playback.SetNextOrder(index.Order(startingOrder)); err != nil {
				return fmt.Errorf("Could not set starting order! err[%v]", err)
			}
		}
		if startingRow != -1 {
			if err := playback.SetNextRow(index.Row(startingRow)); err != nil {
				return fmt.Errorf("Could not set starting row! err[%v]", err)
			}
		}

		var (
			progress  *progressBar.ProgressBar
			lastOrder int
		)

		defer func() {
			if progress != nil {
				progress.Set64(progress.Total)
				progress.Finish()
			}
		}()

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
			return err
		}
		defer waveOut.Close()

		configuration = append(configuration, feature.SongLoop{Enabled: canLoop})
		configuration = append(configuration, feature.IgnoreUnknownEffect{Enabled: !panicOnUnhandledEffect})
		playback.Configure(configuration)

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

		fmt.Printf("Output device: %s\n", waveOut.Name())
		fmt.Printf("Order Looping Enabled: %v\n", playback.CanOrderLoop())
		fmt.Printf("Song: %s\n", playback.GetName())
		outBufs := make(chan *device.PremixData, 64)

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

		p, err := player.NewPlayer(context.TODO(), outBufs, tickInterval)
		if err != nil {
			return err
		}

		if err := p.Play(playback); err != nil {
			return err
		}

		go func() {
			defer close(outBufs)
			if err := p.WaitUntilDone(); err != nil {
				switch {
				case errors.Is(err, song.ErrStopSong):
				case errors.Is(err, context.Canceled):

				default:
					log.Fatalln(err)
				}
			}
		}()

		if err := waveOut.Play(outBufs); err != nil {
			switch {
			case errors.Is(err, song.ErrStopSong):
			case errors.Is(err, context.Canceled):

			default:
				log.Fatalln(err)
			}
		}

		for k, v := range effectMap {
			fmt.Println(k, v)
		}
		fmt.Println()

		fmt.Println("done!")

		return nil
	},
}
