package main_test

import (
	"errors"
	"testing"
	"time"
	"unsafe"

	"gotracker/internal/format"
	"gotracker/internal/player/feature"
	"gotracker/internal/player/intf"

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

	playback.DisableFeatures([]feature.Feature{feature.OrderLoop})

	lastTime := time.Now()
	for err == nil {
		now := time.Now()
		b.Run("generate_s3m", func(b *testing.B) {
			b.Helper()
			b.ReportAllocs()
			var premix *gosound.PremixData
			premix, err = playback.Generate(now.Sub(lastTime))
			if err != nil {
				if !errors.Is(err, intf.ErrStopSong) {
					b.Error(err)
				}
				return
			}
			bb := int64(0)
			if premix != nil {
				for _, d := range premix.Data {
					for _, c := range d {
						for _, f := range c.Data {
							l := len(f)
							if l > 0 {
								bb += int64(l * int(unsafe.Sizeof(f[0])))
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
