package memfs

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FS struct {
	init  sync.Once
	files map[string]storedFile
}

type storedFile struct {
	data    []byte
	name    string
	modTime time.Time
}

func (m *FS) ensure() {
	m.init.Do(func() {
		m.files = make(map[string]storedFile)
	})
}

func (m *FS) CreateFile(name string, data []byte) error {
	m.ensure()

	d, n := filepath.Split(name)
	if d != "" && d != "/" {
		return &fs.PathError{
			Op:   "CreateFile",
			Path: name,
			Err:  os.ErrInvalid,
		}
	}

	blob := make([]byte, len(data))
	copy(blob, data)
	m.files[n] = storedFile{
		data:    blob,
		name:    n,
		modTime: time.Now(),
	}
	return nil
}

func (m *FS) Open(name string) (fs.File, error) {
	m.ensure()

	d, n := filepath.Split(name)
	if d != "" && d != "/" {
		return nil, &fs.PathError{
			Op:   "Open",
			Path: name,
			Err:  os.ErrInvalid,
		}
	}

	sf, found := m.files[n]
	if !found {
		return nil, &fs.PathError{
			Op:   "Open",
			Path: name,
			Err:  os.ErrNotExist,
		}
	}

	f := File{
		fileStat: fileStat{
			name:    sf.name,
			mode:    0777,
			modTime: sf.modTime,
		},
		r: bytes.NewReader(sf.data),
	}
	return &f, nil
}
