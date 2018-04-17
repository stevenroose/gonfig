// Copyright (c) 2017 Steven Roose <steven@stevenroose.org>.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

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
	fieldTagOpts        = "opts"
)

const ( // The values for the struct tag options.
	fieldOptHidden = "hidden"
)

var ( // Some type variables for comparison.
	typeOfTextUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	typeOfByteSlice       = reflect.TypeOf([]byte{})
)

// option holds all useful data and metadata for a single config option variable
// of the config struct.
type option struct {
	value   reflect.Value
	subOpts []*option

	fullIDParts  []string      // full ID of the option with all its parents
	defaultSet   bool          // the default value was set in the structure
	defaultValue reflect.Value // the default value
	isParent     bool          // is nested and has children
	isMap        bool          // is a map type

	// Struct metadata specified by user.
	id     string   // the identifier
	short  string   // the shorthand to be used in CLI flags
	defaul string   // the default value
	desc   string   // the description
	opts   []string // the field opts flags
}

// fullID returns the full ID of the option consisting of all IDs of its parents
// joined by dots.
func (o option) fullID() string {
	return strings.Join(o.fullIDParts, ".")
}

// hasFieldOpt returns whether the given opt flag is set.
func (o option) hasFieldOpt(opt string) bool {
	for _, op := range o.opts {
		if op == opt {
			return true
		}
	}
	return false
}

// optionFromField creates a new option from the field information.
func optionFromField(f reflect.StructField, parent *option) *option {
	opt := new(option)

	id := f.Tag.Get(fieldTagID)
	if len(id) == 0 {
		id = strings.ToLower(f.Name)
	}
	opt.id = id

	if parent == nil {
		opt.fullIDParts = []string{id}
	} else {
		opt.fullIDParts = append(parent.fullIDParts, id)
	}

	opt.short = f.Tag.Get(fieldTagShort)
	opt.defaul, opt.defaultSet = f.Tag.Lookup(fieldTagDefault)
	opt.desc = f.Tag.Get(fieldTagDescription)
	if opts, any := f.Tag.Lookup(fieldTagOpts); any {
		opt.opts = strings.Split(opts, ",")
	}

	return opt
}

// createOptionsFromStruct extracts all options from the struct in a
// recursive manner.
// It returns first a slice of all the options of the struct and second a slice
// of all the options of the slice including all options of the options of the
// slice, in a recursive manner.
func createOptionsFromStruct(v reflect.Value, parent *option) ([]*option, []*option, error) {
	var opts []*option
	var allOpts []*option // recursively includes all subOpts

	for f := 0; f < v.NumField(); f++ {
		field := v.Type().Field(f)
		value := v.Field(f)

		if !value.CanSet() {
			// Unexported field, ignoring.
			continue
		}

		opt := optionFromField(field, parent)
		opt.value = value

		if err := isSupportedType(field.Type); err != nil {
			return nil, nil, fmt.Errorf(
				"type of field %v (%v) is not supported: %v",
				field.Name, field.Type, err)
		}

		var (
			t = field.Type
			k = field.Type.Kind()
		)

		// If it is a pointer, it might be nil. Let's fill it with something.
		if k == reflect.Ptr && opt.value.IsNil() {
			opt.value.Set(reflect.New(t.Elem()))
		}
		if k == reflect.Map && opt.value.IsNil() {
			opt.value.Set(reflect.MakeMap(t))
		}

		var err error
		var allSubOpts []*option
		if t.Implements(typeOfTextUnmarshaler) {
			// TextUnmarshaler is a normal type, should not do more.
		} else if k == reflect.Map {
			opt.isMap = true
		} else if k == reflect.Struct {
			opt.isParent = true
			opt.subOpts, allSubOpts, err = createOptionsFromStruct(opt.value, opt)
			if err != nil {
				return nil, nil, err
			}
		} else if k == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
			opt.isParent = true
			opt.subOpts, allSubOpts, err = createOptionsFromStruct(opt.value.Elem(), opt)
			if err != nil {
				return nil, nil, err
			}
		}

		opts = append(opts, opt)
		allOpts = append(allOpts, append(allSubOpts, opt)...)
	}

	// Check for duplicate values for IDs inside the same struct.
	for i := range opts {
		for j := range opts {
			if i != j {
				if opts[i].id == opts[j].id {
					return nil, nil, errors.New(
						"duplicate config variable: \"" + opts[i].id + "\"")
				}
			}
		}
	}

	return opts, allOpts, nil
}

// inspectConfigStructure inspects the config struct c and inspects it while
// building the set of options and performing sanity checks.
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

	opts, allOpts, err := createOptionsFromStruct(v, nil)
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
