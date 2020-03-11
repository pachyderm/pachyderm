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
apiVersion: apps/v1beta2
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
        image: gcr.io/ml-pipeline/persistenceagent:0.1.23
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
  newTag: '0.1.23'
`)
}

func TestPersistentAgentBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pipeline/persistent-agent/base")
	writePersistentAgentBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pipeline/persistent-agent/base"
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
