package common

// Settings is the settings for configuring an output device
type Settings struct {
	Name             string
	Channels         int
	SamplesPerSecond int
	BitsPerSample    int
	Filepath         string
	OnRowOutput      WrittenCallback
}
