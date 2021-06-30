// SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
// SPDX-License-Identifier: Apache-2.0

package helmtest

import (
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
)

func TestMicrosoft(t *testing.T) {

	type envVarMap struct {
		helmKey string
		envVar  string
		value   string
	}
	testCases := []envVarMap{
		{
			helmKey: "pachd.storage.microsoft.container",
			value:   "fake-bucket",
			envVar:  "MICROSOFT_CONTAINER",
		},
		{
			helmKey: "pachd.storage.microsoft.id",
			value:   "a-fine-id",
			envVar:  "MICROSOFT_ID",
		},
		{
			helmKey: "pachd.storage.microsoft.secret",
			value:   "super-secret-something",
			envVar:  "MICROSOFT_SECRET",
		},
	}
	var (
		expectedProvisioner    = "kubernetes.io/azure-disk"
		expectedStorageBackend = "MICROSOFT"
	)
	helmValues := map[string]string{
		"deployTarget": expectedStorageBackend,
	}
	for _, tc := range testCases {
		helmValues[tc.helmKey] = tc.value
	}

	templatesToCheck := map[string]bool{
		"templates/pachd/storage-secret.yaml":          false,
		"templates/pachd/deployment.yaml":              false,
		"templates/etcd/statefulset.yaml":              false,
		"templates/etcd/storageclass-azure.yaml":       false,
		"templates/postgresql/statefulset.yaml":        false,
		"templates/postgresql/storageclass-azure.yaml": false,
	}

	templatesToRender := []string{}
	for k := range templatesToCheck {
		templatesToRender = append(templatesToRender, k)
	}

	objects, err := manifestToObjects(helm.RenderTemplate(t, &helm.Options{
		SetValues: helmValues,
	}, "../pachyderm", "blah", templatesToRender))

	if err != nil {
		t.Fatal(err)
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

		case *storageV1.StorageClass:
			if resource.Name == "postgresql-storage-class" || resource.Name == "etcd-storage-class" {

				t.Run(fmt.Sprintf("%s storage class annotation equals %s", resource.Name, expectedProvisioner), func(t *testing.T) {
					if resource.Provisioner != expectedProvisioner {
						t.Errorf("expected storageclass provisioner to be %q but it was %q", expectedProvisioner, resource.Provisioner)
					}
				})
				//TODO Check all params in storage class set for microsoft
				if resource.Name == "postgresql-storage-class" {
					templatesToCheck["templates/postgresql/storageclass-azure.yaml"] = true
				}
				if resource.Name == "etcd-storage-class" {
					templatesToCheck["templates/etcd/storageclass-azure.yaml"] = true
				}
			}
		case *appsV1.Deployment:
			if resource.Name != "pachd" {
				continue
			}

			t.Run("pachd deployment env vars", func(t *testing.T) {
				c, ok := GetContainerByName("pachd", resource.Spec.Template.Spec.Containers)
				if !ok {
					t.Errorf("pachd container not found in pachd deployment")
				}
				if err, got := GetEnvVarByName(c.Env, storageBackendEnvVar); err == nil && got != expectedStorageBackend {
					t.Errorf("expected %s to be %q, not %q", storageBackendEnvVar, expectedStorageBackend, got)
				}
			})
			templatesToCheck["templates/pachd/deployment.yaml"] = true

		case *appsV1.StatefulSet:
			if resource.Name == "etcd" || resource.Name == "postgres" {

				for _, pvc := range resource.Spec.VolumeClaimTemplates {
					// Check Default Storage Request
					expectedStorageSize := "256Gi"

					if pvc.Name == "etcd-storage" {
						t.Run(fmt.Sprintf("%s storage class storage resource request equals %s", resource.Name, expectedStorageSize), func(t *testing.T) {
							if got := pvc.Spec.Resources.Requests.Storage().String(); got != expectedStorageSize {
								t.Errorf("expected stateful set storage resource request to be %q but it was %q", expectedStorageSize, got)

							}
						})
						templatesToCheck["templates/etcd/statefulset.yaml"] = true
					}

					if pvc.Name == "postgres-storage" {
						t.Run(fmt.Sprintf("%s storage class storage resource request equals %s", resource.Name, expectedStorageSize), func(t *testing.T) {
							if got := pvc.Spec.Resources.Requests.Storage().String(); got != expectedStorageSize {
								t.Errorf("expected stateful set storage resource request to be %q but it was %q", expectedStorageSize, got)

							}
						})
						templatesToCheck["templates/postgresql/statefulset.yaml"] = true

					}

				}
			}
		}
	}

	for k, ok := range templatesToCheck {
		if !ok {
			t.Errorf("template %q not checked", k)
		}
	}
}
