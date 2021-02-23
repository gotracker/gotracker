package playback_test

import (
	"flag"
	"math"
	"os"
	"testing"
	"time"

	"gotracker/internal/format/xm"
	"gotracker/internal/player/feature"
	"gotracker/internal/player/intf"
)

var (
	enableTremor       bool
	enablePortaLinkMem bool
)

func TestTremor(t *testing.T) {
	if !enableTremor {
		t.Skip()
	}

	fn := "../../../../test/Tremor.xm"

	sampleRate := 44100
	channels := 2
	bitsPerSample := 16

	performChannelComparison(t, fn, sampleRate, channels, bitsPerSample)
}

func TestPortaLinkMem(t *testing.T) {
	if !enablePortaLinkMem {
		t.Skip()
	}

	fn := "../../../../test/Porta-LinkMem.xm"

	sampleRate := 44100
	channels := 2
	bitsPerSample := 16

	performChannelComparison(t, fn, sampleRate, channels, bitsPerSample)
}

func performChannelComparison(t *testing.T, fn string, sampleRate int, channels int, bitsPerSample int) {
	t.Helper()

	playback, err := xm.XM.Load(fn)
	if err != nil {
		t.Fatalf("Could not create song state! err[%v]", err)
	}

	if err := playback.SetupSampler(sampleRate, channels, bitsPerSample); err != nil {
		t.Fatalf("Could not setup playback sampler! err[%v]", err)
	}

	playback.Configure([]feature.Feature{feature.SongLoop{Enabled: false}})

	for {
		premixData, err := playback.Generate(time.Duration(0))
		if err != nil {
			if err == intf.ErrStopSong {
				break
			}
			t.Fatal(err)
		}

		if len(premixData.Data) == 0 {
			continue
		}

		if len(premixData.Data) < 2 {
			t.Fatal("Not enough tracks of data in premix buffer")
		} else if len(premixData.Data) > 2 {
			t.Fatal("Too many tracks of data in premix buffer")
		}

		test := premixData.Data[0]
		control := premixData.Data[1]

		if len(test) < 1 {
			t.Fatal("Not enough blocks of premixed track data in premix buffer")
		} else if len(test) > 1 {
			t.Fatal("Too many blocks of premixed track data in premix buffer")
		}

		tc := test[0]
		cc := control[0]

		if tc.Data == nil && cc.Data == nil {
			continue
		} else if tc.Data == nil {
			t.Fatal("Not enough channel data provided in test track premix buffer")
		} else if cc.Data == nil {
			t.Fatal("Not enough channel data provided in test track premix buffer")
		}

		if len(tc.Data) < channels {
			t.Fatal("Not enough output channels of premixed track data in test track premix buffer")
		} else if len(tc.Data) > channels {
			t.Fatal("Too many output channels of premixed track data in test track premix buffer")
		}

		if len(cc.Data) < channels {
			t.Fatal("Not enough output channels of premixed track data in control track premix buffer")
		} else if len(cc.Data) > channels {
			t.Fatal("Too many output channels of premixed track data in control track premix buffer")
		}

		for c := 0; c < channels; c++ {
			td := tc.Data[c]
			cd := cc.Data[c]

			if len(td) != len(cd) {
				t.Fatal("test track premix buffer length is not the same as for the control track")
			}

			for i, ts := range td {
				cs := cd[i]
				if math.Abs(float64(ts-cs)) >= 0.15 {
					t.Fatal("test track premix buffer data is not the same as for the control track")
				}
			}
		}
	}
}

func TestMain(m *testing.M) {
	flag.BoolVar(&enableTremor, "Tremor", false, "Enable Tremor test")
	flag.BoolVar(&enablePortaLinkMem, "PortaLinkMem", false, "Enable PortaLinkMem test")
	flag.Parse()
	os.Exit(m.Run())
}
