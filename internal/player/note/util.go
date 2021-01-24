package note

import "fmt"

// C2SPD defines the C-2 (or in some players cases C-4) note sampling rate
type C2SPD uint16

// Finetune is a 1/64th of a Semitone
type Finetune int16

// SpaceshipResult is a comparison result for a three-way comparison (<=>)
type SpaceshipResult int

func (sr SpaceshipResult) String() string {
	return fmt.Sprintf("%d", int(sr))
}

const (
	// CompareRightHigher is returned when the right-hand side of a spaceship operator (<=>) is higher than the left-hand side
	CompareRightHigher = SpaceshipResult(-1)
	// CompareEqual is returned when the right-hand side of a spaceship operator (<=>) is equal to the left-hand side
	CompareEqual = SpaceshipResult(0)
	// CompareLeftHigher is returned when the left-hand side of a spaceship operator (<=>) is higher than the right-hand side
	CompareLeftHigher = SpaceshipResult(1)
)
