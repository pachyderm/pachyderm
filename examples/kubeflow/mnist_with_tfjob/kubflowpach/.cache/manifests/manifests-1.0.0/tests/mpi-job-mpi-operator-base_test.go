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

func writeMpiOperatorBase(th *KustTestHarness) {
	th.writeF("/manifests/mpi-job/mpi-operator/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: mpi-operator
  name: mpi-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: mpi-operator
subjects:
- kind: ServiceAccount
  name: mpi-operator
`)
	th.writeF("/manifests/mpi-job/mpi-operator/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app: mpi-operator
  name: mpi-operator
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  - serviceaccounts
  verbs:
  - create
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
- apiGroups:
  - ""
  resources:
  - pods/exec
  verbs:
  - create
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - roles
  - rolebindings
  verbs:
  - create
  - list
  - watch
- apiGroups:
  - apps
  resources:
  - statefulsets
  verbs:
  - create
  - list
  - update
  - watch
- apiGroups:
  - batch
  resources:
  - jobs
  verbs:
  - create
  - list
  - update
  - watch
- apiGroups:
  - policy
  resources:
  - poddisruptionbudgets
  verbs:
  - create
  - list
  - watch
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - create
  - get
- apiGroups:
  - kubeflow.org
  resources:
  - mpijobs
  verbs:
  - '*'
`)
	th.writeF("/manifests/mpi-job/mpi-operator/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: mpijobs.kubeflow.org
spec:
  group: kubeflow.org
  version: v1alpha1
  scope: Namespaced
  names:
    plural: mpijobs
    singular: mpijob
    kind: MPIJob
    shortNames:
    - mj
    - mpij
  validation:
    openAPIV3Schema:
      properties:
        spec:
          title: The MPIJob spec
          description: Only one of gpus, processingUnits, or replicas should be specified
          oneOf:
          - properties:
              gpus:
                title: Total number of GPUs
                description: Valid values are 1, 2, 4, or any multiple of 8
                oneOf:
                - type: integer
                  enum:
                  - 1
                  - 2
                  - 4
                - type: integer
                  multipleOf: 8
                  minimum: 8
              slotsPerWorker:
                title: The number of slots per worker used in hostfile
                description: Defaults to the number of processing units per worker
                type: integer
                minimum: 1
              gpusPerNode:
                title: The maximum number of GPUs available per node
                description: Defaults to the number of GPUs per worker
                type: integer
                minimum: 1
            required:
            - gpus
          - properties:
              processingUnits:
                title: Total number of processing units
                description: Valid values are 1, 2, 4, or any multiple of 8
                oneOf:
                - type: integer
                  enum:
                  - 1
                  - 2
                  - 4
                - type: integer
                  multipleOf: 8
                  minimum: 8
              slotsPerWorker:
                title: The number of slots per worker used in hostfile
                description: Defaults to the number of processing units per worker
                type: integer
                minimum: 1
              processingUnitsPerNode:
                title: The maximum number of processing units available per node
                description: Defaults to the number of processing units per worker
                type: integer
                minimum: 1
              processingResourceType:
                title: The processing resource type, e.g. 'nvidia.com/gpu' or 'cpu'
                description: Defaults to 'nvidia.com/gpu'
                type: string
                enum:
                - nvidia.com/gpu
                - cpu
            required:
            - processingUnits
          - properties:
              replicas:
                title: Total number of replicas
                description: The processing resource limit should be specified for each replica
                type: integer
                minimum: 1
              slotsPerWorker:
                title: The number of slots per worker used in hostfile
                description: Defaults to the number of processing units per worker
                type: integer
                minimum: 1
              processingResourceType:
                title: The processing resource type, e.g. 'nvidia.com/gpu' or 'cpu'
                description: Defaults to 'nvidia.com/gpu'
                type: string
                enum:
                - nvidia.com/gpu
                - cpu
            required:
            - replicas
`)
	th.writeF("/manifests/mpi-job/mpi-operator/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mpi-operator
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mpi-operator
  template:
    metadata:
      labels:
        app: mpi-operator
    spec:
      containers:
      - args:
        - --gpus-per-node
        - "8"
        - --kubectl-delivery-image
        - $(kubectl-delivery-image)
        image: mpioperator/mpi-operator:0.1.0
        imagePullPolicy: Always
        name: mpi-operator
      serviceAccountName: mpi-operator
`)
	th.writeF("/manifests/mpi-job/mpi-operator/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: mpi-operator
  name: mpi-operator
`)
	th.writeF("/manifests/mpi-job/mpi-operator/base/params.env", `
kubectl-delivery-image=mpioperator/kubectl-delivery:latest
`)
	th.writeK("/manifests/mpi-job/mpi-operator/base", `
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
  kustomize.component: mpi-operator
images:
- name: mpioperator/mpi-operator
  newName: mpioperator/mpi-operator
  newTag: 0.1.0
configMapGenerator:
- name: mpi-operator-config
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: kubectl-delivery-image
  objref:
    kind: ConfigMap
    name: mpi-operator-config
    apiVersion: v1
  fieldref:
    fieldpath: data.kubectl-delivery-image
`)
}

func TestMpiOperatorBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/mpi-job/mpi-operator/base")
	writeMpiOperatorBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../mpi-job/mpi-operator/base"
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
