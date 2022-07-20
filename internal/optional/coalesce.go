package optional

// Coalesce will return the first optional value that is set
// (or returns an unset optional value if none is found).
func Coalesce[T any](options ...Value[T]) Value[T] {
	for _, option := range options {
		if option.IsSet() {
			return option
		}
	}

	return Value[T]{}
}

// Coalesce will return the first optional value that is set and not "zero"
// (or returns an unset optional value if none is found).
func CoalesceZero[T any](options ...Value[T]) Value[T] {
	for _, option := range options {
		if !option.IsZero() {
			return option
		}
	}

	return Value[T]{}
}
