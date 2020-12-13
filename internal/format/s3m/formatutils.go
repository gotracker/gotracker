package s3m

import (
	"bytes"
	"gotracker/internal/format/s3m/channel"
	"gotracker/internal/format/s3m/effect"
	"gotracker/internal/format/s3m/util"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/state"
	"os"
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

	ss := s.(*state.Song)

	ss.EffectFactory = effect.Factory
	ss.CalcSemitonePeriod = util.CalcSemitonePeriod
	ss.Pattern.Patterns = s3mSong.GetPatternsInterface()
	ss.Pattern.Orders = s3mSong.Head.OrderList
	ss.Pattern.Row.Ticks = int(s3mSong.Head.Info.InitialSpeed)
	ss.Pattern.Row.Tempo = int(s3mSong.Head.Info.InitialTempo)

	ss.GlobalVolume = util.VolumeFromS3M(s3mSong.Head.Info.GlobalVolume)
	ss.SongData = s3mSong

	for i, cs := range s3mSong.Head.ChannelSettings {
		if cs.IsEnabled() {
			ss.NumChannels = i + 1
		}
	}

	for i := 0; i < ss.NumChannels; i++ {
		cs := &ss.Channels[i]
		cs.Instrument = nil
		cs.Pos = 0
		cs.Period = 0
		cs.SetStoredVolume(64, ss)
		cs.Memory = &channel.Memory{}
		ch := s3mSong.Head.ChannelSettings[i]
		if ch.IsEnabled() {
			pf := s3mSong.Head.Panning[i]
			if pf.IsValid() {
				cs.Pan = util.PanningFromS3M(pf.Value())
			} else {
				l := ch.GetChannel()
				switch l {
				case ChannelIDL1, ChannelIDL2, ChannelIDL3, ChannelIDL4, ChannelIDL5, ChannelIDL6, ChannelIDL7, ChannelIDL8:
					cs.Pan = util.PanningFromS3M(0x03)
				case ChannelIDR1, ChannelIDR2, ChannelIDR3, ChannelIDR4, ChannelIDR5, ChannelIDR6, ChannelIDR7, ChannelIDR8:
					cs.Pan = util.PanningFromS3M(0x0C)
				}
			}
		} else {
			cs.Pan = util.PanningFromS3M(0x08) // center?
		}
		cs.Command = nil

		cs.DisplayNote = note.EmptyNote
		cs.DisplayInst = 0

		cs.TargetPeriod = cs.Period
		cs.TargetPos = cs.Pos
		cs.TargetInst = cs.Instrument
		cs.PortaTargetPeriod = cs.TargetPeriod
		cs.NotePlayTick = 0
		cs.RetriggerCount = 0
		cs.TremorOn = true
		cs.TremorTime = 0
		cs.VibratoDelta = 0
		cs.Cmd = nil
	}

	return nil
}
