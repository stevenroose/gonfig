package gonfig

import (
	"os"
	"strings"
)

// getEnvVar reads the environment variable by an option's fullId and prefix
// by joining all parts together with underscores and putting all to upper case.
func getEnvVar(prefix string, fullId []string) (string, bool) {
	key := strings.Join(fullId, "_")
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

		value, set := getEnvVar(s.conf.EnvPrefix, opt.fullIdParts)
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
	val, found := getEnvVar(s.conf.EnvPrefix, configOpt.fullIdParts)
	if !found {
		return "", nil
	}

	return val, nil
}
