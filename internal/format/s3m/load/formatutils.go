package load

import (
	"gotracker/internal/format/s3m/layout"
	"gotracker/internal/format/s3m/playback"
)

type readerFunc func(filename string) (*layout.Song, error)

func load(filename string, reader readerFunc) (*playback.Manager, error) {
	s3mSong, err := reader(filename)
	if err != nil {
		return nil, err
	}

	m := playback.NewManager(s3mSong)

	m.SetPatterns(s3mSong.Patterns)
	m.SetOrderList(s3mSong.OrderList)
	m.SetTicks(s3mSong.Head.InitialSpeed)
	m.SetTempo(s3mSong.Head.InitialTempo)

	m.SetGlobalVolume(s3mSong.Head.GlobalVolume)

	m.SetNumChannels(len(s3mSong.ChannelSettings))
	for i, ch := range s3mSong.ChannelSettings {
		cs := m.GetChannel(i)
		cs.SetStoredVolume(ch.InitialVolume, s3mSong.Head.GlobalVolume)
		cs.SetPan(ch.InitialPanning)
		cs.SetMemory(&s3mSong.ChannelSettings[i].Memory)
	}

	return m, nil
}
