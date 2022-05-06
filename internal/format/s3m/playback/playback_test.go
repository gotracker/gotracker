package playback_test

import (
	"flag"
	"math"
	"os"
	"testing"
	"time"

	"github.com/gotracker/gotracker/internal/format/s3m"
	"github.com/gotracker/gotracker/internal/format/settings"
	"github.com/gotracker/gotracker/internal/player/feature"
	"github.com/gotracker/gotracker/internal/song"
)

var (
	enableOxxMemory          bool
	enablePeriodLimit        bool
	enablePortaAfterArp      bool
	enableRetrigAfterNoteCut bool
	enableVibratoTypeChange  bool
)

func TestOxxMemory(t *testing.T) {
	if !enableOxxMemory {
		t.Skip()
	}

	fn := "../../../../test/OxxMemory.s3m"

	sampleRate := 44100
	channels := 2
	bitsPerSample := 16

	performSilentChannelsTest(t, fn, sampleRate, channels, bitsPerSample)
}

func TestPeriodLimit(t *testing.T) {
	if !enablePeriodLimit {
		t.Skip()
	}

	fn := "../../../../test/PeriodLimit.s3m"

	sampleRate := 44100
	channels := 2
	bitsPerSample := 16

	performChannelComparison(t, fn, sampleRate, channels, bitsPerSample)
}

func TestPortaAfterArp(t *testing.T) {
	if !enablePortaAfterArp {
		t.Skip()
	}

	fn := "../../../../test/PortaAfterArp.s3m"

	sampleRate := 44100
	channels := 2
	bitsPerSample := 16

	performChannelComparison(t, fn, sampleRate, channels, bitsPerSample)
}

func TestRetrigAfterNoteCut(t *testing.T) {
	if !enableRetrigAfterNoteCut {
		t.Skip()
	}

	fn := "../../../../test/RetrigAfterNoteCut.s3m"

	sampleRate := 44100
	channels := 2
	bitsPerSample := 16

	performChannelComparison(t, fn, sampleRate, channels, bitsPerSample)
}

func TestVibratoTypeChange(t *testing.T) {
	if !enableVibratoTypeChange {
		t.Skip()
	}

	fn := "../../../../test/VibratoTypeChange.s3m"

	sampleRate := 44100
	channels := 2
	bitsPerSample := 16

	performSilentChannelsTest(t, fn, sampleRate, channels, bitsPerSample)
}

func performSilentChannelsTest(t *testing.T, fn string, sampleRate int, channels int, bitsPerSample int) {
	t.Helper()

	s := &settings.Settings{}
	playback, err := s3m.S3M.Load(fn, s)
	if err != nil {
		t.Fatalf("Could not create song state! err[%v]", err)
	}

	if err := playback.SetupSampler(sampleRate, channels, bitsPerSample); err != nil {
		t.Fatalf("Could not setup playback sampler! err[%v]", err)
	}

	playback.Configure([]feature.Feature{feature.SongLoop{Count: 0}})

	for {
		premixData, err := playback.Generate(time.Duration(0))
		if err != nil {
			if err == song.ErrStopSong {
				break
			}
			t.Fatal(err)
		}

		if len(premixData.Data) == 0 {
			continue
		}

		if len(premixData.Data) < 1 {
			t.Fatal("Not enough channels of data in premix buffer")
		}

		for _, test := range premixData.Data {

			if len(test) < 1 {
				t.Fatal("Not enough blocks of premixed track data in premix buffer")
			} else if len(test) > 1 {
				t.Fatal("Too many blocks of premixed track data in premix buffer")
			}

			pm := test[0]

			data := pm.Data
			if data == nil {
				continue
			}

			if len(data) < channels {
				t.Fatal("Not enough output channels of premixed track data in premix buffer")
			} else if len(data) > channels {
				t.Fatal("Too many output channels of premixed track data in premix buffer")
			}

			for _, chdata := range data {
				for i := 0; i < chdata.Channels; i++ {
					s := chdata.StaticMatrix[i]
					if math.Abs(float64(s)) >= 0.5 {
						t.Fatal("expected relative silence, got waveform")
					}
				}
			}
		}
	}
}

func performChannelComparison(t *testing.T, fn string, sampleRate int, channels int, bitsPerSample int) {
	t.Helper()

	s := &settings.Settings{}
	playback, err := s3m.S3M.Load(fn, s)
	if err != nil {
		t.Fatalf("Could not create song state! err[%v]", err)
	}

	if err := playback.SetupSampler(sampleRate, channels, bitsPerSample); err != nil {
		t.Fatalf("Could not setup playback sampler! err[%v]", err)
	}

	playback.Configure([]feature.Feature{feature.SongLoop{Count: 0}})

	for {
		premixData, err := playback.Generate(time.Duration(0))
		if err != nil {
			if err == song.ErrStopSong {
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

			if td.Channels != cd.Channels {
				t.Fatal("test track premix buffer length is not the same as for the control track")
			}

			for i := 0; i < td.Channels; i++ {
				ts := td.StaticMatrix[i]
				cs := cd.StaticMatrix[i]
				if math.Abs(float64(ts-cs)) >= 0.15 {
					t.Fatal("test track premix buffer data is not the same as for the control track")
				}
			}
		}
	}
}

func TestMain(m *testing.M) {
	flag.BoolVar(&enableOxxMemory, "OxxMemory", false, "Enable OxxMemory test")
	flag.BoolVar(&enablePeriodLimit, "PeriodLimit", false, "Enable PeriodLimit test")
	flag.BoolVar(&enablePortaAfterArp, "PortaAfterArp", false, "Enable PortaAfterArp test")
	flag.BoolVar(&enableRetrigAfterNoteCut, "RetrigAfterNoteCut", false, "Enable RetrigAfterNoteCut test")
	flag.BoolVar(&enableVibratoTypeChange, "VibratoTypeChange", false, "Enable VibratoTypeChange test")
	flag.Parse()
	os.Exit(m.Run())
}
