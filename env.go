// Copyright (c) 2017 Steven Roose <steven@stevenroose.org>.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gonfig

import (
	"os"
	"strings"
)

// getEnvVar reads the environment variable by an option's fullId and prefix
// by joining all parts together with underscores and putting all to upper case.
func getEnvVar(prefix string, fullID []string) (string, bool) {
	key := strings.Join(fullID, "_")
	key = strings.Replace(key, "-", "_", -1)
	key = prefix + key
	key = strings.ToUpper(key)

	return os.LookupEnv(key)
}

// parseEnv parses the environment variables for all config options
// and writes the values that have been found in place.
func parseEnv(s *setup) error {
	for _, opt := range s.allOpts {
		if opt.isParent {
			continue
		}

		value, set := getEnvVar(s.conf.EnvPrefix, opt.fullIDParts)
		if !set {
			continue
		}

		if err := opt.setValueByString(value); err != nil {
			return err
		}
	}

	return nil
}

// lookupConfigFileEnv looks for the config file in the environment variables.
func lookupConfigFileEnv(s *setup, configOpt *option) (string, error) {
	val, found := getEnvVar(s.conf.EnvPrefix, configOpt.fullIDParts)
	if !found {
		return "", nil
	}

	return val, nil
}
