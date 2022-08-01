package api

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/playback/player/feature"
	"github.com/heucuva/optional"
)

func getSampleFormat(format optional.Value[string]) (sampling.Format, error) {
	sampFmt := sampling.Format16BitLESigned
	value, set := format.Get()
	if !set {
		return sampFmt, nil
	}

	switch strings.ToLower(value) {
	case "uint8", "u8":
		sampFmt = sampling.Format8BitUnsigned
	case "int8", "i8", "s8":
		sampFmt = sampling.Format8BitSigned
	case "uint16le", "u16le":
		sampFmt = sampling.Format16BitLEUnsigned
	case "int16le", "i16le", "s16le":
		sampFmt = sampling.Format16BitLESigned
	case "uint16be", "u16be":
		sampFmt = sampling.Format16BitBEUnsigned
	case "int16be", "i16be", "s16be":
		sampFmt = sampling.Format16BitBESigned
	case "float32le", "f32le":
		sampFmt = sampling.Format32BitLEFloat
	case "float32be", "f32be":
		sampFmt = sampling.Format32BitBEFloat
	case "float64le", "f64le":
		sampFmt = sampling.Format64BitLEFloat
	case "float64be", "f64be":
		sampFmt = sampling.Format64BitBEFloat
	default:
		return sampFmt, fmt.Errorf("unhandled sampling format %q", value)
	}

	return sampFmt, nil
}

func getFeatureByType[T feature.Feature](features []feature.Feature) (T, bool) {
	var empty T
	if len(features) == 0 {
		return empty, false
	}

	tt := reflect.TypeOf(empty)
	for _, f := range features {
		v := reflect.ValueOf(f)
		if v.CanConvert(tt) {
			return v.Convert(tt).Interface().(T), true
		}
	}

	return empty, false
}
