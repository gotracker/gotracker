// +build windows

package win32

import "syscall"

var (
	user32Dll            = syscall.NewLazyDLL("user32.dll")
	getDesktopWindowProc = user32Dll.NewProc("GetDesktopWindow")
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
