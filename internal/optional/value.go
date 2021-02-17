package optional

// Value is an optional value
type Value struct {
	set   bool
	value interface{}
}

// Set updates the value and sets the set flag
func (o *Value) Set(value interface{}) {
	o.value = value
	o.set = true
}

// Get returns the value and its set flag
func (o *Value) Get() (interface{}, bool) {
	return o.value, o.set
}
