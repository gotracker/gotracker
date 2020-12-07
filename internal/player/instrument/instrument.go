package instrument

import (
	"s3mplayer/internal/s3m"
	"s3mplayer/internal/s3m/util"
)

type InstrumentInfo struct {
	Sample *s3m.SampleFileFormat
	Id     uint8
}

func (i InstrumentInfo) IsInvalid() bool {
	return i.Sample == nil
}

func (i InstrumentInfo) C2Spd() uint16 {
	if i.Sample == nil {
		return util.DefaultC2Spd
	}
	return i.Sample.Info.C2SpdL
}

func (i InstrumentInfo) SetC2Spd(c2spd uint16) {
	if i.Sample != nil {
		i.Sample.Info.C2SpdL = c2spd
	}
}
