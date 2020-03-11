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

func writePipelinesViewerOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/pipeline/pipelines-viewer/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: pipelines-viewer
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
    - pipelines-viewer
    - kubeflow
    links:
    - description: About
      url: ''
    maintainers: []
    owners: []
    type: pipelines-viewer
    version: v1beta1
  selector:
    matchLabels:
      app.kubernetes.io/component: pipelines-viewer
      app.kubernetes.io/instance: pipelines-viewer-0.2.0
      app.kubernetes.io/managed-by: kfctl
      app.kubernetes.io/name: pipelines-viewer
      app.kubernetes.io/part-of: kubeflow
      app.kubernetes.io/version: 0.2.0
`)
	th.writeK("/manifests/pipeline/pipelines-viewer/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
bases:
- ../../base
commonLabels:
  app.kubernetes.io/component: pipelines-viewer
  app.kubernetes.io/instance: pipelines-viewer-0.2.0
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/name: pipelines-viewer
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: 0.2.0
kind: Kustomization
resources:
- application.yaml
`)
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

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-pipeline-viewers-admin
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-admin: "true"
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.authorization.kubeflow.org/aggregate-to-kubeflow-pipeline-viewers-admin: "true"
rules: []

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-pipeline-viewers-edit
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-edit: "true"
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-pipeline-viewers-admin: "true"
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - viewers
  verbs:
  - get
  - list
  - watch
  - create
  - delete
  - deletecollection
  - patch
  - update

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-pipeline-viewers-view
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-view: "true"
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - viewers
  verbs:
  - get
  - list
  - watch
`)
	th.writeF("/manifests/pipeline/pipelines-viewer/base/deployment.yaml", `
apiVersion: apps/v1
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
        image: gcr.io/ml-pipeline/viewer-crd-controller:0.1.31
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
  newTag: 0.2.0
  newName: gcr.io/ml-pipeline/viewer-crd-controller
`)
}

func TestPipelinesViewerOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pipeline/pipelines-viewer/overlays/application")
	writePipelinesViewerOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pipeline/pipelines-viewer/overlays/application"
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
