// +build windows,directsound

package directsound

import (
	"syscall"
	"unsafe"

	"github.com/pkg/errors"

	"gotracker/internal/output/device/win32"
)

var (
	errDirectSoundNotify = errors.Wrap(ErrDirectSound, "DirectSoundNotify")
)

type notifyVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	SetNotificationPositions uintptr
}

type Notify struct {
	vtbl *notifyVtbl
}

type PositionNotify struct {
	Offset      uint32
	EventNotify win32.HANDLE
}

// AddRef references a Notify
func (n *Notify) AddRef() error {
	retVal, _, _ := syscall.Syscall(n.vtbl.AddRef, 1, uintptr(unsafe.Pointer(n)), 0, 0)
	if retVal != 0 {
		return errors.Wrapf(errDirectSoundNotify, "AddRef returned %0.8x", retVal)
	}
	return nil
}

// Release releases a Notify
func (n *Notify) Release() error {
	retVal, _, _ := syscall.Syscall(n.vtbl.Release, 1, uintptr(unsafe.Pointer(n)), 0, 0)
	if retVal != 0 {
		return errors.Wrapf(errDirectSoundNotify, "Release returned %0.8x", retVal)
	}
	return nil
}

// SetNotificationPositions sets up events for notification based on position
func (n *Notify) SetNotificationPositions(events []PositionNotify) error {
	retVal, _, _ := syscall.Syscall(n.vtbl.SetNotificationPositions, 3, uintptr(unsafe.Pointer(n)), uintptr(len(events)), uintptr(unsafe.Pointer(&events[0])))
	if retVal != 0 {
		return errors.Wrapf(errDirectSoundNotify, "SetNotificationPositions returned %0.8x", retVal)
	}

	return nil
}
