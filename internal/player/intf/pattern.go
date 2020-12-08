package intf

// Pattern is an interface for pattern data
type Pattern interface {
	GetRow(uint8) Row
	GetRows() []Row
}
