package memfs

import (
	"bytes"
	"io"
	"io/fs"
)

type File struct {
	fileStat
	r *bytes.Reader
}

func (f File) Stat() (fs.FileInfo, error) {
	return &f, nil
}

func (f *File) Read(out []byte) (int, error) {
	return f.r.Read(out)
}

func (f *File) Close() error {
	return io.NopCloser(f.r).Close()
}

func (f File) Size() int64 {
	// length in bytes for regular files; system-dependent for others
	return f.r.Size()
}
