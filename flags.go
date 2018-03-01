// Copyright (c) 2017 Steven Roose <steven@stevenroose.org>.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gonfig

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/spf13/pflag"
)

const (
	defaultHelpDescription = "print this help menu"
	defaultHelpMessage     = "Usage of __EXEC__:"
)

// addFlag adds a new flag to the flagset for the given option.
// It will try to create a flag with the correct type and fallback to string
// for unsupported types.
func addFlag(flagSet *pflag.FlagSet, opt *option) {
	switch opt.value.Type().Kind() {
	case reflect.Bool:
		var def bool
		if opt.defaultSet {
			def = opt.defaultValue.Bool()
		}
		flagSet.BoolP(opt.fullID(), opt.short, def, opt.desc)

	case reflect.Float32, reflect.Float64:
		var def float64
		if opt.defaultSet {
			def = opt.defaultValue.Float()
		}
		flagSet.Float64P(opt.fullID(), opt.short, def, opt.desc)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var def int64
		if opt.defaultSet {
			def = opt.defaultValue.Int()
		}
		flagSet.Int64P(opt.fullID(), opt.short, def, opt.desc)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var def uint64
		if opt.defaultSet {
			def = opt.defaultValue.Uint()
		}
		flagSet.Uint64P(opt.fullID(), opt.short, def, opt.desc)

	case reflect.Slice:
		if opt.value.Type().Elem().Kind() == reflect.Uint8 {
			// Special case for byte slices.
			flagSet.StringP(opt.fullID(), opt.short, opt.defaul, opt.desc)
			break
		}
		switch opt.value.Type().Elem().Kind() {
		case reflect.Bool:
			var def []bool
			if opt.defaultSet {
				def = opt.defaultValue.Interface().([]bool)
			}
			flagSet.BoolSliceP(opt.fullID(), opt.short, def, opt.desc)

		//TODO pflag.FloatSliceP is missing for now
		//case reflect.Float32, reflect.Float64:
		//	var def []float64
		//	if opt.defaultSet {
		//		def = opt.defaultValue.Convert(reflect.TypeOf(def)).Interface().([]float64)
		//	}
		//	flagSet.Float64P(opt.fullID(), opt.short, def, opt.desc)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var def []int
			if opt.defaultSet {
				slice := reflect.New(reflect.TypeOf(def))
				if err := convertSlice(opt.defaultValue, slice.Elem()); err != nil {
					panic(fmt.Sprintf("Error creating flag for option %s: %s",
						opt.fullID(), err))
				}
				def = slice.Elem().Interface().([]int)
			}
			flagSet.IntSliceP(opt.fullID(), opt.short, def, opt.desc)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			var def []uint
			if opt.defaultSet {
				slice := reflect.New(reflect.TypeOf(def))
				if err := convertSlice(opt.defaultValue, slice.Elem()); err != nil {
					panic(fmt.Sprintf("Error creating flag for option %s: %s",
						opt.fullID(), err))
				}
				def = slice.Elem().Interface().([]uint)
			}
			flagSet.UintSliceP(opt.fullID(), opt.short, def, opt.desc)

		case reflect.String:
			fallthrough
		default:
			defSlice, err := readAsCSV(opt.defaul)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing default value '%s' for slice variable %s: %s",
					opt.defaul, opt.fullID(), err))
			}
			flagSet.StringSliceP(opt.fullID(), opt.short, defSlice, opt.desc)
		}

	case reflect.String:
		fallthrough
	default:
		flagSet.StringP(opt.fullID(), opt.short, opt.defaul, opt.desc)
	}
}

// createFlagSet builds the flagset for the options in the setup.
func createFlagSet(s *setup) *pflag.FlagSet {
	flagSet := pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
	flagSet.SortFlags = false

	for _, opt := range s.allOpts {
		if opt.isParent {
			// Parents are skipped, we should only add the children.
			continue
		}

		addFlag(flagSet, opt)
	}

	if !s.conf.HelpDisable {
		desc := s.conf.HelpDescription
		if desc == "" {
			desc = defaultHelpDescription
		}

		flagSet.BoolP("help", "h", false, desc)
	}

	return flagSet
}

// printHelpAndExit prints the help message and exits the program.
func printHelpAndExit(s *setup) {
	message := s.conf.HelpMessage
	if message == "" {
		exec := path.Base(os.Args[0])
		message = strings.Replace(defaultHelpMessage, "__EXEC__", exec, 1)
	}
	fmt.Println(message)

	flagSet := createFlagSet(s)
	fmt.Println(flagSet.FlagUsages())
	os.Exit(2)
}

// flagFromWord interprets if the word could be a cmd like flag (either
// `--word` or `-w`) and returns the word itself (resp. `word` or `w`).
// It returns an empty string if the word cannot be a flag.
func flagFromWord(w string) string {
	if len(w) > 2 && w[0] == '-' && w[1] == '-' {
		return w[2:]
	} else if len(w) == 2 && w[0] == '-' {
		return w[1:2]
	} else {
		return ""
	}
}

// parseFlagsToMap parses the given command line flags into a string.
func parseFlagsToMap(s *setup, args []string) (map[string]string, error) {
	args = args[1:]
	result := map[string]string{}

	var i = 0
	for i < len(args) {
		arg := args[i]

		if arg == "--help" || arg == "-h" {
			printHelpAndExit(s)
		}

		if arg == "--" {
			// separator that indicates end of flags
			return result, nil
		}

		parts := strings.SplitN(arg, "=", 2)
		key := flagFromWord(parts[0])
		if key == "" {
			return nil, fmt.Errorf(
				"unexpected word while parsing flags: '%s'", arg)
		}

		addValue := func(key, newValue string) {
			value, isSet := result[key]
			if isSet {
				value = value + "," + newValue
			} else {
				value = newValue
			}
			result[key] = value
		}

		if len(parts) == 2 {
			addValue(key, parts[1])
			i += 1
			continue
		}

		if len(args) <= i+1 || flagFromWord(args[i+1]) != "" {
			addValue(key, "true")
			i += 1
			continue
		}

		nextWord := args[i+1]
		addValue(key, nextWord)
		i += 2
	}

	return result, nil
}

// parseFlags parses the command line flags for all config options
// and writes the values that have been found in place.
func parseFlags(s *setup) error {
	flagsMap, err := parseFlagsToMap(s, os.Args)
	if err != nil {
		return err
	}

	for _, opt := range s.allOpts {
		if opt.isParent {
			// Parents are skipped, we should only add the children.
			continue
		}

		stringValue, fullSet := flagsMap[opt.fullID()]
		if fullSet {
			delete(flagsMap, opt.fullID())
		}
		shortValue, shortSet := flagsMap[opt.short]
		if shortSet {
			delete(flagsMap, opt.short)
		}

		if !fullSet && !shortSet {
			continue
		} else if fullSet && shortSet {
			return fmt.Errorf("flag is set with both short and full form: %v",
				opt.fullID())
		}

		if shortSet {
			stringValue = shortValue
		}

		if err := opt.setValueByString(stringValue); err != nil {
			return fmt.Errorf("error parsing flag value for %s: %s",
				opt.fullID(), err)
		}
	}

	// error if there is still something left
	for flag := range flagsMap {
		return fmt.Errorf("unknown flag: %s", flag)
	}

	return nil
}

// lookupConfigFileFlag looks for the config file in the command line flags.
func lookupConfigFileFlag(s *setup, configOpt *option) (string, error) {
	flagsMap, err := parseFlagsToMap(s, os.Args)
	if err != nil {
		return "", nil
	}

	return flagsMap[configOpt.fullID()], nil
}
