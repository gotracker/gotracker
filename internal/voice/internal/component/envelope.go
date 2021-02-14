package component

// Envelope is an envelope component interface
type Envelope interface {
	//Reset(env *envelope.Envelope)
	SetEnabled(enabled bool)
	IsEnabled() bool
	Advance(keyOn bool, prevKeyOn bool)
}
