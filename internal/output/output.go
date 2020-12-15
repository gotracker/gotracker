package output

type createOutputDeviceFunc func(settings Settings) (Device, error)

type outputDeviceKind int

const (
	outputDeviceKindNone = outputDeviceKind(iota)
	outputDeviceKindFile
	outputDeviceKindSoundCard
)

type outputDevicePriority int

// the further down the list, the higher the priority
const (
	outputDevicePriorityNone = outputDevicePriority(iota)
	outputDevicePriorityFile
	outputDevicePriorityPulseAudio
	outputDevicePriorityWinmm
	outputDevicePriorityDSound
)
