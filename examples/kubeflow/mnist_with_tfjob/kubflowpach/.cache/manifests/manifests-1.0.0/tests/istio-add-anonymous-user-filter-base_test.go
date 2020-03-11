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

func writeAddAnonymousUserFilterBase(th *KustTestHarness) {
	th.writeF("/manifests/istio/add-anonymous-user-filter/base/envoy-filter.yaml", `
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: add-user-filter
spec:
  workloadLabels:
    app: istio-ingressgateway
  filters:
  - listenerMatch:
      listenerType: GATEWAY
    filterName: envoy.lua
    filterType: HTTP
    insertPosition:
      index: FIRST
    filterConfig:
      inlineCode: |
        function envoy_on_request(request_handle)
            request_handle:headers():add("kubeflow-userid","anonymous@kubeflow.org")
        end
`)
	th.writeK("/manifests/istio/add-anonymous-user-filter/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: istio-system
resources:
- envoy-filter.yaml
`)
}

func TestAddAnonymousUserFilterBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/istio/add-anonymous-user-filter/base")
	writeAddAnonymousUserFilterBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../istio/add-anonymous-user-filter/base"
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
