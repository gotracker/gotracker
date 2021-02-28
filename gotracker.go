package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"reflect"
	"sort"
	"time"

	progressBar "github.com/cheggaaa/pb"
	device "github.com/gotracker/gosound"

	"gotracker/internal/format"
	itEffect "gotracker/internal/format/it/playback/effect"
	s3mEffect "gotracker/internal/format/s3m/playback/effect"
	xmEffect "gotracker/internal/format/xm/playback/effect"
	"gotracker/internal/output"
	"gotracker/internal/player"
	"gotracker/internal/player/feature"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/render"
)

// flags
var (
	outputSettings         device.Settings
	startingOrder          int
	startingRow            int
	canLoop                bool
	effectCoverage         bool
	panicOnUnhandledEffect bool
	profiler               bool
)

func main() {
	output.Setup()

	flag.IntVar(&outputSettings.SamplesPerSecond, "s", 44100, "sample rate")
	flag.IntVar(&outputSettings.Channels, "c", 2, "channels")
	flag.IntVar(&outputSettings.BitsPerSample, "b", 16, "bits per sample")
	flag.IntVar(&startingOrder, "o", -1, "starting order")
	flag.IntVar(&startingRow, "r", -1, "starting row")
	flag.BoolVar(&canLoop, "l", false, "enable pattern loop")
	flag.StringVar(&outputSettings.Name, "O", output.DefaultOutputDeviceName, "output device")
	flag.StringVar(&outputSettings.Filepath, "f", "output.wav", "output filepath")
	flag.BoolVar(&effectCoverage, "E", false, "gather and display effect coverage data")
	flag.BoolVar(&panicOnUnhandledEffect, "P", false, "panic when an unhandled effect is encountered")
	flag.BoolVar(&profiler, "p", false, "enable profiler (and supporting http server)")

	flag.Parse()

	var fn string
	if len(flag.Args()) > 0 {
		fn = flag.Arg(0)
	}

	if fn == "" {
		flag.Usage()
		return
	}

	playback, songFmt, err := format.Load(fn)
	if err != nil {
		log.Fatalf("Could not create song state! err[%v]", err)
		return
	} else if songFmt != nil {
		if err := playback.SetupSampler(outputSettings.SamplesPerSecond, outputSettings.Channels, outputSettings.BitsPerSample); err != nil {
			log.Fatalf("Could not setup playback sampler! err[%v]", err)
			return
		}
	}
	if startingOrder != -1 {
		if err := playback.SetNextOrder(intf.OrderIdx(startingOrder)); err != nil {
			log.Fatalf("Could not set starting order! err[%v]", err)
			return
		}
	}
	if startingRow != -1 {
		if err := playback.SetNextRow(intf.RowIdx(startingRow)); err != nil {
			log.Fatalf("Could not set starting row! err[%v]", err)
			return
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
		log.Fatalln(err)
		return
	}
	defer waveOut.Close()

	if !canLoop {
		configuration = append(configuration, feature.SongLoop{Enabled: false})
	}
	if panicOnUnhandledEffect {
		configuration = append(configuration, feature.IgnoreUnknownEffect{Enabled: true})
	}
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
		log.Fatalln(err)
		return
	}

	if profiler {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	if err := p.Play(playback); err != nil {
		log.Fatalln(err)
		return
	}

	go func() {
		defer close(outBufs)
		if err := p.WaitUntilDone(); err != nil {
			switch {
			case errors.Is(err, intf.ErrStopSong):
			case errors.Is(err, context.Canceled):

			default:
				log.Fatalln(err)
			}
		}
	}()

	if err := waveOut.Play(outBufs); err != nil {
		switch {
		case errors.Is(err, intf.ErrStopSong):
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
}
