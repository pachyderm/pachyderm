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

func writePytorchJobCrdsOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/pytorch-job/pytorch-job-crds/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: pytorch-job-crds
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: pytorch-job-crds
      app.kubernetes.io/instance: pytorch-job-crds-v1.0.0
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
    kind: ServiceAccount
  - group: kubeflow.org
    kind: PyTorchJob
  descriptor:
    type: "pytorch-job-crds"
    version: "v1"
    description: "Pytorch-job-crds contains the \"PyTorchJob\" custom resource definition."
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
	th.writeK("/manifests/pytorch-job/pytorch-job-crds/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
commonLabels:
  app.kubernetes.io/name: pytorch-job-crds
  app.kubernetes.io/instance: pytorch-job-crds-v1.0.0
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: pytorch
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v1.0.0
`)
	th.writeF("/manifests/pytorch-job/pytorch-job-crds/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: pytorchjobs.kubeflow.org
spec:
  additionalPrinterColumns:
  - JSONPath: .status.conditions[-1:].type
    name: State
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
  group: kubeflow.org
  names:
    kind: PyTorchJob
    plural: pytorchjobs
    singular: pytorchjob
  scope: Namespaced
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      properties:
        spec:
          properties:
            pytorchReplicaSpecs:
              properties:
                Master:
                  properties:
                    replicas:
                      maximum: 1
                      minimum: 1
                      type: integer
                Worker:
                  properties:
                    replicas:
                      minimum: 1
                      type: integer
  versions:
  - name: v1
    served: true
    storage: true
`)
	th.writeK("/manifests/pytorch-job/pytorch-job-crds/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- crd.yaml
`)
}

func TestPytorchJobCrdsOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pytorch-job/pytorch-job-crds/overlays/application")
	writePytorchJobCrdsOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pytorch-job/pytorch-job-crds/overlays/application"
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
