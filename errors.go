package gonfig

import (
	"errors"
	"fmt"
	"reflect"
)

func parseError(v string, t reflect.Type, err error) error {
	msg := fmt.Sprintf(
		"failed to parse '%s' into type %s", v, t)
	if err != nil {
		msg += ": " + err.Error()
	}
	return errors.New(msg)
}

func overflowError(v interface{}, o *option) error {
	msg := fmt.Sprintf(
		"value '%s' overflows type %s of config var %s",
		v, o.value.Type(), o.id)
	return errors.New(msg)
}

func unmarshalError(v string, o *option, err error) error {
	return fmt.Errorf(
		"failed to unmarshal '%s' into type %s of config var %s: %s",
		v, o.value.Type(), o.id, err)
}

func convertibleError(v reflect.Value, typ reflect.Type) error {
	return fmt.Errorf(
		"incompatible type: %s not convertible to %s", v.Type(), typ)
}
