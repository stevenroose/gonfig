// Copyright (c) 2017 Steven Roose <steven@stevenroose.org>.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gonfig

import (
	"errors"
	"fmt"
	"reflect"
)

// setValueByString sets the value by parsing the string.
func setValueByString(v reflect.Value, s string) error {
	if isSlice(v) {
		if err := parseSlice(v, s); err != nil {
			return fmt.Errorf("failed to parse slice value: %v", err)
		}
	} else {
		if err := parseSimpleValue(v, s); err != nil {
			return fmt.Errorf("failed to parse value: %v", err)
		}
	}

	return nil
}

// setValue sets the option value to the given value.
// If the tye of the value is assignable or convertible to the type of the
// option value, it is directly set after optional conversion.
// If not, but the value is a string, it is passed to setValueByString.
// If not, and both v and the option's value are is a slice, we try converting
// the slice elements to the right elemens of the options slice.
func setValue(toSet, v reflect.Value) error {
	t := toSet.Type()
	if v.Type().AssignableTo(t) {
		toSet.Set(v)
		return nil
	}

	if v.Type().ConvertibleTo(t) && toSet.Type() != typeOfByteSlice {
		toSet.Set(v.Convert(t))
		return nil
	}

	if v.Type().Kind() == reflect.String {
		return setValueByString(toSet, v.String())
	}

	if isSlice(toSet) && v.Type().Kind() == reflect.Slice {
		return convertSlice(v, toSet)
	}

	return convertibleError(v, toSet.Type())
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
