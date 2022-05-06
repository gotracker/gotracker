package settings

import (
	"errors"

	"github.com/gotracker/gotracker/internal/optional"

	"github.com/gotracker/voice/pcm"
)

const (
	NamePreferredSampleFormat      string = "PreferredSampleFormat"
	NameUseNativeSampleFormat      string = "UseNativeSampleFormat"
	NameDisplayITLongChannelOutput string = "DisplayITLongChannelOutput"
)

type Settings struct {
	Values map[string]optional.Value[any]
}

// GetOption returns the current option by name
func (s *Settings) GetOption(name string) *optional.Value[any] {
	if s.Values == nil {
		return nil
	}
	if v, ok := s.Values[name]; ok {
		return &v
	}
	return nil
}

// Get returns the current value by name
func (s *Settings) Get(name string) (any, bool) {
	if v := s.GetOption(name); v != nil {
		return v.Get()
	}
	return nil, false
}

// Set sets a value by name
func (s *Settings) Set(name string, value any) error {
	if s.Values == nil {
		s.Values = make(map[string]optional.Value[any])
	}
	var v optional.Value[any]
	v.Set(value)
	s.Values[name] = v
	return nil
}

// OptionFunc is an option setting function
type OptionFunc func(s *Settings) error

// PreferredSampleFormat sets the preferred sample format for the samples in the song
// this will up- and down-convert the existing formats as necessary.
func PreferredSampleFormat(format pcm.SampleDataFormat) OptionFunc {
	return func(s *Settings) error {
		if s == nil {
			return errors.New("settings is nil")
		}
		return s.Set(NamePreferredSampleFormat, format)
	}
}

func UseNativeSampleFormat(enabled bool) OptionFunc {
	return func(s *Settings) error {
		if s == nil {
			return errors.New("settings is nil")
		}
		return s.Set(NameUseNativeSampleFormat, enabled)
	}
}
