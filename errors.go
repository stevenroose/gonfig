// Copyright (c) 2017 Steven Roose <steven@stevenroose.org>.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gonfig

import (
	"errors"
	"fmt"
	"reflect"
)

// parseError returns a nicely formatted error indicating that we failed to
// parse v into type t.
func parseError(v string, t reflect.Type, err error) error {
	msg := fmt.Sprintf("failed to parse '%v' into type %v", v, t)
	if err != nil {
		msg += ": " + err.Error()
	}
	return errors.New(msg)
}

// convertibleError returns a nicely formatted error indicating that the value
// v is not convertible to type t.
func convertibleError(v reflect.Value, t reflect.Type) error {
	return fmt.Errorf(
		"incompatible type: %v not convertible to %v", v.Type(), t)
}
