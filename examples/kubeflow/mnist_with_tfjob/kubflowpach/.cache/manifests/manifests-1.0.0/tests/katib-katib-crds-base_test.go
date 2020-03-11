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

func writeKatibCrdsBase(th *KustTestHarness) {
	th.writeF("/manifests/katib/katib-crds/base/experiment-crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: experiments.kubeflow.org
spec:
  additionalPrinterColumns:
  - JSONPath: .status.conditions[-1:].type
    name: Status
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
  group: kubeflow.org
  version: v1alpha3
  scope: Namespaced
  subresources:
    status: {}
  names:
    kind: Experiment
    singular: experiment
    plural: experiments
    categories:
    - all
    - kubeflow
    - katib
`)
	th.writeF("/manifests/katib/katib-crds/base/suggestion-crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: suggestions.kubeflow.org
spec:
  additionalPrinterColumns:
  - JSONPath: .status.conditions[-1:].type
    name: Type
    type: string
  - JSONPath: .status.conditions[-1:].status
    name: Status
    type: string
  - JSONPath: .spec.requests
    name: Requested
    type: string
  - JSONPath: .status.suggestionCount
    name: Assigned
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
  group: kubeflow.org
  version: v1alpha3
  scope: Namespaced
  subresources:
    status: {}
  names:
    kind: Suggestion
    singular: suggestion
    plural: suggestions
    categories:
    - all
    - kubeflow
    - katib
`)
	th.writeF("/manifests/katib/katib-crds/base/trial-crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: trials.kubeflow.org
spec:
  additionalPrinterColumns:
  - JSONPath: .status.conditions[-1:].type
    name: Type
    type: string
  - JSONPath: .status.conditions[-1:].status
    name: Status
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
  group: kubeflow.org
  version: v1alpha3
  scope: Namespaced
  subresources:
    status: {}
  names:
    kind: Trial
    singular: trial
    plural: trials
    categories:
    - all
    - kubeflow
    - katib
`)
	th.writeK("/manifests/katib/katib-crds/base", `
namespace: kubeflow
resources:
- experiment-crd.yaml
- suggestion-crd.yaml
- trial-crd.yaml
generatorOptions:
  disableNameSuffixHash: true
`)
}

func TestKatibCrdsBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib/katib-crds/base")
	writeKatibCrdsBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib/katib-crds/base"
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
