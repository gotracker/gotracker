//go:build linux || pulseaudio
// +build linux pulseaudio

package pulseaudio

import (
	"bytes"
	"io"

	"github.com/jfreymuth/pulse"
	"github.com/jfreymuth/pulse/proto"
)

type Client struct {
	pc    *pulse.Client
	chmap proto.ChannelMap
	strm  *pulse.PlaybackStream
	ch    chan []byte
	r     bytes.Buffer
}

func New(appName string, sampleRate int, channels int, bitsPerSample int) (*Client, error) {
	pa := Client{}

	switch channels {
	case 1:
		pa.chmap = append(pa.chmap, proto.ChannelMono)
	case 2:
		pa.chmap = append(pa.chmap, proto.ChannelLeft, proto.ChannelRight)
	case 4:
		pa.chmap = append(pa.chmap, proto.ChannelFrontLeft, proto.ChannelFrontRight, proto.ChannelRearLeft, proto.ChannelRearRight)
	}

	var r pulse.Reader
	switch bitsPerSample {
	case 8:
		r = pulse.NewReader(&pa, proto.FormatUint8)
	case 16:
		r = pulse.NewReader(&pa, proto.FormatInt16LE)
	}

	c, err := pulse.NewClient(pulse.ClientApplicationName(appName))
	if err != nil {
		return nil, err
	}
	pa.pc = c

	pa.ch = make(chan []byte)

	strm, err := c.NewPlayback(r,
		pulse.PlaybackSampleRate(sampleRate),
		pulse.PlaybackLatency(0.1),
		pulse.PlaybackChannels(pa.chmap))
	if err != nil {
		c.Close()
		close(pa.ch)
		return nil, err
	}
	pa.strm = strm
	// we need to prime the buffer with empty data, otherwise it'll stall out
	pa.r.Write(make([]byte, pa.strm.BufferSizeBytes()))
	pa.strm.Start()

	return &pa, nil
}

func (pa *Client) Output(data []byte) {
	pa.ch <- data
}

func (pa *Client) Read(p []byte) (int, error) {
	needed := len(p)
	for {
		if pa.r.Len() >= needed {
			return pa.r.Read(p)
		}
		buf, ok := <-pa.ch
		if !ok {
			return 0, io.ErrClosedPipe
		}
		pa.r = *bytes.NewBuffer(pa.r.Bytes())
		pa.r.Write(buf)
	}
}

func (pa *Client) Close() error {
	if pa.strm != nil {
		pa.strm.Close()
	}
	if pa.pc != nil {
		pa.pc.Close()
	}
	close(pa.ch)
	return nil
}
