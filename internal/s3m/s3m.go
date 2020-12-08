// Package s3m does a thing.
package s3m

import (
	"bytes"
	"encoding/binary"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/state"
	"gotracker/internal/player/volume"
	"gotracker/internal/s3m/channel"
	"gotracker/internal/s3m/effect"
	"gotracker/internal/s3m/util"
	"log"
	"os"
)

type ChannelID uint8

const (
	ChannelIDL1 = ChannelID(0 + iota)
	ChannelIDL2
	ChannelIDL3
	ChannelIDL4
	ChannelIDL5
	ChannelIDL6
	ChannelIDL7
	ChannelIDL8
	ChannelIDR1
	ChannelIDR2
	ChannelIDR3
	ChannelIDR4
	ChannelIDR5
	ChannelIDR6
	ChannelIDR7
	ChannelIDR8
	ChannelIDOPL2Melody1
	ChannelIDOPL2Melody2
	ChannelIDOPL2Melody3
	ChannelIDOPL2Melody4
	ChannelIDOPL2Melody5
	ChannelIDOPL2Melody6
	ChannelIDOPL2Melody7
	ChannelIDOPL2Melody8
	ChannelIDOPL2Drums1
	ChannelIDOPL2Drums2
	ChannelIDOPL2Drums3
	ChannelIDOPL2Drums4
	ChannelIDOPL2Drums5
	ChannelIDOPL2Drums6
	ChannelIDOPL2Drums7
	ChannelIDOPL2Drums8
)

type ParaPointer uint16

type ChannelSetting uint8

func (cs ChannelSetting) IsEnabled() bool {
	return (uint8(cs) & 0x80) == 0
}

func (cs ChannelSetting) GetChannel() ChannelID {
	ch := uint8(cs) & 0x7F
	return ChannelID(ch)
}

func (cs ChannelSetting) IsPCM() bool {
	ch := uint8(cs) & 0x7F
	return (ch < 16)
}

func (cs ChannelSetting) IsOPL2() bool {
	ch := uint8(cs) & 0x7F
	return (ch >= 16)
}

type PanningFlags uint8

func (pf PanningFlags) IsValid() bool {
	return (uint8(pf) & 0x20) != 0
}

func (pf PanningFlags) Value() uint8 {
	return uint8(pf) & 0x0F
}

type ModuleHeader struct {
	Name                  [28]byte
	Reserved1C            byte
	Type                  uint8
	Reserved1E            [2]byte
	OrderCount            uint16
	InstrumentCount       uint16
	PatternCount          uint16
	Flags                 uint16
	TrackerVersion        uint16
	FileFormatInformation uint16
	SCRM                  [4]byte
	GlobalVolume          uint8
	InitialSpeed          uint8
	InitialTempo          uint8
	MixingVolume          uint8
	UltraClickRemoval     uint8
	DefaultPanValueFlag   uint8
	Reserved34            [8]byte
	Special               ParaPointer
}

type S3MHeader struct {
	Name               string
	Info               ModuleHeader
	ChannelSettings    [32]ChannelSetting
	OrderList          []uint8
	InstrumentPointers []ParaPointer
	PatternPointers    []ParaPointer
	Panning            [32]PanningFlags
}

type SCRSHeader struct {
	Type          uint8
	Filename      [12]byte
	MemSegH       uint8
	MemSegL       ParaPointer
	Length        uint16
	HiLeng        uint16
	LoopBeginL    uint16
	LoopBeginH    uint16
	LoopEndL      uint16
	LoopEndH      uint16
	Volume        uint8
	Reserved1D    uint8
	PackingScheme uint8
	Flags         uint8
	C2SpdL        uint16
	C2SpdH        uint16
	Reserved24    [4]byte
	IntGp         uint16
	Int512        uint16
	IntLastused   uint32
	SampleName    [28]byte
	SCRS          [4]uint8
}

type SampleFileFormat struct {
	intf.Instrument
	Filename string
	Name     string
	Info     SCRSHeader
	Sample   []uint8
	Id       uint8
}

func (sff *SampleFileFormat) IsInvalid() bool {
	return false
}

func (sff *SampleFileFormat) GetC2Spd() uint16 {
	return sff.Info.C2SpdL
}

func (sff *SampleFileFormat) SetC2Spd(c2spd uint16) {
	sff.Info.C2SpdL = c2spd
}

