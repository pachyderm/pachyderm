package pkgmap

import (
	"fmt"
	"strconv"
)

// StringStringMap is a map from strings to strings.
type StringStringMap map[string]string

// GetString gets the string if it is present, or "" otherwise.
func (s StringStringMap) GetString(key string) (string, error) {
	value, ok := s[key]
	if !ok {
		return "", nil
	}
	return value, nil
}

// GetRequiredString gets the string if it is present, or returns error otherwise.
func (s StringStringMap) GetRequiredString(key string) (string, error) {
	value, ok := s[key]
	if !ok {
		return "", newRequiredError(key)
	}
	return value, nil
}

// GetInt32 gets the int32 if it is present, or 0 otherwise.
func (s StringStringMap) GetInt32(key string) (int32, error) {
	return s.getInt32(key, false)
}

// GetRequiredInt32 gets the int32 if it is present, or returns error otherwise.
func (s StringStringMap) GetRequiredInt32(key string) (int32, error) {
	return s.getInt32(key, true)
}

func (s StringStringMap) getInt32(key string, required bool) (int32, error) {
	value, ok := s[key]
	if !ok {
		if required {
			return 0, newRequiredError(key)
		}
		return 0, nil
	}
	parsedValue, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(parsedValue), nil
}

// GetInt64 gets the int64 if it is present, or 0 otherwise.
func (s StringStringMap) GetInt64(key string) (int64, error) {
	return s.getInt64(key, false)
}

// GetRequiredInt64 gets the int64 if it is present, or returns error otherwise.
func (s StringStringMap) GetRequiredInt64(key string) (int64, error) {
	return s.getInt64(key, true)
}

func (s StringStringMap) getInt64(key string, required bool) (int64, error) {
	value, ok := s[key]
	if !ok {
		if required {
			return 0, newRequiredError(key)
		}
		return 0, nil
	}
	parsedValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return int64(parsedValue), nil
}

// GetUint32 gets the uint32 if it is present, or 0 otherwise.
func (s StringStringMap) GetUint32(key string) (uint32, error) {
	return s.getUint32(key, false)
}

// GetRequiredUint32 gets the uint32 if it is present, or returns error otherwise.
func (s StringStringMap) GetRequiredUint32(key string) (uint32, error) {
	return s.getUint32(key, true)
}

func (s StringStringMap) getUint32(key string, required bool) (uint32, error) {
	value, ok := s[key]
	if !ok {
		if required {
			return 0, newRequiredError(key)
		}
		return 0, nil
	}
	parsedValue, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(parsedValue), nil
}

// GetUint64 gets the uint64 if it is present, or 0 otherwise.
func (s StringStringMap) GetUint64(key string) (uint64, error) {
	return s.getUint64(key, false)
}

// GetRequiredUint64 gets the uint64 if it is present, or returns error otherwise.
func (s StringStringMap) GetRequiredUint64(key string) (uint64, error) {
	return s.getUint64(key, true)
}

func (s StringStringMap) getUint64(key string, required bool) (uint64, error) {
	value, ok := s[key]
	if !ok {
		if required {
			return 0, newRequiredError(key)
		}
		return 0, nil
	}
	parsedValue, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint64(parsedValue), nil
}

// GetBool gets the bool if it is present, or 0 otherwise.
func (s StringStringMap) GetBool(key string) (bool, error) {
	value, ok := s[key]
	if !ok {
		return false, nil
	}
	return strconv.ParseBool(value)
}

// GetRequiredBool gets the bool if it is present, or returns error otherwise.
func (s StringStringMap) GetRequiredBool(key string) (bool, error) {
	value, ok := s[key]
	if !ok {
		return false, newRequiredError(key)
	}
	return strconv.ParseBool(value)
}

// Copy gets a copy of the StringStringMap.
func (s StringStringMap) Copy() StringStringMap {
	m := make(StringStringMap)
	for key, value := range s {
		m[key] = value
	}
	return m
}

func newRequiredError(key string) error {
	return fmt.Errorf("pkgmap: key %s required", key)
}
