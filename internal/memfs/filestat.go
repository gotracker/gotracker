package memfs

import (
	"io/fs"
	"time"
)

type fileStat struct {
	name    string
	mode    fs.FileMode
	modTime time.Time
}

func (s fileStat) Name() string {
	return s.name
}

func (s fileStat) Mode() fs.FileMode {
	return s.mode
}

func (s fileStat) ModTime() time.Time {
	return s.modTime
}

func (s fileStat) IsDir() bool {
	return false
}

func (s fileStat) Sys() any {
	return nil
}
