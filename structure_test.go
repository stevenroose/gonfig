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
				fullIdParts: []string{"empty"},
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
				fullIdParts: []string{"realname"},
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
				fullIdParts: []string{"mother", "father"},
				id:          "father",
			},
			option{
				fullIdParts: []string{"mother", "father", "child"},
				defaultSet:  true,
				isParent:    false,
				id:          "child",
				short:       "S",
				defaul:      "defaultvalue",
				desc:        "testing..",
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
