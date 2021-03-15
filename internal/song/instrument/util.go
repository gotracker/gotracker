package instrument

// ID is an identifier for an instrument/sample that means something to the format
type ID interface {
	IsEmpty() bool
}

// DataIntf is the interface to implementation-specific functions on an instrument
type DataIntf interface{}

// Kind defines the kind of instrument
type Kind int

const (
	// KindPCM defines a PCM instrument
	KindPCM = Kind(iota)
	// KindOPL2 defines an OPL2 instrument
	KindOPL2
)
