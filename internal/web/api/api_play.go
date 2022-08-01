package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gotracker/gomixing/mixing"
	playerFeature "github.com/gotracker/gotracker/internal/feature"
	"github.com/gotracker/gotracker/internal/playlist"
	"github.com/gotracker/gotracker/internal/web/api/content"
	"github.com/gotracker/gotracker/internal/web/api/files"
	"github.com/gotracker/playback"
	"github.com/gotracker/playback/format"
	itFeature "github.com/gotracker/playback/format/it/feature"
	"github.com/gotracker/playback/index"
	playbackOutput "github.com/gotracker/playback/output"
	"github.com/gotracker/playback/player/feature"
	"github.com/gotracker/playback/player/render"
	"github.com/gotracker/playback/song"
	"github.com/heucuva/optional"
)

type apiPlayRowStatus struct {
	File    string        `json:"file"`
	Time    time.Duration `json:"time"`
	Order   int           `json:"order"`
	Row     int           `json:"row"`
	Tick    int           `json:"tick"`
	RowText string        `json:"rowText"`
}

type apiPlaySongStart struct {
	File        string `json:"file"`
	Name        string `json:"name"`
	LoopEnabled bool   `json:"loopEnabled"`
}

type apiPlaySongStop struct {
	File string `json:"file"`
}

func PlayHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	name, ok := vars["file"]
	if !ok || name == "" {
		log.Println("/api/play: file was not provided in vars")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s := playlist.Song{
		Filepath: name,
	}

	var req playRequest
	if val, ok := vars["samplerate"]; ok {
		if v, err := strconv.Atoi(val); err == nil {
			req.SamplesPerSecond.Set(v)
		}
	}

	if val, ok := vars["channels"]; ok {
		if v, err := strconv.Atoi(val); err == nil {
			req.Channels.Set(v)
		}
	}

	if val, ok := vars["format"]; ok && val != "" {
		req.SampleFormat.Set(val)
	}

	if val := r.URL.Query().Get("format"); val != "" {
		req.ContainerFormat.Set(val)
	}

	if err := playSong(s, req, w, r); err != nil {
		log.Printf("/api/play: playSong failed with %v\n", err)
		if errors.Is(err, os.ErrNotExist) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
}

type playRequest struct {
	Channels             optional.Value[int]           `yaml:"channels" json:"channels"`
	SamplesPerSecond     optional.Value[int]           `yaml:"sampleRate" json:"sampleRate"`
	SampleFormat         optional.Value[string]        `yaml:"sampleFormat" json:"sampleFormat"`
	ContainerFormat      optional.Value[string]        `yaml:"containerFormat" json:"containerFormat"`
	Song                 playlist.Song                 `yaml:"song" json:"song"`
	LoopPlaylist         optional.Value[bool]          `yaml:"loopPlaylist" json:"loopPlaylist"`
	Randomized           optional.Value[bool]          `yaml:"randomized" json:"randomized"`
	NumPremixedBuffers   optional.Value[int]           `yaml:"numPremixedBuffers" json:"numPremixedBuffers"`
	DisableNativeSamples optional.Value[bool]          `yaml:"disableNativeSamples" json:"disableNativeSamples"`
	PlayerSleepInterval  optional.Value[time.Duration] `yaml:"playerSleepInterval" json:"playerSleepInterval"`
	ITLongChannelOutput  optional.Value[bool]          `yaml:"itLongChannelOutput" json:"itLongChannelOutput"`
	ITEnableNNA          optional.Value[bool]          `yaml:"itEnableNNA" json:"itEnableNNA"`
}

func playSong(s playlist.Song, request playRequest, w http.ResponseWriter, r *http.Request) error {
	var features []feature.Feature
	if disableNativeSamples, set := request.DisableNativeSamples.Get(); !set || !disableNativeSamples {
		features = append(features, feature.UseNativeSampleFormat(true))
	}
	if playerSleepInterval, set := request.PlayerSleepInterval.Get(); !set {
		features = append(features, playerFeature.PlayerSleepInterval{Enabled: false})
	} else {
		features = append(features, playerFeature.PlayerSleepInterval{
			Enabled:  true,
			Interval: playerSleepInterval,
		})
	}

	sampFmt, err := getSampleFormat(request.SampleFormat)
	if err != nil {
		return err
	}

	var (
		samplesPerSecond = 44100
		channels         = 2
	)

	if value, set := request.SamplesPerSecond.Get(); set && value > 0 {
		samplesPerSecond = value
	}

	if value, set := request.Channels.Get(); set && value > 0 {
		channels = value
	}

	premixBuffers := 64
	if value, set := request.NumPremixedBuffers.Get(); set {
		premixBuffers = value
	}
	outBufs := make(chan *playbackOutput.PremixData, premixBuffers)

	playCtx, playCancel := context.WithCancel(r.Context())
	defer playCancel()

	var (
		wg              sync.WaitGroup
		songTime        time.Duration
		sampleTimeCoeff = time.Second / time.Duration(samplesPerSecond)
		container       content.Type
	)

	containerFormat, _ := request.ContainerFormat.Get()
	switch containerFormat {
	case "streaming-wave":
		container = &content.AudioWav{
			Channels:   channels,
			SampleRate: samplesPerSecond,
			Format:     sampFmt,
		}
	default:
		container = &content.AudioPCM{
			Channels:   channels,
			SampleRate: samplesPerSecond,
			Format:     sampFmt,
		}
	}

	wg.Add(1)
	go func() {
		defer func() {
			playCancel()
			log.Printf("/api/play/%s: stopped mixing.\n", s.Filepath)
			wg.Done()
		}()

		m := mixing.Mixer{
			Channels: channels,
		}

		panmixer := mixing.GetPanMixer(channels)

		log.Printf("/api/play/%s: started mixing...\n", s.Filepath)

		for {
			select {
			case <-playCtx.Done():
				return
			case buf, ok := <-outBufs:
				if !ok {
					return
				}
				if buf == nil {
					continue
				}

				data := m.Flatten(panmixer, buf.SamplesLen, buf.Data, buf.MixerVolume, sampFmt)

				if _, err := container.Write(w, data); err != nil {
					return
				}

				songTime += time.Duration(buf.SamplesLen) * sampleTimeCoeff

				if row, ok := buf.Userdata.(*render.RowRender); ok && row != nil {
					if row.RowText != nil {
						playbackStatusGetBroker(s.Filepath).Notify("rowText", apiPlayRowStatus{
							File:    s.Filepath,
							Time:    songTime,
							Order:   row.Order,
							Row:     row.Row,
							Tick:    row.Tick,
							RowText: row.RowText.String(),
						})
					}
				}
			}
		}
	}()

	features = append(features, feature.IgnoreUnknownEffect{Enabled: true})

	log.Printf("/api/play/%s: started playback...\n", s.Filepath)
	defer log.Printf("/api/play/%s: stopped playback.\n", s.Filepath)

	container.WriteHeader(w)
	err = renderSong(s, samplesPerSecond, channels, features, request, func(pb playback.Playback, tickInterval time.Duration) error {
		if err := playbackStatusGetBroker(s.Filepath).Notify("songStart", apiPlaySongStart{
			File:        s.Filepath,
			Name:        pb.GetName(),
			LoopEnabled: pb.CanOrderLoop(),
		}); err != nil {
			return err
		}

		var tick func() (time.Duration, bool)

		if tickInterval == 0 {
			tick = func() (time.Duration, bool) {
				return 0, true
			}
		} else {
			ticker := time.NewTicker(tickInterval)
			defer ticker.Stop()

			tick = func() (time.Duration, bool) {
				start := time.Now()
				now, ok := <-ticker.C
				if !ok {
					return 0, false
				}

				return now.Sub(start), true
			}
		}

		imedBuf := make(chan *playbackOutput.PremixData, 1)
		defer close(imedBuf)

		readyChan := make(chan struct{}, 1)
		defer close(readyChan)
		readyChan <- struct{}{}

		for {
			select {
			case <-playCtx.Done():
				if err := playCtx.Err(); err != nil && !errors.Is(err, context.Canceled) {
					return err
				}
				return nil

			case <-readyChan:
				interval, ok := tick()
				if !ok {
					playCancel()
					return nil
				}

				err := pb.Update(interval, imedBuf)
				if errors.Is(err, song.ErrStopSong) {
					playCancel()
					return nil
				}

			case buf := <-imedBuf:
				select {
				case <-playCtx.Done():
					if err := playCtx.Err(); err != nil && !errors.Is(err, context.Canceled) {
						return err
					}
					return nil

				case outBufs <- buf:
					readyChan <- struct{}{}
				}
			}
		}
	})

	close(outBufs)

	wg.Wait()

	stopErr := playbackStatusGetBroker(s.Filepath).Notify("songStopped", apiPlaySongStop{
		File: s.Filepath,
	})
	_ = stopErr

	return err
}

func renderSong(ps playlist.Song, samplesPerSecond, channels int, features []feature.Feature, request playRequest, startPlayingCB func(pb playback.Playback, tickInterval time.Duration) error) error {
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

	f, err := files.GetFS().Open(ps.Filepath)
	if err != nil {
		return err
	}

	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	s, err := format.LoadFromReader("", bytes.NewReader(data), features...)
	if err != nil {
		return fmt.Errorf("could not create song state! err[%v]", err)
	} else if s == nil {
		return nil
	}

	pb, err := s.ConstructPlayer()
	if err != nil {
		return fmt.Errorf("could not construct playback sampler! err[%v]", err)
	}

	if err := pb.SetupSampler(samplesPerSecond, channels); err != nil {
		return fmt.Errorf("could not setup playback sampler! err[%v]", err)
	}

	cfg := features

	startOrder, startOrderSet := ps.Start.Order.Get()
	startRow, startRowSet := ps.Start.Row.Get()
	if startOrderSet || startRowSet {
		txn := pb.StartPatternTransaction()
		if startOrderSet && startOrder >= 0 {
			txn.SetNextOrder(index.Order(startOrder))
		}
		if startRowSet && startRow >= 0 {
			txn.SetNextRow(index.Row(startRow))
		}
		if err := txn.Commit(); err != nil {
			return err
		}
	}

	endOrder, endOrderSet := ps.End.Order.Get()
	endRow, endRowSet := ps.End.Row.Get()
	if endOrderSet && endRowSet && endOrder >= 0 && endRow >= 0 {
		cfg = append(cfg, feature.PlayUntilOrderAndRow{
			Order: endOrder,
			Row:   endRow,
		})
	}

	if tempo, ok := ps.Tempo.Get(); ok {
		cfg = append(cfg, feature.SetDefaultTempo{Tempo: tempo})
	}

	if bpm, ok := ps.BPM.Get(); ok {
		cfg = append(cfg, feature.SetDefaultBPM{BPM: bpm})
	}

	var loopCount int
	if canPossiblyLoop {
		if l, ok := ps.Loop.Count.Get(); ok {
			loopCount = l
		}
	}
	cfg = append(cfg, feature.SongLoop{Count: loopCount})

	if itLongChannelOutput, set := request.ITLongChannelOutput.Get(); set {
		cfg = append(cfg, itFeature.LongChannelOutput{Enabled: itLongChannelOutput})
	}
	if itEnableNNA, set := request.ITEnableNNA.Get(); set {
		cfg = append(cfg, itFeature.NewNoteActions{Enabled: itEnableNNA})
	}

	if err := pb.Configure(cfg); err != nil {
		return err
	}

	if err = startPlayingCB(pb, tickInterval); err != nil {
		return err
	}

	return nil
}
