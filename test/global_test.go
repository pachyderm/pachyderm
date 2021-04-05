// SPDX-FileCopyrightText: 2021 Pachyderm, Inc. <info@pachyderm.com>
// SPDX-License-Identifier: Apache-2.0

package helmtest

import (
	"bytes"
	"fmt"
	"io"

	goyaml "github.com/go-yaml/yaml"
	"github.com/gruntwork-io/terratest/modules/logger"
	"k8s.io/kubectl/pkg/scheme"
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

func manifestToObjects(manifest string) ([]interface{}, error) {
	files, err := splitYAML(manifest)
	if err != nil {
		return nil, fmt.Errorf("couldn’t split YAML: %w", err)
	}
	var objects []interface{}
	for i, f := range files {
		object, _, err := scheme.Codecs.UniversalDeserializer().Decode([]byte(f), nil, nil)
		if err != nil {
			return nil, fmt.Errorf("couldn’t decode file %d: %w", i, err)
		}
		objects = append(objects, object)
	}
	return objects, nil
}
