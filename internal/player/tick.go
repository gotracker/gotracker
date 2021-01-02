package player

// TickableIntf is an interface which exposes the OnTick call
type TickableIntf interface {
	OnTick() error
}

// DoTick calls the OnTick() function on the interface, if possible
func DoTick(t TickableIntf) error {
	if t != nil {
		return t.OnTick()
	}
	return nil
}
