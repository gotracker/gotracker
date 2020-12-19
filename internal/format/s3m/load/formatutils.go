package load

import (
	"gotracker/internal/format/s3m/layout"
	"gotracker/internal/format/s3m/playback/effect"
	"gotracker/internal/format/s3m/playback/util"
	"gotracker/internal/player/intf"
)

type readerFunc func(filename string) (*layout.Song, error)

func load(s intf.Song, filename string, reader readerFunc) error {
	s3mSong, err := reader(filename)
	if err != nil {
		return err
	}

	s.SetEffectFactory(effect.Factory)
	s.SetCalcSemitonePeriod(util.CalcSemitonePeriod)
	s.SetPatterns(s3mSong.Patterns)
	s.SetOrderList(s3mSong.OrderList)
	s.SetTicks(s3mSong.Head.InitialSpeed)
	s.SetTempo(s3mSong.Head.InitialTempo)

	s.SetGlobalVolume(s3mSong.Head.GlobalVolume)
	s.SetSongData(s3mSong)

	s.SetNumChannels(len(s3mSong.ChannelSettings))
	for i, ch := range s3mSong.ChannelSettings {
		cs := s.GetChannel(i)
		cs.SetStoredVolume(ch.InitialVolume, s)
		cs.SetPan(ch.InitialPanning)
		cs.SetMemory(&s3mSong.ChannelSettings[i].Memory)
	}

	return nil
}
