package files

import (
	"io/fs"

	"github.com/gotracker/gotracker/internal/memfs"
)

var filesystem memfs.FS

func GetFS() fs.FS {
	return &filesystem
}

func AddFile(filename string, data []byte) error {
	return filesystem.CreateFile(filename, data)
}
