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

func writeNotebookControllerOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/jupyter/notebook-controller/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: notebook-controller
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: notebook-controller
      app.kubernetes.io/instance: notebook-controller
      app.kubernetes.io/managed-by: kfctl
      app.kubernetes.io/component: notebook
      app.kubernetes.io/part-of: kubeflow
      app.kubernetes.io/version: v0.6
  componentKinds:
    - group: core
      kind: Service
    - group: apps
      kind: Deployment
    - group: core
      kind: ServiceAccount
  descriptor: 
    type: notebook-controller
    version: v0.6.2
    description: Notebooks controller allows users to create a custom resource \"Notebook\" (jupyter notebook).
    maintainers:
    - name: Lun-kai Hsu
      email: lunkai@google.com
    owners:
    - name: Lun-kai Hsu
      email: lunkai@gogle.com
    keywords:
     - jupyter
     - notebook
     - notebook-controller
     - jupyterhub  
    links:
    - description: About
      url: "https://github.com/kubeflow/kubeflow/tree/master/components/notebook-controller"
  addOwnerRef: true
`)
	th.writeK("/manifests/jupyter/notebook-controller/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
commonLabels:
  app.kubernetes.io/name: notebook-controller
  app.kubernetes.io/instance: notebook-controller
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: notebook
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.6
`)
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
  - kubeflow.org
  resources:
  - notebooks
  - notebooks/status
  verbs:
  - '*'
- apiGroups:
  - networking.istio.io
  resources:
  - virtualservices
  verbs:
  - '*'
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
  version: v1alpha1
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
apiVersion: apps/v1beta1
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
    newTag: v20190603-v0-175-geeca4530-e3b0c4
configMapGenerator:
- name: parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: POD_LABELS
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.POD_LABELS
- name: USE_ISTIO
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.USE_ISTIO
`)
}

func TestNotebookControllerOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/jupyter/notebook-controller/overlays/application")
	writeNotebookControllerOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../jupyter/notebook-controller/overlays/application"
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
