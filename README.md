gonfig
======

Not being very happy with the current options for configurating Go programs, 
I created gonfig with the following two promises in mind:

1. The configuration variables are fully specified in a struct.
2. Next to this struct, you only need one statement to load the configuration.

It has the following features:

- loading config in struct
- having default values
- reading from command line flags
- reading from configuration files in TOML, YAML or JSON
- reading from environment variables
- printing help message
- supported types for interpreting:
  - native Go types: all `int`, `uint`, `string`, `bool`
  - types that implement `TextUnmarshaler` from the "encoding" package
  - slices of the above mentioned types
- config file location can be provided as environment variable or command line 
  flag

Intended usage:

```go
// config here is created inline.  You can also perfectly define a type for it:
//   type Config struct {
//       StringSetting string `id:"stringsetting",short:"s",default:"myString!",desc:"Value for the string"`
//   }
//   var config Config
var config = struct{
	StringSetting string `id:"stringsetting" short:"s" default:"myString!" desc:"Value for the string"`
	IntSetting    int    `id:"intsetting" short:"i" desc:"Value for the int"`
}{
	IntSetting: 42, // alternative way to set defaults
}

func main() {
	err := gonfig.Load(&config, gonfig.Conf{
		ConfigFileFlag: "configfile",
		ConfigFileFlagShort: "c",

		FileEnable: true,
		FileName: "myapp.conf",
		FileType: "yaml", // json, toml
		FileDirectory: "~/.config/myapp",

		FlagEnable: true,

		EnvEnable: true,
		EnvPrefix: "MYAPP_",

		HelpEnable: true,
	})
}
```
