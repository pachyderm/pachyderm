package env

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
)

type internalDecoder interface {
	decode() (map[string]string, error)
}

type decoder struct {
	internalDecoder

	once  *sync.Once
	value *atomic.Value
	err   *atomic.Value
}

func newDecoder(internalDecoder internalDecoder) *decoder {
	return &decoder{
		internalDecoder,
		&sync.Once{},
		&atomic.Value{},
		&atomic.Value{},
	}
}

func (d *decoder) Decode() (map[string]string, error) {
	d.once.Do(func() {
		value, err := d.internalDecoder.decode()
		if value != nil {
			d.value.Store(value)
		}
		if err != nil {
			d.err.Store(err)
		}
	})
	valueObj := d.value.Load()
	errObj := d.err.Load()
	var value map[string]string
	var err error
	if valueObj != nil {
		value = valueObj.(map[string]string)
	}
	if errObj != nil {
		err = errObj.(error)
	}
	return value, err
}

type envFileDecoder struct {
	reader io.Reader
}

func newEnvFileDecoder(reader io.Reader) *decoder {
	return newDecoder(
		&envFileDecoder{
			reader,
		},
	)
}

func (e *envFileDecoder) decode() (map[string]string, error) {
	bufReader := bufio.NewReader(e.reader)
	env := make(map[string]string)
	for line, err := bufReader.ReadString('\n'); true; line, err = bufReader.ReadString('\n') {
		if len(line) > 0 {
			trimmedLine := strings.TrimSpace(line)
			if len(trimmedLine) > 0 {
				split := strings.SplitN(line, "=", 2)
				if len(split) == 2 {
					env[split[0]] = strings.TrimSpace(split[1])
				} else {
					env[split[0]] = ""
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return env, nil
}

type jsonDecoder struct {
	reader io.Reader
}

func newJSONDecoder(reader io.Reader) *decoder {
	return newDecoder(
		&jsonDecoder{
			reader,
		},
	)
}

func (j *jsonDecoder) decode() (map[string]string, error) {
	data := make(map[string]interface{})
	if err := json.NewDecoder(j.reader).Decode(&data); err != nil {
		return nil, err
	}
	env := make(map[string]string)
	for key, value := range data {
		switch value.(type) {
		case string:
			env[key] = value.(string)
		case int64:
			env[key] = fmt.Sprintf("%d", value.(int64))
		case float64:
			env[key] = fmt.Sprintf("%d", int64(value.(float64)))
		case bool:
			env[key] = fmt.Sprintf("%v", value.(bool))
		default:
			return nil, fmt.Errorf("unhandled type: %T", value)
		}
	}
	return env, nil
}
