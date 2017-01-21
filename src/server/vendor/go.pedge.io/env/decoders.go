package env

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type envFileDecoder struct {
	reader io.Reader
}

func newEnvFileDecoder(reader io.Reader) *envFileDecoder {
	return &envFileDecoder{reader}
}

func (e *envFileDecoder) Decode() (map[string]string, error) {
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

func newJSONDecoder(reader io.Reader) *jsonDecoder {
	return &jsonDecoder{reader}
}

func (j *jsonDecoder) Decode() (map[string]string, error) {
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
			return nil, fmt.Errorf("env: unhandled JSON type: %T", value)
		}
	}
	return env, nil
}
