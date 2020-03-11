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

func writeSpartakusOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/common/spartakus/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: spartakus
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: spartakus
      app.kubernetes.io/instance: spartakus-v1.0.0
      app.kubernetes.io/managed-by: kfctl
      app.kubernetes.io/component: spartakus
      app.kubernetes.io/part-of: kubeflow
      app.kubernetes.io/version: v1.0.0
  componentKinds:
  - group: core
    kind: ConfigMap
  - group: apps
    kind: Deployment
  - group: core
    kind: ServiceAccount
  descriptor:
    type: spartakus
    version: v1beta1
    description: ""
    maintainers: []
    owners: []
    keywords:
     - spartakus
     - kubeflow
    links:
    - description: About
      url: ""
  addOwnerRef: true
`)
	th.writeK("/manifests/common/spartakus/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
commonLabels:
  app.kubernetes.io/name: spartakus
  app.kubernetes.io/instance: spartakus-v1.0.0
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: spartakus
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v1.0.0
`)
	th.writeF("/manifests/common/spartakus/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    app: spartakus
  name: spartakus
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: spartakus
subjects:
- kind: ServiceAccount
  name: spartakus
`)
	th.writeF("/manifests/common/spartakus/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    app: spartakus
  name: spartakus
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
  - list
`)
	th.writeF("/manifests/common/spartakus/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: spartakus
  name: spartakus-volunteer
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: spartakus-volunteer
    spec:
      containers:
      - args:
        - volunteer
        - --cluster-id=$(usageId)
        - --database=https://stats-collector.kubeflow.org
        image: gcr.io/google_containers/spartakus-amd64:v1.1.0
        name: volunteer
      serviceAccountName: spartakus
`)
	th.writeF("/manifests/common/spartakus/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: spartakus
  name: spartakus
`)
	th.writeF("/manifests/common/spartakus/base/params.yaml", `
varReference:
- path: spec/template/spec/containers/0/args/1
  kind: Deployment
`)
	th.writeF("/manifests/common/spartakus/base/params.env", `
usageId=unknown_cluster
`)
	th.writeK("/manifests/common/spartakus/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- deployment.yaml
- service-account.yaml
commonLabels:
  kustomize.component: spartakus
images:
- name: gcr.io/google_containers/spartakus-amd64
  newName: gcr.io/google_containers/spartakus-amd64
  newTag: v1.1.0
configMapGenerator:
- name: spartakus-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: usageId
  objref:
    kind: ConfigMap
    name: spartakus-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.usageId
configurations:
- params.yaml
`)
}

func TestSpartakusOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/common/spartakus/overlays/application")
	writeSpartakusOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../common/spartakus/overlays/application"
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
