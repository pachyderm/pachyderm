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

func writeMetacontrollerBase(th *KustTestHarness) {
	th.writeF("/manifests/metacontroller/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: meta-controller-cluster-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: meta-controller-service
`)
	th.writeF("/manifests/metacontroller/base/crd.yaml", `
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: compositecontrollers.metacontroller.k8s.io
spec:
  group: metacontroller.k8s.io
  names:
    kind: CompositeController
    plural: compositecontrollers
    shortNames:
    - cc
    - cctl
    singular: compositecontroller
  scope: Cluster
  version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: controllerrevisions.metacontroller.k8s.io
spec:
  group: metacontroller.k8s.io
  names:
    kind: ControllerRevision
    plural: controllerrevisions
    singular: controllerrevision
  scope: Namespaced
  version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: decoratorcontrollers.metacontroller.k8s.io
spec:
  group: metacontroller.k8s.io
  names:
    kind: DecoratorController
    plural: decoratorcontrollers
    shortNames:
    - dec
    - decorators
    singular: decoratorcontroller
  scope: Cluster
  version: v1alpha1
`)
	th.writeF("/manifests/metacontroller/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: meta-controller-service
`)
	th.writeF("/manifests/metacontroller/base/stateful-set.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    app: metacontroller
  name: metacontroller
spec:
  replicas: 1
  selector:
    matchLabels:
      app: metacontroller
  serviceName: ""
  template:
    metadata:
      labels:
        app: metacontroller
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      containers:
      - command:
        - /usr/bin/metacontroller
        - --logtostderr
        - -v=4
        - --discovery-interval=20s
        image: metacontroller/metacontroller:v0.3.0
        imagePullPolicy: Always
        name: metacontroller
        ports:
        - containerPort: 2345
        resources:
          limits:
            cpu: "4"
            memory: 4Gi
          requests:
            cpu: 500m
            memory: 1Gi
        securityContext:
          allowPrivilegeEscalation: true
          privileged: true
      serviceAccountName: meta-controller-service
  # Workaround for https://github.com/kubernetes-sigs/kustomize/issues/677
  volumeClaimTemplates: []
`)
	th.writeK("/manifests/metacontroller/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cluster-role-binding.yaml
- crd.yaml
- service-account.yaml
- stateful-set.yaml
commonLabels:
  kustomize.component: metacontroller
images:
- name: metacontroller/metacontroller
  newName: metacontroller/metacontroller
  newTag: v0.3.0
`)
}

func TestMetacontrollerBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/metacontroller/base")
	writeMetacontrollerBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../metacontroller/base"
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
