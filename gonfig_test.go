// Copyright (c) 2017 Steven Roose <steven@stevenroose.org>.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gonfig

import (
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testTimeStr = "2009-11-10T23:00:00Z"
	testTime    *time.Time
)

func init() {
	testTime = &time.Time{}
	if err := testTime.UnmarshalText([]byte(testTimeStr)); err != nil {
		panic(err)
	}
}

func stringPointer(s string) *string {
	return &s
}

type NotSupported interface {
	DoStuff() error
}

type MarshaledUpper []byte

func (m MarshaledUpper) String() string {
	return string(m)
}

func (m *MarshaledUpper) UnmarshalText(t []byte) error {
	*m = MarshaledUpper(strings.ToLower(string(t)))
	return nil
}

type ErrorMarshaler string

func (m *ErrorMarshaler) UnmarshalText(t []byte) error {
	return errors.New("error")
}

type HexEncoded []byte

func (h HexEncoded) String() string {
	s, _ := h.MarshalText()
	return string(s)
}

func (h HexEncoded) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString([]byte(h))), nil
}

func (h *HexEncoded) UnmarshalText(t []byte) error {
	decoded, err := hex.DecodeString(string(t))
	if err != nil {
		return err
	}
	*h = decoded
	return nil
}

type TestStruct struct {
	StringVar  string  `default:"defstring" short:"s" desc:"descstring"`
	UintVar    uint    `default:"42"`
	IntVar     int     `default:"-42"`
	BoolVar1   bool    `short:"b"`
	BoolVar2   bool    `default:"true"`
	Float32Var float32 `id:"float" default:"0.50"`
	Float64Var float64 `default:"0.25"`

	Uint8Var      uint8
	Uint16Var     uint16
	Uint32Var     uint32
	Uint64Var     uint64
	Int8Var       int8
	Int16Var      int16
	Int32Var      int32 `id:"int-32-var"`
	Int64Var      int64
	ByteSliceVar1 []byte `id:"bytes1"`
	ByteSliceVar2 []byte `id:"bytes2" default:"AQID"`

	Strings1 []string `default:"string1,string2"`
	Strings2 []string `default:"string1,string2"`
	Strings3 []string `default:"string1,string2"`
	Strings4 []string `default:"string1,string2"`
	Ints1    []int    `id:"ints" default:"42,43"`
	Ints2    []int    `default:"42,43"`

	Nested NestedTestStruct `id:"nestedid"`

	Marshaled *MarshaledUpper `id:"upper1"`
	HexData   *HexEncoded     `id:"hex"`
}

type NestedTestStruct struct {
	StringVar string `default:"defstring2" short:"n" desc:"descstring2"`
	IntVar    int    `id:"int"`
	BoolVar1  bool   `id:"boolvar"`
}

func setOS(args []string, env map[string]string) {
	// Set command line args.
	os.Args = append([]string{"test"}, args...)

	// Set environment.
	os.Clearenv()
	for k, v := range env {
		os.Setenv(k, v)
	}
}

