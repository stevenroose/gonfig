package gonfig

import (
	"encoding/hex"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MarshaledUpper []byte

func (m MarshaledUpper) String() string {
	return string(m)
}

func (m *MarshaledUpper) UnmarshalText(t []byte) error {
	*m = MarshaledUpper(strings.ToLower(string(t)))
	return nil
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

	Uint8Var  uint8
	Uint16Var uint16
	Uint32Var uint32
	Uint64Var uint64
	Int8Var   int8
	Int16Var  int16
	Int32Var  int32 `id:"int-32-var"`
	Int64Var  int64

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
	StringVar string `default:"defstring2" short:"h" desc:"descstring2"`
	IntVar    int    `id:"int"`
	BoolVar1  bool   `id:"boolvar"`
}

func TestGonfig(t *testing.T) {
	testCases := []struct {
		desc string

		args        []string
		env         map[string]string
		fileContent string

		conf Conf

		config     interface{}
		shouldFail bool
		validate   func(t *testing.T, config interface{})
	}{
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
				FlagEnable:   true,
				EnvEnable:    true,
				FileEnable:   true,
				FileEncoding: "json",
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
				FlagEnable:   true,
				EnvEnable:    true,
				FileEnable:   true,
				FileEncoding: "yaml",
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
				"strings1 = [\"one\", \"two\", \"three\"]\n" +
				"ints = [1, 2, 3]\n" +
				"upper1 = \"TEST\"\n" +
				"hex = \"010203\"\n" +
				"[nestedid]\n" +
				"stringvar = \"otherstringvalue\"\n" +
				"int = 42\n",
			conf: Conf{
				FlagEnable:   true,
				EnvEnable:    true,
				FileEnable:   true,
				FileEncoding: "toml",
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
			desc: "TestStruct_flags_env",
			args: []string{"-b",
				"--int16var", "42",
				"--float", "-0.25",
				"--nestedid.int", "42",
				"-h", "otherstringvalue",
				"--strings1", "one", "--strings1", "two", "--strings1", "three",
				"--ints", "3", "--ints", "2", "--ints", "1",
				"--upper1", "TEST",
				"--hex", "010203",
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
				FlagEnable: true,
				EnvEnable:  true,
				EnvPrefix:  "PREF_",
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
				assert.EqualValues(t, 42, c.Uint64Var)
				assert.EqualValues(t, 42, c.Int32Var)
				assert.EqualValues(t, "otherstringvalue", c.Nested.StringVar)
				assert.EqualValues(t, true, c.Nested.BoolVar1)
				assert.EqualValues(t, 42, c.Nested.IntVar)
				slice123 := []string{"one", "two", "three"}
				assert.EqualValues(t, slice123, c.Strings1)
				assert.EqualValues(t, slice123, c.Strings2)
				assert.EqualValues(t, []int{3, 2, 1}, c.Ints1)
				assert.EqualValues(t, []int{42, 43}, c.Ints2)
				assert.EqualValues(t, "test", c.Marshaled.String())
				assert.EqualValues(t, "010203", c.HexData.String())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Set command line args.
			os.Args = append([]string{"test"}, tc.args...)

			// Set environment.
			os.Clearenv()
			for k, v := range tc.env {
				os.Setenv(k, v)
			}

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
			conf.FileDirectory, conf.FileDefaultFilename = path.Split(filename)

			require.NoError(t, Load(tc.config, conf))

			tc.validate(t, tc.config)
		})
	}
}
