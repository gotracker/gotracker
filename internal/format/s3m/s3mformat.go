package s3m

import (
	"bytes"
	"encoding/binary"
	"errors"
	"gotracker/internal/format/s3m/channel"
	"gotracker/internal/format/s3m/util"
	"gotracker/internal/player/note"
	"log"
)

// ModuleHeader is the initial header definition of an S3M file
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

// SCRSFlags is a bitset for the S3M instrument/sample header definition
type SCRSFlags uint8

// IsLooped returns true if bit 0 is set
func (f SCRSFlags) IsLooped() bool {
	return (uint8(f) & 0x01) != 0
}

// IsStereo returns true if bit 1 is set
func (f SCRSFlags) IsStereo() bool {
	return (uint8(f) & 0x02) != 0
}

// Is16BitSample returns true if bit 2 is set
func (f SCRSFlags) Is16BitSample() bool {
	return (uint8(f) & 0x04) != 0
}

// SCRSHeader is the S3M instrument/sample header definition
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
	Flags         SCRSFlags
	C2SpdL        uint16
	C2SpdH        uint16
	Reserved24    [4]byte
	IntGp         uint16
	Int512        uint16
	IntLastused   uint32
	SampleName    [28]byte
	SCRS          [4]uint8
}

// PackedPattern is the S3M packed pattern definition
type PackedPattern struct {
	Length uint16
	Data   []byte
}

func readS3MHeader(buffer *bytes.Buffer) (*Header, error) {
	var head = Header{}
	binary.Read(buffer, binary.LittleEndian, &head.Info)
	if getString(head.Info.SCRM[:]) != "SCRM" {
		return nil, errors.New("invalid file format")
	}
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
	return &head, nil
}

func readS3MSample(data []byte, ptr ParaPointer) *Instrument {
	pos := int(ptr) * 16
	if pos >= len(data) {
		return nil
	}
	buffer := bytes.NewBuffer(data[pos:])
	sample := Instrument{}
	si := SCRSHeader{}
	binary.Read(buffer, binary.LittleEndian, &si)
	sample.Filename = getString(si.Filename[:])
	sample.Name = getString(si.SampleName[:])
	sample.Looped = si.Flags.IsLooped()
	sample.LoopBegin = int(si.LoopBeginL)
	sample.LoopEnd = int(si.LoopEndL)
	if si.C2SpdL != 0 {
		sample.C2Spd = note.C2SPD(si.C2SpdL)
	} else {
		sample.C2Spd = util.DefaultC2Spd
	}

	sample.Volume = util.VolumeFromS3M(si.Volume)
	sample.NumChannels = 1
	if si.Flags.IsStereo() {
		sample.NumChannels = 2
	}
	sample.BitsPerSample = 8
	if si.Flags.Is16BitSample() {
		sample.BitsPerSample = 16
	}

	sample.Length = int(si.Length)
	sample.Sample = make([]uint8, sample.Length)
	pos = (int(si.MemSegL) + int(si.MemSegH)*65536) * 16
	dataLen := sample.Length * sample.NumChannels * sample.BitsPerSample / 8
	copy(sample.Sample, data[pos:pos+dataLen])
	return &sample
}

func readS3MPattern(data []byte, ptr ParaPointer) *Pattern {
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

func readS3M(filename string) (*Song, error) {
	buffer, err := readFile(filename)
	if err != nil {
		return nil, err
	}
	data := buffer.Bytes()

	song := Song{}
	if h, err := readS3MHeader(buffer); err != nil {
		return nil, err
	} else if h != nil {
		song.Head = *h
	}

	song.Instruments = make([]Instrument, len(song.Head.InstrumentPointers))
	for instNum, ptr := range song.Head.InstrumentPointers {
		var sample = readS3MSample(data, ptr)
		if sample == nil {
			continue
		}
		sample.ID = uint8(instNum + 1)
		song.Instruments[instNum] = *sample
	}

	song.Patterns = make([]Pattern, len(song.Head.PatternPointers))
	for patNum, ptr := range song.Head.PatternPointers {
		var pattern = readS3MPattern(data, ptr)
		if pattern == nil {
			continue
		}
		song.Patterns[patNum] = *pattern
	}

	return &song, nil
}