func TestGonfig(t *testing.T) {
	testCases := []struct {
		desc string

		args        []string
		env         map[string]string
		fileContent string

		conf Conf

		config      interface{}
		shouldError bool
		shouldPanic bool
		validate    func(t *testing.T, config interface{})
	}{
		{
			desc: "default env and cli",
			args: []string{"-b",
				"--int16var", "42",
				"--int64var", "42",
				"--uint32var", "42",
				"--float", "-0.25",
				"--nestedid.int", "42",
				"-n", "otherstringvalue",
				"--strings1", "one", "--strings1", "two", "--strings1", "three",
				"--ints", "3", "--ints", "2", "--ints", "1",
				"--upper1", "TEST",
				"--hex", "010203",
				"--bytes1", "AQID",
			},
			env: map[string]string{
				"INT8VAR":               "42",
				"PREF_UINT64VAR":        "42",
				"PREF_INT_32_VAR":       "42",
				"PREF_INT16VAR":         "32",
				"PREF_NESTEDID_BOOLVAR": "true",
				"PREF_STRINGS2":         "one,two,three",
				"PREF_INTS":             "1,2,3",
			},
			conf: Conf{
				FileDisable: true,
				EnvPrefix:   "PREF_",
			},
			config: &TestStruct{},
			validate: func(t *testing.T, config interface{}) {
				c, success := config.(*TestStruct)
				require.True(t, success)

				assert.EqualValues(t, "defstring", c.StringVar)
				assert.EqualValues(t, 42, c.UintVar)
				assert.EqualValues(t, -42, c.IntVar)
				assert.EqualValues(t, true, c.BoolVar1)
				assert.EqualValues(t, true, c.BoolVar2)
				assert.EqualValues(t, -0.25, c.Float32Var)
				assert.EqualValues(t, 0.25, c.Float64Var)
				assert.EqualValues(t, 0, c.Int8Var)
				assert.EqualValues(t, 42, c.Int16Var)
				assert.EqualValues(t, 42, c.Int64Var)
				assert.EqualValues(t, 42, c.Uint32Var)
				assert.EqualValues(t, 42, c.Uint64Var)
				assert.EqualValues(t, 42, c.Int32Var)
				assert.EqualValues(t, "otherstringvalue", c.Nested.StringVar)
				assert.EqualValues(t, true, c.Nested.BoolVar1)
				assert.EqualValues(t, 42, c.Nested.IntVar)
				assert.EqualValues(t, []byte{1, 2, 3}, c.ByteSliceVar1)
				assert.EqualValues(t, []byte{1, 2, 3}, c.ByteSliceVar2)
				slice123 := []string{"one", "two", "three"}
				assert.EqualValues(t, slice123, c.Strings1)
				assert.EqualValues(t, slice123, c.Strings2)
				assert.EqualValues(t, []string{"string1", "string2"}, c.Strings3)
				assert.EqualValues(t, []int{3, 2, 1}, c.Ints1)
				assert.EqualValues(t, []int{42, 43}, c.Ints2)
				assert.EqualValues(t, "test", c.Marshaled.String())
				assert.EqualValues(t, "010203", c.HexData.String())
			},
		},
		{
			desc: "json",
			args: []string{"--uint8var", "42"},
			env:  map[string]string{"UINT16VAR": "42"},
			fileContent: `{
				"stringvar": "stringvalue",
				"uintvar": 43,
				"intvar": -43,
				"boolvar1": true,
				"float": -0.5,
				"float64var": -0.25,
				"uint8var": 42,
				"int8var": 42,
				"int-32-var": 42,
				"bytes1": "AQID",
				"strings1": ["one", "two", "three"],
				"ints": [1,2,3],
				"nestedid": {
					"stringvar": "otherstringvalue",
					"int": 42
				},
				"upper1": "TEST",
				"hex": "010203"
			}`,
			conf: Conf{
				FileDecoder: DecoderJSON,
			},
			config: &TestStruct{},
			validate: func(t *testing.T, config interface{}) {
				c, success := config.(*TestStruct)
				require.True(t, success)

				assert.EqualValues(t, "stringvalue", c.StringVar)
				assert.EqualValues(t, 43, c.UintVar)
				assert.EqualValues(t, -43, c.IntVar)
				assert.EqualValues(t, true, c.BoolVar1)
				assert.EqualValues(t, true, c.BoolVar2)
				assert.EqualValues(t, -0.5, c.Float32Var)
				assert.EqualValues(t, -0.25, c.Float64Var)
				assert.EqualValues(t, 42, c.Uint8Var)
				assert.EqualValues(t, 42, c.Int8Var)
				assert.EqualValues(t, 0, c.Int16Var)
				assert.EqualValues(t, 42, c.Int32Var)
				assert.EqualValues(t, []byte{1, 2, 3}, c.ByteSliceVar1)
				assert.EqualValues(t, []byte{1, 2, 3}, c.ByteSliceVar2)
				assert.EqualValues(t, "otherstringvalue", c.Nested.StringVar)
				assert.EqualValues(t, false, c.Nested.BoolVar1)
				assert.EqualValues(t, 42, c.Nested.IntVar)
				slice123 := []string{"one", "two", "three"}
				assert.EqualValues(t, slice123, c.Strings1)
				assert.EqualValues(t, []int{1, 2, 3}, c.Ints1)
				assert.EqualValues(t, "test", c.Marshaled.String())
				assert.EqualValues(t, "010203", c.HexData.String())
			},
		},
		{
			desc: "yaml",
			args: []string{"--uint8var", "42"},
			env:  map[string]string{"UINT16VAR": "42"},
			fileContent: "stringvar: stringvalue\n" +
				"uintvar: 43\n" +
				"intvar: -43\n" +
				"boolvar1: true\n" +
				"float: -0.5\n" +
				"float64var: -0.25\n" +
				"uint8var: 42\n" +
				"int8var: 42\n" +
				"int-32-var: 42\n" +
				"bytes1: AQID\n" +
				"strings1:\n" +
				"  - one\n" +
				"  - two\n" +
				"  - three\n" +
				"ints: [1, 2, 3]\n" +
				"nestedid:\n" +
				"  stringvar: otherstringvalue\n" +
				"  int: 42\n" +
				"upper1: TEST\n" +
				"hex: \"010203\"\n",
			conf: Conf{
				FileDecoder: DecoderYAML,
			},
			config: &TestStruct{},
			validate: func(t *testing.T, config interface{}) {
				c, success := config.(*TestStruct)
				require.True(t, success)

				assert.EqualValues(t, "stringvalue", c.StringVar)
				assert.EqualValues(t, 43, c.UintVar)
				assert.EqualValues(t, -43, c.IntVar)
				assert.EqualValues(t, true, c.BoolVar1)
				assert.EqualValues(t, true, c.BoolVar2)
				assert.EqualValues(t, -0.5, c.Float32Var)
				assert.EqualValues(t, -0.25, c.Float64Var)
				assert.EqualValues(t, 42, c.Uint8Var)
				assert.EqualValues(t, 42, c.Int8Var)
				assert.EqualValues(t, 0, c.Int16Var)
				assert.EqualValues(t, 42, c.Int32Var)
				assert.EqualValues(t, []byte{1, 2, 3}, c.ByteSliceVar1)
				assert.EqualValues(t, []byte{1, 2, 3}, c.ByteSliceVar2)
				assert.EqualValues(t, "otherstringvalue", c.Nested.StringVar)
				assert.EqualValues(t, false, c.Nested.BoolVar1)
				assert.EqualValues(t, 42, c.Nested.IntVar)
				slice123 := []string{"one", "two", "three"}
				assert.EqualValues(t, slice123, c.Strings1)
				assert.EqualValues(t, []int{1, 2, 3}, c.Ints1)
				assert.EqualValues(t, "test", c.Marshaled.String())
				assert.EqualValues(t, "010203", c.HexData.String())
			},
		},
		{
			desc: "toml",
			args: []string{"--uint8var", "42"},
			env:  map[string]string{"UINT16VAR": "42"},
			fileContent: "stringvar = \"stringvalue\"\n" +
				"uintvar = 43\n" +
				"intvar = -43\n" +
				"boolvar1 = true\n" +
				"float = -0.5\n" +
				"float64var = -0.25\n" +
				"uint8var = 42\n" +
				"int8var = 42\n" +
				"int-32-var = 42\n" +
				"bytes1 = \"AQID\"\n" +
				"strings1 = [\"one\", \"two\", \"three\"]\n" +
				"ints = [1, 2, 3]\n" +
				"upper1 = \"TEST\"\n" +
				"hex = \"010203\"\n" +
				"[nestedid]\n" +
				"stringvar = \"otherstringvalue\"\n" +
				"int = 42\n",
			conf: Conf{
				FileDecoder: DecoderTOML,
			},
			config: &TestStruct{},
			validate: func(t *testing.T, config interface{}) {
				c, success := config.(*TestStruct)
				require.True(t, success)

				assert.EqualValues(t, "stringvalue", c.StringVar)
				assert.EqualValues(t, 43, c.UintVar)
				assert.EqualValues(t, -43, c.IntVar)
				assert.EqualValues(t, true, c.BoolVar1)
				assert.EqualValues(t, true, c.BoolVar2)
				assert.EqualValues(t, -0.5, c.Float32Var)
				assert.EqualValues(t, -0.25, c.Float64Var)
				assert.EqualValues(t, 42, c.Uint8Var)
				assert.EqualValues(t, 42, c.Int8Var)
				assert.EqualValues(t, 0, c.Int16Var)
				assert.EqualValues(t, 42, c.Int32Var)
				assert.EqualValues(t, []byte{1, 2, 3}, c.ByteSliceVar1)
				assert.EqualValues(t, []byte{1, 2, 3}, c.ByteSliceVar2)
				assert.EqualValues(t, "otherstringvalue", c.Nested.StringVar)
				assert.EqualValues(t, false, c.Nested.BoolVar1)
				assert.EqualValues(t, 42, c.Nested.IntVar)
				slice123 := []string{"one", "two", "three"}
				assert.EqualValues(t, slice123, c.Strings1)
				assert.EqualValues(t, []int{1, 2, 3}, c.Ints1)
				assert.EqualValues(t, "test", c.Marshaled.String())
				assert.EqualValues(t, "010203", c.HexData.String())
			},
		},
		{
			desc: "interface type not supported",
			config: &struct {
				Var1 NotSupported
			}{},
			shouldPanic: true,
		},
		{
			desc: "map not supported",
			config: &struct {
				Map map[string]string
			}{},
			shouldPanic: true,
		},
		{
			desc: "invalid default value bool",
			config: &struct {
				V bool `default:"strng"`
			}{},
			shouldPanic: true,
		},
		{
			desc: "invalid default value int",
			config: &struct {
				V int `default:"strng"`
			}{},
			shouldPanic: true,
		},
		{
			desc: "invalid default value uint",
			config: &struct {
				V uint `default:"-1"`
			}{},
			shouldPanic: true,
		},
		{
			desc: "invalid default value int8",
			config: &struct {
				V int8 `default:"9999999"`
			}{},
			shouldPanic: true,
		},
		{
			desc: "invalid default value float",
			config: &struct {
				V float64 `default:"strng"`
			}{},
			shouldPanic: true,
		},
		{
			desc: "invalid default value byte slice",
			config: &struct {
				V []byte `default:"strng"`
			}{},
			shouldPanic: true,
		},
		{
			desc: "default value on nested id",
			config: &struct {
				V TestStruct `default:"strng"`
			}{},
			shouldPanic: true,
		},
		{
			desc: "incorrect value passed to bool",
			config: &struct {
				Var bool
			}{},
			conf:        Conf{EnvDisable: true, FileDisable: true},
			args:        []string{"--var=strng"},
			shouldError: true,
		},
		{
			desc: "incorrect value passed to int",
			config: &struct {
				Var int
			}{},
			conf:        Conf{EnvDisable: true, FileDisable: true},
			args:        []string{"--var", "strng"},
			shouldError: true,
		},
		{
			desc: "incorrect value passed to uint",
			config: &struct {
				Var uint
			}{},
			conf:        Conf{EnvDisable: true, FileDisable: true},
			args:        []string{"--var", "-1"},
			shouldError: true,
		},
		{
			desc: "incorrect value passed to int8",
			config: &struct {
				Var int8
			}{},
			conf:        Conf{EnvDisable: true, FileDisable: true},
			args:        []string{"--var", "9999999"},
			shouldError: true,
		},
		{
			desc: "incorrect value passed to float",
			config: &struct {
				Var float64
			}{},
			conf:        Conf{EnvDisable: true, FileDisable: true},
			args:        []string{"--var", "strng"},
			shouldError: true,
		},
		{
			desc: "incorrect value passed to byte slice",
			config: &struct {
				Var []byte
			}{},
			conf:        Conf{EnvDisable: true, FileDisable: true},
			args:        []string{"--var", "strng"},
			shouldError: true,
		},
		{
			desc: "incorrect value passed via flags",
			config: &struct {
				Var int
			}{},
			conf:        Conf{EnvDisable: true, FileDisable: true},
			args:        []string{"--var", "strng"},
			shouldError: true,
		},
		{
			desc: "incorrect value passed via env",
			config: &struct {
				Var int
			}{},
			conf: Conf{FlagDisable: true, FileDisable: true},
			env: map[string]string{
				"VAR": "strng",
			},
			shouldError: true,
		},
		{
			desc: "value passed into nested ID",
			config: &struct {
				Var struct {
					Inner int
				}
			}{},
			conf:        Conf{EnvDisable: true, FileDisable: true},
			args:        []string{"--var", "strng"},
			shouldError: true,
		},
		{
			desc: "pointer to struct",
			config: &struct {
				Var *struct {
					Inner int
				}
			}{},
			conf: Conf{EnvDisable: true, FileDisable: true},
			args: []string{"--var.inner", "5"},
			validate: func(t *testing.T, config interface{}) {
				c, success := config.(*struct {
					Var *struct {
						Inner int
					}
				})
				require.True(t, success)

				assert.Equal(t, 5, c.Var.Inner)
			},
		},
		{
			desc: "duplicate id",
			config: &struct {
				V   int
				NoV int `id:"v"`
			}{},
			shouldPanic: true,
		},
		{
			desc: "duplicate shorthand",
			config: &struct {
				V   int `short:"s"`
				NoV int `short:"s"`
			}{},
			shouldPanic: true,
		},
		{
			desc: "duplicate shorthand nested",
			config: &struct {
				V   int `short:"s"`
				Var *struct {
					Inner int `short:"s"`
				}
			}{},
			shouldPanic: true,
		},
		{
			desc: "config no pointer",
			config: struct {
				V int
			}{},
			shouldPanic: true,
		},
		{
			desc:        "config pointer to no struct",
			config:      stringPointer("s"),
			shouldPanic: true,
		},
		{
			desc: "unmarshaller error",
			config: &struct {
				V *ErrorMarshaler
			}{},
			conf:        Conf{EnvDisable: true, FileDisable: true},
			args:        []string{"--v", "ss"},
			shouldError: true,
		},
		{
			desc: "element in slice not convertible",
			config: &struct {
				V []int
			}{},
			conf:        Conf{EnvDisable: true, FileDisable: true},
			args:        []string{"--v", "5", "--v", "ss"},
			shouldError: true,
		},
		{
			desc: "error in inner struct",
			config: &struct {
				V struct {
					Inner NotSupported
				}
			}{},
			shouldPanic: true,
		},
		{
			desc: "error in inner struct pointer",
			config: &struct {
				V *struct {
					Inner NotSupported
				}
			}{},
			shouldPanic: true,
		},
		{
			desc: "time.Time",
			config: &struct {
				Tm *time.Time
			}{},
			conf: Conf{EnvDisable: true, FileDisable: true},
			args: []string{"--tm", testTimeStr},
			validate: func(t *testing.T, config interface{}) {
				c, success := config.(*struct {
					Tm *time.Time
				})
				require.True(t, success)

				assert.EqualValues(t, testTime, c.Tm)
			},
		},
		{
			desc: "find file encoding",
			config: &struct {
				V int
			}{},
			conf:        Conf{EnvDisable: true, FlagDisable: true},
			fileContent: `{"v": 5}`,
			validate: func(t *testing.T, config interface{}) {
				c, success := config.(*struct {
					V int
				})
				require.True(t, success)

				assert.Equal(t, 5, c.V)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			setOS(tc.args, tc.env)

			// Write config file.
			var filename string
			if tc.fileContent != "" {
				file, err := ioutil.TempFile("", "gonfig")
				require.NoError(t, err)
				_, err = file.WriteString(tc.fileContent)
				require.NoError(t, err)
				filename = file.Name()
				t.Logf("Config file created at %s", filename)
			}

			conf := tc.conf
			conf.FileDefaultFilename = filename

			if tc.shouldPanic {
				require.Panics(t, func() { Load(tc.config, conf) })
			} else if tc.shouldError {
				require.Error(t, Load(tc.config, conf))
			} else {
				require.NoError(t, Load(tc.config, conf))
				tc.validate(t, tc.config)
			}
		})
	}
}