func (sff *SampleFileFormat) GetVolume() volume.Volume {
	return util.VolumeFromS3M(sff.Info.Volume)
}

func (sff *SampleFileFormat) IsLooped() bool {
	return (sff.Info.Flags & 1) != 0
}

func (sff *SampleFileFormat) GetLoopBegin() int {
	return int(sff.Info.LoopBeginL)
}

func (sff *SampleFileFormat) GetLoopEnd() int {
	return int(sff.Info.LoopEndL)
}

func (sff *SampleFileFormat) GetLength() int {
	return len(sff.Sample)
}

func (sff *SampleFileFormat) GetSample(pos int) volume.Volume {
	return volume.FromS3MSample(sff.Sample[pos])
}

func (sff *SampleFileFormat) GetId() int {
	return int(sff.Id)
}

type PackedPattern struct {
	Length uint16
	Data   []byte
}

type RowData struct {
	intf.Row
	Channels [32]channel.Data
}

func (r RowData) GetChannels() []intf.ChannelData {
	c := make([]intf.ChannelData, len(r.Channels))
	for i, _ := range r.Channels {
		c[i] = &r.Channels[i]
	}

	return c
}

type Pattern struct {
	intf.Pattern
	Packed PackedPattern
	Rows   [64]RowData
}

func (p Pattern) GetRow(row uint8) intf.Row {
	return &p.Rows[row]
}

func (p Pattern) GetRows() []intf.Row {
	rows := make([]intf.Row, len(p.Rows))
	for i, pr := range p.Rows {
		rows[i] = pr
	}
	return rows
}

type S3M struct {
	intf.SongData
	Head        S3MHeader
	Instruments []SampleFileFormat
	Patterns    []Pattern
}

func (s *S3M) GetOrderList() []uint8 {
	return s.Head.OrderList
}

func (s *S3M) GetPatternsInterface() *[]intf.Pattern {
	p := make([]intf.Pattern, len(s.Patterns))
	for i, sp := range s.Patterns {
		p[i] = sp
	}
	return &p
}

func (s *S3M) GetPattern(patNum uint8) intf.Pattern {
	if int(patNum) >= len(s.Patterns) {
		return nil
	}
	return &s.Patterns[patNum]
}

func (s *S3M) IsChannelEnabled(channelNum int) bool {
	return s.Head.ChannelSettings[channelNum].IsEnabled()
}

func (s *S3M) NumInstruments() int {
	return len(s.Instruments)
}

func (s *S3M) GetInstrument(instNum int) intf.Instrument {
	return &s.Instruments[instNum]
}

func (s *S3M) GetName() string {
	return s.Head.Name
}

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

func readHeader(buffer *bytes.Buffer) *S3MHeader {
	var head = S3MHeader{}
	binary.Read(buffer, binary.LittleEndian, &head.Info)
	head.Name = getString(head.Info.Name[:])
	head.OrderList = make([]uint8, head.Info.OrderCount)
	binary.Read(buffer, binary.LittleEndian, &head.ChannelSettings)
	binary.Read(buffer, binary.LittleEndian, &head.OrderList)
	head.InstrumentPointers = make([]ParaPointer, head.Info.InstrumentCount)
	binary.Read(buffer, binary.LittleEndian, &head.InstrumentPointers)
	head.PatternPointers = make([]ParaPointer, head.Info.PatternCount)
	binary.Read(buffer, binary.LittleEndian, &head.PatternPointers)
	if head.Info.DefaultPanValueFlag == 252 {
		binary.Read(buffer, binary.LittleEndian, &head.Panning)
	}
	return &head
}

func readSample(data []byte, ptr ParaPointer) *SampleFileFormat {
	pos := int(ptr) * 16
	if pos >= len(data) {
		return nil
	}
	buffer := bytes.NewBuffer(data[pos:])
	var sample = SampleFileFormat{}
	binary.Read(buffer, binary.LittleEndian, &sample.Info)
	if sample.Info.C2SpdL == 0 {
		sample.Info.C2SpdL = util.DefaultC2Spd
	}
	sample.Filename = getString(sample.Info.Filename[:])
	sample.Name = getString(sample.Info.SampleName[:])
	sample.Sample = make([]uint8, sample.Info.Length)

	pos = (int(sample.Info.MemSegL) + int(sample.Info.MemSegH)*65536) * 16
	copy(sample.Sample, data[pos:pos+int(sample.Info.Length)])
	return &sample
}

