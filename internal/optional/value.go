package optional

// Value is an optional value
type Value[T any] struct {
	set   bool
	value T
}

// NewValue constructs a Value structure with a value already set into it
func NewValue[T any](value T) Value[T] {
	var v Value[T]
	v.Set(value)
	return v
}

// IsZero is used by the yaml marshaller to determine "zero"-ness for omitempty
// we're using it for the `set` bool
func (o Value[T]) IsZero() bool {
	return !o.set
}

// MarshalYAML outputs the value of the Value, if `set` is set.
// otherwise, it returns nil
func (o Value[T]) MarshalYAML() (T, error) {
	if o.set {
		return o.value, nil
	}
	var empty T
	return empty, nil
}

// UnmarshalYAML unmarshals a value out of yaml and safely into our struct
func (o *Value[T]) UnmarshalYAML(unmarshal func(any) error) error {
	var val T
	if err := unmarshal(&val); err != nil {
		return err
	}
	o.Set(val)
	return nil
}

// Reset clears the memory on the value
func (o *Value[T]) Reset() {
	var empty T
	o.value = empty
	o.set = false
}

// Set updates the value and sets the set flag
func (o *Value[T]) Set(value T) {
	o.value = value
	o.set = true
}

func (o Value[T]) IsSet() bool {
	return o.set
}

// Get returns the value and its set flag
func (o Value[T]) Get() (T, bool) {
	return o.value, o.set
}
