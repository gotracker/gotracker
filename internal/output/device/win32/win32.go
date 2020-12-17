// +build windows

package win32

import (
	"errors"
	"os"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	user32Dll            = syscall.NewLazyDLL("user32.dll")
	getDesktopWindowProc = user32Dll.NewProc("GetDesktopWindow")

	kernel32Dll     = syscall.NewLazyDLL("kernel32.dll")
	createEventProc = kernel32Dll.NewProc("CreateEventA")
	closeHandleProc = kernel32Dll.NewProc("CloseHandle")
)

const (
	// WAVERR_STILLPLAYING defines a value equal to the WAVERR_STILLPLAYING error code from winmm api
	WAVERR_STILLPLAYING = uintptr(33)

	// WAVE_FORMAT_PCM specifies PCM wave format
	WAVE_FORMAT_PCM = uint16(0x0001)

	// DSSCL_PRIORITY allows for setting the wave format
	DSSCL_PRIORITY = uint32(0x00000002)

	// DSBPLAY_LOOPING loops a buffer until told not to
	DSBPLAY_LOOPING = uint32(0x00000001)

	WAVE_MAPPER = uint32(0xFFFFFFFF)
)

// DSBCAPS directsound buffer capabilities
type DSBCAPS uint32

const (
	// DSBCAPS_PRIMARYBUFFER primary buffer
	DSBCAPS_PRIMARYBUFFER = DSBCAPS(0x00000001)
	// DSBCAPS_STATIC static
	DSBCAPS_STATIC = DSBCAPS(0x00000002)
	// DSBCAPS_LOCHARDWARE loc hardware
	DSBCAPS_LOCHARDWARE = DSBCAPS(0x00000004)
	// DSBCAPS_LOCSOFTWARE loc software
	DSBCAPS_LOCSOFTWARE = DSBCAPS(0x00000008)
	// DSBCAPS_CTRLFREQUENCY control frequency
	DSBCAPS_CTRLFREQUENCY = DSBCAPS(0x00000020)
	// DSBCAPS_CTRLPAN control pan
	DSBCAPS_CTRLPAN = DSBCAPS(0x00000040)
	// DSBCAPS_CTRLVOLUME control volume
	DSBCAPS_CTRLVOLUME = DSBCAPS(0x00000080)
	// DSBCAPS_CTRLDEFAULT control pan + volume + frequency
	DSBCAPS_CTRLDEFAULT = DSBCAPS(0x000000E0)
	// DSBCAPS_CTRLPOSITIONNOTIFY control position notify
	DSBCAPS_CTRLPOSITIONNOTIFY = DSBCAPS(0x00000100)
	// DSBCAPS_CTRLALL control all capabilities
	DSBCAPS_CTRLALL = DSBCAPS(0x000001E0)
	// DSBCAPS_STICKYFOCUS sticky focus
	DSBCAPS_STICKYFOCUS = DSBCAPS(0x00004000)
	// DSBCAPS_GLOBALFOCUS global focus
	DSBCAPS_GLOBALFOCUS = DSBCAPS(0x00008000)
	// DSBCAPS_GETCURRENTPOSITION2 more accurate play cursor under emulation
	DSBCAPS_GETCURRENTPOSITION2 = DSBCAPS(0x00010000)
	// DSBCAPS_MUTE3DATMAXDISTANCE  mute 3d at max distance
	DSBCAPS_MUTE3DATMAXDISTANCE = DSBCAPS(0x00020000)
)

type DSBSTATUS uint32

const (
	// DSBSTATUS_PLAYING playing
	DSBSTATUS_PLAYING = DSBSTATUS(0x00000001)
	// DSBSTATUS_BUFFERLOST buffer lost
	DSBSTATUS_BUFFERLOST = DSBSTATUS(0x00000002)
	// DSBSTATUS_LOOPING looping
	DSBSTATUS_LOOPING = DSBSTATUS(0x00000004)
)

// HANDLE is a handle value
type HANDLE uintptr

// HWND is a window handle value
type HWND HANDLE

// HWAVEOUT is a handle for a WAVEOUT device
type HWAVEOUT HANDLE

// WAVEHDR is a structure containing the details about a circular buffer of wave data
type WAVEHDR struct {
	LpData          uintptr
	DwBufferLength  uint32
	DwBytesRecorded uint32
	DwUser          uintptr
	DwFlags         uint32
	DwLoops         uint32
	LpNext          uintptr
	Reserved        uintptr
}

// WAVEFORMATEX is a structure containing data about a wave format
type WAVEFORMATEX struct {
	WFormatTag      uint16
	NChannels       uint16
	NSamplesPerSec  uint32
	NAvgBytesPerSec uint32
	NBlockAlign     uint16
	WBitsPerSample  uint16
	CbSize          uint16
}

// GetDesktopWindow returns the handle of the desktop window
func GetDesktopWindow() HWND {
	result, _, _ := getDesktopWindowProc.Call()
	return HWND(result)
}

// WaitForSingleObjectInfinite will wait infinitely for a single handle value to become available
func WaitForSingleObjectInfinite(handle HANDLE) error {
	h := atomic.LoadUintptr((*uintptr)(&handle))
	s, e := syscall.WaitForSingleObject(syscall.Handle(h), syscall.INFINITE)
	switch s {
	case syscall.WAIT_OBJECT_0:
		break
	case syscall.WAIT_FAILED:
		return os.NewSyscallError("WaitForSingleObject", e)
	default:
		return errors.New("os: unexpected result from WaitForSingleObject")
	}
	return nil
}

// WaitForSingleObject will wait for a single handle value to become available up to a total of `duration` milliseconds
func WaitForSingleObject(handle HANDLE, duration time.Duration) error {
	h := atomic.LoadUintptr((*uintptr)(&handle))
	s, e := syscall.WaitForSingleObject(syscall.Handle(h), uint32(duration.Milliseconds()))
	switch s {
	case syscall.WAIT_OBJECT_0:
		break
	case syscall.WAIT_FAILED:
		return os.NewSyscallError("WaitForSingleObject", e)
	default:
		return errors.New("os: unexpected result from WaitForSingleObject")
	}
	return nil
}

// EventToChannel turns an event handle into a channel
func EventToChannel(event HANDLE) (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	done := make(chan struct{}, 1)
	go func() {
		defer close(ch)
		for {
			select {
			case <-done:
				return
			default:
			}
			if err := WaitForSingleObjectInfinite(event); err != nil {
				panic(err)
				return
			}
			ch <- struct{}{}
		}
	}()
	return ch, func() {
		done <- struct{}{}
		close(done)
	}
}

// CreateEvent creates a handle for event operations
func CreateEvent() (HANDLE, error) {
	retVal, _, _ := createEventProc.Call(0, 0, 0)
	if retVal == 0 {
		return HANDLE(0), errors.New("failed to create a new event")
	}

	return HANDLE(retVal), nil
}

// CloseHandle closes a handle
func CloseHandle(handle HANDLE) error {
	retVal, _, _ := closeHandleProc.Call(uintptr(handle))
	if retVal == 0 {
		return errors.New("failed to close handle")
	}

	return nil
}
