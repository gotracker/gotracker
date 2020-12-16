// +build windows,directsound

package directsound

import (
	"gotracker/internal/output/win32"
	"reflect"
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
)

type directSoundBufferVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	GetCaps            uintptr
	GetCurrentPosition uintptr
	GetFormat          uintptr
	GetVolume          uintptr
	GetPan             uintptr
	GetFrequency       uintptr
	GetStatus          uintptr
	Initialize         uintptr
	Lock               uintptr
	Play               uintptr
	SetCurrentPosition uintptr
	SetFormat          uintptr
	SetVolume          uintptr
	SetPan             uintptr
	SetFrequency       uintptr
	Stop               uintptr
	Unlock             uintptr
	Restore            uintptr
}

type Buffer struct {
	vtbl *directSoundBufferVtbl
}

// AddRef references a buffer
func (b *Buffer) AddRef() error {
	retVal, _, _ := syscall.Syscall(b.vtbl.AddRef, 1, uintptr(unsafe.Pointer(b)), 0, 0)
	if retVal != 0 {
		return errors.Errorf("DirectSoundBuffer.AddRef returned %0.8x", retVal)
	}
	return nil
}

// Release releases a buffer
func (b *Buffer) Release() error {
	retVal, _, _ := syscall.Syscall(b.vtbl.Release, 1, uintptr(unsafe.Pointer(b)), 0, 0)
	if retVal != 0 {
		return errors.Errorf("DirectSoundBuffer.Release returned %0.8x", retVal)
	}
	return nil
}

// GetNotify returns the notification interface
func (b *Buffer) GetNotify() (*Notify, error) {
	guidIDirectSoundNotify := syscall.GUID{0xb0210783, 0x89cd, 0x11d0, [...]byte{0xaf, 0x8, 0x0, 0xa0, 0xc9, 0x25, 0xcd, 0x16}}
	var notify *Notify
	retVal, _, _ := syscall.Syscall(b.vtbl.QueryInterface, 3, uintptr(unsafe.Pointer(b)), uintptr(unsafe.Pointer(&guidIDirectSoundNotify)), uintptr(unsafe.Pointer(&notify)))
	if retVal != 0 {
		return nil, errors.Errorf("DirectSoundBuffer.QueryInterface returned %0.8x", retVal)
	}

	return notify, nil
}

func (b *Buffer) setFormat(wfx win32.WAVEFORMATEX) error {
	retVal, _, _ := syscall.Syscall(b.vtbl.SetFormat, 2, uintptr(unsafe.Pointer(b)), uintptr(unsafe.Pointer(&wfx)), 0)
	if retVal != 0 {
		return errors.Errorf("DirectSoundBuffer.SetFormat returned %0.8x", retVal)
	}
	return nil
}

// Play sets a buffer into playing mode
func (b *Buffer) Play(looping bool) error {
	var flags uint32
	if looping {
		flags = flags | win32.DSBPLAY_LOOPING
	}
	retVal, _, _ := syscall.Syscall6(b.vtbl.Play, 4, uintptr(unsafe.Pointer(b)), 0, 0, uintptr(flags), 0, 0)
	if retVal != 0 {
		return errors.Errorf("DirectSoundBuffer.Play returned %0.8x", retVal)
	}
	return nil
}

// GetCurrentPosition returns the current play and write position cursors
func (b *Buffer) GetCurrentPosition() (uint32, uint32, error) {
	var currentPlayCursor uint32
	var currentWriteCursor uint32
	retVal, _, _ := syscall.Syscall(b.vtbl.GetCurrentPosition, 3, uintptr(unsafe.Pointer(b)), uintptr(unsafe.Pointer(&currentPlayCursor)), uintptr(unsafe.Pointer(&currentWriteCursor)))
	if retVal != 0 {
		return 0, 0, errors.Errorf("DirectSoundBuffer.GetCurrentPosition returned %0.8x", retVal)
	}
	return currentPlayCursor, currentWriteCursor, nil
}

// GetStatus returns the status of the buffer
func (b *Buffer) GetStatus() (win32.DSBSTATUS, error) {
	var status win32.DSBSTATUS
	retVal, _, _ := syscall.Syscall(b.vtbl.GetStatus, 2, uintptr(unsafe.Pointer(b)), uintptr(unsafe.Pointer(&status)), 0)
	if retVal != 0 {
		return 0, errors.Errorf("DirectSoundBuffer.GetStatus returned %0.8x", retVal)
	}
	return status, nil
}

// Lock locks the buffer for writing
func (b *Buffer) Lock(offset int, numBytes int) ([][]byte, error) {
	var flags uint32
	segments := make([][]byte, 2)
	segs := []*reflect.SliceHeader{
		(*reflect.SliceHeader)(unsafe.Pointer(&segments[0])),
		(*reflect.SliceHeader)(unsafe.Pointer(&segments[1])),
	}
	retVal, _, _ := syscall.Syscall9(b.vtbl.Lock, 8, uintptr(unsafe.Pointer(b)), uintptr(offset), uintptr(numBytes),
		uintptr(unsafe.Pointer(&segs[0].Data)), uintptr(unsafe.Pointer(&segs[0].Len)),
		uintptr(unsafe.Pointer(&segs[1].Data)), uintptr(unsafe.Pointer(&segs[1].Len)),
		uintptr(flags), 0)
	if retVal != 0 {
		return nil, errors.Errorf("DirectSoundBuffer.Lock returned %0.8x", retVal)
	}
	for i, _ := range segs {
		segs[i].Cap = segs[i].Len
	}
	return segments, nil
}

func (b *Buffer) Unlock(segments [][]byte) error {
	segs := []*reflect.SliceHeader{
		(*reflect.SliceHeader)(unsafe.Pointer(&segments[0])),
		(*reflect.SliceHeader)(unsafe.Pointer(&segments[1])),
	}
	retVal, _, _ := syscall.Syscall6(b.vtbl.Unlock, 5, uintptr(unsafe.Pointer(b)),
		segs[0].Data, uintptr(segs[0].Len),
		segs[1].Data, uintptr(segs[1].Len),
		0)
	if retVal != 0 {
		return errors.Errorf("DirectSoundBuffer.Unlock returned %0.8x", retVal)
	}
	return nil
}
