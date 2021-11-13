package helmtest

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	v1 "k8s.io/api/core/v1"
)

func TestServiceAccount(t *testing.T) {
	helmChartPath := "../pachyderm"

	options := &helm.Options{
		SetValues: map[string]string{
			"deployTarget": "LOCAL",
		},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "blah", []string{"templates/pachd/rbac/serviceaccount.yaml"})

	var serviceAccount v1.ServiceAccount

	helm.UnmarshalK8SYaml(t, output, &serviceAccount)

	if len(serviceAccount.Annotations) != 0 {
		t.Fatalf("Expected no service account annotations, found %s", serviceAccount.Annotations)
	}
}

func TestWorkerServiceAccount(t *testing.T) {
	helmChartPath := "../pachyderm"

	options := &helm.Options{
		SetValues: map[string]string{
			"deployTarget": "LOCAL",
		},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "blah", []string{"templates/pachd/rbac/worker-serviceaccount.yaml"})

	var serviceAccount v1.ServiceAccount

	helm.UnmarshalK8SYaml(t, output, &serviceAccount)

	if len(serviceAccount.Annotations) != 0 {
		t.Fatalf("Expected no service account annotations, found %s", serviceAccount.Annotations)
	}
}

func TestServiceAccountAdditionalAnnotations(t *testing.T) {
	helmChartPath := "../pachyderm"

	options := &helm.Options{
		SetValues: map[string]string{
			"deployTarget": "LOCAL",
		},
		ValuesFiles: []string{"exampleAnnotations.yaml"},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "blah", []string{"templates/pachd/rbac/serviceaccount.yaml"})

	var serviceAccount v1.ServiceAccount

	helm.UnmarshalK8SYaml(t, output, &serviceAccount)

	if serviceAccount.Annotations["eks.amazonaws.com/role-arn"] != "blah123" {
		t.Fatalf("Service Account annotations not expected %s", serviceAccount.Annotations)
	}
}

func TestWorkerServiceAccountAdditionalAnnotations(t *testing.T) {
	helmChartPath := "../pachyderm"

	options := &helm.Options{
		SetValues: map[string]string{
			"deployTarget": "LOCAL",
		},
		ValuesFiles: []string{"exampleWorkerAnnotations.yaml"},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "blah", []string{"templates/pachd/rbac/worker-serviceaccount.yaml"})

	var serviceAccount v1.ServiceAccount

	helm.UnmarshalK8SYaml(t, output, &serviceAccount)

	if serviceAccount.Annotations["eks.amazonaws.com/role-arn"] != "blah123" {
		t.Fatalf("Service Account annotations not expected %s", serviceAccount.Annotations)
	}
}
