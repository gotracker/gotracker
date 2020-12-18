package s3mfile

// PanningFlags is a flagset and panning value for the panning system
type PanningFlags uint8

const (
	// PanningFlagValid is the flag used to determine that the panning value is valid
	PanningFlagValid = PanningFlags(0x20)

	// DefaultPanningLeft is the default panning value for left
	DefaultPanningLeft = PanningFlagValid | PanningFlags(0x03)
	// DefaultPanningRight is the default panning value for right
	DefaultPanningRight = PanningFlagValid | PanningFlags(0x0C)
)

// IsValid returns true if bit 5 is set
func (pf PanningFlags) IsValid() bool {
	return uint8(pf&PanningFlagValid) != 0
}

// Value returns the panning position (0=full left, 15=full right)
func (pf PanningFlags) Value() uint8 {
	return uint8(pf) & 0x0F
}
