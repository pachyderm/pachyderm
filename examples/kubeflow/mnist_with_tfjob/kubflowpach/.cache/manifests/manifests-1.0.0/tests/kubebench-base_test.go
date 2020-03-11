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

func writeKubebenchBase(th *KustTestHarness) {
	th.writeF("/manifests/kubebench/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: kubebench-operator
  name: kubebench-operator
`)
	th.writeF("/manifests/kubebench/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: kubebench-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kubebench-operator
subjects:
- kind: ServiceAccount
  name: kubebench-operator
`)
	th.writeF("/manifests/kubebench/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: kubebench-operator
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - kubebenchjobs
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - configmaps
  - pods
  - pods/exec
  - services
  - endpoints
  - persistentvolumeclaims
  - events
  - secrets
  verbs:
  - '*'
- apiGroups:
  - kubeflow.org
  resources:
  - tfjobs
  - pytorchjobs
  - mpijobs
  verbs:
  - '*'
- apiGroups:
  - argoproj.io
  resources:
  - workflows
  verbs:
  - '*'
`)
	th.writeF("/manifests/kubebench/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: kubebenchjobs.kubeflow.org
spec:
  group: kubeflow.org
  names:
    kind: KubebenchJob
    plural: kubebenchjobs
  scope: Namespaced
  version: v1alpha2
`)
	th.writeF("/manifests/kubebench/base/config-map.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubebench-config
data:
  kubebenchconfig.yaml: |
    defaultWorkflowAgent:
      container:
        name: kubebench-workflow-agent
        image: gcr.io/kubeflow-images-public/kubebench/workflow-agent:bc682c1
    defaultManagedVolumes:
      experimentVolume:
        name: kubebench-experiment-volume
        emptyDir: {}
      workflowVolume:
        name: kubebench-workflow-volume
        emptyDir: {}
`)
	th.writeF("/manifests/kubebench/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubebench-operator
spec:
  selector:
    matchLabels:
      app: kubebench-operator
  template:
    metadata:
      labels:
        app: kubebench-operator
    spec:
      volumes:
      - name: kubebench-config
        configMap:
          name: kubebench-config
      containers:
      - image: gcr.io/kubeflow-images-public/kubebench/kubebench-operator-v1alpha2
        name: kubebench-operator
        command:
        - /app/kubebench-operator-v1alpha2
        args:
        - --config=/config/kubebenchconfig.yaml
        volumeMounts:
        - mountPath: /config
          name: kubebench-config
      serviceAccountName: kubebench-operator
`)
	th.writeF("/manifests/kubebench/base/params.yaml", `
varReference:
- path: metadata/annotations/getambassador.io\/config
  kind: Service
`)
	th.writeF("/manifests/kubebench/base/params.env", `
namespace=
clusterDomain=cluster.local
`)
	th.writeK("/manifests/kubebench/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- service-account.yaml
- cluster-role-binding.yaml
- cluster-role.yaml
- crd.yaml
- config-map.yaml
- deployment.yaml
namespace: kubeflow
commonLabels:
  kustomize.component: kubebench
configMapGenerator:
- name: parameters
  env: params.env
images:
  # NOTE: the image for workflow agent should be configured in config-map.yaml
  - name:  gcr.io/kubeflow-images-public/kubebench/kubebench-operator-v1alpha2
    newName: gcr.io/kubeflow-images-public/kubebench/kubebench-operator-v1alpha2
    newTag: bc682c1
vars:
- name: clusterDomain
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.clusterDomain
configurations:
- params.yaml
`)
}

func TestKubebenchBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/kubebench/base")
	writeKubebenchBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../kubebench/base"
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
