[![Build Status](https://travis-ci.org/stevenroose/gonfig.svg?branch=master)](https://travis-ci.org/stevenroose/gonfig)
[![Coverage Status](https://coveralls.io/repos/github/stevenroose/gonfig/badge.svg?branch=master)](https://coveralls.io/github/stevenroose/gonfig?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/stevenroose/gonfig)](https://goreportcard.com/report/github.com/stevenroose/gonfig)
[![GoDoc](https://godoc.org/github.com/stevenroose/gonfig?status.svg)](https://godoc.org/github.com/stevenroose/gonfig)


Description
===========

gonfig is a configuration library designed using the following principles:

1. The configuration variables are fully specified and loaded into  a struct.
2. You only need one statement to load the configuration fully.
3. Configuration variables can be retrieved from various sources, in this order
   of priority:
   - default values
   - config file in either YAML, TOML or JSON
   - environment variables
   - command line flags

Furthermore, it has the following features:

- supported types for interpreting:
  - native Go types: all `int`, `uint`, `string`, `bool`
  - types that implement `TextUnmarshaler` from the "encoding" package
  - byte slices are interpreted as base64
  - slices of the above mentioned types

- the location of the config file can be passed through command line flags or
  environment variables

- printing help message


Usage
=====

```go
var config = struct{
	StringSetting string `id:"stringsetting" short:"s" default:"myString!" desc:"Value for the string"`
	IntSetting    int    `id:"intsetting" short:"i" desc:"Value for the int"`

	ConfigFile    string `short:"c"`
}{
	IntSetting: 42, // alternative way to set default values
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