func readPattern(data []byte, ptr ParaPointer) *Pattern {
	pos := int(ptr) * 16
	if pos >= len(data) {
		return nil
	}
	buffer := bytes.NewBuffer(data[pos:])
	var pattern = new(Pattern)
	binary.Read(buffer, binary.LittleEndian, &pattern.Packed.Length)
	pattern.Packed.Data = make([]byte, pattern.Packed.Length-2)
	binary.Read(buffer, binary.LittleEndian, &pattern.Packed.Data)

	buffer = bytes.NewBuffer(pattern.Packed.Data)

	rowNum := 0
	for rowNum < len(pattern.Rows) {
		row := &pattern.Rows[rowNum]
		for {
			var what channel.What
			err := binary.Read(buffer, binary.LittleEndian, &what)

			if err != nil {
				log.Fatal(err)
			}
			if what == 0 {
				rowNum++
				break
			}

			channelNum := what.Channel()
			temp := &row.Channels[channelNum]

			temp.What = what
			temp.Note = 0
			temp.Instrument = 0
			temp.Volume = 255
			temp.Command = 0
			temp.Info = 0

			if temp.What.HasNote() {
				binary.Read(buffer, binary.LittleEndian, &temp.Note)
				binary.Read(buffer, binary.LittleEndian, &temp.Instrument)
			}

			if temp.What.HasVolume() {
				binary.Read(buffer, binary.LittleEndian, &temp.Volume)
			}

			if temp.What.HasCommand() {
				binary.Read(buffer, binary.LittleEndian, &temp.Command)
				binary.Read(buffer, binary.LittleEndian, &temp.Info)
			}
		}
	}

	return pattern
}

func ReadS3M(filename string) (*S3M, error) {
	buffer, err := readFile(filename)
	if err != nil {
		return nil, err
	}
	data := buffer.Bytes()

	var song = new(S3M)
	song.Head = *readHeader(buffer)

	song.Instruments = make([]SampleFileFormat, len(song.Head.InstrumentPointers))
	for instNum, ptr := range song.Head.InstrumentPointers {
		var sample = readSample(data, ptr)
		if sample == nil {
			continue
		}
		sample.Id = uint8(instNum + 1)
		song.Instruments[instNum] = *sample
	}

	song.Patterns = make([]Pattern, len(song.Head.PatternPointers))
	for patNum, ptr := range song.Head.PatternPointers {
		var pattern = readPattern(data, ptr)
		if pattern == nil {
			continue
		}
		song.Patterns[patNum] = *pattern
	}

	return song, nil
}

func GetBaseClockRate() float32 {
	return util.S3MBaseClock
}

func Load(ss *state.Song, filename string) error {
	s3mSong, err := ReadS3M(filename)
	if err != nil {
		return err
	}

	ss.EffectFactory = effect.EffectFactory
	ss.Pattern.Patterns = s3mSong.GetPatternsInterface()
	ss.Pattern.Orders = &s3mSong.Head.OrderList
	ss.Pattern.Row.Ticks = int(s3mSong.Head.Info.InitialSpeed)
	ss.Pattern.Row.Tempo = int(s3mSong.Head.Info.InitialTempo)

	ss.GlobalVolume = util.VolumeFromS3M(s3mSong.Head.Info.GlobalVolume)
	ss.SongData = s3mSong

	// new method for determining active channels (uses S3M data I somehow overlooked before)
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
		ch := s3mSong.Head.ChannelSettings[i]
		if ch.IsEnabled() {
			pf := s3mSong.Head.Panning[i]
			if pf.IsValid() {
				cs.Pan = pf.Value()
			} else {
				l := ch.GetChannel()
				switch l {
				case ChannelIDL1, ChannelIDL2, ChannelIDL3, ChannelIDL4, ChannelIDL5, ChannelIDL6, ChannelIDL7, ChannelIDL8:
					cs.Pan = 0x03
				case ChannelIDR1, ChannelIDR2, ChannelIDR3, ChannelIDR4, ChannelIDR5, ChannelIDR6, ChannelIDR7, ChannelIDR8:
					cs.Pan = 0x0C
				}
			}
		} else {
			cs.Pan = 0x08 // center?
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