func TestFindConfigFile_NoVariable(t *testing.T) {
	setOS(nil, nil)
	s := &setup{
		conf: &Conf{
			FlagDisable:         true,
			EnvDisable:          true,
			FileDefaultFilename: "/default.conf",
		},
	}

	filename, err := findCustomConfigFile(s)
	require.NoError(t, err)
	assert.Empty(t, filename)
}

func TestFindConfigFile_WithFlag(t *testing.T) {
	setOS([]string{"--configfile", "fromflag.conf"}, nil)
	s := &setup{
		conf: &Conf{
			FlagDisable:         true,
			EnvDisable:          true,
			FileDefaultFilename: "default.conf",
			ConfigFileVariable:  "configfile",
		},
	}
	require.NoError(t, inspectConfigStructure(s, &struct {
		ConfigFile string
	}{}))

	filename, err := findCustomConfigFile(s)
	require.NoError(t, err)
	wd, err := os.Getwd()
	require.NoError(t, err)
	assert.Equal(t, path.Join(wd, "fromflag.conf"), filename)
}

func TestFindConfigFile_WithEnv(t *testing.T) {
	setOS(nil, map[string]string{"CONFIGFILE": "fromenv.conf"})
	s := &setup{
		conf: &Conf{
			FlagDisable:         true,
			EnvDisable:          true,
			FileDefaultFilename: "default.conf",
			ConfigFileVariable:  "configfile",
		},
	}
	require.NoError(t, inspectConfigStructure(s, &struct {
		ConfigFile string
	}{}))

	filename, err := findCustomConfigFile(s)
	require.NoError(t, err)
	wd, err := os.Getwd()
	require.NoError(t, err)
	assert.Equal(t, path.Join(wd, "fromenv.conf"), filename)
}

func TestFindConfigFile_VariableNotSet(t *testing.T) {
	setOS(nil, nil)
	s := &setup{
		conf: &Conf{
			FlagDisable:         true,
			EnvDisable:          true,
			FileDefaultFilename: "default.conf",
			ConfigFileVariable:  "configfile",
		},
	}
	require.NoError(t, inspectConfigStructure(s, &struct {
		ConfigFile string
	}{}))

	filename, err := findCustomConfigFile(s)
	require.NoError(t, err)
	assert.Empty(t, filename)
}

func TestFindConfigFile_VariableNotProvided(t *testing.T) {
	setOS(nil, nil)
	s := &setup{
		conf: &Conf{
			FlagDisable:         true,
			EnvDisable:          true,
			FileDefaultFilename: "default.conf",
			ConfigFileVariable:  "configfile",
		},
	}
	require.NoError(t, inspectConfigStructure(s, &struct {
		ConfigFileX string
	}{}))

	assert.Panics(t, func() { findCustomConfigFile(s) })
}
