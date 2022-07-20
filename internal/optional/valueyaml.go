package optional

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
