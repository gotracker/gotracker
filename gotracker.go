package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"gotracker/internal/player"
	"gotracker/internal/player/render"
	"gotracker/internal/player/state"
	"gotracker/internal/s3m"
	"gotracker/internal/winmm"
)

var (
	sampler       render.Sampler
	startingOrder int
)

// Play starts a song playing
func Play(ss *state.Song) <-chan render.RowRender {
	out := make(chan render.RowRender, 64)
	go func() {
		for {
			renderData := player.RenderOneRow(ss, &sampler)
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

// WaveOut is a wave output device
type WaveOut winmm.Device

func openWaveOut() *WaveOut {
	var handle = winmm.WaveOutOpen(sampler.Channels, sampler.SampleRate, sampler.BitsPerSample)
	if handle == nil {
		log.Fatal("Could not open WAVE_MAPPER interface for output")
		return nil
	}
	return (*WaveOut)(handle)
}

// Play starts the wave output device playing
func (waveOut *WaveOut) Play(in <-chan render.RowRender) {
	type RowWave struct {
		Wave *winmm.Wave
		Text string
	}
	out := make(chan RowWave, 3)
	go func() {
		ok := true
		var row render.RowRender
		for ok {
			row, ok = <-in
			if ok {
				var rowWave RowWave
				rowWave.Wave = winmm.WaveOutWrite(winmm.Device(*waveOut), row.RenderData)
				rowWave.Text = fmt.Sprintf("[%0.2d:%0.2d] %s", row.Order, row.Row, row.RowText.String(13))
				out <- rowWave
			}
		}
		close(out)
	}()
	ok := true
	var rowWave RowWave
	for ok {
		rowWave, ok = <-out
		if ok {
			fmt.Println(rowWave.Text)
			for !winmm.WaveOutFinished(winmm.Device(*waveOut), rowWave.Wave) {
				time.Sleep(time.Microsecond * 1)
			}
		}
	}
}

func main() {
	flag.IntVar(&sampler.SampleRate, "s", 44100, "sample rate")
	flag.IntVar(&sampler.Channels, "c", 2, "channels")
	flag.IntVar(&sampler.BitsPerSample, "b", 16, "bits per sample")
	flag.IntVar(&startingOrder, "o", -1, "starting order")

	flag.Parse()

	var fn string
	if len(flag.Args()) > 0 {
		fn = flag.Arg(0)
	}

	if fn == "" {
		flag.Usage()
		return
	}

	ss := state.NewSong()
	if err := s3m.Load(ss, fn); err != nil {
		log.Fatalf("Could not create song state! err[%v]", err)
		return
	}
	sampler.BaseClockRate = s3m.GetBaseClockRate()
	if startingOrder != -1 {
		ss.Pattern.CurrentOrder = uint8(startingOrder)
	}

	fmt.Println(ss.SongData.GetName())

	waveOut := openWaveOut()

	var buffers <-chan render.RowRender

	buffers = Play(ss)
	waveOut.Play(buffers)
}
