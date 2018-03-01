// Copyright (c) 2017 Steven Roose <steven@stevenroose.org>.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gonfig

import (
	"errors"
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
func isSupportedType(t reflect.Type) error {
	if t.Implements(typeOfTextUnmarshaler) {
		return nil
	}

	if t == typeOfByteSlice {
		return nil
	}

	switch t.Kind() {
	case reflect.Bool:
	case reflect.String:
	case reflect.Float32, reflect.Float64:
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:

	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if err := isSupportedType(t.Field(i).Type); err != nil {
				return fmt.Errorf("struct with unsupported type: %v", err)
			}
		}

	case reflect.Slice:
		// All but the fixed-bitsize types.
		if err := isSupportedType(t.Elem()); err != nil {
			return fmt.Errorf("slice of unsupported type: %v", err)
		}

	case reflect.Ptr:
		if err := isSupportedType(t.Elem()); err != nil {
			return fmt.Errorf("pointer to unsupported type: %v", err)
		}

	case reflect.Map:
		if t.Key().Kind() != reflect.String || t.Elem().Kind() != reflect.Interface {
			return errors.New("only maps of type map[string]interface{} are supported")
		}

	default:
		return errors.New("type not supported")
	}

	return nil
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

// setSimpleMapValue trues to add the key and value to the map.
func setSimpleMapValue(mapValue reflect.Value, key, value string) error {
	v := reflect.New(mapValue.Type().Elem()).Elem()
	if err := parseSimpleValue(v, value); err != nil {
		return err
	}
	mapValue.SetMapIndex(reflect.ValueOf(key), v)
	return nil
}
