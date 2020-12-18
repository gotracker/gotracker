package s3mfile

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"gotracker/internal/format/s3m/util"
)

// File is an S3M internal file representation
type File struct {
	Head               ModuleHeader
	ChannelSettings    [32]ChannelSetting
	OrderList          []uint8
	InstrumentPointers []ParaPointer16
	PatternPointers    []ParaPointer16
	Panning            [32]PanningFlags
	Instruments        []SCRSFull
	Patterns           []PackedPattern
}

// SCRSFull is a full SCRS header + sample data (if applicable)
type SCRSFull struct {
	SCRS
	Sample []uint8
}

// Read reads an S3M file from the reader `r` and creates an internal File representation
func Read(r io.Reader) (*File, error) {
	buffer := &bytes.Buffer{}
	if _, err := buffer.ReadFrom(r); err != nil {
		return nil, err
	}
	data := buffer.Bytes()

	fh, err := ReadModuleHeader(buffer)
	if err != nil {
		return nil, err
	}
	if util.GetString(fh.SCRM[:]) != "SCRM" {
		return nil, errors.New("invalid file format")
	}

	f := File{
		Head:               *fh,
		OrderList:          make([]uint8, fh.OrderCount),
		InstrumentPointers: make([]ParaPointer16, fh.InstrumentCount),
		PatternPointers:    make([]ParaPointer16, fh.PatternCount),
		Instruments:        make([]SCRSFull, 0),
		Patterns:           make([]PackedPattern, 0),
	}
	if err := binary.Read(buffer, binary.LittleEndian, &f.ChannelSettings); err != nil {
		return nil, err
	}
	if err := binary.Read(buffer, binary.LittleEndian, &f.OrderList); err != nil {
		return nil, err
	}
	if err := binary.Read(buffer, binary.LittleEndian, &f.InstrumentPointers); err != nil {
		return nil, err
	}
	if err := binary.Read(buffer, binary.LittleEndian, &f.PatternPointers); err != nil {
		return nil, err
	}
	if fh.DefaultPanValueFlag == 0xFC {
		if err := binary.Read(buffer, binary.LittleEndian, &f.Panning); err != nil {
			return nil, err
		}
	}

	for _, ptr := range f.InstrumentPointers {
		sample, err := readS3MSample(data, ptr)
		if err != nil {
			return nil, err
		}
		if sample == nil {
			continue
		}
		f.Instruments = append(f.Instruments, *sample)
	}

	for _, ptr := range f.PatternPointers {
		pattern, err := readS3MPattern(data, ptr)
		if err != nil {
			return nil, err
		}
		if pattern == nil {
			continue
		}
		f.Patterns = append(f.Patterns, *pattern)
	}

	return &f, nil
}

func readS3MSample(data []byte, ptr ParaPointer) (*SCRSFull, error) {
	pos := ptr.Offset()
	if pos >= len(data) {
		return nil, errors.New("data out of range")
	}
	buffer := bytes.NewBuffer(data[pos:])
	scrs, err := ReadSCRS(buffer)
	if err != nil {
		return nil, err
	}

	s := SCRSFull{
		SCRS: *scrs,
	}

	switch si := s.Ancillary.(type) {
	case *SCRSDigiplayerHeader:
		numChannels := 1
		if si.Flags.IsStereo() {
			numChannels = 2
		}
		bitsPerSample := 8
		if si.Flags.Is16BitSample() {
			bitsPerSample = 16
		}
		filePos := si.MemSeg.Offset()
		dataLen := int(si.Length.Lo) * numChannels * bitsPerSample / 8
		s.Sample = data[filePos : filePos+dataLen]

	default:
		// do nothing
	}

	return &s, nil
}

func readS3MPattern(data []byte, ptr ParaPointer) (*PackedPattern, error) {
	pos := ptr.Offset()
	if pos >= len(data) {
		return nil, errors.New("data out of range")
	}
	p := PackedPattern{}
	p.Length = binary.LittleEndian.Uint16(data[pos:])
	p.Data = data[pos+2 : pos+int(p.Length)]

	return &p, nil
}
