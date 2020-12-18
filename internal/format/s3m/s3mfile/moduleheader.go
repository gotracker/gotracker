package s3mfile

import (
	"encoding/binary"
	"io"
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
	Special               ParaPointer16
}

// ReadModuleHeader reads a ModuleHeader from the input stream
func ReadModuleHeader(r io.Reader) (*ModuleHeader, error) {
	var mh ModuleHeader
	if err := binary.Read(r, binary.LittleEndian, &mh); err != nil {
		return nil, err
	}

	return &mh, nil
}
