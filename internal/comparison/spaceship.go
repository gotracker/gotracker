package comparison

import (
	"fmt"
)

// Spaceship is a comparison result for a three-way comparison (<=>)
type Spaceship int

func (s Spaceship) String() string {
	return fmt.Sprintf("%d", int(s))
}

const (
	// SpaceshipRightGreater is returned when the right-hand side of a spaceship operator (<=>) is greater than the left-hand side
	SpaceshipRightGreater = Spaceship(-1)
	// SpaceshipEqual is returned when the right-hand side of a spaceship operator (<=>) is equal to the left-hand side
	SpaceshipEqual = Spaceship(0)
	// SpaceshipLeftGreater is returned when the left-hand side of a spaceship operator (<=>) is greater than the right-hand side
	SpaceshipLeftGreater = Spaceship(1)
)
