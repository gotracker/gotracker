package optional_test

import (
	"testing"

	"github.com/gotracker/gotracker/internal/optional"
	"golang.org/x/exp/constraints"
)

func expect[TValue constraints.Ordered | constraints.Complex | ~bool](t *testing.T, valueName string, expected, encountered TValue) {
	t.Helper()
	if expected != encountered {
		t.Errorf("expected %s to be %v, but encountered %v", valueName, expected, encountered)
	}
}

func TestValue(t *testing.T) {
	t.Run("SetInt", func(t *testing.T) {
		var target optional.Value[int]
		expectedValue := 5
		expectedSet := true
		target.Set(expectedValue)
		encounteredValue, encounteredSet := target.Get()
		expect(t, "set", expectedSet, encounteredSet)
		expect(t, "value", expectedValue, encounteredValue)
	})
	t.Run("SetBool", func(t *testing.T) {
		var target optional.Value[bool]
		expectedValue := true
		expectedSet := true
		target.Set(expectedValue)
		encounteredValue, encounteredSet := target.Get()
		expect(t, "set", expectedSet, encounteredSet)
		expect(t, "value", expectedValue, encounteredValue)
	})
	t.Run("SetString", func(t *testing.T) {
		var target optional.Value[string]
		expectedValue := "Foo"
		expectedSet := true
		target.Set(expectedValue)
		encounteredValue, encounteredSet := target.Get()
		expect(t, "set", expectedSet, encounteredSet)
		expect(t, "value", expectedValue, encounteredValue)
	})
	t.Run("SetComplex128", func(t *testing.T) {
		var target optional.Value[complex128]
		expectedValue := complex(2.71828, 3.14159)
		expectedSet := true
		target.Set(expectedValue)
		encounteredValue, encounteredSet := target.Get()
		expect(t, "set", expectedSet, encounteredSet)
		expect(t, "value", expectedValue, encounteredValue)
	})
	t.Run("SetStruct", func(t *testing.T) {
		type TestStruct struct {
			ValInt     int
			ValString  string
			ValBool    bool
			ValComplex complex128
		}
		var target optional.Value[TestStruct]
		expectedValue := TestStruct{
			ValInt:     5,
			ValString:  "Foo",
			ValBool:    true,
			ValComplex: complex(2.71828, 3.14159),
		}
		expectedSet := true
		target.Set(expectedValue)
		encounteredValue, encounteredSet := target.Get()
		expect(t, "set", expectedSet, encounteredSet)
		expect(t, "value.ValInt", expectedValue.ValInt, encounteredValue.ValInt)
		expect(t, "value.ValString", expectedValue.ValString, encounteredValue.ValString)
		expect(t, "value.ValBool", expectedValue.ValBool, encounteredValue.ValBool)
		expect(t, "value.ValComplex", expectedValue.ValComplex, encounteredValue.ValComplex)
	})
}
