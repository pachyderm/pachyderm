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

func TestAWS(t *testing.T) {

	type envVarMap struct {
		helmKey string
		envVar  string
		value   string
	}

	testCases := []envVarMap{
		{
			helmKey: "pachd.storage.amazon.bucket",
			envVar:  "AMAZON_BUCKET",
			value:   "my-bucket",
		},
		{
			helmKey: "pachd.storage.amazon.cloudFrontDistribution",
			envVar:  "AMAZON_DISTRIBUTION",
			value:   "test-cf",
		},
		{
			helmKey: "pachd.storage.amazon.id",
			envVar:  "AMAZON_ID",
			value:   "testazid123",
		},
		{
			helmKey: "pachd.storage.amazon.region",
			envVar:  "AMAZON_REGION",
			value:   "us-east-1",
		},
		{
			helmKey: "pachd.storage.amazon.secret",
			envVar:  "AMAZON_SECRET",
			value:   "a-fine-secret",
		},
		{
			helmKey: "pachd.storage.amazon.token",
			envVar:  "AMAZON_TOKEN",
			value:   "a-fine-token",
		},
		{
			helmKey: "pachd.storage.amazon.customEndpoint",
			envVar:  "CUSTOM_ENDPOINT",
			value:   "https://myfineendpoint.com",
		},
		{
			helmKey: "pachd.storage.amazon.disableSSL",
			envVar:  "DISABLE_SSL",
			value:   "false",
		},
		{
			helmKey: "pachd.storage.amazon.logOptions",
			envVar:  "OBJ_LOG_OPTS",
			value:   "Debug",
		},
		/*{
			helmKey: "pachd.storage.amazon.maxUploadParts",
			envVar:  "MAX_UPLOAD_PARTS",
			value:   "", //TODO Int
		},*/
		/*{
			helmKey: "pachd.storage.amazon.verifySSL",
			envVar:  "NO_VERIFY_SSL",
			value:   "", //TODO bool
		},*/
		/*{
			helmKey: "pachd.storage.amazon.partSize",
			envVar:  "PART_SIZE",
			value:   "", //TODO int
		},*/
		/*{
			helmKey: "pachd.storage.amazon.retries",
			envVar:  "RETRIES",
			value:   "", //TODO int
		},*/
		/*{
			helmKey: "pachd.storage.amazon.reverse",
			envVar:  "REVERSE",
			value:   "", //TODO bool
		},*/
		{
			helmKey: "pachd.storage.amazon.timeout",
			envVar:  "TIMEOUT",
			value:   "10m",
		},
		{
			helmKey: "pachd.storage.amazon.uploadACL",
			envVar:  "UPLOAD_ACL",
			value:   "read-only",
		},
	}
	var (
		expectedServiceAccount = "my-fine-sa"
		expectedProvisioner    = "ebs.csi.aws.com"
		expectedStorageBackend = "AMAZON"
	)

	helmValues := map[string]string{
		"pachd.storage.backend": "AMAZON",
		`pachd.serviceAccount.additionalAnnotations.eks\.amazonaws\.com/role-arn`:  expectedServiceAccount,
		`worker.serviceAccount.additionalAnnotations.eks\.amazonaws\.com/role-arn`: expectedServiceAccount,
	}
	for _, tc := range testCases {
		helmValues[tc.helmKey] = tc.value
	}

	templatesToCheck := map[string]bool{
		"templates/pachd/storage-secret.yaml":             false,
		"templates/pachd/deployment.yaml":                 false,
		"templates/pachd/rbac/serviceaccount.yaml":        false,
		"templates/pachd/rbac/worker-serviceaccount.yaml": false,
		"templates/etcd/statefulset.yaml":                 false,
		"templates/etcd/storageclass-aws.yaml":            false,
		"templates/postgresql/statefulset.yaml":           false,
		"templates/postgresql/storageclass-aws.yaml":      false,
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

		case *v1.ServiceAccount:
			if resource.Name == "pachyderm-worker" || resource.Name == "pachyderm" {

				t.Run(fmt.Sprintf("%s service account annotation equals %s", resource.Name, expectedServiceAccount), func(t *testing.T) {
					if sa := resource.Annotations["eks.amazonaws.com/role-arn"]; sa != expectedServiceAccount {
						t.Errorf("expected service account to be %q but was %q", expectedServiceAccount, sa)
					}
				})

				if resource.Name == "pachyderm-worker" {
					templatesToCheck["templates/pachd/rbac/worker-serviceaccount.yaml"] = true
				}
				if resource.Name == "pachyderm" {
					templatesToCheck["templates/pachd/rbac/serviceaccount.yaml"] = true
				}
			}
		case *storageV1.StorageClass:
			if resource.Name == "postgresql-storage-class" || resource.Name == "etcd-storage-class" {

				t.Run(fmt.Sprintf("%s storage class annotation equals %s", resource.Name, expectedProvisioner), func(t *testing.T) {
					if resource.Provisioner != expectedProvisioner {
						t.Errorf("expected storageclass provisioner to be %q but it was %q", expectedProvisioner, resource.Provisioner)
					}
				})

				if resource.Name == "postgresql-storage-class" {
					templatesToCheck["templates/postgresql/storageclass-aws.yaml"] = true
				}
				if resource.Name == "etcd-storage-class" {
					templatesToCheck["templates/etcd/storageclass-aws.yaml"] = true
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
					expectedStorageSize := "10Gi"

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
