package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"time"

	progressBar "github.com/cheggaaa/pb"
	device "github.com/gotracker/gosound"

	"gotracker/internal/format"
	"gotracker/internal/output"
	"gotracker/internal/player"
	"gotracker/internal/player/feature"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/render"
)

// flags
var (
	outputSettings device.Settings
	startingOrder  int
	startingRow    int
	canLoop        bool
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
		playback.SetNextOrder(intf.OrderIdx(startingOrder))
	}
	if startingRow != -1 {
		playback.SetNextRow(intf.RowIdx(startingRow))
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

	waveOut, disableFeatures, err := output.CreateOutputDevice(outputSettings)
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer waveOut.Close()

	if !canLoop {
		disableFeatures = append(disableFeatures, feature.OrderLoop)
	}
	playback.DisableFeatures(disableFeatures)

	fmt.Printf("Output device: %s\n", waveOut.Name())
	fmt.Printf("Order Looping Enabled: %v\n", playback.CanOrderLoop())
	fmt.Printf("Song: %s\n", playback.GetName())
	outBufs := make(chan *device.PremixData, 64)

	tickInterval := time.Duration(5) * time.Millisecond
	disableSleepIdx := sort.Search(len(disableFeatures), func(i int) bool {
		return disableFeatures[i] == feature.PlayerSleepInterval
	})
	if disableSleepIdx < len(disableFeatures) && disableFeatures[disableSleepIdx] == feature.PlayerSleepInterval {
		tickInterval = time.Duration(0)
	}

	p, err := player.NewPlayer(nil, outBufs, tickInterval)
	if err != nil {
		log.Fatalln(err)
		return
	}

	if err := p.Play(playback); err != nil {
		log.Fatalln(err)
		return
	}

	go func() {
		defer close(outBufs)
		p.WaitUntilDone()
	}()

	waveOut.Play(outBufs)
	fmt.Println("done!")
}
