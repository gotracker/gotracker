package device

import (
	"bufio"
	"encoding/binary"
	"errors"
	"os"

	"gotracker/internal/audio/mixing"
)

type fileDeviceWav struct {
	device
	mix mixing.Mixer

	f  *os.File
	w  *bufio.Writer
	sz uint32
}

const (
	wavFileChunkSizePos     = 4
	wavFileSubchunk2SizePos = 40
)

func newFileWavDevice(settings Settings) (Device, error) {
	fd := fileDeviceWav{
		device: device{
			onRowOutput: settings.OnRowOutput,
		},
		mix: mixing.Mixer{
			Channels:      settings.Channels,
			BitsPerSample: settings.BitsPerSample,
		},
	}
	f, err := os.OpenFile(settings.Filepath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	if f == nil {
		return nil, errors.New("unexpected file error")
	}

	byteRate := settings.SamplesPerSecond * settings.Channels * settings.BitsPerSample / 8
	blockAlign := settings.Channels * settings.BitsPerSample / 8

	w := bufio.NewWriter(f)
	// RIFF header
	w.Write([]byte{'R', 'I', 'F', 'F'})             // ChunkID
	binary.Write(w, binary.LittleEndian, uint32(0)) // ChunkSize
	w.Write([]byte{'W', 'A', 'V', 'E'})             // Format

	// fmt header
	w.Write([]byte{'f', 'm', 't', ' '})              // Subchunk1ID
	binary.Write(w, binary.LittleEndian, uint32(16)) // Subchunk1Size
	// = win32.WAVEFORMATEX (before the CbSize)
	binary.Write(w, binary.LittleEndian, uint16(0x001))                     // AudioFormat // = win32.WAVE_FORMAT_PCM
	binary.Write(w, binary.LittleEndian, uint16(settings.Channels))         // NumChannels
	binary.Write(w, binary.LittleEndian, uint32(settings.SamplesPerSecond)) // SampleRate
	binary.Write(w, binary.LittleEndian, uint32(byteRate))                  // ByteRate
	binary.Write(w, binary.LittleEndian, uint16(blockAlign))                // BlockAlign
	binary.Write(w, binary.LittleEndian, uint16(settings.BitsPerSample))    // BitsPerSample

	// data header
	w.Write([]byte{'d', 'a', 't', 'a'})             // Subchunk2ID
	binary.Write(w, binary.LittleEndian, uint32(0)) // Subchunk2Size

	fd.f = f
	fd.w = w

	return &fd, nil
}

// Play starts the wave output device playing
func (d *fileDeviceWav) Play(in <-chan *PremixData) {
	panmixer := mixing.GetPanMixer(d.mix.Channels)
	for row := range in {
		mixedData := d.mix.Flatten(panmixer, row.SamplesLen, row.Data)
		d.w.Write(mixedData)
		d.sz += uint32(len(mixedData))
		if d.onRowOutput != nil {
			d.onRowOutput(KindFile, row)
		}
	}
}

// Close closes the wave output device
func (d *fileDeviceWav) Close() {
	d.w.Flush()
	chunkSize := 36 + d.sz
	d.f.Seek(wavFileChunkSizePos, 0)
	binary.Write(d.w, binary.LittleEndian, uint32(chunkSize)) // ChunkSize
	d.f.Seek(wavFileSubchunk2SizePos, 0)
	binary.Write(d.w, binary.LittleEndian, uint32(d.sz)) // Subchunk2Size
	d.w.Flush()
	d.w = nil
	d.f.Close()
}

func init() {
	fileDeviceMap[".wav"] = newFileWavDevice
}
