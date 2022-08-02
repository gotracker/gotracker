package content

import (
	"encoding/binary"
	"net/http"
	"sync"

	"github.com/gotracker/gomixing/sampling"
)

type AudioWav struct {
	SampleRate int
	Channels   int
	Format     sampling.Format

	once sync.Once
}

func (a AudioWav) WriteHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "audio/wav")
}

func (a *AudioWav) Write(w http.ResponseWriter, data []byte) (int, error) {
	var (
		n   int
		err error
	)
	a.once.Do(func() {
		header := a.generateHeader()
		n, err = w.Write(header)
	})
	if err != nil {
		return n, err
	}

	return w.Write(data)
}

func (a AudioWav) generateHeader() []byte {
	var header = [...]byte{
		0x52, 0x49, 0x46, 0x46, //  0: ChunkID
		0xFF, 0xFF, 0xFF, 0xFF, //  4: ChunkSize
		0x57, 0x41, 0x56, 0x45, //  8: Format
		0x66, 0x6d, 0x74, 0x20, // 12: Subchunk1ID
		0x10, 0x00, 0x00, 0x00, // 16: Subchunk1Size
		0x00, 0x00, // 20: AudioFormat
		0x00, 0x00, // 22: NumChannels
		0x00, 0x00, 0x00, 0x00, // 24: SampleRate
		0x00, 0x00, 0x00, 0x00, // 28: ByteRate
		0x00, 0x00, // 32: BlockAlign
		0x00, 0x00, // 34: BitsPerSample
		0x64, 0x61, 0x74, 0x61, // 36: Subchunk2ID
		0xFF, 0xFF, 0xFF, 0xFF, // 40: Subchunk2Size
	}

	var (
		audioFormat   int
		bitsPerSample int
	)

	switch a.Format {
	case sampling.Format8BitUnsigned:
		audioFormat = 1
		bitsPerSample = 8
	case sampling.Format8BitSigned:
		audioFormat = 1
		bitsPerSample = 8
	case sampling.Format16BitLEUnsigned:
		audioFormat = 1
		bitsPerSample = 16
	case sampling.Format16BitLESigned:
		audioFormat = 1
		bitsPerSample = 16
	case sampling.Format16BitBEUnsigned:
		audioFormat = 1
		bitsPerSample = 16
	case sampling.Format16BitBESigned:
		audioFormat = 1
		bitsPerSample = 16
	case sampling.Format32BitLEFloat:
		audioFormat = 3
		bitsPerSample = 32
	case sampling.Format32BitBEFloat:
		audioFormat = 3
		bitsPerSample = 32
	case sampling.Format64BitLEFloat:
		audioFormat = 3
		bitsPerSample = 64
	case sampling.Format64BitBEFloat:
		audioFormat = 3
		bitsPerSample = 64
	default:
		audioFormat = 1
		bitsPerSample = 8
	}

	byteRate := (a.SampleRate * a.Channels * bitsPerSample) >> 3
	blockAlign := (a.Channels * bitsPerSample) >> 3

	// AudioFormat
	binary.LittleEndian.PutUint16(header[20:], uint16(audioFormat))

	// NumChannels
	binary.LittleEndian.PutUint16(header[22:], uint16(a.Channels))

	// SampleRate
	binary.LittleEndian.PutUint32(header[24:], uint32(a.SampleRate))

	// ByteRate
	binary.LittleEndian.PutUint32(header[28:], uint32(byteRate))

	// BlockAlign
	binary.LittleEndian.PutUint16(header[32:], uint16(blockAlign))

	// BitsPerSample
	binary.LittleEndian.PutUint16(header[34:], uint16(bitsPerSample))

	return header[:]
}
