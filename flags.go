// Copyright (c) 2017 Steven Roose <steven@stevenroose.org>.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gonfig

import (
	"fmt"
	"os"
	"strings"
)

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
				"unexpected word while parsing flags: '%v'", arg)
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

		if opt.isMap {
			// An exception for maps, we need to look for all prefixed flags.
			for flag, value := range flagsMap {
				if strings.HasPrefix(flag, opt.fullID()+".") {
					key := strings.TrimPrefix(flag, opt.fullID()+".")
					if err := setSimpleMapValue(opt.value, key, value); err != nil {
						return fmt.Errorf(
							"error parsing map value '%v' for config var %v: %v",
							value, opt.fullID(), err)
					}
					delete(flagsMap, flag)
				}
			}
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
		} else if shortSet {
			stringValue = shortValue
		}

		if err := setValueByString(opt.value, stringValue); err != nil {
			return fmt.Errorf("error parsing flag value '%v' of option '%v': %v",
				stringValue, opt.fullID(), err)
		}
	}

	if !s.conf.FlagIgnoreUnknown {
		// error if there is still something left
		for flag := range flagsMap {
			return fmt.Errorf("unknown flag: %v", flag)
		}
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
