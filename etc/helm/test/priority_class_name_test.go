package helmtest

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	appsV1 "k8s.io/api/apps/v1"
)

func TestSetPachdPriorityClass(t *testing.T) {
	helmChartPath := "../pachyderm"
	options := &helm.Options{
		SetValues: map[string]string{
			"deployTarget": "LOCAL",
		},
		ValuesFiles: []string{"priorityClassNameValues.yaml"},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "deployment", []string{"templates/pachd/deployment.yaml"})

	var deployment *appsV1.Deployment
	helm.UnmarshalK8SYaml(t, output, &deployment)

	if deployment.Spec.Template.Spec.PriorityClassName != "highest" {
		t.Fatalf("Priority Class Name of Deployment not expected. got: %v, want: %s", deployment.Spec.Template.Spec.PriorityClassName, "highest")
	}
}

func TestSetEtcdPriorityClass(t *testing.T) {
	helmChartPath := "../pachyderm"
	options := &helm.Options{
		SetValues: map[string]string{
			"deployTarget": "LOCAL",
		},
		ValuesFiles: []string{"priorityClassNameValues.yaml"},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "statefulset", []string{"templates/etcd/statefulset.yaml"})

	var etcd *appsV1.StatefulSet
	helm.UnmarshalK8SYaml(t, output, &etcd)

	if etcd.Spec.Template.Spec.PriorityClassName != "lowest" {
		t.Fatalf("Priority Class Name of Etcd statefulset not expected. got: %v, want: %s", etcd.Spec.Template.Spec.PriorityClassName, "lowest")
	}
}
