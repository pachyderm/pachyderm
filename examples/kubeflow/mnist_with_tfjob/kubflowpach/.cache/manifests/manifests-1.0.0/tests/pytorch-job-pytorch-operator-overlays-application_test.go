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

func writePytorchOperatorOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/pytorch-job/pytorch-operator/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: pytorch-operator
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: pytorch-operator
      app.kubernetes.io/instance: pytorch-operator-v1.0.0
      app.kubernetes.io/version: v1.0.0
      app.kubernetes.io/component: pytorch
      app.kubernetes.io/part-of: kubeflow
      app.kubernetes.io/managed-by: kfctl
  componentKinds:
  - group: core
    kind: Service
  - group: apps
    kind: Deployment
  - group: core
    kind: ConfigMap
  - group: core
    kind: ServiceAccount
  - group: kubeflow.org
    kind: PyTorchJob
  descriptor:
    type: "pytorch-operator"
    version: "v1"
    description: "Pytorch-operator allows users to create and manage the \"PyTorchJob\" custom resource."
    maintainers:
    - name: Johnu George
      email: johnugeo@cisco.com
    owners:
    - name: Johnu George
      email: johnugeo@cisco.com
    keywords:
    - "pytorchjob"
    - "pytorch-operator"
    - "pytorch-training"
    links:
    - description: About
      url: "https://github.com/kubeflow/pytorch-operator"
    - description: Docs
      url: "https://www.kubeflow.org/docs/reference/pytorchjob/v1/pytorch/"
  addOwnerRef: true
`)
	th.writeK("/manifests/pytorch-job/pytorch-operator/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
commonLabels:
  app.kubernetes.io/name: pytorch-operator
  app.kubernetes.io/instance: pytorch-operator-v1.0.0
  app.kubernetes.io/version: v1.0.0
  app.kubernetes.io/component: pytorch
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/managed-by: kfctl
`)
	th.writeF("/manifests/pytorch-job/pytorch-operator/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    app: pytorch-operator
  name: pytorch-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: pytorch-operator
subjects:
- kind: ServiceAccount
  name: pytorch-operator
`)
	th.writeF("/manifests/pytorch-job/pytorch-operator/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    app: pytorch-operator
  name: pytorch-operator
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - pytorchjobs
  - pytorchjobs/status
  verbs:
  - '*'
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - pods
  - services
  - endpoints
  - events
  verbs:
  - '*'
---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-pytorchjobs-admin
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-admin: "true"
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.authorization.kubeflow.org/aggregate-to-kubeflow-pytorchjobs-admin: "true"
rules: []

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-pytorchjobs-edit
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-edit: "true"
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-pytorchjobs-admin: "true"
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - pytorchjobs
  - pytorchjobs/status
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
  name: kubeflow-pytorchjobs-view
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-view: "true"
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - pytorchjobs
  - pytorchjobs/status
  verbs:
  - get
  - list
  - watch
`)
	th.writeF("/manifests/pytorch-job/pytorch-operator/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pytorch-operator
spec:
  replicas: 1
  selector:
    matchLabels:
      name: pytorch-operator
  template:
    metadata:
      labels:
        name: pytorch-operator
    spec:
      containers:
      - command:
        - /pytorch-operator.v1
        - --alsologtostderr
        - -v=1
        - --monitoring-port=8443
        env:
        - name: MY_POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: MY_POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        image: gcr.io/kubeflow-images-public/pytorch-operator:v0.6.0-18-g5e36a57
        name: pytorch-operator
      serviceAccountName: pytorch-operator
`)
	th.writeF("/manifests/pytorch-job/pytorch-operator/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: pytorch-operator
  name: pytorch-operator
`)
	th.writeF("/manifests/pytorch-job/pytorch-operator/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  annotations:
    prometheus.io/path: /metrics
    prometheus.io/port: "8443"
    prometheus.io/scrape: "true"
  labels:
    app: pytorch-operator
  name: pytorch-operator
spec:
  ports:
  - name: monitoring-port
    port: 8443
    targetPort: 8443
  selector:
    name: pytorch-operator
  type: ClusterIP

`)
	th.writeF("/manifests/pytorch-job/pytorch-operator/base/params.env", `
pytorchDefaultImage=null
deploymentScope=cluster
deploymentNamespace=null
`)
	th.writeK("/manifests/pytorch-job/pytorch-operator/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kubeflow
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- deployment.yaml
- service-account.yaml
- service.yaml
commonLabels:
  kustomize.component: pytorch-operator
images:
- name: gcr.io/kubeflow-images-public/pytorch-operator
  newName: gcr.io/kubeflow-images-public/pytorch-operator
  newTag: v1.0.0-g047cf0f
`)
}

func TestPytorchOperatorOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pytorch-job/pytorch-operator/overlays/application")
	writePytorchOperatorOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pytorch-job/pytorch-operator/overlays/application"
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
