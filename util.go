package gonfig

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func parseInt(v reflect.Value, s string) error {
	var bitSize int
	switch v.Type().Kind() {
	case reflect.Int:
		bitSize = 0
	case reflect.Int64:
		bitSize = 64
	case reflect.Int32:
		bitSize = 32
	case reflect.Int16:
		bitSize = 16
	case reflect.Int8:
		bitSize = 8
	default:
		panic("not an int")
	}
	p, err := strconv.ParseInt(s, 10, bitSize)
	if err != nil {
		return parseError(s, v.Type(), err)
	}
	v.SetInt(p)
	return nil
}

func parseUint(v reflect.Value, s string) error {
	var bitSize int
	switch v.Type().Kind() {
	case reflect.Uint:
		bitSize = 0
	case reflect.Uint64:
		bitSize = 64
	case reflect.Uint32:
		bitSize = 32
	case reflect.Uint16:
		bitSize = 16
	case reflect.Uint8:
		bitSize = 8
	default:
		panic("not a uint")
	}
	p, err := strconv.ParseUint(s, 10, bitSize)
	if err != nil {
		return parseError(s, v.Type(), err)
	}
	v.SetUint(p)
	return nil
}

func parseFloat(v reflect.Value, s string) error {
	var bitSize int
	switch v.Type().Kind() {
	case reflect.Float32:
		bitSize = 32
	case reflect.Float64:
		bitSize = 64
	default:
		panic("not a float")
	}
	p, err := strconv.ParseFloat(s, bitSize)
	if err != nil {
		return parseError(s, v.Type(), err)
	}
	v.SetFloat(p)
	return nil
}

func parseSimpleValue(v reflect.Value, s string) error {
	if v.Type() == typeOfByteSlice {
		decoded, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return parseError(s, v.Type(), err)
		}
		v.Set(reflect.ValueOf(decoded))
		return nil
	}

	switch v.Type().Kind() {
	case reflect.String:
		v.SetString(s)

	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return parseError(s, v.Type(), err)
		}
		v.SetBool(b)

	case reflect.Float32, reflect.Float64:
		if err := parseFloat(v, s); err != nil {
			return err
		}

	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		if err := parseInt(v, s); err != nil {
			return err
		}

	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		if err := parseUint(v, s); err != nil {
			return err
		}

	default:
		panic("not simple value")
	}

	return nil
}

func parseSlice(v reflect.Value, s string) error {
	vals, err := readAsCSV(s)
	if err != nil {
		return fmt.Errorf("error parsing comma separated value '%s': %s", s, err)
	}

	slice := reflect.MakeSlice(v.Type(), len(vals), len(vals))
	for i := 0; i < len(vals); i++ {
		if err := parseSimpleValue(slice.Index(i), vals[i]); err != nil {
			return err
		}
	}

	v.Set(slice)
	return nil
}

func cleanUpYAML(v interface{}) interface{} {
	switch v := v.(type) {

	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range v {
			result[k] = cleanUpYAML(v)
		}
		return result

	case map[interface{}]interface{}:
		result := make(map[string]interface{})
		for k, v := range v {
			result[fmt.Sprintf("%v", k)] = cleanUpYAML(v)
		}
		return result

	case []interface{}:
		result := make([]interface{}, len(v))
		for i, v := range v {
			result[i] = cleanUpYAML(v)
		}
		return result

	default:
		return v
	}
}

func readAsCSV(val string) ([]string, error) {
	if val == "" {
		return []string{}, nil
	}
	stringReader := strings.NewReader(val)
	csvReader := csv.NewReader(stringReader)
	return csvReader.Read()
}

func writeAsCSV(vals []string) (string, error) {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)
	err := w.Write(vals)
	if err != nil {
		return "", err
	}
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n"), nil
}
