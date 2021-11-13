package helmtest

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func TestEtcdPodLabels(t *testing.T) {
	helmChartPath := "../pachyderm"

	options := &helm.Options{
		SetValues: map[string]string{
			"deployTarget": "LOCAL",
		},
		ValuesFiles: []string{"exampleEtcdPodLabels.yaml"},
	}
	expectedLabels := map[string]string{
		"app":   "etcd",
		"suite": "pachyderm",
		"myApp": "Blah123",
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "blah", []string{"templates/etcd/statefulset.yaml"})

	var etcd appsV1.StatefulSet

	helm.UnmarshalK8SYaml(t, output, &etcd)

	if !reflect.DeepEqual(expectedLabels, etcd.Spec.Template.Labels) {
		t.Fatalf("Wanted Etcd Pod Labels %+v, Got, %+v", expectedLabels, etcd.Spec.Template.Labels)
	}
}

func TestPachdPodLabels(t *testing.T) {
	helmChartPath := "../pachyderm"

	options := &helm.Options{
		SetValues: map[string]string{
			"deployTarget": "LOCAL",
		},
		ValuesFiles: []string{"examplePachdPodLabels.yaml"},
	}
	expectedLabels := map[string]string{
		"app":   "pachd",
		"suite": "pachyderm",
		"myApp": "Blah123",
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "blah", []string{"templates/pachd/deployment.yaml"})

	var pachd appsV1.Deployment

	helm.UnmarshalK8SYaml(t, output, &pachd)
	fmt.Printf("%+v\n", pachd.Spec.Template.Labels)

	if !reflect.DeepEqual(expectedLabels, pachd.Spec.Template.Labels) {
		t.Fatalf("Wanted Pachd Pod Labels %+v, Got, %+v", expectedLabels, pachd.Spec.Template.Labels)
	}
}

func TestPachdServiceLabels(t *testing.T) {
	helmChartPath := "../pachyderm"

	options := &helm.Options{
		SetValues: map[string]string{
			"deployTarget": "LOCAL",
		},
		ValuesFiles: []string{"examplePachdSvcLabels.yaml"},
	}
	expectedLabels := map[string]string{
		"app":   "pachd",
		"suite": "pachyderm",
		"myApp": "Blah123",
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "blah", []string{"templates/pachd/service.yaml"})

	var svc v1.Service

	helm.UnmarshalK8SYaml(t, output, &svc)

	if !reflect.DeepEqual(expectedLabels, svc.Labels) {
		t.Fatalf("Wanted Pachd Svc Labels %+v, Got, %+v", expectedLabels, svc.Labels)
	}
}

func TestEtcdServiceLabels(t *testing.T) {
	helmChartPath := "../pachyderm"

	options := &helm.Options{
		SetValues: map[string]string{
			"deployTarget": "LOCAL",
		},
		ValuesFiles: []string{"exampleEtcdSvcLabels.yaml"},
	}
	expectedLabels := map[string]string{
		"app":   "etcd",
		"suite": "pachyderm",
		"myApp": "Blah123",
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "blah", []string{"templates/etcd/service.yaml"})

	var svc v1.Service

	helm.UnmarshalK8SYaml(t, output, &svc)

	if !reflect.DeepEqual(expectedLabels, svc.Labels) {
		t.Fatalf("Wanted Etcd Svc Labels %+v, Got, %+v", expectedLabels, svc.Labels)
	}
}
