package util

import (
	"bytes"
	"os"
)

// ReadFile will open a file for reading, then return a bytestream reader for it
func ReadFile(filename string) (*bytes.Buffer, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	buffer := &bytes.Buffer{}
	buffer.ReadFrom(file)
	return buffer, nil
}

var (
	protrackerSineTable = [32]uint8{
		0, 24, 49, 74, 97, 120, 141, 161, 180, 197, 212, 224, 235, 244, 250, 253,
		255, 253, 250, 244, 235, 224, 212, 197, 180, 161, 141, 120, 97, 74, 49, 24,
	}
)

// GetProtrackerSine returns the sine value for a particular position using the
// Protracker-compliant half-period sine table
func GetProtrackerSine(pos int) float32 {
	sin := float32(protrackerSineTable[pos&0x1f]) / 255
	if pos > 32 {
		return -sin
	}
	return sin
}
