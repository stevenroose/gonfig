// Copyright (c) 2017 Steven Roose <steven@stevenroose.org>.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gonfig

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	yaml "gopkg.in/yaml.v2"
)

// FileDecoderFn represents a method that translates the content of a config file
// to a map[string]interface{}.  It is important that in this map, all
// interface{} types are either:
// - a simple Go type (intX, uintX, bool, string, floatXX)
// - a value implementing encoding.TextUnmarshaler
// - a []byte
// - a slice of one of those
// - a map[string]interface{}
type FileDecoderFn func(content []byte) (map[string]interface{}, error)

// DecoderJSON is the JSON decoding function for config files.
var DecoderJSON FileDecoderFn = func(c []byte) (map[string]interface{}, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(c, &m); err != nil {
		return nil, fmt.Errorf("error parsing JSON config file: %s", err)
	}

	return m, nil
}

// DecoderTOML is the TOML decoding function for config files.
var DecoderTOML FileDecoderFn = func(c []byte) (map[string]interface{}, error) {
	var m map[string]interface{}
	if err := toml.Unmarshal(c, &m); err != nil {
		return nil, fmt.Errorf("error parsing TOML config file: %s", err)
	}

	return m, nil
}

// DecoderYAML is the YAML decoding function for config files.
var DecoderYAML FileDecoderFn = func(c []byte) (map[string]interface{}, error) {
	var m map[string]interface{}
	if err := yaml.Unmarshal(c, &m); err != nil {
		return nil, fmt.Errorf("error parsing YAML config file: %s", err)
	}
	// Cast map[interface{}]interface{} to map[string]interface{}.
	m = cleanUpYAML(m).(map[string]interface{})
	return m, nil
}

// DecoderTryAll is an encoding function that tries all other existing encoding
// functions and uses the first one that does not produce an error.
var DecoderTryAll FileDecoderFn = func(c []byte) (map[string]interface{}, error) {
	decoders := []FileDecoderFn{
		DecoderYAML,
		DecoderTOML,
		DecoderJSON,
	}

	errs := make([]string, len(decoders))
	for i, decoder := range decoders {
		m, err := decoder(c)
		if err == nil {
			return m, nil
		}
		errs[i] = err.Error()
	}

	errStr := fmt.Sprintf("[\"%s\"]", strings.Join(errs, "\", \""))
	return nil, fmt.Errorf("config file failed to decode with decoders for "+
		"YAML, TOML and JSON: %s", errStr)
}
