package gonfig

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFile_FileNotExist(t *testing.T) {
	require.Error(t, parseFile(&setup{
		configFilePath: "/doesntexist.conf",
	}))
}

func TestParseFile_InvalidEncoding(t *testing.T) {
	file, err := ioutil.TempFile("", "gonfig")
	require.NoError(t, err)

	require.Panics(t, func() {
		parseFile(&setup{
			configFilePath: file.Name(),
			conf: &Conf{
				FileEncoding: "nonexistent",
			},
		})
	})
}

func TestParseFile_InvalidJSON(t *testing.T) {
	file, err := ioutil.TempFile("", "gonfig")
	require.NoError(t, err)

	_, err = file.WriteString(`{
		"key": "value",
	}`)
	require.NoError(t, err)

	require.Error(t, parseFile(&setup{
		configFilePath: file.Name(),
		conf: &Conf{
			FileEncoding: "json",
		},
	}))
}

func TestParseFile_InvalidYAML(t *testing.T) {
	file, err := ioutil.TempFile("", "gonfig")
	require.NoError(t, err)

	_, err = file.WriteString("test: \"value\n")
	require.NoError(t, err)

	require.Error(t, parseFile(&setup{
		configFilePath: file.Name(),
		conf: &Conf{
			FileEncoding: "yaml",
		},
	}))
}

func TestParseFile_InvalidTOML(t *testing.T) {
	file, err := ioutil.TempFile("", "gonfig")
	require.NoError(t, err)

	_, err = file.WriteString("test = value\n")
	require.NoError(t, err)

	require.Error(t, parseFile(&setup{
		configFilePath: file.Name(),
		conf: &Conf{
			FileEncoding: "toml",
		},
	}))
}
