package tests_test

import (
	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/target"
	"testing"
)

func writeMetricsCollectorBase(th *KustTestHarness) {
	th.writeF("/manifests/katib-v1alpha2/metrics-collector/base/metrics-collector-rbac.yaml", `
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: metrics-collector
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - pods/log
  - pods/status
  verbs:
  - "*"
- apiGroups:
  - batch
  resources:
  - jobs
  verbs:
  - "*"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: metrics-collector
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: metrics-collector
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: metrics-collector
subjects:
- kind: ServiceAccount
  name: metrics-collector
`)
	th.writeF("/manifests/katib-v1alpha2/metrics-collector/base/metrics-collector-template-configmap.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: metrics-collector-template
data:
  defaultMetricsCollectorTemplate.yaml : |-
    apiVersion: batch/v1beta1
    kind: CronJob
    metadata:
      name: {{.Trial}}
      namespace: {{.NameSpace}}
    spec:
      schedule: "*/1 * * * *"
      successfulJobsHistoryLimit: 0
      failedJobsHistoryLimit: 1
      jobTemplate:
        spec:
          backoffLimit: 0
          template:
            spec:
              serviceAccountName: metrics-collector
              containers:
              - name: {{.Trial}}
                image: gcr.io/kubeflow-images-public/katib/v1alpha2/metrics-collector:v0.1.2-alpha-289-g14dad8b
                imagePullPolicy: IfNotPresent
                command: ["./metricscollector"]
                args:
                - "-e"
                - "{{.Experiment}}"
                - "-t"
                - "{{.Trial}}"
                - "-k"
                - "{{.JobKind}}"
                - "-n"
                - "{{.NameSpace}}"
                - "-m"
                - "{{.ManagerService}}"
                - "-mn"
                - "{{.MetricNames}}"
              restartPolicy: Never
`)
	th.writeK("/manifests/katib-v1alpha2/metrics-collector/base", `
namespace: kubeflow
resources:
- metrics-collector-rbac.yaml
- metrics-collector-template-configmap.yaml
generatorOptions:
  disableNameSuffixHash: true
images:
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/metrics-collector
    newTag: v0.6.0-rc.0
`)
}

func TestMetricsCollectorBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib-v1alpha2/metrics-collector/base")
	writeMetricsCollectorBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib-v1alpha2/metrics-collector/base"
	fsys := fs.MakeRealFS()
	_loader, loaderErr := loader.NewLoader(targetPath, fsys)
	if loaderErr != nil {
		t.Fatalf("could not load kustomize loader: %v", loaderErr)
	}
	rf := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()))
	kt, err := target.NewKustTarget(_loader, rf, transformer.NewFactoryImpl())
	if err != nil {
		th.t.Fatalf("Unexpected construction error %v", err)
	}
	n, err := kt.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := n.EncodeAsYaml()
	th.assertActualEqualsExpected(m, string(expected))
}
