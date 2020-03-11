package tests_test

import (
	"sigs.k8s.io/kustomize/v3/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/v3/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/loader"
	"sigs.k8s.io/kustomize/v3/pkg/plugins"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
	"sigs.k8s.io/kustomize/v3/pkg/target"
	"sigs.k8s.io/kustomize/v3/pkg/validators"
	"testing"
)

func writePipelinesRunnerOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/pipeline/pipelines-runner/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: pipelines-runner
spec:
  addOwnerRef: true
  componentKinds:
  - group: core
    kind: ConfigMap
  - group: apps
    kind: Deployment
  descriptor:
    description: ''
    keywords:
    - pipelines-runner
    - kubeflow
    links:
    - description: About
      url: ''
    maintainers: []
    owners: []
    type: pipelines-runner
    version: v1beta1
  selector:
    matchLabels:
      app.kubernetes.io/component: pipelines-runner
      app.kubernetes.io/instance: pipelines-runner-0.2.0
      app.kubernetes.io/managed-by: kfctl
      app.kubernetes.io/name: pipelines-runner
      app.kubernetes.io/part-of: kubeflow
      app.kubernetes.io/version: 0.2.0
`)
	th.writeK("/manifests/pipeline/pipelines-runner/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
bases:
- ../../base
commonLabels:
  app.kubernetes.io/component: pipelines-runner
  app.kubernetes.io/instance: pipelines-runner-0.2.0
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/name: pipelines-runner
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: 0.2.0
kind: Kustomization
resources:
- application.yaml
`)
	th.writeF("/manifests/pipeline/pipelines-runner/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: pipeline-runner
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: pipeline-runner
subjects:
- kind: ServiceAccount
  name: pipeline-runner
`)
	th.writeF("/manifests/pipeline/pipelines-runner/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: pipeline-runner
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - watch
  - list
- apiGroups:
  - ""
  resources:
  - persistentvolumeclaims
  verbs:
  - create
  - delete
  - get
- apiGroups:
  - snapshot.storage.k8s.io
  resources:
  - volumesnapshots
  verbs:
  - create
  - delete
  - get
- apiGroups:
  - argoproj.io
  resources:
  - workflows
  verbs:
  - get
  - list
  - watch
  - update
  - patch
- apiGroups:
  - ""
  resources:
  - pods
  - pods/exec
  - pods/log
  - services
  verbs:
  - '*'
- apiGroups:
  - ""
  - apps
  - extensions
  resources:
  - deployments
  - replicasets
  verbs:
  - '*'
- apiGroups:
  - kubeflow.org
  - serving.kubeflow.org
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - batch
  resources:
  - jobs
  verbs:
  - '*'
`)
	th.writeF("/manifests/pipeline/pipelines-runner/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pipeline-runner
`)
	th.writeK("/manifests/pipeline/pipelines-runner/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kubeflow
commonLabels:
  app: pipeline-runner
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- service-account.yaml
`)
}

func TestPipelinesRunnerOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pipeline/pipelines-runner/overlays/application")
	writePipelinesRunnerOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pipeline/pipelines-runner/overlays/application"
	fsys := fs.MakeRealFS()
	lrc := loader.RestrictionRootOnly
	_loader, loaderErr := loader.NewLoader(lrc, validators.MakeFakeValidator(), targetPath, fsys)
	if loaderErr != nil {
		t.Fatalf("could not load kustomize loader: %v", loaderErr)
	}
	rf := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())
	pc := plugins.DefaultPluginConfig()
	kt, err := target.NewKustTarget(_loader, rf, transformer.NewFactoryImpl(), plugins.NewLoader(pc, rf))
	if err != nil {
		th.t.Fatalf("Unexpected construction error %v", err)
	}
	actual, err := kt.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(actual, string(expected))
}
