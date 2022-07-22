package main_test

import (
	"errors"
	"testing"
	"time"
	"unsafe"

	"github.com/gotracker/playback"
	"github.com/gotracker/playback/format"
	"github.com/gotracker/playback/output"
	"github.com/gotracker/playback/player/feature"
	"github.com/gotracker/playback/song"
)

func BenchmarkPlayerS3M(b *testing.B) {
	var pb playback.Playback
	var err error
	b.Run("load_s3m", func(b *testing.B) {
		pb, _, err = format.Load("test/celestial_fantasia.s3m")
		if err != nil {
			b.Error(err)
		}
	})

	if err := pb.SetupSampler(44100, 2); err != nil {
		b.Error(err)
	}

	if err := pb.Configure([]feature.Feature{feature.SongLoop{Count: 0}}); err != nil {
		b.Error(err)
	}

	lastTime := time.Now()
	for err == nil {
		now := time.Now()
		b.Run("generate_s3m", func(b *testing.B) {
			b.Helper()
			b.ReportAllocs()
			var premix *output.PremixData
			premix, err = pb.Generate(now.Sub(lastTime))
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
	var pb playback.Playback
	var err error
	b.Run("load_it", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pb, _, err = format.Load(".vscode/test/beyond_the_network.it")
			if err != nil {
				b.Error(err)
			}
		}
	})

	if err := pb.SetupSampler(44100, 2); err != nil {
		b.Error(err)
	}

	if err := pb.Configure([]feature.Feature{feature.SongLoop{Count: 0}}); err != nil {
		b.Error(err)
	}
	pb.SetNextOrder(38)

	var step time.Duration
	for err == nil {
		b.Run("generate_it", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			var premix *output.PremixData
			for i := 0; i < b.N; i++ {
				premix, err = pb.Generate(step)
				if err != nil {
					if !errors.Is(err, song.ErrStopSong) {
						b.Error(err)
					}
					return
				}
				if premix != nil {
					step += time.Duration(premix.SamplesLen) / time.Duration(int(pb.GetSampleRate())*pb.GetNumChannels()*2)
					b.SetBytes(int64(premix.SamplesLen))
				}
			}
		})
	}
}
