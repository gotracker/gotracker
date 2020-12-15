package main

import (
	"flag"
	"fmt"
	"log"

	"gotracker/internal/format"
	"gotracker/internal/output"
	"gotracker/internal/player"
	"gotracker/internal/player/render"
	"gotracker/internal/player/state"
)

// flags
var (
	outputSettings output.Settings
	startingOrder  int
)

// local vars
var (
	sampler *render.Sampler
)

// Play starts a song playing
func Play(ss *state.Song) <-chan render.RowRender {
	out := make(chan render.RowRender, 64)
	go func() {
		for {
			renderData := player.RenderOneRow(ss, sampler)
			if renderData != nil {
				if renderData.Stop {
					break
				} else if renderData.RenderData != nil && len(renderData.RenderData) != 0 {
					out <- *renderData
				}
			}
		}
		close(out)
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

	outputSettings.OnRowOutput = func(row render.RowRender) {
		fmt.Printf("[%0.2d:%0.2d] %s\n", row.Order, row.Row, row.RowText.String(13))
	}

	waveOut, disableFeatures, err := output.CreateOutputDevice(outputSettings)
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer waveOut.Close()

	var buffers <-chan render.RowRender

	ss.DisableFeatures(disableFeatures)

	buffers = Play(ss)
	waveOut.Play(buffers)
}
