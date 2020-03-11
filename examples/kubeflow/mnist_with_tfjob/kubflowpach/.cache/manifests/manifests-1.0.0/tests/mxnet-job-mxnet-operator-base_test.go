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

func writeMxnetOperatorBase(th *KustTestHarness) {
	th.writeF("/manifests/mxnet-job/mxnet-operator/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    app: mxnet-operator
  name: mxnet-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: mxnet-operator
subjects:
- kind: ServiceAccount
  name: mxnet-operator`)
	th.writeF("/manifests/mxnet-job/mxnet-operator/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    app: mxnet-operator
  name: mxnet-operator
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - mxjobs
  verbs:
  - '*'
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - '*'
- apiGroups:
  - storage.k8s.io
  resources:
  - storageclasses
  verbs:
  - '*'
- apiGroups:
  - batch
  resources:
  - jobs
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - configmaps
  - pods
  - services
  - endpoints
  - persistentvolumeclaims
  - events
  verbs:
  - '*'
- apiGroups:
  - apps
  - extensions
  resources:
  - deployments
  verbs:
  - '*'`)
	th.writeF("/manifests/mxnet-job/mxnet-operator/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: mxjobs.kubeflow.org
spec:
  group: kubeflow.org
  names:
    kind: MXJob
    plural: mxjobs
    singular: mxjob
  version: v1beta1
  scope: Namespaced
`)
	th.writeF("/manifests/mxnet-job/mxnet-operator/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mxnet-operator
spec:
  replicas: 1
  template:
    metadata:
      labels:
        name: mxnet-operator
    spec:
      containers:
      - command:
        - /opt/kubeflow/mxnet-operator.v1beta1
        - --alsologtostderr
        - -v=1
        env:
        - name: MY_POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: MY_POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        image: mxjob/mxnet-operator:v1beta1
        imagePullPolicy: Always
        name: mxnet-operator
      serviceAccountName: mxnet-operator
`)
	th.writeF("/manifests/mxnet-job/mxnet-operator/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: mxnet-operator
  name: mxnet-operator`)
	th.writeK("/manifests/mxnet-job/mxnet-operator/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kubeflow
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- crd.yaml
- deployment.yaml
- service-account.yaml
commonLabels:
  kustomize.component: mxnet-operator
images:
- name: mxjob/mxnet-operator
  newName: mxjob/mxnet-operator
  newTag: v1beta1
`)
}

func TestMxnetOperatorBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/mxnet-job/mxnet-operator/base")
	writeMxnetOperatorBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../mxnet-job/mxnet-operator/base"
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
