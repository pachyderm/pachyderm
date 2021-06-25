// SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
// SPDX-License-Identifier: Apache-2.0

package helmtest

import (
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func TestMinio(t *testing.T) {
	var (
		expectedStorageBackend = "MINIO"
	)

	type envVarMap struct {
		helmKey string
		envVar  string
		value   string
	}

	testCases := []envVarMap{
		{
			helmKey: "pachd.storage.minio.bucket",
			envVar:  "MINIO_BUCKET",
			value:   "my-bucket",
		},
		{
			helmKey: "pachd.storage.minio.endpoint",
			envVar:  "MINIO_ENDPOINT",
			value:   "https://endpoint.test/",
		},
		{
			helmKey: "pachd.storage.minio.id",
			envVar:  "MINIO_ID",
			value:   "my-fine-id",
		},
		{
			helmKey: "pachd.storage.minio.secret",
			envVar:  "MINIO_SECRET",
			value:   "my-fine-secret",
		},
		{
			helmKey: "pachd.storage.minio.secure",
			envVar:  "MINIO_SECURE",
			value:   "", //TODO
		},
		{
			helmKey: "pachd.storage.minio.signature",
			envVar:  "MINIO_SIGNATURE",
			value:   "", //TODO
		},
	}

	helmValues := map[string]string{
		"pachd.storage.backend": expectedStorageBackend,
	}
	for _, tc := range testCases {
		helmValues[tc.helmKey] = tc.value
	}
	templatesToCheck := map[string]bool{
		"templates/pachd/storage-secret.yaml": false,
		"templates/pachd/deployment.yaml":     false,
	}

	templatesToRender := []string{}
	for k := range templatesToCheck {
		templatesToRender = append(templatesToRender, k)
	}

	objects, err := manifestToObjects(helm.RenderTemplate(t,
		&helm.Options{
			SetStrValues: helmValues,
		}, "../pachyderm/", "release-name", templatesToRender),
	)

	if err != nil {
		t.Fatalf("could not render templates to objects: %v", err)
	}
	for _, object := range objects {
		switch resource := object.(type) {
		case *v1.Secret:
			if resource.Name != "pachyderm-storage-secret" {
				continue
			}
			for _, tc := range testCases {
				t.Run(fmt.Sprintf("%s equals %s", tc.envVar, tc.value), func(t *testing.T) {
					if got := string(resource.Data[tc.envVar]); got != tc.value {
						t.Errorf("got %s; want %s", got, tc.value)
					}
				})
			}
			templatesToCheck["templates/pachd/storage-secret.yaml"] = true
		case *appsV1.Deployment:
			if resource.Name != "pachd" {
				continue
			}
			c, ok := GetContainerByName("pachd", resource.Spec.Template.Spec.Containers)
			if !ok {
				t.Errorf("pachd container not found in pachd deployment")
			}
			if err, got := GetEnvVarByName(c.Env, storageBackendEnvVar); err == nil && got != expectedStorageBackend {
				t.Errorf("expected %q to be %q, not %q", storageBackendEnvVar, expectedStorageBackend, got)
			}
			templatesToCheck["templates/pachd/deployment.yaml"] = true
		}
	}

	for k, ok := range templatesToCheck {
		if !ok {
			t.Errorf("template %q not checked", k)
		}
	}
}
