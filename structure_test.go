// Copyright (c) 2017 Steven Roose <steven@stevenroose.org>.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gonfig

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptionFromField(t *testing.T) {
	testCases := []struct {
		desc string

		fieldName string
		fieldTag  string
		parent    option

		expected option
	}{
		{
			"empty",
			"Empty",
			``,
			option{},
			option{
				fullIDParts: []string{"empty"},
				defaultSet:  false,
				isParent:    false,
				id:          "empty",
				short:       "",
				defaul:      "",
				desc:        "",
			},
		},
		{
			"normal",
			"name",
			`id:"realname" short:"s" default:"defaultvalue" desc:"testing.."`,
			option{},
			option{
				fullIDParts: []string{"realname"},
				defaultSet:  true,
				isParent:    false,
				id:          "realname",
				short:       "s",
				defaul:      "defaultvalue",
				desc:        "testing..",
			},
		},
		{
			"with parent",
			"child",
			`short:"S" default:"defaultvalue" desc:"testing.."`,
			option{
				isParent:    true,
				fullIDParts: []string{"mother", "father"},
				id:          "father",
			},
			option{
				fullIDParts: []string{"mother", "father", "child"},
				defaultSet:  true,
				isParent:    false,
				id:          "child",
				short:       "S",
				defaul:      "defaultvalue",
				desc:        "testing..",
			},
		},
		{
			"with option",
			"name",
			`id:"realname" opts:"hidden" short:"s" default:"defaultvalue" desc:"testing.."`,
			option{},
			option{
				fullIDParts: []string{"realname"},
				defaultSet:  true,
				isParent:    false,
				id:          "realname",
				short:       "s",
				defaul:      "defaultvalue",
				desc:        "testing..",
				opts:        []string{"hidden"},
			},
		},
		{
			"with capital letters (not first)",
			"myField",
			"",
			option{},
			option{
				fullIDParts: []string{"my-field"},
				defaultSet:  false,
				isParent:    false,
				id:          "my-field",
				short:       "",
				defaul:      "",
				desc:        "",
			},
		},
		{
			"with capital letters (even first)",
			"MyField",
			"",
			option{},
			option{
				fullIDParts: []string{"my-field"},
				defaultSet:  false,
				isParent:    false,
				id:          "my-field",
				short:       "",
				defaul:      "",
				desc:        "",
			},
		},
		{
			"with capital letters (id overriden)",
			"MyField",
			`id:"field"`,
			option{},
			option{
				fullIDParts: []string{"field"},
				defaultSet:  false,
				isParent:    false,
				id:          "field",
				short:       "",
				defaul:      "",
				desc:        "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			field := reflect.StructField{
				Name: tc.fieldName,
				Tag:  reflect.StructTag(tc.fieldTag),
			}

			result := optionFromField(field, &tc.parent)

			assert.Equal(t, &tc.expected, result)
		})
	}
}
