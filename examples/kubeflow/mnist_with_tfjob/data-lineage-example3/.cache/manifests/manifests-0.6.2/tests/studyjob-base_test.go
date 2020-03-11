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

func writeStudyjobBase(th *KustTestHarness) {
	th.writeF("/manifests/katib-v1alpha1/studyjob/base/studyjob-controller-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: studyjob-controller
  labels:
    app: studyjob-controller
spec:
  replicas: 1
  selector:
    matchLabels:
      app: studyjob-controller
  template:
    metadata:
      labels:
        app: studyjob-controller
    spec:
      serviceAccountName: studyjob-controller
      containers:
      - name: studyjob-controller
        image: gcr.io/kubeflow-images-public/katib/studyjob-controller:v0.1.2-alpha-156-g4ab3dbd
        imagePullPolicy: Always
        ports:
        - containerPort: 443
          name: validating
          protocol: TCP
        env:
        - name: VIZIER_CORE_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
`)
	th.writeF("/manifests/katib-v1alpha1/studyjob/base/studyjob-crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: studyjobs.kubeflow.org
spec:
  group: kubeflow.org
  version: v1alpha1
  scope: Namespaced
  names:
    kind: StudyJob
    singular: studyjob
    plural: studyjobs
  additionalPrinterColumns:
  - JSONPath: .status.condition
    name: Condition
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
`)
	th.writeF("/manifests/katib-v1alpha1/studyjob/base/studyjob-rbac.yaml", `
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: studyjob-controller
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  - serviceaccounts
  - services
  verbs:
  - "*"
- apiGroups:
  - ""
  resources:
  - pods
  - pods/log
  - pods/status
  verbs:
  - "*"
- apiGroups:
  - batch
  resources:
  - jobs
  - cronjobs
  verbs:
  - "*"
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - create
  - get
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - validatingwebhookconfigurations
  verbs:
  - '*'
- apiGroups:
  - kubeflow.org
  resources:
  - studyjobs
  verbs:
  - "*"
- apiGroups:
  - kubeflow.org
  resources:
  - tfjobs
  - pytorchjobs
  verbs:
  - "*"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: studyjob-controller
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: studyjob-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: studyjob-controller
subjects:
- kind: ServiceAccount
  name: studyjob-controller
`)
	th.writeF("/manifests/katib-v1alpha1/studyjob/base/studyjob-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: studyjob-controller
spec:
  ports:
  - port: 443
    protocol: TCP
    targetPort: 443
  selector:
    app: studyjob-controller
`)
	th.writeF("/manifests/katib-v1alpha1/studyjob/base/studyjob-worker-template.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: worker-template
data:
  defaultWorkerTemplate.yaml : |-
    apiVersion: batch/v1
    kind: Job
    metadata:
      name: {{.WorkerID}}
      namespace: {{.NameSpace}}
    spec:
      template:
        spec:
          containers:
          - name: {{.WorkerID}}
            image: alpine
          restartPolicy: Never
`)
	th.writeK("/manifests/katib-v1alpha1/studyjob/base", `
namespace: kubeflow
resources:
- studyjob-controller-deployment.yaml
- studyjob-crd.yaml
- studyjob-rbac.yaml
- studyjob-service.yaml
- studyjob-worker-template.yaml
generatorOptions:
  disableNameSuffixHash: true
images:
  - name: gcr.io/kubeflow-images-public/katib/studyjob-controller
    newTag: v0.1.2-alpha-157-g3d4cd04
`)
}

func TestStudyjobBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib-v1alpha1/studyjob/base")
	writeStudyjobBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib-v1alpha1/studyjob/base"
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
