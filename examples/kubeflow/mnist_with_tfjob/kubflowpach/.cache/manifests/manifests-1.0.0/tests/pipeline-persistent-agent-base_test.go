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

func writePersistentAgentBase(th *KustTestHarness) {
	th.writeF("/manifests/pipeline/persistent-agent/base/clusterrole-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: persistenceagent
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: persistenceagent
`)
	th.writeF("/manifests/pipeline/persistent-agent/base/clusterrole.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: persistenceagent
rules:
- apiGroups:
  - argoproj.io
  resources:
  - workflows
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - kubeflow.org
  resources:
  - scheduledworkflows
  verbs:
  - get
  - list
  - watch
`)
	th.writeF("/manifests/pipeline/persistent-agent/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: persistenceagent
spec:
  template:
    spec:
      containers:
      - name: ml-pipeline-persistenceagent
        env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: gcr.io/ml-pipeline/persistenceagent
        imagePullPolicy: IfNotPresent
      serviceAccountName: ml-pipeline-persistenceagent
`)
	th.writeF("/manifests/pipeline/persistent-agent/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: persistenceagent
`)
	th.writeK("/manifests/pipeline/persistent-agent/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
nameprefix: ml-pipeline-
commonLabels:
  app: ml-pipeline-persistenceagent
resources:
- clusterrole-binding.yaml
- clusterrole.yaml
- deployment.yaml
- service-account.yaml
images:
- name: gcr.io/ml-pipeline/persistenceagent
  newTag: 0.2.0
  newName: gcr.io/ml-pipeline/persistenceagent
`)
}

func TestPersistentAgentBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pipeline/persistent-agent/base")
	writePersistentAgentBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pipeline/persistent-agent/base"
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
