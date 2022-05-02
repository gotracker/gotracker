package logging

import (
	"fmt"
)

type Squelchable struct {
	Squelch bool
}

func (s *Squelchable) Printf(format string, args ...any) {
	if s.Squelch {
		return
	}
	fmt.Printf(format, args...)
}

func (s *Squelchable) Println(args ...any) {
	if s.Squelch {
		return
	}
	fmt.Println(args...)
}

func (s *Squelchable) Print(args ...any) {
	if s.Squelch {
		return
	}
	fmt.Print(args...)
}
