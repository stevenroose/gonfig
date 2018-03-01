// Copyright (c) 2017 Steven Roose <steven@stevenroose.org>.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gonfig

import (
	"fmt"
	"reflect"
)

// setValueByString sets the value of the option by parsing the string.
func (o *option) setValueByString(s string) error {
	if o.isSlice {
		if err := parseSlice(o.value, s); err != nil {
			return fmt.Errorf("failed to set value of %s: %s", o.fullID(), err)
		}
	} else {
		if err := parseSimpleValue(o.value, s); err != nil {
			return fmt.Errorf("failed to set value of %s: %s", o.fullID(), err)
		}
	}

	return nil
}

// setValue sets the value of option to the given value.
// If the tye of the value is assignable or convertible to the type of the
// options value, it is directly set after optional conversion.
// If not, but the value is a string, it is passed to setValueByString.
// If not, and both v and the option's value are is a slice, we try converting
// the slice elements to the right elemens of the options slice.
func (o *option) setValue(v reflect.Value) error {
	t := o.value.Type()
	if v.Type().AssignableTo(t) {
		o.value.Set(v)
		return nil
	}

	if v.Type().ConvertibleTo(t) && o.value.Type() != typeOfByteSlice {
		o.value.Set(v.Convert(t))
		return nil
	}

	if v.Type().Kind() == reflect.String {
		return o.setValueByString(v.String())
	}

	if o.isSlice && v.Type().Kind() == reflect.Slice {
		return convertSlice(v, o.value)
	}

	return convertibleError(v, o.value.Type())
}

// isSupportedType returns whether the type t is supported by gonfig for parsing.
func isSupportedType(t reflect.Type) bool {
	if t.Implements(typeOfTextUnmarshaler) {
		return true
	}

	if t == typeOfByteSlice {
		return true
	}

	switch t.Kind() {
	case reflect.Bool:
		return true
	case reflect.String:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true

	case reflect.Struct:
		return true

	case reflect.Slice:
		// All but the fixed-bitsize types.
		return isSupportedType(t.Elem())

	case reflect.Ptr:
		return isSupportedType(t.Elem())

	default:
		return false
	}
}

// isZero checks if the value is the zero value for its type.
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && isZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		z := true
		for i := 0; i < v.NumField(); i++ {
			z = z && isZero(v.Field(i))
		}
		return z
	}
	// Compare other types directly:
	z := reflect.Zero(v.Type())
	return v.Interface() == z.Interface()
}
