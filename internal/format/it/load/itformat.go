package load

import (
	"bytes"
	"errors"
	"fmt"

	itfile "github.com/gotracker/goaudiofile/music/tracked/it"
	"github.com/gotracker/gomixing/volume"

	formatutil "gotracker/internal/format/internal/util"
	"gotracker/internal/format/it/layout"
	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/format/it/playback/util"
	"gotracker/internal/instrument"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/pattern"
)

func moduleHeaderToHeader(fh *itfile.ModuleHeader) (*layout.Header, error) {
	if fh == nil {
		return nil, errors.New("file header is nil")
	}
	head := layout.Header{
		Name:         fh.GetName(),
		InitialSpeed: int(fh.InitialSpeed),
		InitialTempo: int(fh.InitialTempo),
		GlobalVolume: volume.Volume(fh.GlobalVolume.Value()),
		MixingVolume: volume.Volume(fh.MixingVolume.Value()),
	}
	return &head, nil
}

func dumpBytes(data []byte, s int) {
	r := bytes.NewReader(data)
	i := 0
	fmt.Printf("%0.4X  ", i+s)
	for {
		b, err := r.ReadByte()
		if err != nil {
			break
		}
		fmt.Printf("%0.2X", b)
		i++
		if i&15 == 0 {
			fmt.Printf("\n%0.4X  ", i+s)
		} else {
			fmt.Printf(" ")
		}
	}
	if i%15 != 0 {
		fmt.Println()
	}
}

func convertItPattern(pkt itfile.PackedPattern, channels int) (*pattern.Pattern, int, error) {
	pat := &pattern.Pattern{
		Orig: pkt,
	}

	//dumpBytes(pkt.Data, 0)

	channelMem := make([]itfile.ChannelData, channels)
	maxCh := uint8(0)
	pos := 0
	for rowNum := 0; rowNum < int(pkt.Rows); rowNum++ {
		pat.Rows = append(pat.Rows, pattern.RowData{})
		row := &pat.Rows[rowNum]
		row.Channels = make([]intf.ChannelData, channels)
	channelLoop:
		for {
			sz, chn, err := pkt.ReadChannelData(pos, channelMem)
			if err != nil {
				return nil, 0, err
			}
			//dumpBytes(pkt.Data[pos:pos+sz], pos)
			pos += sz
			if chn == nil {
				break channelLoop
			}

			channelNum := int(chn.ChannelNumber)

			cd := channel.Data{
				What:            chn.Flags,
				Note:            chn.Note,
				Instrument:      chn.Instrument,
				VolPan:          chn.VolPan,
				Effect:          chn.Command,
				EffectParameter: chn.CommandData,
			}

			row.Channels[channelNum] = &cd
			if maxCh < uint8(channelNum) {
				maxCh = uint8(channelNum)
			}
		}
	}

	return pat, int(maxCh), nil
}

func convertItFileToSong(f *itfile.File) (*layout.Song, error) {
	h, err := moduleHeaderToHeader(&f.Head)
	if err != nil {
		return nil, err
	}

	linearFrequencySlides := f.Head.Flags.IsLinearSlides()
	oldEffectMode := f.Head.Flags.IsOldEffects()
	efgLinkMode := f.Head.Flags.IsEFGLinking()

	song := layout.Song{
		Head:              *h,
		Instruments:       make(map[uint8]*instrument.Instrument),
		InstrumentNoteMap: make(map[uint8]map[note.Semitone]layout.NoteInstrument),
		Patterns:          make([]pattern.Pattern, len(f.Patterns)),
		OrderList:         make([]intf.PatternIdx, int(f.Head.OrderCount)),
	}

	for i := 0; i < int(f.Head.OrderCount); i++ {
		song.OrderList[i] = intf.PatternIdx(f.OrderList[i])
	}

	if f.Head.Flags.IsUseInstruments() {
		for instNum, inst := range f.Instruments {
			switch ii := inst.(type) {
			case *itfile.IMPIInstrumentOld:
				instMap, err := convertITInstrumentOldToInstrument(ii, f.Samples, linearFrequencySlides)
				if err != nil {
					return nil, err
				}

				for _, ci := range instMap {
					addSampleWithNoteMapToSong(&song, ci.Inst, ci.NR, instNum)
				}

			case *itfile.IMPIInstrument:
				instMap, err := convertITInstrumentToInstrument(ii, f.Samples, linearFrequencySlides)
				if err != nil {
					return nil, err
				}

				for _, ci := range instMap {
					addSampleWithNoteMapToSong(&song, ci.Inst, ci.NR, instNum)
				}
			}
		}
	}

	lastEnabledChannel := 0
	song.Patterns = make([]pattern.Pattern, len(f.Patterns))
	for patNum, pkt := range f.Patterns {
		pattern, maxCh, err := convertItPattern(pkt, len(f.Head.ChannelVol))
		if err != nil {
			return nil, err
		}
		if pattern == nil {
			continue
		}
		if lastEnabledChannel < maxCh {
			lastEnabledChannel = maxCh
		}
		song.Patterns[patNum] = *pattern
	}

	channels := make([]layout.ChannelSetting, lastEnabledChannel+1)
	for chNum := range channels {
		cs := layout.ChannelSetting{
			OutputChannelNum: chNum,
			Enabled:          true,
			InitialVolume:    volume.Volume(1),
			ChannelVolume:    volume.Volume(f.Head.ChannelVol[chNum].Value()),
			InitialPanning:   util.PanningFromIt(f.Head.ChannelPan[chNum]),
			Memory: channel.Memory{
				LinearFreqSlides: linearFrequencySlides,
				OldEffectMode:    oldEffectMode,
				EFGLinkMode:      efgLinkMode,
			},
		}

		channels[chNum] = cs
	}

	song.ChannelSettings = channels

	return &song, nil
}

type noteRemap struct {
	Orig  note.Semitone
	Remap note.Semitone
}

func addSampleWithNoteMapToSong(song *layout.Song, sample *instrument.Instrument, sts []noteRemap, instNum int) {
	if sample == nil {
		return
	}
	id := channel.SampleID{
		InstID: uint8(instNum + 1),
	}
	sample.ID = id
	song.Instruments[id.InstID] = sample

	id, ok := sample.ID.(channel.SampleID)
	if !ok {
		return
	}
	inm, ok := song.InstrumentNoteMap[id.InstID]
	if !ok {
		inm = make(map[note.Semitone]layout.NoteInstrument)
		song.InstrumentNoteMap[id.InstID] = inm
	}
	for _, st := range sts {
		inm[st.Orig] = layout.NoteInstrument{
			NoteRemap: st.Remap,
			Inst:      sample,
		}
	}
}

func readIT(filename string) (*layout.Song, error) {
	buffer, err := formatutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	s, err := itfile.Read(buffer)
	if err != nil {
		return nil, err
	}

	return convertItFileToSong(s)
}
