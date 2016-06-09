package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	cannotParseErr                     = "cannot parse"
	cannotSetBothRequiredAndDefaultErr = "cannot set both required and default"
	duplicateRestrictToKeyErr          = "duplicate restrict to key"
	envKeyNotSetWhenRequiredErr        = "env key not set when required"
	expectedPointerErr                 = "expected pointer"
	expectedStructErr                  = "expected struct"
	fieldTypeNotAllowedErr             = "field type not allowed"
	invalidTagErr                      = "invalid tag, must be KEY,{required},{default=DEFAULT_VALUE}"
)

func populate(object interface{}, decoders []Decoder) error {
	decoderMap, err := getDecoderMap(decoders)
	if err != nil {
		return err
	}
	return populateInternal(reflect.ValueOf(object), decoderMap, false)
}

func populateInternal(reflectValue reflect.Value, decoderMap map[string]string, recursive bool) error {
	if reflectValue.Type().Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	} else if !recursive {
		return fmt.Errorf("%s: %v", expectedPointerErr, reflectValue.Type())
	}
	if reflectValue.Type().Kind() != reflect.Struct {
		return fmt.Errorf("%s: %v", expectedStructErr, reflectValue.Type())
	}
	numField := reflectValue.NumField()
	for i := 0; i < numField; i++ {
		structField := reflectValue.Type().Field(i)
		if structField.Type.Kind() == reflect.Struct {
			if err := populateInternal(reflectValue.Field(i), decoderMap, true); err != nil {
				return err
			}
			continue
		}
		envTag, err := getEnvTag(structField)
		if err != nil {
			return err
		}
		if envTag == nil {
			continue
		}
		value := getValue(envTag.key, envTag.defaultValue, decoderMap)
		if value == "" {
			if envTag.required {
				return fmt.Errorf("%s: %s %v", envKeyNotSetWhenRequiredErr, envTag.key, reflectValue.Type())
			}
			continue
		}
		parsedValue, err := parseField(structField, value)
		if err != nil {
			return err
		}
		reflectValue.Field(i).Set(reflect.ValueOf(parsedValue))
	}
	return nil
}

func getDecoderMap(decoders []Decoder) (map[string]string, error) {
	env := make(map[string]string)
	for _, decoder := range decoders {
		subEnv, err := decoder.Decode()
		if err != nil {
			return nil, err
		}
		for key, value := range subEnv {
			if value != "" {
				if _, ok := env[key]; !ok {
					env[key] = value
				}
			}
		}
	}
	return env, nil
}

func getValue(key string, defaultValue string, decoderMap map[string]string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	value = decoderMap[key]
	if value != "" {
		return value
	}
	return defaultValue
}

type envTag struct {
	key          string
	required     bool
	defaultValue string
}

func getEnvTag(structField reflect.StructField) (*envTag, error) {
	tag := structField.Tag.Get("env")
	if tag == "" {
		return nil, nil
	}
	split := strings.SplitN(tag, ",", 2)
	envTag := &envTag{
		key: split[0],
	}
	if len(split) == 1 {
		return envTag, nil
	}
	split = strings.SplitN(split[1], "=", 2)
	switch split[0] {
	case "required":
		envTag.required = true
	case "default":
		if len(split) != 2 {
			return nil, fmt.Errorf("%s: %s", invalidTagErr, tag)
		}
		envTag.defaultValue = split[1]
	}
	return envTag, nil
}

func parseField(structField reflect.StructField, value string) (interface{}, error) {
	fieldKind := structField.Type.Kind()
	switch fieldKind {
	case reflect.Bool:
		if value == "" {
			return false, nil
		}
		parsedValue, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", cannotParseErr, err.Error())
		}
		return parsedValue, nil
	case reflect.Int:
		parsedValue, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", cannotParseErr, err.Error())
		}
		return int(parsedValue), nil
	case reflect.Int8:
		parsedValue, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", cannotParseErr, err.Error())
		}
		return int8(parsedValue), nil
	case reflect.Int16:
		parsedValue, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", cannotParseErr, err.Error())
		}
		return int16(parsedValue), nil
	case reflect.Int32:
		parsedValue, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", cannotParseErr, err.Error())
		}
		return int32(parsedValue), nil

	case reflect.Int64:
		parsedValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", cannotParseErr, err.Error())
		}
		return int64(parsedValue), nil
	case reflect.Uint:
		parsedValue, err := strconv.ParseUint(value, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", cannotParseErr, err.Error())
		}
		return uint(parsedValue), nil
	case reflect.Uint8:
		parsedValue, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", cannotParseErr, err.Error())
		}
		return uint8(parsedValue), nil
	case reflect.Uint16:
		parsedValue, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", cannotParseErr, err.Error())
		}
		return uint16(parsedValue), nil
	case reflect.Uint32:
		parsedValue, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", cannotParseErr, err.Error())
		}
		return uint32(parsedValue), nil

	case reflect.Uint64:
		parsedValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", cannotParseErr, err.Error())
		}
		return uint64(parsedValue), nil
	case reflect.Float32:
		parsedValue, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", cannotParseErr, err.Error())
		}
		return float32(parsedValue), nil

	case reflect.Float64:
		parsedValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", cannotParseErr, err.Error())
		}
		return float64(parsedValue), nil
	case reflect.String:
		return value, nil
	default:
		return nil, fmt.Errorf("%s: %v", fieldTypeNotAllowedErr, fieldKind)
	}
}
