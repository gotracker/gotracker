package voice

// Controller is the instrument actuation control interface
type Controller interface {
	Attack()
	Release()
	Fadeout()
	IsKeyOn() bool
	IsFadeout() bool
}
