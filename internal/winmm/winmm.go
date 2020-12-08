// +build windows

package winmm

// #cgo LDFLAGS: -lwinmm
// #include <Windows.h>
// #include <Mmreg.h>
import "C"

import (
	"log"
	"unsafe"
)

// Wave is the holder for a single buffer that is heading out to the wave mapper device
type Wave struct {
	hdr    C.WAVEHDR
	lpData unsafe.Pointer
}

// Device is the holder for the wave mapper device
type Device struct {
	handle *C.HWAVEOUT
}

// WaveOutOpen starts up a wave mapper device
func WaveOutOpen(channels int, samplesPerSec int, bitsPerSample int) *Device {
	phwo := Device{}
	phwo.handle = (*C.HWAVEOUT)(C.malloc(512))
	uDeviceID := C.WAVE_MAPPER
	pwfx := C.WAVEFORMATEX{}
	pwfx.wFormatTag = C.WAVE_FORMAT_PCM
	pwfx.nChannels = C.WORD(channels)
	pwfx.nSamplesPerSec = C.DWORD(samplesPerSec)
	pwfx.wBitsPerSample = C.WORD(bitsPerSample)
	pwfx.nBlockAlign = C.WORD(pwfx.nChannels * pwfx.wBitsPerSample / 8)
	pwfx.nAvgBytesPerSec = pwfx.nSamplesPerSec * C.DWORD(pwfx.nBlockAlign)
	pwfx.cbSize = C.WORD(unsafe.Sizeof(pwfx))
	dwCallback := C.DWORD_PTR(0)
	dwCallbackInstance := C.DWORD_PTR(0)
	fdwOpen := C.DWORD(C.CALLBACK_NULL)
	result := C.waveOutOpen(phwo.handle, C.UINT(uDeviceID), &pwfx, dwCallback, dwCallbackInstance, fdwOpen)
	if result != C.MMSYSERR_NOERROR {
		log.Panicf("WinMM Err: %d", result)
		return nil
	}
	return &phwo
}

// WaveOutWrite writes data out to the wave mapper device
func WaveOutWrite(hwo Device, data []byte) *Wave {
	wave := Wave{}
	wave.lpData = C.CBytes(data)
	wave.hdr.lpData = C.LPSTR(wave.lpData)
	wave.hdr.dwBufferLength = C.DWORD(len(data))
	wave.hdr.dwBytesRecorded = C.DWORD(0)
	wave.hdr.dwUser = C.DWORD_PTR(0)
	wave.hdr.dwFlags = C.DWORD(0)
	wave.hdr.dwLoops = C.DWORD(0)
	wave.hdr.lpNext = nil
	wave.hdr.reserved = C.DWORD_PTR(0)
	szHdr := C.UINT(unsafe.Sizeof(wave.hdr))
	C.waveOutPrepareHeader(*hwo.handle, &wave.hdr, szHdr)
	C.waveOutWrite(*hwo.handle, &wave.hdr, szHdr)
	return &wave
}

// WaveOutFinished disposes of a wave output buffer
func WaveOutFinished(hwo Device, wave *Wave) bool {
	szHdr := C.UINT(unsafe.Sizeof(wave.hdr))
	result := C.waveOutUnprepareHeader(*hwo.handle, &wave.hdr, szHdr)
	//C.free(wave.lpData)
	return result == C.MMSYSERR_NOERROR
}

// WaveOutClose disposes of a wave mapper device
func WaveOutClose(hwo Device) {
	C.waveOutClose(*hwo.handle)
}
