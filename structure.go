package gonfig

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const ( // The values for the struct field tags that we use.
	fieldTagID          = "id"
	fieldTagShort       = "short"
	fieldTagDefault     = "default"
	fieldTagDescription = "desc"
)

var (
	typeOfTextUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	typeOfByteSlice       = reflect.TypeOf([]byte{})
)

type option struct {
	value   reflect.Value
	subOpts []*option

	// fullIdParts holds the hierarchical ID of the option with all the names of its
	// parent options
	fullIdParts []string
	defaultSet  bool
	isParent    bool
	isSlice     bool

	// Metadata specified by user.
	id     string // the identifier
	short  string // the shorthand to be used in CLI flags
	defaul string // the default value
	desc   string // the description
}

func (o option) fullId() string {
	return strings.Join(o.fullIdParts, ".")
}

func (o *option) setValueByString(s string) error {
	t := o.value.Type()
	if t.Implements(typeOfTextUnmarshaler) {
		unmarshaler := o.value.Interface().(encoding.TextUnmarshaler)
		if err := unmarshaler.UnmarshalText([]byte(s)); err != nil {
			return unmarshalError(s, o, err)
		}
	}

	if o.isSlice {
		if err := parseSlice(o.value, s); err != nil {
			return fmt.Errorf("failed to set value of %s: %s", o.fullId(), err)
		}
	} else {
		if err := parseSimpleValue(o.value, s); err != nil {
			return fmt.Errorf("failed to set value of %s: %s", o.fullId(), err)
		}
	}

	return nil
}

func convertSlice(from, to reflect.Value) error {
	subType := to.Type().Elem()
	converted := reflect.MakeSlice(to.Type(), from.Len(), from.Len())
	for i := 0; i < from.Len(); i++ {
		elem := from.Index(i)
		if elem.Type().Kind() == reflect.Interface {
			elem = elem.Elem()
		}

		if !elem.Type().ConvertibleTo(subType) {
			return convertibleError(elem, subType)
		}

		converted.Index(i).Set(elem.Convert(subType))
	}

	to.Set(converted)
	return nil
}

func (o *option) setValue(v reflect.Value) error {
	if o.isSlice && v.Type().Kind() == reflect.Slice {
		return convertSlice(v, o.value)
	}

	if !v.Type().ConvertibleTo(o.value.Type()) {
		if v.Type().Kind() == reflect.String {
			// Try setting by string.
			return o.setValueByString(v.String())
		}
		return convertibleError(v, o.value.Type())
	}

	if v.Type().AssignableTo(o.value.Type()) {
		o.value.Set(v)
	} else {
		o.value.Set(v.Convert(o.value.Type()))
	}

	return nil
}

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

func optionFromField(f reflect.StructField, parent *option) *option {
	opt := new(option)

	id := f.Tag.Get(fieldTagID)
	if len(id) == 0 {
		id = strings.ToLower(f.Name)
	}
	opt.id = id

	if parent == nil {
		opt.fullIdParts = []string{id}
	} else {
		opt.fullIdParts = append(parent.fullIdParts, id)
	}

	opt.short = f.Tag.Get(fieldTagShort)
	opt.defaul, opt.defaultSet = f.Tag.Lookup(fieldTagDefault)
	opt.desc = f.Tag.Get(fieldTagDescription)

	return opt
}

func parseOptionsFromStruct(v reflect.Value, parent *option) ([]*option, []*option, error) {
	var opts []*option
	var allOpts []*option // recursively includes all subOpts

	for f := 0; f < v.NumField(); f++ {
		field := v.Type().Field(f)
		opt := optionFromField(field, parent)

		if !isSupportedType(field.Type) {
			return nil, nil, fmt.Errorf(
				"type of field %s (%s) is not supported",
				field.Name, field.Type)
		}

		opt.value = v.Field(f)

		// If a struct type, recursively add values for the inner struct.
		var subOpts, allSubOpts []*option
		var err error
		switch field.Type.Kind() {
		case reflect.Ptr:
			if field.Type.Elem().Kind() != reflect.Struct {
				break
			}
			opt.isParent = true
			subOpts, allSubOpts, err = parseOptionsFromStruct(opt.value.Elem(), opt)
		case reflect.Struct:
			opt.isParent = true
			subOpts, allSubOpts, err = parseOptionsFromStruct(opt.value, opt)
		case reflect.Slice:
			opt.isSlice = true
		}
		if err != nil {
			return nil, nil, err
		}
		opt.subOpts = subOpts

		opts = append(opts, opt)
		allOpts = append(allOpts, opt)
		allOpts = append(allOpts, allSubOpts...)
	}

	// Check for duplicate values for IDs.
	for i := range opts {
		for j := range opts {
			if i != j {
				if opts[i].id == opts[j].id {
					return nil, nil, errors.New(
						"duplicate config variable: " + opts[i].id)
				}
			}
		}
	}

	return opts, allOpts, nil
}

func inspectConfigStructure(s *setup, c interface{}) error {
	// First make sure that we have a pointer to a struct.
	if reflect.TypeOf(c).Kind() != reflect.Ptr {
		return errors.New("config variable must be a pointer to a struct")
	}
	v := reflect.ValueOf(c).Elem()
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return errors.New("config variable must be a pointer to a struct")
	}

	opts, allOpts, err := parseOptionsFromStruct(v, nil)
	if err != nil {
		return err
	}

	// The method for getting the options from a struct already checks for
	// duplicate IDs.
	// Here we check for duplicate shorts among all options.
	for i := range allOpts {
		for j := range allOpts {
			if i != j {
				if allOpts[i].short != "" && allOpts[i].short == allOpts[j].short {
					return errors.New(
						"duplicate config variable shorthand: " + opts[i].short)
				}
			}
		}
	}

	s.opts = opts
	s.allOpts = allOpts
	return nil
}
