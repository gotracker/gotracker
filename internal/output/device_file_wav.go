package output

import (
	"bufio"
	"gotracker/internal/player/render"
	"gotracker/internal/player/render/mixer"
	"os"

	"github.com/pkg/errors"
)

type fileDeviceWav struct {
	device
	mix mixer.Mixer

	f  *os.File
	w  *bufio.Writer
	sz uint32
}

const (
	fileChunkSizePos     = 4
	fileSubchunk2SizePos = 40
)

func newFileWavDevice(settings Settings) (Device, error) {
	fd := fileDeviceWav{
		mix: mixer.Mixer{
			Channels:      settings.Channels,
			BitsPerSample: settings.BitsPerSample,
		},
	}
	f, err := os.OpenFile(settings.Filepath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0)
	if err != nil {
		return nil, err
	}

	if f == nil {
		return nil, errors.New("unexpected file error")
	}

	w := bufio.NewWriter(f)
	// RIFF header
	w.Write([]byte{'R', 'I', 'F', 'F'}) // ChunkID
	w.Write([]byte{0, 0, 0, 0})         // ChunkSize
	w.Write([]byte{'W', 'A', 'V', 'E'}) // Format

	// fmt header
	w.Write([]byte{'f', 'm', 't', ' '})                                      // Subchunk1ID
	w.Write([]byte{16, 0, 0, 0})                                             // Subchunk1Size
	w.Write([]byte{1, 0})                                                    // AudioFormat (1 = PCM)
	w.Write([]byte{uint8(settings.Channels), uint8(settings.Channels >> 8)}) // NumChannels
	w.Write([]byte{uint8(settings.SamplesPerSecond), uint8(settings.SamplesPerSecond >> 8),
		uint8(settings.SamplesPerSecond >> 16), uint8(settings.SamplesPerSecond >> 24)}) // SampleRate
	byteRate := settings.SamplesPerSecond * settings.Channels * settings.BitsPerSample / 8
	w.Write([]byte{uint8(byteRate), uint8(byteRate >> 8), uint8(byteRate >> 16), uint8(byteRate >> 24)}) // ByteRate
	blockAlign := settings.Channels * settings.BitsPerSample / 8
	w.Write([]byte{uint8(blockAlign), uint8(blockAlign >> 8)})                         // BlockAlign
	w.Write([]byte{uint8(settings.BitsPerSample), uint8(settings.BitsPerSample >> 8)}) // BitsPerSample

	// data header
	w.Write([]byte{'d', 'a', 't', 'a'}) // Subchunk2ID
	w.Write([]byte{0, 0, 0, 0})         // Subchunk2Size

	fd.f = f
	fd.w = w

	return &fd, nil
}

// Play starts the wave output device playing
func (d *fileDeviceWav) Play(in <-chan render.RowRender) {
	panmixer := mixer.GetPanMixer(d.mix.Channels)
	for row := range in {
		data := d.mix.NewMixBuffer(row.SamplesLen)
		for _, rdata := range row.RenderData {
			pos := 0
			for _, cdata := range rdata {
				if cdata.Flush != nil {
					cdata.Flush()
				}
				if len(cdata.Data) > 0 {
					volMtx := cdata.Volume.Apply(panmixer.GetMixingMatrix(cdata.Pan)...)
					data.Add(pos, cdata.Data, volMtx)
				}
				pos += cdata.SamplesLen
			}
		}
		mixedData := data.ToRenderData(row.SamplesLen, d.mix.BitsPerSample, len(row.RenderData))
		d.w.Write(mixedData)
		d.sz += uint32(len(mixedData))
	}
}

// Close closes the wave output device
func (d *fileDeviceWav) Close() {
	d.w.Flush()
	d.w = nil
	d.f.Seek(fileChunkSizePos, 0)
	chunkSize := 36 + d.sz
	d.f.Write([]byte{uint8(chunkSize), uint8(chunkSize >> 8), uint8(chunkSize >> 16), uint8(chunkSize >> 24)}) // ChunkSize
	d.f.Seek(fileSubchunk2SizePos, 0)
	d.f.Write([]byte{uint8(d.sz), uint8(d.sz >> 8), uint8(d.sz >> 16), uint8(d.sz >> 24)}) // Subchunk2Size
	d.f.Close()
}

func init() {
	fileDeviceMap[".wav"] = newFileWavDevice
}
