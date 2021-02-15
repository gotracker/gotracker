package voice

// FilterEnveloper is a filter envelope interface
type FilterEnveloper interface {
	EnableFilterEnvelope(enabled bool)
	IsFilterEnvelopeEnabled() bool
	GetCurrentFilterEnvelope() float32
	SetFilterEnvelopePosition(pos int)
}
