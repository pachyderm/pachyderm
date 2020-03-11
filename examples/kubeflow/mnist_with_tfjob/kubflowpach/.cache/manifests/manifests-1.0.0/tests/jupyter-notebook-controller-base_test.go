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

func writeNotebookControllerBase(th *KustTestHarness) {
	th.writeF("/manifests/jupyter/notebook-controller/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: role
subjects:
- kind: ServiceAccount
  name: service-account
`)
	th.writeF("/manifests/jupyter/notebook-controller/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: role
rules:
- apiGroups:
  - apps
  resources:
  - statefulsets
  - deployments
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - get
  - list
  - watch
  - create
- apiGroups:
  - kubeflow.org
  resources:
  - notebooks
  - notebooks/status
  - notebooks/finalizers
  verbs:
  - '*'
- apiGroups:
  - networking.istio.io
  resources:
  - virtualservices
  verbs:
  - '*'

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-notebooks-admin
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-admin: "true"
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.authorization.kubeflow.org/aggregate-to-kubeflow-notebooks-admin: "true"
rules: []

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-notebooks-edit
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-edit: "true"
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-notebooks-admin: "true"
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - notebooks
  - notebooks/status
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
  name: kubeflow-notebooks-view
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-view: "true"
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - notebooks
  - notebooks/status
  verbs:
  - get
  - list
  - watch
`)
	th.writeF("/manifests/jupyter/notebook-controller/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: notebooks.kubeflow.org
spec:
  group: kubeflow.org
  names:
    kind: Notebook
    plural: notebooks
    singular: notebook
  scope: Namespaced
  subresources:
    status: {}
  versions:
  - name: v1alpha1
    served: true
    storage: false
  - name: v1beta1
    served: true
    storage: true
  - name: v1
    served: true
    storage: false
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          properties:
            template:
              description: 'INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
                Important: Run "make" to regenerate code after modifying this file'
              properties:
                spec:
                  type: object
              type: object
          type: object
        status:
          properties:
            conditions:
              description: Conditions is an array of current conditions
              items:
                properties:
                  type:
                    description: Type of the confition/
                    type: string
                required:
                - type
                type: object
              type: array
          required:
          - conditions
          type: object
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []

`)
	th.writeF("/manifests/jupyter/notebook-controller/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment
spec:
  template:
    spec:
      containers:
      - name: manager
        image: gcr.io/kubeflow-images-public/notebook-controller:v20190614-v0-160-g386f2749-e3b0c4
        command:
          - /manager
        env:
          - name: USE_ISTIO
            value: "false"
          - name: POD_LABELS
            value: $(POD_LABELS)
        imagePullPolicy: Always
        livenessProbe:
          httpGet:
            path: /metrics
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 30
      serviceAccountName: service-account
`)
	th.writeF("/manifests/jupyter/notebook-controller/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: service-account
`)
	th.writeF("/manifests/jupyter/notebook-controller/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: service
spec:
  ports:
  - port: 443
`)
	th.writeF("/manifests/jupyter/notebook-controller/base/params.env", `
POD_LABELS=gcp-cred-secret=user-gcp-sa,gcp-cred-secret-filename=user-gcp-sa.json
USE_ISTIO=false
ISTIO_GATEWAY=kubeflow/kubeflow-gateway
`)
	th.writeK("/manifests/jupyter/notebook-controller/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- crd.yaml
- deployment.yaml
- service-account.yaml
- service.yaml
namePrefix: notebook-controller-
namespace: kubeflow
commonLabels:
  app: notebook-controller
  kustomize.component: notebook-controller
images:
- name: gcr.io/kubeflow-images-public/notebook-controller
  newName: gcr.io/kubeflow-images-public/notebook-controller
  newTag: v1.0.0-gcd65ce25
configMapGenerator:
- envs:
  - params.env
  name: parameters
generatorOptions:
  disableNameSuffixHash: true
vars:
- fieldref:
    fieldPath: data.POD_LABELS
  name: POD_LABELS
  objref:
    apiVersion: v1
    kind: ConfigMap
    name: parameters
- fieldref:
    fieldPath: data.USE_ISTIO
  name: USE_ISTIO
  objref:
    apiVersion: v1
    kind: ConfigMap
    name: parameters
- fieldref:
    fieldPath: data.ISTIO_GATEWAY
  name: ISTIO_GATEWAY
  objref:
    apiVersion: v1
    kind: ConfigMap
    name: parameters
`)
}

func TestNotebookControllerBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/jupyter/notebook-controller/base")
	writeNotebookControllerBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../jupyter/notebook-controller/base"
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
