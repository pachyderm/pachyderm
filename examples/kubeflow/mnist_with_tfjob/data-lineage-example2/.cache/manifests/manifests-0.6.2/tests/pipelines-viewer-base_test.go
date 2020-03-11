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

func writePipelinesViewerBase(th *KustTestHarness) {
	th.writeF("/manifests/pipeline/pipelines-viewer/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: viewers.kubeflow.org
spec:
  group: kubeflow.org
  names:
    kind: Viewer
    listKind: ViewerList
    plural: viewers
    shortNames:
    - vi
    singular: viewer
  scope: Namespaced
  versions:
  - name: v1beta1
    served: true
    storage: true
`)
	th.writeF("/manifests/pipeline/pipelines-viewer/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: crd-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: controller-role
subjects:
- kind: ServiceAccount
  name: crd-service-account
`)
	th.writeF("/manifests/pipeline/pipelines-viewer/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: controller-role
rules:
- apiGroups:
  - '*'
  resources:
  - deployments
  - services
  verbs:
  - create
  - get
  - list
  - watch
  - update
  - patch
  - delete
- apiGroups:
  - kubeflow.org
  resources:
  - viewers
  verbs:
  - create
  - get
  - list
  - watch
  - update
  - patch
  - delete
`)
	th.writeF("/manifests/pipeline/pipelines-viewer/base/deployment.yaml", `
apiVersion: apps/v1beta2
kind: Deployment
metadata:
  name: controller-deployment
spec:
  template:
    spec:
      containers:
      - env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: gcr.io/ml-pipeline/viewer-crd-controller:0.1.23
        imagePullPolicy: Always
        name: ml-pipeline-viewer-controller
      serviceAccountName: crd-service-account
`)
	th.writeF("/manifests/pipeline/pipelines-viewer/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: crd-service-account
`)
	th.writeK("/manifests/pipeline/pipelines-viewer/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kubeflow
nameprefix: ml-pipeline-viewer-
commonLabels:
  app: ml-pipeline-viewer-crd
resources:
- crd.yaml
- cluster-role-binding.yaml
- cluster-role.yaml
- deployment.yaml
- service-account.yaml
images:
- name: gcr.io/ml-pipeline/viewer-crd-controller
  newTag: '0.1.23'
`)
}

func TestPipelinesViewerBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pipeline/pipelines-viewer/base")
	writePipelinesViewerBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pipeline/pipelines-viewer/base"
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
