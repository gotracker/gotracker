package s3m

import (
	"bytes"
	"os"

	"gotracker/internal/format/s3m/effect"
	"gotracker/internal/format/s3m/util"
	"gotracker/internal/player/intf"
)

func readFile(filename string) (*bytes.Buffer, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	buffer := &bytes.Buffer{}
	buffer.ReadFrom(file)
	return buffer, nil
}

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
		cs.SetMemory(&ch.Memory)
	}

	return nil
}
