package pcm

// SampleDataFormat is the format of the sample data
type SampleDataFormat uint8

const (
	// SampleDataFormat8BitUnsigned is for unsigned 8-bit data
	SampleDataFormat8BitUnsigned = SampleDataFormat(iota)
	// SampleDataFormat8BitSigned is for signed 8-bit data
	SampleDataFormat8BitSigned
	// SampleDataFormat16BitLEUnsigned is for unsigned, little-endian, 16-bit data
	SampleDataFormat16BitLEUnsigned
	// SampleDataFormat16BitLESigned is for signed, little-endian, 16-bit data
	SampleDataFormat16BitLESigned
	// SampleDataFormat16BitBEUnsigned is for unsigned, big-endian, 16-bit data
	SampleDataFormat16BitBEUnsigned
	// SampleDataFormat16BitBESigned is for signed, big-endian, 16-bit data
	SampleDataFormat16BitBESigned
)
