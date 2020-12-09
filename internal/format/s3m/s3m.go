// Package s3m does a thing.
package s3m

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"gotracker/internal/format/s3m/channel"
	"gotracker/internal/format/s3m/effect"
	"gotracker/internal/format/s3m/util"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/state"
	"gotracker/internal/player/volume"
	"log"
	"os"
)

// ChannelID is the S3M value for a channel specification (found within the ChanenlSetting header block)
type ChannelID uint8

const (
	// ChannelIDL1 is the Left Channel 1
	ChannelIDL1 = ChannelID(0 + iota)
	// ChannelIDL2 is the Left Channel 2
	ChannelIDL2
	// ChannelIDL3 is the Left Channel 3
	ChannelIDL3
	// ChannelIDL4 is the Left Channel 4
	ChannelIDL4
	// ChannelIDL5 is the Left Channel 5
	ChannelIDL5
	// ChannelIDL6 is the Left Channel 6
	ChannelIDL6
	// ChannelIDL7 is the Left Channel 7
	ChannelIDL7
	// ChannelIDL8 is the Left Channel 8
	ChannelIDL8
	// ChannelIDR1 is the Right Channel 1
	ChannelIDR1
	// ChannelIDR2 is the Right Channel 2
	ChannelIDR2
	// ChannelIDR3 is the Right Channel 3
	ChannelIDR3
	// ChannelIDR4 is the Right Channel 4
	ChannelIDR4
	// ChannelIDR5 is the Right Channel 5
	ChannelIDR5
	// ChannelIDR6 is the Right Channel 6
	ChannelIDR6
	// ChannelIDR7 is the Right Channel 7
	ChannelIDR7
	// ChannelIDR8 is the Right Channel 8
	ChannelIDR8
	// ChannelIDOPL2Melody1 is the Adlib/OPL2 Melody Channel 1
	ChannelIDOPL2Melody1
	// ChannelIDOPL2Melody2 is the Adlib/OPL2 Melody Channel 2
	ChannelIDOPL2Melody2
	// ChannelIDOPL2Melody3 is the Adlib/OPL2 Melody Channel 3
	ChannelIDOPL2Melody3
	// ChannelIDOPL2Melody4 is the Adlib/OPL2 Melody Channel 4
	ChannelIDOPL2Melody4
	// ChannelIDOPL2Melody5 is the Adlib/OPL2 Melody Channel 5
	ChannelIDOPL2Melody5
	// ChannelIDOPL2Melody6 is the Adlib/OPL2 Melody Channel 6
	ChannelIDOPL2Melody6
	// ChannelIDOPL2Melody7 is the Adlib/OPL2 Melody Channel 7
	ChannelIDOPL2Melody7
	// ChannelIDOPL2Melody8 is the Adlib/OPL2 Melody Channel 8
	ChannelIDOPL2Melody8
	// ChannelIDOPL2Drums1 is the Adlib/OPL2 Drums Channel 1
	ChannelIDOPL2Drums1
	// ChannelIDOPL2Drums2 is the Adlib/OPL2 Drums Channel 2
	ChannelIDOPL2Drums2
	// ChannelIDOPL2Drums3 is the Adlib/OPL2 Drums Channel 3
	ChannelIDOPL2Drums3
	// ChannelIDOPL2Drums4 is the Adlib/OPL2 Drums Channel 4
	ChannelIDOPL2Drums4
	// ChannelIDOPL2Drums5 is the Adlib/OPL2 Drums Channel 5
	ChannelIDOPL2Drums5
	// ChannelIDOPL2Drums6 is the Adlib/OPL2 Drums Channel 6
	ChannelIDOPL2Drums6
	// ChannelIDOPL2Drums7 is the Adlib/OPL2 Drums Channel 7
	ChannelIDOPL2Drums7
	// ChannelIDOPL2Drums8 is the Adlib/OPL2 Drums Channel 8
	ChannelIDOPL2Drums8
)

