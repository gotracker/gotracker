package intf

type Pattern interface {
	GetRow(uint8) Row
	GetRows() []Row
}
