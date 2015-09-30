package dockervolume

import (
	"fmt"
	"strconv"
)

type opts struct {
	m map[string]string
}

func newOpts(m map[string]string) *opts {
	return &opts{m}
}

func (o *opts) GetRequiredString(key string) (string, error) {
	if o.m == nil {
		return "", fmt.Errorf("dockervolume: must pass opt %s (--opt %s=VALUE)", key, key)
	}
	value, ok := o.m[key]
	if !ok {
		return "", fmt.Errorf("dockervolume: must pass opt %s (--opt %s=VALUE)", key, key)
	}
	return value, nil
}

func (o *opts) GetOptionalString(key string, defaultValue string) (string, error) {
	if o.m == nil {
		return defaultValue, nil
	}
	value, ok := o.m[key]
	if !ok {
		return defaultValue, nil
	}
	return value, nil
}

func (o *opts) GetRequiredUInt64(key string) (uint64, error) {
	valueObj, err := o.GetRequiredString(key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(valueObj, 10, 64)
}

func (o *opts) GetOptionalUInt64(key string, defaultValue uint64) (uint64, error) {
	valueObj, err := o.GetOptionalString(key, strconv.FormatUint(defaultValue, 10))
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(valueObj, 10, 64)
}
