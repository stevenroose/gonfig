[![Build Status](https://travis-ci.org/stevenroose/gonfig.svg?branch=master)](https://travis-ci.org/stevenroose/gonfig)
[![Coverage Status](https://coveralls.io/repos/github/stevenroose/gonfig/badge.svg?branch=master)](https://coveralls.io/github/stevenroose/gonfig?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/stevenroose/gonfig)](https://goreportcard.com/report/github.com/stevenroose/gonfig)
[![GoDoc](https://godoc.org/github.com/stevenroose/gonfig?status.svg)](https://godoc.org/github.com/stevenroose/gonfig)


Description
===========

gonfig is a configuration library designed using the following principles:

1. The configuration variables are fully specified and loaded into a struct 
   variable.
2. You only need one statement to load the configuration fully.
3. Configuration variables can be retrieved from various sources, in this order
   of increasing priority:
   - default values from the struct definition
   - the value already in the object when passed into `Load()`
   - config file in either YAML, TOML, JSON or a custom decoder
   - environment variables
   - command line flags

Furthermore, it has the following features:

- supported types for interpreting:
  - native Go types: all `int`, `uint`, `string`, `bool`
  - types that implement `TextUnmarshaler` from the "encoding" package
  - byte slices (`[]byte`) are interpreted as base64
  - slices of the above mentioned types
  - `map[string]interface{}`

- the location of the config file can be passed through command line flags or
  environment variables

- printing help message (and hiding individual flags)


Documentation
=============

Documentation can be found on godoc.org: https://godoc.org/github.com/stevenroose/gonfig

```go
// Load loads the configuration of your program in the struct at c.
// Use conf to specify how gonfig should look for configuration variables.
// This method can panic if there was a problem in the configuration struct that
// is used (which should not happen at runtime), but will always try to produce
// an error instead if the user provided incorrect values.
//
// The recognised tags on the exported struct variables are:
//  - id: the keyword identifier (defaults to lowercase of variable name)
//  - default: the default value of the variable
//  - short: the shorthand used for command line flags (like -h)
//  - desc: the description of the config var, used in --help
//  - opts: comma-separated flags.  Supported flags are:
//     - hidden: Hides the option from help outputs.
func Load(c interface{}, conf Conf) error

// Conf is used to specify the intended behavior of gonfig.
type Conf struct {
	// ConfigFileVariable is the config variable that will be read before looking
	// for a config file.  If no value is specified in the environment variables
	// of the command line flags, the default config file will be read.
	// This flag should be defined in the config file struct and referred to here
	// by its ID.  The default value for this variable is obviously ignored.
	ConfigFileVariable string

	// FileDisable disabled reading config variables from the config file.
	FileDisable bool
	// FileDefaultFilename is the default filename to look for for the config
	// file.  If this is empty and no filename is explicitly provided, parsing
	// a config file is skipped.
	FileDefaultFilename string
	// FileDecoder specifies the decoder function to be used for decoding the
	// config file.  The following decoders are provided, but the user can also
	// specify a custom decoder function:
	//  - DecoderYAML
	//  - DecoderTOML
	//  - DecoderJSON
	// If no decoder function is provided, gonfig tries to guess the function
	// based on the file extension and otherwise tries them all in the above
	// mentioned order.
	FileDecoder FileDecoderFn

	// FlagDisable disabled reading config variables from the command line flags.
	FlagDisable bool
	// FlagIgnoreUnknown ignores unknown command line flags instead of stopping
	// with an error message.
	FlagIgnoreUnknown bool

	// EnvDisables disables reading config variables from the environment
	// variables.
	EnvDisable bool
	// EnvPrefix is the prefix to use for the the environment variables.
	// gonfig does not add an underscore after the prefix.
	EnvPrefix string

	// HelpDisable disables printing the help message when the --help or -h flag
	// is provided.  If this is false, an explicit --help flag will be added.
	HelpDisable bool
	// HelpMessage is the message printed before the list of the flags when the
	// user sets the --help flag.
	// The default is "Usage of [executable name]:".
	HelpMessage string
	// HelpDescription is the description to print for the help flag.
	// By default, this is "show this help menu".
	HelpDescription string
}
```

Usage
=====

```go
var config = struct{
	StringSetting string `id:"stringsetting" short:"s" default:"myString!" desc:"Value for the string"`
	IntSetting    int    `id:"intsetting" short:"i" desc:"Value for the int"`

	ConfigFile    string `short:"c"`
}{
	// alternative way to set default values; they overwrite the ones in the struct
	IntSetting: 42, 
}
// config here is created inline.  You can also perfectly define a type for it:
//   type Config struct {
//       StringSetting string `id:"str",short:"s",default:"myString!",desc:"Value for the string"`
//   }
//   var config Config

func main() {
	err := gonfig.Load(&config, gonfig.Conf{
		ConfigFileVariable: "configfile", // enables passing --configfile myfile.conf

		FileDefaultFilename: "myapp.conf",
		FileDecoder: gonfig.DecoderTOML,

		EnvPrefix: "MYAPP_",
	})
}
```

License
=======

gonfig is licensed by an MIT license as can be found in the LICENSE file.
