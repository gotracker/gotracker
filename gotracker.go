package main

import (
	"flag"
	"fmt"
	"log"

	progressBar "github.com/cheggaaa/pb"
	"github.com/pkg/errors"

	"gotracker/internal/format"
	"gotracker/internal/module/player"
	"gotracker/internal/module/player/render"
	"gotracker/internal/module/player/state"
	"gotracker/internal/output"
	"gotracker/internal/output/device"
)

// flags
var (
	outputSettings device.Settings
	startingOrder  int
)

// local vars
var (
	sampler *render.Sampler
)

// Play starts a song playing
func Play(ss *state.Song) <-chan *device.PremixData {
	out := make(chan *device.PremixData, 64)
	go func() {
		defer close(out)
		for {
			premix, err := player.RenderOneRow(ss, sampler)
			if err != nil {
				if errors.Is(err, state.ErrStopSong) {
					break
				}
				log.Fatal(err)
			}
			if premix != nil && premix.Data != nil && len(premix.Data) != 0 {
				out <- premix
			}
		}
	}()
	return out
}

func main() {
	output.Setup()

	flag.IntVar(&outputSettings.SamplesPerSecond, "s", 44100, "sample rate")
	flag.IntVar(&outputSettings.Channels, "c", 2, "channels")
	flag.IntVar(&outputSettings.BitsPerSample, "b", 16, "bits per sample")
	flag.IntVar(&startingOrder, "o", -1, "starting order")
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
		ss.Pattern.CurrentOrder = uint8(startingOrder)
	}

	fmt.Println(ss.SongData.GetName())

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

	ss.DisableFeatures(disableFeatures)

	buffers := Play(ss)
	waveOut.Play(buffers)
}
