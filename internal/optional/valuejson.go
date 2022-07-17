package optional

import "encoding/json"

// MarshalJSON outputs the value of the Value, if `set` is set.
// otherwise, it returns nil
func (o Value[T]) MarshalJSON() ([]byte, error) {
	if o.set {
		return json.Marshal(o.value)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals a value out of json and safely into our struct
func (o *Value[T]) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		o.Reset()
		return nil
	}
	var val T
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	o.Set(val)
	return nil
}
