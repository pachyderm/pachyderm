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

func writeIngressgatewaySelfSignedCertBase(th *KustTestHarness) {
	th.writeF("/manifests/istio/ingressgateway-self-signed-cert/base/certificate.yaml", `
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  name: istio-ingress-crt
spec:
  secretName: istio-ingressgateway-certs
  domains:
  - $(domain)
  commonName: "istio-ingressgateway-root-ca"
  isCA: true
  issuerRef:
    name: kubeflow-self-signing-issuer
    kind: ClusterIssuer
`)
	th.writeF("/manifests/istio/ingressgateway-self-signed-cert/base/params.yaml", `
varReference:
- path: spec/domains
  kind: Certificate
`)
	th.writeF("/manifests/istio/ingressgateway-self-signed-cert/base/params.env", `
domain=example.org
`)
	th.writeK("/manifests/istio/ingressgateway-self-signed-cert/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: istio-system
resources:
- certificate.yaml

configMapGenerator:
- name: ingressgateway-self-signed-cert-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true

vars:
- name: domain
  objref:
    kind: ConfigMap
    name: ingressgateway-self-signed-cert-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.domain
configurations:
- params.yaml
`)
}

func TestIngressgatewaySelfSignedCertBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/istio/ingressgateway-self-signed-cert/base")
	writeIngressgatewaySelfSignedCertBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../istio/ingressgateway-self-signed-cert/base"
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
