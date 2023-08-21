package cmdutil

import (
	"context"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Decoder decodes an env file.
type Decoder interface {
	Decode() (map[string]string, error)
}

// Populate populates an object with environment variables.
//
// The environment has precedence over the decoders, earlier
// decoders have precedence over later decoders.
func Populate(object interface{}, decoders ...Decoder) error {
	return populate(object, decoders)
}

// Main runs the common functionality needed in a go main function.
// appEnv will be populated and passed to do, defaultEnv can be nil
// if there is an error, os.Exit(1) will be called.
func Main[T any](ctx context.Context, do func(context.Context, T) error, appEnv T, decoders ...Decoder) {
	if err := Populate(appEnv, decoders...); err != nil {
		mainError(err)
	}
	if err := do(ctx, appEnv); err != nil {
		mainError(err)
	}
	os.Exit(0)
}

func mainError(err error) {
	ErrorAndExitf("%v\n", err)
}

const (
	cannotParseErr              = "cannot parse"
	envKeyNotSetWhenRequiredErr = "env key not set when required"
	expectedPointerErr          = "expected pointer"
	expectedStructErr           = "expected struct"
	fieldTypeNotAllowedErr      = "field type not allowed"
	invalidTagErr               = "invalid tag, must be KEY,{required},{default=DEFAULT_VALUE}"
)

// These are structs types that we unmarshal directly, without recursing into them.
var knownStructs = map[reflect.Type]func(string) (any, error){
	reflect.TypeOf(resource.Quantity{}): func(x string) (any, error) {
		return resource.ParseQuantity(x) //nolint:wrapcheck
	},
}

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
		return errors.Errorf("%s: %v", expectedPointerErr, reflectValue.Type())
	}
	if reflectValue.Type().Kind() != reflect.Struct {
		return errors.Errorf("%s: %v", expectedStructErr, reflectValue.Type())
	}

	for i := 0; i < reflectValue.NumField(); i++ {
		structField := reflectValue.Type().Field(i)
		ptrToStruct := structField.Type.Kind() == reflect.Ptr && structField.Type.Elem().Kind() == reflect.Struct
		if structField.Type.Kind() == reflect.Struct || ptrToStruct {
			if _, ok := knownStructs[structField.Type]; !ok {
				if err := populateInternal(reflectValue.Field(i), decoderMap, true); err != nil {
					return err
				}
				continue
			}
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
				return errors.Errorf("%s: %s %v", envKeyNotSetWhenRequiredErr, envTag.key, reflectValue.Type())
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

// PopulateDefaults will parse the tags of the given structure and populate each
// field with a default value (if specified in the tags). This is meant for use
// by tests, which do not want to read from env vars.
func PopulateDefaults(object interface{}) error {
	return populateDefaultsInternal(reflect.ValueOf(object), false)
}

func populateDefaultsInternal(reflectValue reflect.Value, recursive bool) error {
	if reflectValue.Type().Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	} else if !recursive {
		return errors.Errorf("%s: %v", expectedPointerErr, reflectValue.Type())
	}
	if reflectValue.Type().Kind() != reflect.Struct {
		return errors.Errorf("%s: %v", expectedStructErr, reflectValue.Type())
	}

	for i := 0; i < reflectValue.NumField(); i++ {
		structField := reflectValue.Type().Field(i)
		ptrToStruct := structField.Type.Kind() == reflect.Ptr && structField.Type.Elem().Kind() == reflect.Struct
		if structField.Type.Kind() == reflect.Struct || ptrToStruct {
			if _, ok := knownStructs[structField.Type]; !ok {
				if err := populateDefaultsInternal(reflectValue.Field(i), true); err != nil {
					return err
				}
				continue
			}
		}
		envTag, err := getEnvTag(structField)
		if err != nil {
			return err
		}
		if envTag == nil || envTag.defaultValue == "" {
			continue
		}
		parsedValue, err := parseField(structField, envTag.defaultValue)
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
			return nil, errors.EnsureStack(err)
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
	split = strings.SplitN(strings.TrimSpace(split[1]), "=", 2)
	switch split[0] {
	case "required":
		envTag.required = true
	case "default":
		if len(split) != 2 {
			return nil, errors.Errorf("%s: %s", invalidTagErr, tag)
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
			return nil, errors.Wrapf(err, cannotParseErr)
		}
		return parsedValue, nil
	case reflect.Int:
		parsedValue, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return nil, errors.Wrapf(err, cannotParseErr)
		}
		return int(parsedValue), nil
	case reflect.Int8:
		parsedValue, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return nil, errors.Wrapf(err, cannotParseErr)
		}
		return int8(parsedValue), nil
	case reflect.Int16:
		parsedValue, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return nil, errors.Wrapf(err, cannotParseErr)
		}
		return int16(parsedValue), nil
	case reflect.Int32:
		parsedValue, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, errors.Wrapf(err, cannotParseErr)
		}
		return int32(parsedValue), nil

	case reflect.Int64:
		parsedValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, cannotParseErr)
		}
		return int64(parsedValue), nil
	case reflect.Uint:
		parsedValue, err := strconv.ParseUint(value, 10, 0)
		if err != nil {
			return nil, errors.Wrapf(err, cannotParseErr)
		}
		return uint(parsedValue), nil
	case reflect.Uint8:
		parsedValue, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return nil, errors.Wrapf(err, cannotParseErr)
		}
		return uint8(parsedValue), nil
	case reflect.Uint16:
		parsedValue, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return nil, errors.Wrapf(err, cannotParseErr)
		}
		return uint16(parsedValue), nil
	case reflect.Uint32:
		parsedValue, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return nil, errors.Wrapf(err, cannotParseErr)
		}
		return uint32(parsedValue), nil

	case reflect.Uint64:
		parsedValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, cannotParseErr)
		}
		return uint64(parsedValue), nil
	case reflect.Float32:
		parsedValue, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return nil, errors.Wrapf(err, cannotParseErr)
		}
		return float32(parsedValue), nil

	case reflect.Float64:
		parsedValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, errors.Wrapf(err, cannotParseErr)
		}
		return float64(parsedValue), nil

	case reflect.String:
		return value, nil

	case reflect.Struct:
		// Already checked by the caller.
		parser := knownStructs[structField.Type]
		v, err := parser(value)
		if err != nil {
			return nil, errors.Wrapf(err, cannotParseErr)
		}
		return v, nil
	default:
		return nil, errors.Errorf("%s: %v", fieldTypeNotAllowedErr, fieldKind)
	}
}
