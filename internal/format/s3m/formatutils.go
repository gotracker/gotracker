package s3m

import (
	"bytes"
	"os"

	"gotracker/internal/format/s3m/channel"
	"gotracker/internal/format/s3m/effect"
	"gotracker/internal/format/s3m/util"
	"gotracker/internal/player/intf"
)

func readFile(filename string) (*bytes.Buffer, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(file)
	return buffer, nil
}

func getString(bytearray []byte) string {
	n := bytes.Index(bytearray, []byte{0})
	if n == -1 {
		n = len(bytearray)
	}
	s := string(bytearray[:n])
	return s
}

func load(s intf.Song, filename string, reader readerFunc) error {
	s3mSong, err := reader(filename)
	if err != nil {
		return err
	}

	s.SetEffectFactory(effect.Factory)
	s.SetCalcSemitonePeriod(util.CalcSemitonePeriod)
	s.SetPatterns(s3mSong.Patterns)
	s.SetOrderList(s3mSong.Head.OrderList)
	s.SetTicks(int(s3mSong.Head.Info.InitialSpeed))
	s.SetTempo(int(s3mSong.Head.Info.InitialTempo))

	s.SetGlobalVolume(util.VolumeFromS3M(s3mSong.Head.Info.GlobalVolume))
	s.SetSongData(s3mSong)

	numCh := 0
	for i, cs := range s3mSong.Head.ChannelSettings {
		if cs.IsEnabled() {
			numCh = i + 1
		}
	}
	s.SetNumChannels(numCh)

	for i := 0; i < numCh; i++ {
		cs := s.GetChannel(i)
		cs.SetStoredVolume(util.VolumeFromS3M(64), s)
		cs.SetMemory(&channel.Memory{})
		ch := s3mSong.Head.ChannelSettings[i]
		if ch.IsEnabled() {
			pf := s3mSong.Head.Panning[i]
			if pf.IsValid() {
				cs.SetPan(util.PanningFromS3M(pf.Value()))
			} else {
				l := ch.GetChannel()
				switch l {
				case ChannelIDL1, ChannelIDL2, ChannelIDL3, ChannelIDL4, ChannelIDL5, ChannelIDL6, ChannelIDL7, ChannelIDL8:
					cs.SetPan(util.PanningFromS3M(0x03))
				case ChannelIDR1, ChannelIDR2, ChannelIDR3, ChannelIDR4, ChannelIDR5, ChannelIDR6, ChannelIDR7, ChannelIDR8:
					cs.SetPan(util.PanningFromS3M(0x0C))
				}
			}
		} else {
			cs.SetPan(util.PanningFromS3M(0x08)) // center?
		}
	}

	return nil
}
