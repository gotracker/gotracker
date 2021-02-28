package intf

// Memory is an interface for storing effect data on the channel state
type Memory interface {
	StartOrder()
	Retrigger()
}
