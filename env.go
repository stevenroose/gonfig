// Copyright (c) 2017 Steven Roose <steven@stevenroose.org>.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gonfig

import (
	"fmt"
	"os"
	"strings"
)

// makeEnvKey creates the environment variable key with the opts fullId and
// prefix by joining all parts together with underscores and putting all to
// upper case.
func makeEnvKey(prefix string, fullID []string) string {
	key := strings.Join(fullID, "_")
	key = strings.Replace(key, "-", "_", -1)
	key = prefix + key
	key = strings.ToUpper(key)
	return key
}

// parseEnv parses the environment variables for all config options
// and writes the values that have been found in place.
func parseEnv(s *setup) error {
	for _, opt := range s.allOpts {
		if opt.isParent {
			continue
		}

		envKey := makeEnvKey(s.conf.EnvPrefix, opt.fullIDParts)

		if opt.isMap {
			// An exception for maps, we need to look for all prefixed vars.
			pref := envKey + "_"
			for _, env := range os.Environ() {
				split := strings.SplitN(env, "=", 2)
				key, value := split[0], split[1]
				if strings.HasPrefix(key, pref) {
					mapKey := strings.ToLower(strings.TrimPrefix(key, pref))
					if err := setSimpleMapValue(opt.value, mapKey, value); err != nil {
						return fmt.Errorf(
							"error parsing map value '%v' for config var %v: %v",
							value, opt.fullID(), err)
					}
				}
			}
			continue
		}

		value, set := os.LookupEnv(envKey)
		if !set {
			continue
		}

		if err := setValueByString(opt.value, value); err != nil {
			return fmt.Errorf("failed to set option '%v' with value '%v': %v",
				opt.fullID(), value, err)
		}
	}

	return nil
}

// lookupConfigFileEnv looks for the config file in the environment variables.
func lookupConfigFileEnv(s *setup, configOpt *option) (string, error) {
	val, found := os.LookupEnv(makeEnvKey(s.conf.EnvPrefix, configOpt.fullIDParts))
	if !found {
		return "", nil
	}

	return val, nil
}
