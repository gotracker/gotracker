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
	device "github.com/gotracker/gosound"

	"gotracker/internal/command/internal/logging"
	"gotracker/internal/command/internal/playlist"
	"gotracker/internal/format"
	itEffect "gotracker/internal/format/it/playback/effect"
	s3mEffect "gotracker/internal/format/s3m/playback/effect"
	"gotracker/internal/format/settings"
	xmEffect "gotracker/internal/format/xm/playback/effect"
	"gotracker/internal/output"
	"gotracker/internal/player"
	"gotracker/internal/player/feature"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/render"
	"gotracker/internal/song"
	"gotracker/internal/song/index"
)

func Playlist(pl *playlist.Playlist, options []settings.OptionFunc, settings *Settings, logger logging.Log) (bool, error) {
	var (
		playback  intf.Playback
		progress  *progressBar.ProgressBar
		lastOrder int
	)

	settings.Output.OnRowOutput = func(deviceKind device.Kind, premix *device.PremixData) {
		row := premix.Userdata.(*render.RowRender)
		switch deviceKind {
		case device.KindSoundCard:
			if row.RowText != nil {
				logger.Printf("[%0.3d:%0.3d] %s\n", row.Order, row.Row, row.RowText.String())
			}
		case device.KindFile:
			if progress == nil {
				progress = progressBar.StartNew(playback.GetNumOrders())
				lastOrder = row.Order
			}
			if lastOrder != row.Order {
				progress.Increment()
				lastOrder = row.Order
			}
		}
	}

	waveOut, configuration, err := output.CreateOutputDevice(settings.Output)
	if err != nil {
		return false, err
	}
	defer waveOut.Close()

	outBufs := make(chan *device.PremixData, settings.NumPremixBuffers)

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

	configuration = append(configuration, feature.IgnoreUnknownEffect{Enabled: !settings.PanicOnUnhandledEffect})

	logger.Printf("Output device: %s\n", waveOut.Name())

	playedAtLeastOne, err := renderSongs(pl, outBufs, options, configuration, settings, func(pb intf.Playback, tickInterval time.Duration) error {
		playback = pb
		defer func() {
			if progress != nil {
				progress.Set64(progress.Total)
				progress.Finish()
			}
		}()

		var effectMap map[string]int
		if settings.GatherEffectCoverage {
			effectMap = make(map[string]int)
			playback.SetOnEffect(func(e intf.Effect) {
				var name string
				switch t := e.(type) {
				case *xmEffect.VolEff:
					for _, eff := range t.Effects {
						typ := reflect.TypeOf(eff)
						name = typ.Name()
						effectMap[name]++
					}
				case *itEffect.VolEff:
					for _, eff := range t.Effects {
						typ := reflect.TypeOf(eff)
						name = typ.Name()
						effectMap[name]++
					}
				case *s3mEffect.UnhandledCommand:
					name = fmt.Sprintf("UnhandledCommand(%c)", t.Command+'@')
					effectMap[name]++
				default:
					typ := reflect.TypeOf(t)
					name = typ.Name()
					effectMap[name]++
				}
			})
		}

		logger.Printf("Order Looping Enabled: %v\n", playback.CanOrderLoop())
		logger.Printf("Song: %s\n", playback.GetName())

		p, err := player.NewPlayer(context.TODO(), outBufs, tickInterval)
		if err != nil {
			return err
		}

		if err := p.Play(playback); err != nil {
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

func findFeatureByName(configuration []feature.Feature, name string) (feature.Feature, bool) {
	for _, feature := range configuration {
		tf := reflect.TypeOf(feature)
		if tf.Name() == name {
			return feature, true
		}
	}
	return nil, false
}

func renderSongs(pl *playlist.Playlist, outBufs chan<- *device.PremixData, options []settings.OptionFunc, configuration []feature.Feature, settings *Settings, startPlayingCB func(pb intf.Playback, tickInterval time.Duration) error) (bool, error) {
	defer close(outBufs)

	tickInterval := time.Duration(5) * time.Millisecond
	if feat, found := findFeatureByName(configuration, "PlayerSleepInterval"); found {
		if f, ok := feat.(feature.PlayerSleepInterval); ok {
			if f.Enabled {
				tickInterval = f.Interval
			} else {
				tickInterval = time.Duration(0)
			}
		}
	}

	canPossiblyLoop := true
	if feat, found := findFeatureByName(configuration, "SongLoop"); found {
		if f, ok := feat.(feature.SongLoop); ok {
			canPossiblyLoop = f.Enabled
		}
	}

	var playedAtLeastOne bool
playlistLoop:
	for _, songIdx := range pl.GetPlaylist() {
		song := pl.GetSong(songIdx)
		if song == nil {
			continue
		}
		playback, songFmt, err := format.Load(song.Filepath, options...)
		if err != nil {
			return playedAtLeastOne, fmt.Errorf("Could not create song state! err[%v]", err)
		} else if songFmt != nil {
			if err := playback.SetupSampler(settings.Output.SamplesPerSecond, settings.Output.Channels, settings.Output.BitsPerSample); err != nil {
				return playedAtLeastOne, fmt.Errorf("Could not setup playback sampler! err[%v]", err)
			}
		}
		startOrder, startOrderSet := song.Start.Order.GetInt()
		startRow, startRowSet := song.Start.Row.GetInt()
		if startOrderSet || startRowSet {
			txn := playback.StartPatternTransaction()
			defer txn.Cancel()
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

		cfg := append([]feature.Feature{}, configuration...)
		endOrder, endOrderSet := song.End.Order.GetInt()
		endRow, endRowSet := song.End.Row.GetInt()
		if endOrderSet && endRowSet && endOrder >= 0 && endRow >= 0 {
			cfg = append(cfg, feature.PlayUntilOrderAndRow{
				Order: endOrder,
				Row:   endRow,
			})
		}
		var looping bool
		if canPossiblyLoop {
			if l, ok := song.Loop.GetBool(); ok {
				looping = l
			}
		}
		cfg = append(cfg,
			feature.SongLoop{Enabled: looping},
			feature.ITLongChannelOutput{Enabled: settings.ITLongChannelOutput})

		playback.Configure(cfg)

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
