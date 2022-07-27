package play

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	progressBar "github.com/cheggaaa/pb"
	"github.com/gotracker/playback"

	playerFeature "github.com/gotracker/gotracker/internal/feature"
	"github.com/gotracker/gotracker/internal/logging"
	"github.com/gotracker/gotracker/internal/output"
	deviceCommon "github.com/gotracker/gotracker/internal/output/device/common"
	"github.com/gotracker/gotracker/internal/playlist"
	"github.com/gotracker/playback/format"
	itEffect "github.com/gotracker/playback/format/it/effect"
	itFeature "github.com/gotracker/playback/format/it/feature"
	s3mEffect "github.com/gotracker/playback/format/s3m/effect"
	xmEffect "github.com/gotracker/playback/format/xm/effect"
	"github.com/gotracker/playback/index"
	playbackOutput "github.com/gotracker/playback/output"
	"github.com/gotracker/playback/player/feature"
	"github.com/gotracker/playback/player/render"
	"github.com/gotracker/playback/song"
)

func Playlist(pl *playlist.Playlist, features []feature.Feature, settings *Settings, logger logging.Log) (bool, error) {
	var (
		play      playback.Playback
		progress  *progressBar.ProgressBar
		lastOrder int
	)

	settings.Output.OnRowOutput = func(kind deviceCommon.Kind, premix *playbackOutput.PremixData) {
		row := premix.Userdata.(*render.RowRender)
		switch kind {
		case deviceCommon.KindSoundCard:
			if row.RowText != nil {
				logger.Printf("[%0.3d:%0.3d] %s\n", row.Order, row.Row, row.RowText.String())
			}
		case deviceCommon.KindFile:
			if progress == nil {
				progress = progressBar.StartNew(play.GetNumOrders())
				lastOrder = row.Order
			}
			if lastOrder != row.Order {
				progress.Increment()
				lastOrder = row.Order
			}
		}
	}

	waveOut, outFeatures, err := output.CreateOutputDevice(settings.Output)
	if err != nil {
		return false, err
	}
	defer waveOut.Close()

	outBufs := make(chan *playbackOutput.PremixData, settings.NumPremixBuffers)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := waveOut.Play(outBufs); err != nil {
			switch {
			case errors.Is(err, song.ErrStopSong):
			case errors.Is(err, context.Canceled):

			default:
				log.Fatalln(err)
			}
		}
	}()

	features = append(features, outFeatures...)
	features = append(features, feature.IgnoreUnknownEffect{Enabled: !settings.PanicOnUnhandledEffect})

	if settings.Tracing {
		features = append(features, feature.EnableTracing{
			Filename: settings.TracingFile,
		})
	}

	logger.Printf("Output device: %s\n", waveOut.Name())

	playedAtLeastOne, err := renderSongs(pl, outBufs, features, settings, func(pb playback.Playback, tickInterval time.Duration) error {
		play = pb
		defer func() {
			if progress != nil {
				progress.Set64(progress.Total)
				progress.Finish()
			}
		}()

		var effectMap map[string]int
		if settings.GatherEffectCoverage {
			effectMap = make(map[string]int)
			play.SetOnEffect(func(e playback.Effect) {
				var name string
				switch t := e.(type) {
				case *s3mEffect.UnhandledCommand:
					name = fmt.Sprintf("UnhandledCommand(%s)", t.String())
					effectMap[name]++

				case *xmEffect.VolEff:
					for _, eff := range t.Effects {
						typ := reflect.TypeOf(eff)
						name = typ.Name()
						effectMap[name]++
					}
				case *xmEffect.UnhandledCommand:
					name = fmt.Sprintf("UnhandledCommand(%c)", t.Command.ToRune())
					effectMap[name]++
				case *xmEffect.UnhandledVolCommand:
					name = fmt.Sprintf("UnhandledVolCommand(%s)", t.String())
					effectMap[name]++

				case *itEffect.VolEff:
					for _, eff := range t.Effects {
						typ := reflect.TypeOf(eff)
						name = typ.Name()
						effectMap[name]++
					}
				case *itEffect.UnhandledCommand:
					name = fmt.Sprintf("UnhandledCommand(%c)", t.Command.ToRune())
					effectMap[name]++
				case *itEffect.UnhandledVolCommand:
					name = fmt.Sprintf("UnhandledVolCommand(%s)", t.String())
					effectMap[name]++

				default:
					typ := reflect.TypeOf(t)
					name = typ.Name()
					effectMap[name]++
				}
			})
		}

		logger.Printf("Order Looping Enabled: %v\n", play.CanOrderLoop())
		logger.Printf("Song: %s\n", play.GetName())

		p, err := NewPlayer(context.TODO(), outBufs, tickInterval)
		if err != nil {
			return err
		}

		if err := p.Play(play); err != nil {
			return err
		}

		if err := p.WaitUntilDone(); err != nil {
			switch {
			case errors.Is(err, song.ErrStopSong):
			case errors.Is(err, context.Canceled):

			default:
				return err
			}
		}

		return nil
	})
	if !playedAtLeastOne || err != nil {
		return playedAtLeastOne, err
	}

	wg.Wait()

	logger.Println()
	logger.Println("done!")

	return true, nil
}