// ParaPointer is a pointer offset within the S3M file format
type ParaPointer uint16

// ChannelSetting is a full channel setting (with flags) definition from the S3M header
type ChannelSetting uint8

// IsEnabled returns the enabled flag (bit 7 is unset)
func (cs ChannelSetting) IsEnabled() bool {
	return (uint8(cs) & 0x80) == 0
}

// GetChannel returns the ChannelID for the channel
func (cs ChannelSetting) GetChannel() ChannelID {
	ch := uint8(cs) & 0x7F
	return ChannelID(ch)
}

// IsPCM returns true if the channel is one of the left or right channels (non-Adlib/OPL2)
func (cs ChannelSetting) IsPCM() bool {
	ch := uint8(cs) & 0x7F
	return (ch < 16)
}

// IsOPL2 returns true if the channel is an Adlib/OPL2 channel (non-PCM)
func (cs ChannelSetting) IsOPL2() bool {
	ch := uint8(cs) & 0x7F
	return (ch >= 16)
}

// PanningFlags is a flagset and panning value for the panning system
type PanningFlags uint8

// IsValid returns true if bit 5 is set
func (pf PanningFlags) IsValid() bool {
	return (uint8(pf) & 0x20) != 0
}

// Value returns the panning position (0=full left, 15=full right)
func (pf PanningFlags) Value() uint8 {
	return uint8(pf) & 0x0F
}

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

