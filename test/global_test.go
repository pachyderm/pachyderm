// SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
// SPDX-License-Identifier: Apache-2.0

package helmtest

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/gruntwork-io/terratest/modules/logger"
	"gopkg.in/yaml.v3"
	goyaml "gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
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
		var value yaml.Node
		if err := dec.Decode(&value); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		//fmt.Printf("%+v\n", value.Content[0].Content[0].HeadComment)
		b, err := goyaml.Marshal(&value)
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

func TestEnsureVolumeMountPresent(t *testing.T) {
	volumeMounts := []v1.VolumeMount{
		{
			Name:      "volmount1",
			MountPath: "/my/mount",
		},
		{
			Name:      "volmount2",
			MountPath: "/my/other/mount",
		},
	}
	want := v1.VolumeMount{
		Name:      "volmount1",
		MountPath: "/my/mount",
	}

	if !ensureVolumeMountPresent(want, volumeMounts) {
		t.Error("Volume mount not found")
	}
}

func ensureVolumeMountPresent(matchMount v1.VolumeMount, mounts []v1.VolumeMount) bool {
	present := false

	for _, vm := range mounts {
		if reflect.DeepEqual(matchMount, vm) {
			present = true
			break
		}
	}
	return present
}

func TestEnsureVolumePresent(t *testing.T) {
	volumes := []v1.Volume{
		{
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: "A Fine Secret",
				},
			},
			Name: "pachd-tls-cert",
		},
		{
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: "Another Secret",
				},
			},
			Name: "another-secret",
		},
	}

	want := v1.Volume{
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: "A Fine Secret",
			},
		},
		Name: "pachd-tls-cert",
	}
	if !ensureVolumePresent(want, volumes) {
		t.Error("Volume mount not found")
	}
}

func ensureVolumePresent(matchVol v1.Volume, volumes []v1.Volume) bool {
	present := false

	for _, v := range volumes {
		if reflect.DeepEqual(matchVol, v) {
			present = true
			break
		}
	}
	return present
}

// GetContainerByName returns container or nil
func GetContainerByName(name string, containers []v1.Container) (*v1.Container, bool) {
	for _, c := range containers {
		if c.Name == name {
			return &c, true
		}
	}
	return &v1.Container{}, false
}