func getFeatureByType[T feature.Feature](features []feature.Feature) (T, bool) {
	var empty T
	if len(features) == 0 {
		return empty, false
	}

	tt := reflect.TypeOf(empty)
	for _, f := range features {
		v := reflect.ValueOf(f)
		if v.CanConvert(tt) {
			return v.Convert(tt).Interface().(T), true
		}
	}

	return empty, false
}

func renderSongs(pl *playlist.Playlist, outBufs chan<- *playbackOutput.PremixData, features []feature.Feature, settings *Settings, startPlayingCB func(pb playback.Playback, tickInterval time.Duration) error) (bool, error) {
	defer close(outBufs)

	tickInterval := time.Duration(5) * time.Millisecond
	if setting, ok := getFeatureByType[playerFeature.PlayerSleepInterval](features); ok {
		if setting.Enabled {
			tickInterval = setting.Interval
		} else {
			tickInterval = 0
		}
	}

	canPossiblyLoop := true
	if setting, ok := getFeatureByType[feature.SongLoop](features); ok {
		canPossiblyLoop = (setting.Count != 0)
	}

	var playedAtLeastOne bool
playlistLoop:
	for _, songIdx := range pl.GetPlaylist() {
		song := pl.GetSong(songIdx)
		if song == nil {
			continue
		}
		s, err := format.Load(song.Filepath, features...)
		if err != nil {
			return playedAtLeastOne, fmt.Errorf("could not create song state! err[%v]", err)
		} else if s == nil {
			return playedAtLeastOne, fmt.Errorf("unexpectedly empty song state! file[%s]", song.Filepath)
		}

		playback, err := s.ConstructPlayer()
		if err != nil {
			return playedAtLeastOne, fmt.Errorf("could not construct playback sampler! err[%v]", err)
		}

		if err := playback.SetupSampler(settings.Output.SamplesPerSecond, settings.Output.Channels); err != nil {
			return playedAtLeastOne, fmt.Errorf("could not setup playback sampler! err[%v]", err)
		}

		cfg := features

		startOrder, startOrderSet := song.Start.Order.Get()
		startRow, startRowSet := song.Start.Row.Get()
		if startOrderSet || startRowSet {
			txn := playback.StartPatternTransaction()
			if startOrderSet && startOrder >= 0 {
				txn.SetNextOrder(index.Order(startOrder))
			}
			if startRowSet && startRow >= 0 {
				txn.SetNextRow(index.Row(startRow))
			}
			if err := txn.Commit(); err != nil {
				return playedAtLeastOne, err
			}
		}

		endOrder, endOrderSet := song.End.Order.Get()
		endRow, endRowSet := song.End.Row.Get()
		if endOrderSet && endRowSet && endOrder >= 0 && endRow >= 0 {
			cfg = append(cfg, feature.PlayUntilOrderAndRow{
				Order: endOrder,
				Row:   endRow,
			})
		}

		if tempo, ok := song.Tempo.Get(); ok {
			cfg = append(cfg, feature.SetDefaultTempo{Tempo: tempo})
		}

		if bpm, ok := song.BPM.Get(); ok {
			cfg = append(cfg, feature.SetDefaultBPM{BPM: bpm})
		}

		var loopCount int
		if canPossiblyLoop {
			if l, ok := song.Loop.Count.Get(); ok {
				loopCount = l
			}
		}
		cfg = append(cfg,
			feature.SongLoop{Count: loopCount},
			itFeature.LongChannelOutput{Enabled: settings.ITLongChannelOutput},
			itFeature.NewNoteActions{Enabled: settings.ITEnableNNA})

		if setting, ok := getFeatureByType[playerFeature.SoloChannels](features); ok && len(setting.Channels) > 0 {
			cm := make([]feature.ChannelMute, playback.GetNumChannels())
			for i := range cm {
				cm[i].Channel = i + 1
				cm[i].Muted = true
			}
			for _, solo := range setting.Channels {
				if solo > 0 && solo <= len(cm) {
					cm[solo-1].Muted = false
				}
			}
			for _, f := range cm {
				cfg = append(cfg, f)
			}
		}

		if err := playback.Configure(cfg); err != nil {
			return playedAtLeastOne, err
		}

		if err = startPlayingCB(playback, tickInterval); err != nil {
			continue
		}

		pl.MarkPlayed(song)

		playedAtLeastOne = true
	}

	if pl.IsLooping() {
		goto playlistLoop
	}

	return playedAtLeastOne, nil
}
