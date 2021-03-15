package instrument

// InstrumentID is an identifier for an instrument/sample that means something to the format
type InstrumentID interface {
	IsEmpty() bool
}

// InstrumentDataIntf is the interface to implementation-specific functions on an instrument
type InstrumentDataIntf interface{}

// InstrumentKind defines the kind of instrument
type InstrumentKind int

const (
	// InstrumentKindPCM defines a PCM instrument
	InstrumentKindPCM = InstrumentKind(iota)
	// InstrumentKindOPL2 defines an OPL2 instrument
	InstrumentKindOPL2
)
