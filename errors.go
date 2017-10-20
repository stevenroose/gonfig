package gonfig

import (
	"errors"
	"fmt"
	"reflect"
)

// parseError returns a nicely formatted error indicating that we failed to
// parse v into type t.
func parseError(v string, t reflect.Type, err error) error {
	msg := fmt.Sprintf("failed to parse '%s' into type %s", v, t)
	if err != nil {
		msg += ": " + err.Error()
	}
	return errors.New(msg)
}

// convertibleError returns a nicely formatted error indicating that the value
// v is not convertible to type t.
func convertibleError(v reflect.Value, t reflect.Type) error {
	return fmt.Errorf(
		"incompatible type: %s not convertible to %s", v.Type(), t)
}
