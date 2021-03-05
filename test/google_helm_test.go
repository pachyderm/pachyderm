package helmtest

import (
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	v1 "k8s.io/api/core/v1"
)

/*
pachd:
  image:
    tag: 1.12.3
  storage:
    backend: GOOGLE
    google:
      googleBucket: "fake-bucket"
      googleCred: "fake-creds"

*/
func TestGoogleServiceAccount(t *testing.T) {
	helmChartPath := "../pachyderm"

	expectedServiceAccount := "my-fine-sa"
	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.image.tag":                         "1.12.3",
			"pachd.storage.backend":                   "GOOGLE",
			"pachd.storage.google.googleBucket":       "fake-bucket",
			"pachd.storage.google.googleCred":         "fake-creds",
			"pachd.storage.google.serviceAccountName": expectedServiceAccount,
		},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "blah", []string{"templates/pachd/rbac/serviceaccount.yaml"})

	var serviceAccount v1.ServiceAccount

	helm.UnmarshalK8SYaml(t, output, &serviceAccount)

	fmt.Printf("%+v", serviceAccount.Annotations)
	manifestServiceAccount := serviceAccount.Annotations["iam.gke.io/gcp-service-account"]
	if manifestServiceAccount != expectedServiceAccount {
		t.Fatalf("Google service account expected (%s) actual (%s) ", expectedServiceAccount, manifestServiceAccount)
	}
}

func TestGoogleWorkerServiceAccount(t *testing.T) {
	helmChartPath := "../pachyderm"

	expectedServiceAccount := "my-fine-sa"
	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.image.tag":                         "1.12.3",
			"pachd.storage.backend":                   "GOOGLE",
			"pachd.storage.google.googleBucket":       "fake-bucket",
			"pachd.storage.google.googleCred":         "fake-creds",
			"pachd.storage.google.serviceAccountName": expectedServiceAccount,
		},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "blah", []string{"templates/pachd/rbac/worker-serviceaccount.yaml"})

	var serviceAccount v1.ServiceAccount

	helm.UnmarshalK8SYaml(t, output, &serviceAccount)

	fmt.Printf("%+v", serviceAccount.Annotations)
	manifestServiceAccount := serviceAccount.Annotations["iam.gke.io/gcp-service-account"]
	if manifestServiceAccount != expectedServiceAccount {
		t.Fatalf("Google service account expected (%s) actual (%s) ", expectedServiceAccount, manifestServiceAccount)
	}
}
