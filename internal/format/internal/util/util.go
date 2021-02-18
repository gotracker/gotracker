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
	if _, err := buffer.ReadFrom(file); err != nil {
		return nil, err
	}
	return buffer, nil
}
