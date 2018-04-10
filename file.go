// Copyright (c) 2017 Steven Roose <steven@stevenroose.org>.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gonfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
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
			if casted, ok := val.(map[string]interface{}); ok {
				if err := parseMapOpts(casted, opt.subOpts); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("error parsing config file: "+
					"value of type %v given for composite config var %v",
					reflect.TypeOf(val), opt.fullID())
			}
		} else {
			if err := setValue(opt.value, reflect.ValueOf(val)); err != nil {
				return fmt.Errorf("failed to set option '%v': %v",
					opt.fullID(), err)
			}
		}
	}

	return nil
}

// parseFileContent parses the config file given its content.
func parseFileContent(s *setup, content []byte) error {
	decoder := s.conf.FileDecoder
	if decoder == nil {
		// Look for the config file extension to determine the encoding.
		switch path.Ext(s.configFilePath) {
		case "json":
			decoder = DecoderJSON
		case "toml":
			decoder = DecoderTOML
		case "yaml", "yml":
			decoder = DecoderYAML
		default:
			decoder = DecoderTryAll
		}
	}

	m, err := decoder(content)
	if err != nil {
		return fmt.Errorf("failed to parse file at %v: %v",
			s.configFilePath, err)
	}

	// Parse the map for the options.
	if err := parseMapOpts(m, s.opts); err != nil {
		return fmt.Errorf("error loading config vars from config file: %v", err)
	}

	return nil
}

// parseFile parses the config file for all config options by delegating
// the call to the method specific to the config file encoding specified.
func parseFile(s *setup) error {
	if _, err := os.Stat(s.configFilePath); os.IsNotExist(err) {
		// Config file is not present.  We ignore this when we are using
		// the default config file, but we escalate if the user provided
		// the config file explicitely.
		if s.customConfigFile {
			return fmt.Errorf(
				"config file at %v does not exist", s.configFilePath)
		} else {
			return nil
		}
	}

	content, err := ioutil.ReadFile(s.configFilePath)
	if err != nil {
		return fmt.Errorf(
			"error reading config file at %v: %v", s.configFilePath, err)
	}

	return parseFileContent(s, content)
}
