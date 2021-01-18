package note

// InstrumentKind defines the kind of instrument
type InstrumentKind int

const (
	// InstrumentKindPCM defines a PCM instrument
	InstrumentKindPCM = InstrumentKind(iota)
	// InstrumentKindOPL2 defines an OPL2 instrument
	InstrumentKindOPL2
)
