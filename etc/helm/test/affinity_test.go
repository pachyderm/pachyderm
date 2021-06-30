package helmtest

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	appsV1 "k8s.io/api/apps/v1"
)

func TestSetPachdAffinity(t *testing.T) {
	helmChartPath := "../pachyderm"
	options := &helm.Options{
		SetValues: map[string]string{
			"deployTarget": "LOCAL",
		},
		ValuesFiles: []string{"pachdAffinity.yaml"},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "deployment", []string{"templates/pachd/deployment.yaml"})

	var deployment *appsV1.Deployment
	helm.UnmarshalK8SYaml(t, output, &deployment)

	if deployment.Spec.Template.Spec.Affinity == nil {
		t.Fatalf("Affinity of Deployment not found")
	}
}

func TestSetEtcdAffinity(t *testing.T) {
	helmChartPath := "../pachyderm"
	options := &helm.Options{
		SetValues: map[string]string{
			"deployTarget": "LOCAL",
		},
		ValuesFiles: []string{"etcdAffinity.yaml"},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "statefulset", []string{"templates/etcd/statefulset.yaml"})

	var etcd *appsV1.StatefulSet
	helm.UnmarshalK8SYaml(t, output, &etcd)

	if etcd.Spec.Template.Spec.Affinity == nil {
		t.Fatalf("Affinity of Etcd statefulset not found")
	}
}
