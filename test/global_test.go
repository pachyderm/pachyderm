package helmtest

import (
	"bytes"
	"io"

	goyaml "github.com/go-yaml/yaml"
	"github.com/gruntwork-io/terratest/modules/logger"
)

func init() {
	logger.Default = logger.Discard
}

// adapted from https://play.golang.org/p/MZNwxdUzxPo
func splitYAML(manifest string) ([]string, error) {
	dec := goyaml.NewDecoder(bytes.NewReader([]byte(manifest)))
	var res []string
	for {
		var value interface{}
		if err := dec.Decode(&value); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		b, err := goyaml.Marshal(value)
		if err != nil {
			return nil, err
		}
		res = append(res, string(b))
	}
	return res, nil
}
