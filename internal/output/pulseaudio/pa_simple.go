// +build pulseaudio

package pulseaudio

// #cgo LDFLAGS: -lpulse-simple -lpulse
// #include <pulse/sample.h>
// #include <pulse/simple.h>
// #include <pulse/stream.h>
// #include <pulse/error.h>
// #include <stdlib.h>
import "C"

import (
	"fmt"
	"unsafe"
)

type Client struct {
	simple *C.pa_simple
}

func New(streamName string, sampleRate int, channels int, bitsPerSample int) (*Client, error) {
	var (
		server *C.char                 = nil
		name   *C.char                 = C.CString("gotracker")
		dev    *C.char                 = nil
		stream *C.char                 = C.CString(streamName)
		dir    C.pa_stream_direction_t = C.PA_STREAM_PLAYBACK
		cMap   *C.pa_channel_map       = nil
		cAttr  *C.pa_buffer_attr       = nil
		code   C.int
	)

	defer func() {
		C.free(unsafe.Pointer(server))
		C.free(unsafe.Pointer(name))
		C.free(unsafe.Pointer(dev))
		C.free(unsafe.Pointer(stream))
	}()

	css := C.pa_sample_spec{}
	css.rate = C.uint32_t(sampleRate)
	css.channels = C.uint8_t(channels)

	switch bitsPerSample {
	case 8:
		css.format = C.PA_SAMPLE_U8
	case 16:
		css.format = C.PA_SAMPLE_S16LE
	case 24:
		css.format = C.PA_SAMPLE_S24LE
	case 32:
		css.format = C.PA_SAMPLE_S32LE
	default:
		css.format = C.PA_SAMPLE_INVALID
	}

	pCss := (*C.pa_sample_spec)(unsafe.Pointer(&css))

	s := Client{}
	s.simple = C.pa_simple_new(server, name, dir, dev, stream, pCss, cMap, cAttr, &code)
	if code != 0 {
		return nil, fmt.Errorf("pulseaudio simple failure code: %v", C.GoString(C.pa_strerror(code)))
	}
	return &s, nil
}

func (p *Client) Output(data []byte) {
	var code C.int
	buf := C.CBytes(data)
	bufLen := C.size_t(len(data))
	C.pa_simple_write(p.simple, buf, bufLen, &code)
	C.free(buf)
}

func (p *Client) Close() {
	C.pa_simple_free(p.simple)
}
