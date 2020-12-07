package volume

type Volume float32

func FromS3M(vol uint8) Volume {
	var v Volume
	switch {
	case vol < 63:
		v = Volume(vol) / 64.0
	case vol >= 63:
		v = Volume(63.0) / 64.0
	default:
		v = 0.0
	}
	return v
}

func (v Volume) ToByte() uint8 {
	return uint8(v * 64.0)
}
