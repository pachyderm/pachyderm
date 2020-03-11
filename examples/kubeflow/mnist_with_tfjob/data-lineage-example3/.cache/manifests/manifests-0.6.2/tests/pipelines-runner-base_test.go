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

func writePipelinesRunnerBase(th *KustTestHarness) {
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

func TestPipelinesRunnerBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pipeline/pipelines-runner/base")
	writePipelinesRunnerBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pipeline/pipelines-runner/base"
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