// Header is a mildly-decoded S3M header definition
type Header struct {
	Name               string
	Info               ModuleHeader
	ChannelSettings    [32]ChannelSetting
	OrderList          []uint8
	InstrumentPointers []ParaPointer
	PatternPointers    []ParaPointer
	Panning            [32]PanningFlags
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

// SampleFileFormat is the mildly-decoded S3M instrument/sample header
type SampleFileFormat struct {
	intf.Instrument
	Filename string
	Name     string
	Info     SCRSHeader
	Sample   []uint8
	ID       uint8
	C2Spd    note.C2SPD
	Volume   volume.Volume
}

// IsInvalid always returns false (valid)
func (sff *SampleFileFormat) IsInvalid() bool {
	return false
}

// GetC2Spd returns the C2SPD value for the instrument
// This may get mutated if a finetune command is processed
func (sff *SampleFileFormat) GetC2Spd() note.C2SPD {
	return sff.C2Spd
}

// SetC2Spd sets the C2SPD value for the instrument
func (sff *SampleFileFormat) SetC2Spd(c2spd note.C2SPD) {
	sff.C2Spd = c2spd
}

// GetVolume returns the default volume value for the instrument
func (sff *SampleFileFormat) GetVolume() volume.Volume {
	return sff.Volume
}

// IsLooped returns true if the instrument has the loop flag set
func (sff *SampleFileFormat) IsLooped() bool {
	return sff.Info.Flags.IsLooped()
}

// GetLoopBegin returns the loop start position
func (sff *SampleFileFormat) GetLoopBegin() int {
	return int(sff.Info.LoopBeginL)
}

// GetLoopEnd returns the loop end position
func (sff *SampleFileFormat) GetLoopEnd() int {
	return int(sff.Info.LoopEndL)
}

// GetLength returns the length of the instrument
func (sff *SampleFileFormat) GetLength() int {
	return len(sff.Sample)
}

// GetSample returns the sample at position `pos` in the instrument
func (sff *SampleFileFormat) GetSample(pos int) volume.Volume {
	return util.VolumeFromS3M8BitSample(sff.Sample[pos])
}

// GetID returns the instrument number (1-based)
func (sff *SampleFileFormat) GetID() int {
	return int(sff.ID)
}

// PackedPattern is the S3M packed pattern definition
type PackedPattern struct {
	Length uint16
	Data   []byte
}

// RowData is the data for each row
type RowData struct {
	intf.Row
	Channels [32]channel.Data
}

// GetChannels returns an interface to all the channels in the row
func (r RowData) GetChannels() []intf.ChannelData {
	c := make([]intf.ChannelData, len(r.Channels))
	for i := range r.Channels {
		c[i] = &r.Channels[i]
	}

	return c
}

// Pattern is the data for each pattern
type Pattern struct {
	intf.Pattern
	Packed PackedPattern
	Rows   [64]RowData
}

// GetRow returns the interface to the row at index `row`
func (p Pattern) GetRow(row uint8) intf.Row {
	return &p.Rows[row]
}

// GetRows returns the interfaces to all the rows in the pattern
func (p Pattern) GetRows() []intf.Row {
	rows := make([]intf.Row, len(p.Rows))
	for i, pr := range p.Rows {
		rows[i] = pr
	}
	return rows
}

// Song is the full definition of the song data of an Song file
type Song struct {
	intf.SongData
	Head        Header
	Instruments []SampleFileFormat
	Patterns    []Pattern
}

// GetOrderList returns the list of all pattern orders for the song
func (s *Song) GetOrderList() []uint8 {
	return s.Head.OrderList
}

// GetPatternsInterface returns an interface to all the patterns
func (s *Song) GetPatternsInterface() []intf.Pattern {
	p := make([]intf.Pattern, len(s.Patterns))
	for i, sp := range s.Patterns {
		p[i] = sp
	}
	return p
}

// GetPattern returns an interface to a specific pattern indexed by `patNum`
func (s *Song) GetPattern(patNum uint8) intf.Pattern {
	if int(patNum) >= len(s.Patterns) {
		return nil
	}
	return &s.Patterns[patNum]
}

// IsChannelEnabled returns true if the channel at index `channelNum` is enabled
func (s *Song) IsChannelEnabled(channelNum int) bool {
	return s.Head.ChannelSettings[channelNum].IsEnabled()
}

// NumInstruments returns the number of instruments in the song
func (s *Song) NumInstruments() int {
	return len(s.Instruments)
}

// GetInstrument returns the instrument interface indexed by `instNum` (0-based)
func (s *Song) GetInstrument(instNum int) intf.Instrument {
	return &s.Instruments[instNum]
}

// GetName returns the name of the song
func (s *Song) GetName() string {
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

type modHeader struct {
	Name       [20]byte
	Samples    [31]modSample
	SongLen    uint8
	RestartPos uint8
	Order      [128]uint8
	Sig        [4]uint8
}

type modSample struct {
	Name      [22]byte
	Len       uint16
	FineTune  uint8
	Volume    uint8
	LoopStart uint16
	LoopEnd   uint16
}

func readMODHeader(channels int, buffer *bytes.Buffer) (*Header, *modHeader, error) {
	head := Header{}

	mh := modHeader{}
	binary.Read(buffer, binary.LittleEndian, &mh)

	numPatterns := 0
	head.OrderList = make([]uint8, mh.SongLen)
	for i, o := range mh.Order {
		if i < int(mh.SongLen) {
			head.OrderList[i] = o
		}
		if numPatterns-1 < int(o) {
			numPatterns = int(o) + 1
		}
	}

	head.Name = getString(mh.Name[:])
	head.Info.PatternCount = uint16(numPatterns)
	head.Info.OrderCount = uint16(mh.SongLen)
	head.Info.InitialSpeed = 6
	head.Info.InitialTempo = 125
	head.Info.GlobalVolume = 64

	for i := 0; i < 32; i++ {
		if i >= channels {
			head.ChannelSettings[i] = 255
		} else if i%1 == 0 {
			head.ChannelSettings[i] = ChannelSetting(uint8(ChannelIDL1) + uint8(i)>>1)
		} else {
			head.ChannelSettings[i] = ChannelSetting(uint8(ChannelIDR1) + uint8(i)>>1)
		}
	}

	return &head, &mh, nil
}

func readMODPattern(buffer *bytes.Buffer, channels int) *Pattern {
	pattern := Pattern{}

	for r := 0; r < 64; r++ {
		for c := 0; c < channels; c++ {
			var data [4]uint8
			buffer.Read(data[:])
			sampleNumber := (data[0] & 0xF0) | (data[2] >> 4)
			samplePeriod := (uint16(data[0]&0x0F) << 8) | uint16(data[1])
			effect := (data[2] & 0x0F)
			effectParameter := data[3]

			cd := &pattern.Rows[r].Channels[c]
			cd.What = channel.What(c)

			if sampleNumber != 0 || samplePeriod != 0 {
				cd.What = cd.What | channel.What(0x20)
				cd.Instrument = sampleNumber
				cd.Note = util.ModPeriodToNote(samplePeriod * 4)
			}
			cd.Volume = 255
			if effect != 0 || cd.Info != 0 {
				cd.Info = effectParameter
				switch effect {
				case 0xF: // Set Speed / Tempo
					cd.What = cd.What | channel.What(0x80)
					if cd.Info < 0x20 {
						cd.Command = 'A' - '@' // Set Speed
					} else {
						cd.Command = 'T' - '@' // Tempo
					}
				case 0xB: // Pattern Jump
					cd.What = cd.What | channel.What(0x80)
					cd.Command = 'B' - '@'
				case 0xD: // Pattern Break
					cd.What = cd.What | channel.What(0x80)
					cd.Command = 'C' - '@'
				case 0xA: // Volume Slide
					cd.What = cd.What | channel.What(0x80)
					cd.Command = 'D' - '@'
				case 0x2: // Porta Down
					cd.What = cd.What | channel.What(0x80)
					cd.Command = 'E' - '@'
				case 0x1: // Porta Up
					cd.What = cd.What | channel.What(0x80)
					cd.Command = 'F' - '@'
				case 0x3: // Porta to Note
					cd.What = cd.What | channel.What(0x80)
					cd.Command = 'G' - '@'
				case 0x4: // Vibrato
					cd.What = cd.What | channel.What(0x80)
					cd.Command = 'H' - '@'
				case 0x0: // Arpeggio
					cd.What = cd.What | channel.What(0x80)
					cd.Command = 'J' - '@'
				case 0x6: // Vibrato+VolSlide
					cd.What = cd.What | channel.What(0x80)
					cd.Command = 'K' - '@'
				case 0x5: // Porta+VolSlide
					cd.What = cd.What | channel.What(0x80)
					cd.Command = 'L' - '@'
				case 0x9: // Sample Offset
					cd.What = cd.What | channel.What(0x80)
					cd.Command = 'O' - '@'
				case 0x7: // Tremolo
					cd.What = cd.What | channel.What(0x80)
					cd.Command = 'R' - '@'
				case 0xC: // Set Volume
					cd.What = cd.What | channel.What(0x40)
					cd.Volume = cd.Info
				}

				if effect == 0xE {
					// special
					switch effectParameter >> 4 {
					case 0xA: // Fine VolSlide down
						cd.What = cd.What | channel.What(0x80)
						cd.Command = 'D' - '@'
						cd.Info = 0xF0 | (effectParameter & 0x0F)
					case 0xB: // Fine VolSlide up
						cd.What = cd.What | channel.What(0x80)
						cd.Command = 'S' - '@'
						cd.Info = ((effectParameter & 0x0F) << 4) | 0x0F
					case 0x2: // Fine Porta Down
						cd.What = cd.What | channel.What(0x80)
						cd.Command = 'E' - '@'
						cd.Info = 0xF0 | (effectParameter & 0x0F)
					case 0x1: // Fine Porta Up
						cd.What = cd.What | channel.What(0x80)
						cd.Command = 'F' - '@'
						cd.Info = 0xF0 | (effectParameter & 0x0F)
					case 0x9: // Retrig+VolSlide
						cd.What = cd.What | channel.What(0x80)
						cd.Command = 'Q' - '@'
						cd.Info = (effectParameter & 0x0F)
					case 0x0: // Set Filter on/off
						cd.What = cd.What | channel.What(0x80)
						cd.Command = 'S' - '@'
						cd.Info = 0x00 | (effectParameter & 0x0F)
					case 0x3: // Set Glissando on/off
						cd.What = cd.What | channel.What(0x80)
						cd.Command = 'S' - '@'
						cd.Info = 0x10 | (effectParameter & 0x0F)
					case 0x5: // Set FineTune
						cd.What = cd.What | channel.What(0x80)
						cd.Command = 'S' - '@'
						cd.Info = 0x20 | (effectParameter & 0x0F)
					case 0x4: // Set Vibrato Waveform
						cd.What = cd.What | channel.What(0x80)
						cd.Command = 'S' - '@'
						cd.Info = 0x30 | (effectParameter & 0x0F)
					case 0x7: // Set Tremolo Waveform
						cd.What = cd.What | channel.What(0x80)
						cd.Command = 'S' - '@'
						cd.Info = 0x40 | (effectParameter & 0x0F)
					case 0x8: // Set Pan Position
						cd.What = cd.What | channel.What(0x80)
						cd.Command = 'S' - '@'
						cd.Info = 0x80 | (effectParameter & 0x0F)
					case 0x6: // Pattern Loop
						cd.What = cd.What | channel.What(0x80)
						cd.Command = 'S' - '@'
						cd.Info = 0xB0 | (effectParameter & 0x0F)
					case 0xC: // Note Cut
						cd.What = cd.What | channel.What(0x80)
						cd.Command = 'S' - '@'
						cd.Info = 0xC0 | (effectParameter & 0x0F)
					case 0xD: // Note Delay
						cd.What = cd.What | channel.What(0x80)
						cd.Command = 'S' - '@'
						cd.Info = 0xD0 | (effectParameter & 0x0F)
					case 0xE: // Pattern Delay
						cd.What = cd.What | channel.What(0x80)
						cd.Command = 'S' - '@'
						cd.Info = 0xE0 | (effectParameter & 0x0F)
					case 0xF: // Funk Repeat
						cd.What = cd.What | channel.What(0x80)
						cd.Command = 'S' - '@'
						cd.Info = 0xF0 | (effectParameter & 0x0F)
					}
				}
			}
		}
	}

	return &pattern
}

func readMODSample(buffer *bytes.Buffer, num int, inst modSample) *SampleFileFormat {
	var sample = SampleFileFormat{}
	sample.Filename = fmt.Sprintf("inst%0.2d.bin", num+1)
	sample.Name = getString(inst.Name[:])
	sl := util.BE16ToLE16(inst.Len)

	switch inst.FineTune {
	case 0:
		sample.C2Spd = 8363
	case 1:
		sample.C2Spd = 8413
	case 2:
		sample.C2Spd = 8463
	case 3:
		sample.C2Spd = 8529
	case 4:
		sample.C2Spd = 8581
	case 5:
		sample.C2Spd = 8651
	case 6:
		sample.C2Spd = 8723
	case 7:
		sample.C2Spd = 8757
	case 8:
		sample.C2Spd = 7895
	case 9:
		sample.C2Spd = 7941
	case 10:
		sample.C2Spd = 7985
	case 11:
		sample.C2Spd = 8046
	case 12:
		sample.C2Spd = 8107
	case 13:
		sample.C2Spd = 8169
	case 14:
		sample.C2Spd = 8232
	case 15:
		sample.C2Spd = 8280
	default:
		sample.C2Spd = 8363
	}

	sample.Volume = util.VolumeFromS3M(inst.Volume)
	sample.Info.LoopBeginL = util.BE16ToLE16(inst.LoopStart)
	sample.Info.LoopEndL = util.BE16ToLE16(inst.LoopEnd)
	if sample.Info.LoopBeginL != sample.Info.LoopEndL {
		sample.Info.Flags = 1
	}

	samps := make([]uint8, sl)
	buffer.Read(samps)
	sample.Sample = samps
	for i, s := range samps {
		sample.Sample[i] = util.MODSampleToS3MSample(s)
	}
	return &sample
}

func readS3MSample(data []byte, ptr ParaPointer) *SampleFileFormat {
	pos := int(ptr) * 16
	if pos >= len(data) {
		return nil
	}
	buffer := bytes.NewBuffer(data[pos:])
	var sample = SampleFileFormat{}
	binary.Read(buffer, binary.LittleEndian, &sample.Info)
	sample.Filename = getString(sample.Info.Filename[:])
	sample.Name = getString(sample.Info.SampleName[:])
	sample.Sample = make([]uint8, sample.Info.Length)
	if sample.Info.C2SpdL != 0 {
		sample.C2Spd = note.C2SPD(sample.Info.C2SpdL)
	} else {
		sample.C2Spd = util.DefaultC2Spd
	}

	sample.Volume = util.VolumeFromS3M(sample.Info.Volume)

	pos = (int(sample.Info.MemSegL) + int(sample.Info.MemSegH)*65536) * 16
	copy(sample.Sample, data[pos:pos+int(sample.Info.Length)])
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

type format struct {
	intf.Format
}

var (
	// S3M is the exported interface to the S3M file loader
	S3M = format{}
)

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

	song.Instruments = make([]SampleFileFormat, len(song.Head.InstrumentPointers))
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

type modSig struct {
	sig      string
	channels int
}

var (
	sigChannels = [...]modSig{
		// amiga / protracker
		{"M.K.", 4},
		// fasttracker
		{"6CHN", 6}, {"8CHN", 8},
		// (unusual)
		{"10CH", 10}, {"11CH", 11}, {"12CH", 12}, {"13CH", 13}, {"14CH", 14},
		{"15CH", 15}, {"16CH", 16}, {"17CH", 17}, {"18CH", 18}, {"19CH", 19},
		{"20CH", 20}, {"21CH", 21}, {"22CH", 22}, {"23CH", 23}, {"24CH", 24},
		{"25CH", 25}, {"26CH", 26}, {"27CH", 27}, {"28CH", 28}, {"29CH", 29},
		{"30CH", 30}, {"31CH", 31}, {"32CH", 32},
	}
)

func readMOD(filename string) (*Song, error) {
	buffer, err := readFile(filename)
	if err != nil {
		return nil, err
	}
	data := buffer.Bytes()

	song := Song{}

	sig := getString(data[1080:1084])
	numCh := 0
	for _, s := range sigChannels {
		if s.sig == sig {
			numCh = s.channels
			break
		}
	}

	if numCh == 0 {
		return nil, errors.New("invalid file format")
	}

	h, mh, err := readMODHeader(numCh, buffer)
	if err != nil {
		return nil, err
	}
	song.Head = *h

	song.Patterns = make([]Pattern, song.Head.Info.PatternCount)
	for i := 0; i < int(song.Head.Info.PatternCount); i++ {
		var pattern = readMODPattern(buffer, numCh)
		if pattern == nil {
			continue
		}
		song.Patterns[i] = *pattern
	}

	song.Instruments = make([]SampleFileFormat, len(mh.Samples))
	for instNum, inst := range mh.Samples {
		var sample = readMODSample(buffer, instNum, inst)
		if sample == nil {
			continue
		}
		sample.ID = uint8(instNum + 1)
		song.Instruments[instNum] = *sample
	}

	return &song, nil
}

// LoadMOD loads a MOD file and upgrades it into an S3M file internally
func LoadMOD(s intf.Song, filename string) error {
	return load(s, filename, readMOD)
}

type readerFunc func(filename string) (*Song, error)

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

// GetBaseClockRate returns the base clock rate for the S3M player
func (f format) GetBaseClockRate() float32 {
	return util.S3MBaseClock
}

// Load loads an S3M file into the song state `s`
func (f format) Load(s intf.Song, filename string) error {
	return load(s, filename, readS3M)
}
