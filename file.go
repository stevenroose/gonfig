package gonfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/BurntSushi/toml"

	yaml "gopkg.in/yaml.v2"
)

// parseMapOpts parses options from a map[string]interface{}.  This is used
// for configuration file encodings that can decode to such a map.
func parseMapOpts(j map[string]interface{}, opts []*option) error {
	for _, opt := range opts {
		val, set := j[opt.id]
		if !set {
			continue
		}

		if opt.isParent {
			switch val.(type) {
			case map[string]interface{}:
				casted := val.(map[string]interface{})
				if err := parseMapOpts(casted, opt.subOpts); err != nil {
					return err
				}
			default:
				return fmt.Errorf("error parsing config file: "+
					"value of type %s given for composite config var %s",
					reflect.TypeOf(val), opt.fullId())
			}
		} else {
			if err := opt.setValue(reflect.ValueOf(val)); err != nil {
				return err
			}
		}
	}

	return nil
}

// openConfigFile trues to open the config file at path.  If it fails
// it returns a nice error.
func openConfigFile(path string) ([]byte, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file at %s does not exist", path)
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(
			"error reading config file at %s: %s", path, err)
	}

	return content, nil
}

// parseFile parses the config file for all config options by delegating
// the call to the method specific to the config file encoding specified.
func parseFile(s *setup) error {
	content, err := openConfigFile(s.configFilePath)
	if err != nil {
		return err
	}

	// Decode to a map using the given encoding.
	var m map[string]interface{}
	switch s.conf.FileEncoding {

	case "json":
		if err := json.Unmarshal(content, &m); err != nil {
			return fmt.Errorf("error parsing JSON config file: %s", err)
		}

	case "toml":
		if err := toml.Unmarshal(content, &m); err != nil {
			return fmt.Errorf("error parsing TOML config file: %s", err)
		}

	case "yaml":
		if err := yaml.Unmarshal(content, &m); err != nil {
			return fmt.Errorf("error parsing YAML config file: %s", err)
		}
		// Cast map[interface{}]interface{} to map[string]interface{}.
		m = cleanUpYAML(m).(map[string]interface{})

	default:
		return fmt.Errorf(
			"unknown config file encoding: %s", s.conf.FileEncoding)
	}

	// Parse the map for the options.
	if err := parseMapOpts(m, s.opts); err != nil {
		return err
	}

	return nil
}
