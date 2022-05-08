package main_test

import (
	"errors"
	"testing"
	"time"
	"unsafe"

	"github.com/gotracker/gotracker/internal/format"
	"github.com/gotracker/gotracker/internal/player/feature"
	"github.com/gotracker/gotracker/internal/player/intf"
	"github.com/gotracker/gotracker/internal/song"

	"github.com/gotracker/gosound"
)

func BenchmarkPlayerS3M(b *testing.B) {
	var playback intf.Playback
	var err error
	b.Run("load_s3m", func(b *testing.B) {
		playback, _, err = format.Load("test/celestial_fantasia.s3m")
		if err != nil {
			b.Error(err)
		}
	})

	if err := playback.SetupSampler(44100, 2, 16); err != nil {
		b.Error(err)
	}

	playback.Configure([]feature.Feature{feature.SongLoop{Count: 0}})

	lastTime := time.Now()
	for err == nil {
		now := time.Now()
		b.Run("generate_s3m", func(b *testing.B) {
			b.Helper()
			b.ReportAllocs()
			var premix *gosound.PremixData
			premix, err = playback.Generate(now.Sub(lastTime))
			if err != nil {
				if !errors.Is(err, song.ErrStopSong) {
					b.Error(err)
				}
				return
			}
			bb := int64(0)
			if premix != nil {
				for _, d := range premix.Data {
					for _, c := range d {
						for _, f := range c.Data {
							l := f.Channels
							if l > 0 {
								bb += int64(l * int(unsafe.Sizeof(f.StaticMatrix[0])))
							}
						}
					}
				}
			}
			b.SetBytes(bb)
		})
		lastTime = now
	}
}

func BenchmarkIT(b *testing.B) {
	var playback intf.Playback
	var err error
	b.Run("load_it", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			playback, _, err = format.Load(".vscode/test/beyond_the_network.it")
			if err != nil {
				b.Error(err)
			}
		}
	})

	if err := playback.SetupSampler(44100, 2, 16); err != nil {
		b.Error(err)
	}

	playback.Configure([]feature.Feature{feature.SongLoop{Count: 0}})
	playback.SetNextOrder(38)

	var step time.Duration
	for err == nil {
		b.Run("generate_it", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			var premix *gosound.PremixData
			for i := 0; i < b.N; i++ {
				premix, err = playback.Generate(step)
				if err != nil {
					if !errors.Is(err, song.ErrStopSong) {
						b.Error(err)
					}
					return
				}
				if premix != nil {
					step += time.Duration(premix.SamplesLen) / time.Duration(int(playback.GetSampleRate())*playback.GetNumChannels()*2)
					b.SetBytes(int64(premix.SamplesLen))
				}
			}
		})
	}
}
