package content

import (
	"fmt"
	"net/http"

	"github.com/gotracker/gomixing/sampling"
)

type AudioPCM struct {
	SampleRate int
	Channels   int
	Format     sampling.Format
}

func (a AudioPCM) WriteHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", a.getContentType())
}

func (a *AudioPCM) Write(w http.ResponseWriter, data []byte) (int, error) {
	return w.Write(data)
}

func (a AudioPCM) getContentType() string {
	switch a.Format {
	case sampling.Format8BitUnsigned:
		return fmt.Sprintf("audio/pcm;rate=%d;channels=%d;encoding=unsigned-int;bits=8", a.SampleRate, a.Channels)
	case sampling.Format8BitSigned:
		return fmt.Sprintf("audio/pcm;rate=%d;channels=%d;encoding=signed-int;bits=8", a.SampleRate, a.Channels)
	case sampling.Format16BitLEUnsigned:
		return fmt.Sprintf("audio/pcm;rate=%d;channels=%d;encoding=unsigned-int;bits=16;big-endian=false", a.SampleRate, a.Channels)
	case sampling.Format16BitLESigned:
		return fmt.Sprintf("audio/pcm;rate=%d;channels=%d;encoding=signed-int;bits=16;big-endian=false", a.SampleRate, a.Channels)
	case sampling.Format16BitBEUnsigned:
		return fmt.Sprintf("audio/pcm;rate=%d;channels=%d;encoding=unsigned-int;bits=16;big-endian=true", a.SampleRate, a.Channels)
	case sampling.Format16BitBESigned:
		return fmt.Sprintf("audio/pcm;rate=%d;channels=%d;encoding=signed-int;bits=16;big-endian=true", a.SampleRate, a.Channels)
	case sampling.Format32BitLEFloat:
		return fmt.Sprintf("audio/pcm;rate=%d;channels=%d;encoding=float;bits=32;big-endian=false", a.SampleRate, a.Channels)
	case sampling.Format32BitBEFloat:
		return fmt.Sprintf("audio/pcm;rate=%d;channels=%d;encoding=float;bits=32;big-endian=true", a.SampleRate, a.Channels)
	case sampling.Format64BitLEFloat:
		return fmt.Sprintf("audio/pcm;rate=%d;channels=%d;encoding=float;bits=64;big-endian=false", a.SampleRate, a.Channels)
	case sampling.Format64BitBEFloat:
		return fmt.Sprintf("audio/pcm;rate=%d;channels=%d;encoding=float;bits=64;big-endian=true", a.SampleRate, a.Channels)
	default:
		return fmt.Sprintf("audio/pcm;rate=%d;channels=%d", a.SampleRate, a.Channels)
	}
}
