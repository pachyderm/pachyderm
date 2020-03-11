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

func writeTfJobOperatorOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/tf-training/tf-job-operator/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: tf-job-operator
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: tf-job-operator
      app.kubernetes.io/instance: tf-job-operator-v1.0.0
      app.kubernetes.io/managed-by: kfctl
      app.kubernetes.io/component: tfjob
      app.kubernetes.io/part-of: kubeflow
      app.kubernetes.io/version: v1.0.0
  componentKinds:
  - group: core
    kind: Service
  - group: apps
    kind: Deployment
  - group: core
    kind: ServiceAccount
  - group: kubeflow.org
    kind: TFJob
  descriptor:
    type: "tf-job-operator"
    version: "v1"
    description: "Tf-operator allows users to create and manage the \"TFJob\" custom resource."
    maintainers:
    - name: Richard Liu
      email: ricliu@google.com
    owners:
    - name: Richard Liu
      email: ricliu@google.com
    keywords:
    - "tfjob"
    - "tf-operator"
    - "tf-training"
    links:
    - description: About
      url: "https://github.com/kubeflow/tf-operator"
    - description: Docs
      url: "https://www.kubeflow.org/docs/reference/tfjob/v1/tensorflow/"
  addOwnerRef: true
`)
	th.writeK("/manifests/tf-training/tf-job-operator/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
commonLabels:
  app.kubernetes.io/name: tf-job-operator
  app.kubernetes.io/instance: tf-job-operator-v1.0.0
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: tfjob
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v1.0.0
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/cluster-role-binding.yaml", `
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    app: tf-job-operator
  name: tf-job-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: tf-job-operator
subjects:
- kind: ServiceAccount
  name: tf-job-operator
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/cluster-role.yaml", `
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    app: tf-job-operator
  name: tf-job-operator
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - tfjobs
  - tfjobs/status
  - tfjobs/finalizers
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
- apiGroups:
  - apps
  - extensions
  resources:
  - deployments
  verbs:
  - '*'

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-tfjobs-admin
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-admin: "true"
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.authorization.kubeflow.org/aggregate-to-kubeflow-tfjobs-admin: "true"
rules: []

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-tfjobs-edit
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-edit: "true"
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-tfjobs-admin: "true"
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - tfjobs
  - tfjobs/status
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
  name: kubeflow-tfjobs-view
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-view: "true"
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - tfjobs
  - tfjobs/status
  verbs:
  - get
  - list
  - watch
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/deployment.yaml", `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tf-job-operator
spec:
  replicas: 1
  template:
    metadata:
      labels:
        name: tf-job-operator
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      containers:
      - args:
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
        image: gcr.io/kubeflow-images-public/tf_operator:kubeflow-tf-operator-postsubmit-v1-5adee6f-6109-a25c
        name: tf-job-operator
      serviceAccountName: tf-job-operator
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/service-account.yaml", `
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: tf-job-dashboard
  name: tf-job-dashboard
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: tf-job-operator
  name: tf-job-operator
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/service.yaml", `
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    prometheus.io/path: /metrics
    prometheus.io/scrape: "true"
    prometheus.io/port: "8443"
  labels:
    app: tf-job-operator
  name: tf-job-operator
spec:
  ports:
  - name: monitoring-port
    port: 8443
    targetPort: 8443
  selector:
    name: tf-job-operator
  type: ClusterIP
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/params.env", `
namespace=
`)
	th.writeK("/manifests/tf-training/tf-job-operator/base", `
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
  kustomize.component: tf-job-operator
images:
- name: gcr.io/kubeflow-images-public/tf_operator
  newName: gcr.io/kubeflow-images-public/tf_operator
  newTag: v1.0.0-g92389064
`)
}

func TestTfJobOperatorOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/tf-training/tf-job-operator/overlays/application")
	writeTfJobOperatorOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../tf-training/tf-job-operator/overlays/application"
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
