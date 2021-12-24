package optional

import (
	"gotracker/internal/song/index"
	"gotracker/internal/song/note"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"
)

// Value is an optional value
type Value struct {
	set   bool
	value any
}

// NewValue constructs a Value structure with a value already set into it
func NewValue(value any) Value {
	var v Value
	v.Set(value)
	return v
}

// IsZero is used by the yaml marshaller to determine "zero"-ness for omitempty
// we're using it for the `set` bool
func (o Value) IsZero() bool {
	return !o.set
}

// MarshalYAML outputs the value of the Value, if `set` is set.
// otherwise, it returns nil
func (o Value) MarshalYAML() (any, error) {
	if o.set {
		return o.value, nil
	}
	return nil, nil
}

// UnmarshalYAML unmarshals a value out of yaml and safely into our struct
func (o *Value) UnmarshalYAML(unmarshal func(any) error) error {
	var val any
	if err := unmarshal(&val); err != nil {
		return err
	}
	o.Set(val)
	return nil
}

// Reset clears the memory on the value
func (o *Value) Reset() {
	o.value = nil
	o.set = false
}

// Set updates the value and sets the set flag
func (o *Value) Set(value any) {
	o.value = value
	o.set = true
}

func (o *Value) IsSet() bool {
	return o.set
}

// Get returns the value and its set flag
func (o *Value) Get() (any, bool) {
	return o.value, o.set
}

// GetBool returns the stored value as a boolean and if it has been set
func (o *Value) GetBool() (bool, bool) {
	if v, ok := o.value.(bool); ok {
		return v, o.set
	}
	return false, false
}

// GetInt returns the stored value as an integer and if it has been set
func (o *Value) GetInt() (int, bool) {
	if v, ok := o.value.(int); ok {
		return v, o.set
	}
	return 0, false
}

// GetVolume returns the stored value as a volume and if it has been set
func (o *Value) GetVolume() (volume.Volume, bool) {
	if v, ok := o.value.(volume.Volume); ok {
		return v, o.set
	}
	return volume.Volume(1), false
}

// GetPeriod returns the stored value as a period and if it has been set
func (o *Value) GetPeriod() (note.Period, bool) {
	if v, ok := o.value.(note.Period); ok {
		return v, o.set
	}
	return nil, false
}

// GetPeriodDelta returns the stored value as a period and if it has been set
func (o *Value) GetPeriodDelta() (note.PeriodDelta, bool) {
	if v, ok := o.value.(note.PeriodDelta); ok {
		return v, o.set
	}
	return note.PeriodDelta(0), false
}

// GetPanning returns the stored value as a panning position and if it has been set
func (o *Value) GetPanning() (panning.Position, bool) {
	if v, ok := o.value.(panning.Position); ok {
		return v, o.set
	}
	return panning.CenterAhead, false
}

// GetPosition returns the stored value as a sample position and if it has been set
func (o *Value) GetPosition() (sampling.Pos, bool) {
	if v, ok := o.value.(sampling.Pos); ok {
		return v, o.set
	}
	return sampling.Pos{}, false
}

// GetOrderIdx returns the stored value as an order index and if it has been set
func (o *Value) GetOrderIdx() (index.Order, bool) {
	if v, ok := o.value.(index.Order); ok {
		return v, o.set
	}
	return index.Order(0), false
}

// GetRowIdx returns the stored value as a row index and if it has been set
func (o *Value) GetRowIdx() (index.Row, bool) {
	if v, ok := o.value.(index.Row); ok {
		return v, o.set
	}
	return index.Row(0), false
}

// GetFinetune returns the stored value as a finetune value and if it has been set
func (o *Value) GetFinetune() (note.Finetune, bool) {
	if v, ok := o.value.(note.Finetune); ok {
		return v, o.set
	}
	return note.Finetune(0), false
}
