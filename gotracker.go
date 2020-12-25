package main

import (
	"flag"
	"fmt"
	"log"

	progressBar "github.com/cheggaaa/pb"
	device "github.com/gotracker/gosound"

	"gotracker/internal/format"
	"gotracker/internal/output"
	"gotracker/internal/player"
	"gotracker/internal/player/feature"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/render"
	"gotracker/internal/player/state"
)

// flags
var (
	outputSettings device.Settings
	startingOrder  int
	startingRow    int
	canLoop        bool
)

// local vars
var (
	sampler *render.Sampler
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

	sampler = render.NewSampler(outputSettings.SamplesPerSecond, outputSettings.Channels, outputSettings.BitsPerSample)

	ss := state.NewSong()
	if fmt, err := format.Load(ss, fn); err != nil {
		log.Fatalf("Could not create song state! err[%v]", err)
		return
	} else if fmt != nil {
		sampler.BaseClockRate = fmt.GetBaseClockRate()
	}
	if startingOrder != -1 {
		ss.Pattern.CurrentOrder = intf.OrderIdx(startingOrder)
	}
	if startingRow != -1 {
		ss.Pattern.CurrentRow = intf.RowIdx(startingRow)
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
			fmt.Printf("[%0.2d:%0.2d] %s\n", row.Order, row.Row, row.RowText.String())
		case device.KindFile:
			if progress == nil {
				progress = progressBar.StartNew(len(ss.SongData.GetOrderList()))
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
		disableFeatures = append(disableFeatures, feature.PatternLoop)
	}
	ss.DisableFeatures(disableFeatures)

	fmt.Printf("Output device: %s\n", waveOut.Name())
	fmt.Printf("Song: %s\n", ss.SongData.GetName())
	outBufs := make(chan *device.PremixData, 64)
	defer close(outBufs)

	p, err := player.NewPlayer(nil, outBufs, sampler)
	if err != nil {
		log.Fatalln(err)
		return
	}

	if err := p.Play(ss); err != nil {
		log.Fatalln(err)
		return
	}

	//if err := p.WaitUntilDone(); err != nil {
	//	log.Fatalln(err)
	//	return
	//}

	waveOut.Play(outBufs)
}
