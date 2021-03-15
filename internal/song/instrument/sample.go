package instrument

import (
	"github.com/gotracker/gomixing/volume"
	"github.com/gotracker/voice/pcm"

	"gotracker/internal/format/settings"
)

func NewSample(data []byte, length int, channels int, format pcm.SampleDataFormat, s *settings.Settings) (pcm.Sample, error) {
	sf := format
	if v, ok := s.Get(settings.NamePreferredSampleFormat); ok {
		if val, ok := v.(pcm.SampleDataFormat); ok {
			sf = val
		}
	}

	var sample pcm.Sample
	if sf == format {
		sample = pcm.NewSample(data, length, channels, format)
	} else {
		inSample := pcm.NewSample(data, length, channels, format)
		outSample, err := pcm.ConvertTo(inSample, sf)
		if err != nil {
			return nil, err
		}
		sample = outSample
	}

	if v, ok := s.Get(settings.NameUseNativeSampleFormat); ok {
		if val, ok := v.(bool); ok && val {
			inSample := sample
			in := make(volume.Matrix, channels)
			var nativeData []volume.Volume
			for i := 0; i < length; i++ {
				d, err := inSample.Read(in)
				if err != nil {
					return nil, err
				}
				nativeData = append(nativeData, d...)
			}
			sample = pcm.NewSampleNative(nativeData, length, channels)
		}
	}

	return sample, nil
}
