// +build windows,dsound

package dsound

import (
	"gotracker/internal/output/win32"
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
)

var (
	dsoundDll              = syscall.NewLazyDLL("dsound.dll")
	directSoundCreate8Proc = dsoundDll.NewProc("DirectSoundCreate8")
)

type directSoundVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	CreateSoundBuffer    uintptr
	GetCaps              uintptr
	DuplicateSoundBuffer uintptr
	SetCooperativeLevel  uintptr
	Compact              uintptr
	GetSpeakerConfig     uintptr
	SetSpeakerConfig     uintptr
	Initialize           uintptr
}

type DirectSound struct {
	vtbl *directSoundVtbl
}

func createDevice(deviceID *syscall.GUID) (*DirectSound, error) {
	var obj *DirectSound
	deviceIDPtr := uintptr(unsafe.Pointer(deviceID))
	objPtr := uintptr(unsafe.Pointer(&obj))
	retVal, _, _ := directSoundCreate8Proc.Call(deviceIDPtr, objPtr, 0)
	if retVal != 0 {
		return nil, errors.Errorf("DirectSoundCreate8 returned %0.8x", retVal)
	}

	hwnd := win32.GetDesktopWindow()

	if err := obj.setCooperativeLevel(hwnd, win32.DSSCL_PRIORITY); err != nil {
		obj.release()
		return nil, err
	}
	return obj, nil
}

// NewDSound returns a new DirectSound interface for the preferred device
func NewDSound(preferredDeviceName string) (*DirectSound, error) {
	var deviceID *syscall.GUID
	if preferredDeviceName != "" {
		// TODO: determine GUID for provided preferred device name here
		// preferredDeviceName = &syscall.GUID{ ... }
	}
	return createDevice(deviceID)
}

func (ds *DirectSound) addRef() error {
	retVal, _, _ := syscall.Syscall(ds.vtbl.AddRef, 1, uintptr(unsafe.Pointer(ds)), 0, 0)
	if retVal != 0 {
		return errors.Errorf("DirectSound.AddRef returned %0.8x", retVal)
	}
	return nil
}

func (ds *DirectSound) release() error {
	retVal, _, _ := syscall.Syscall(ds.vtbl.Release, 1, uintptr(unsafe.Pointer(ds)), 0, 0)
	if retVal != 0 {
		return errors.Errorf("DirectSound.Release returned %0.8x", retVal)
	}
	return nil
}

func (ds *DirectSound) setCooperativeLevel(hwnd win32.HWND, level uint32) error {
	retVal, _, _ := syscall.Syscall(ds.vtbl.SetCooperativeLevel, 3, uintptr(unsafe.Pointer(ds)), uintptr(hwnd), uintptr(level))
	if retVal != 0 {
		return errors.Errorf("DirectSound.Release returned %0.8x", retVal)
	}
	return nil
}

type dsBufferDesc struct {
	Size        uint32
	Flags       win32.DSBCAPS
	BufferBytes uint32
	Reserved    uint32
	WfxFormat   *win32.WAVEFORMATEX
}

// CreateSoundBufferPrimary creates a primary sound buffer
func (ds *DirectSound) CreateSoundBufferPrimary(channels int, samplesPerSec int, bitsPerSample int) (*Buffer, *win32.WAVEFORMATEX, error) {
	bd := dsBufferDesc{
		Flags: win32.DSBCAPS_PRIMARYBUFFER,
	}
	bd.Size = uint32(unsafe.Sizeof(bd))

	var buffer *Buffer
	retVal, _, _ := syscall.Syscall6(ds.vtbl.CreateSoundBuffer, 4, uintptr(unsafe.Pointer(ds)), uintptr(unsafe.Pointer(&bd)), uintptr(unsafe.Pointer(&buffer)), 0, 0, 0)
	if retVal != 0 {
		return nil, nil, errors.Errorf("DirectSound.CreateSoundBuffer returned %0.8x", retVal)
	}

	wfx := win32.WAVEFORMATEX{
		WFormatTag:     win32.WAVE_FORMAT_PCM,
		NChannels:      uint16(channels),
		NSamplesPerSec: uint32(samplesPerSec),
		WBitsPerSample: uint16(bitsPerSample),
	}
	wfx.CbSize = uint16(unsafe.Sizeof(wfx))
	wfx.NBlockAlign = uint16(channels * bitsPerSample / 8)
	wfx.NAvgBytesPerSec = wfx.NSamplesPerSec * uint32(wfx.NBlockAlign)

	if err := buffer.setFormat(wfx); err != nil {
		buffer.Release()
		return nil, nil, err
	}

	return buffer, &wfx, nil
}

// CreateSoundBufferSecondary creates a secondary sound buffer
func (ds *DirectSound) CreateSoundBufferSecondary(wfx *win32.WAVEFORMATEX, bufferSize int) (*Buffer, error) {
	bd := dsBufferDesc{
		Flags:       win32.DSBCAPS_GETCURRENTPOSITION2 | win32.DSBCAPS_GLOBALFOCUS | win32.DSBCAPS_CTRLPOSITIONNOTIFY,
		BufferBytes: uint32(bufferSize),
		WfxFormat:   wfx,
	}
	bd.Size = uint32(unsafe.Sizeof(bd))

	var buffer *Buffer
	retVal, _, _ := syscall.Syscall6(ds.vtbl.CreateSoundBuffer, 4, uintptr(unsafe.Pointer(ds)), uintptr(unsafe.Pointer(&bd)), uintptr(unsafe.Pointer(&buffer)), 0, 0, 0)
	if retVal != 0 {
		return nil, errors.Errorf("DirectSound.CreateSoundBuffer returned %0.8x", retVal)
	}

	return buffer, nil
}

// Close cleans up the DirectSound device
func (ds *DirectSound) Close() error {
	return ds.release()
}
