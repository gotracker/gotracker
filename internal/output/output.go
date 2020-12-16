package output

type createOutputDeviceFunc func(settings Settings) (Device, error)

// DeviceKind is an enumeration of the device type
type DeviceKind int

const (
	// DeviceKindNone is nothing!
	DeviceKindNone = DeviceKind(iota)
	// DeviceKindFile is a file device type
	DeviceKindFile
	// DeviceKindSoundCard is an active sound playback device (e.g.: a sound card attached to speakers)
	DeviceKindSoundCard
)

type devicePriority int

// the further down the list, the higher the priority
const (
	devicePriorityNone = devicePriority(iota)
	devicePriorityFile
	devicePriorityPulseAudio
	devicePriorityWinmm
	devicePriorityDirectSound
)
