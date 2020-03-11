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

func writeIstioIngressOverlaysOidc(th *KustTestHarness) {
	th.writeF("/manifests/aws/istio-ingress/overlays/oidc/ingress.yaml", `
apiVersion: extensions/v1beta1 # networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: istio-ingress
  namespace: istio-system
  annotations:
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/auth-type: oidc
    alb.ingress.kubernetes.io/auth-idp-cognito: '{"Issuer":"$(oidcIssuer)","AuthorizationEndpoint":"$(oidcAuthorizationEndpoint)","TokenEndpoint":"$(oidcTokenEndpoint)","UserInfoEndpoint":"$(oidcUserInfoEndpoint)","SecretName":"$(oidcSecretName)"}'
    alb.ingress.kubernetes.io/certificate-arn: $(certArn)
    alb.ingress.kubernetes.io/listen-ports: '[{"HTTPS":443}]'
`)
	th.writeF("/manifests/aws/istio-ingress/overlays/oidc/params.yaml", `
varReference:
- path: metadata/annotations
  kind: Ingress

`)
	th.writeF("/manifests/aws/istio-ingress/overlays/oidc/params.env", `
oidcIssuer=
oidcAuthorizationEndpoint=
oidcTokenEndpoint=
oidcUserInfoEndpoint=
oidcSecretName=istio-oidc-secret
certArn=`)
	th.writeF("/manifests/aws/istio-ingress/overlays/oidc/secrets.env", `
clientId=
clientSecret=
`)
	th.writeK("/manifests/aws/istio-ingress/overlays/oidc", `
bases:
- ../../base
patchesStrategicMerge:
- ingress.yaml
#- oidc-secret.yaml
secretGenerator:
- name: istio-oidc-secret
  env: secrets.env
  namespace: istio-system
configMapGenerator:
- name: istio-ingress-oidc-parameters
  env: params.env
vars:
- name: oidcIssuer
  objref:
    kind: ConfigMap
    name: istio-ingress-oidc-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.oidcIssuer
- name: oidcAuthorizationEndpoint
  objref:
    kind: ConfigMap
    name: istio-ingress-oidc-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.oidcAuthorizationEndpoint
- name: oidcTokenEndpoint
  objref:
    kind: ConfigMap
    name: istio-ingress-oidc-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.oidcTokenEndpoint
- name: oidcUserInfoEndpoint
  objref:
    kind: ConfigMap
    name: istio-ingress-oidc-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.oidcUserInfoEndpoint
- name: oidcSecretName
  objref:
    kind: ConfigMap
    name: istio-ingress-oidc-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.oidcSecretName
- name: certArn
  objref:
    kind: ConfigMap
    name: istio-ingress-oidc-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.certArn
configurations:
- params.yaml
`)
	th.writeF("/manifests/aws/istio-ingress/base/ingress.yaml", `
apiVersion: extensions/v1beta1 # networking.k8s.io/v1beta1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/listen-ports: '[{"HTTP": 80}]'
  name: istio-ingress
  namespace: istio-system
spec:
  rules:
    - http:
        paths:
          - backend:
              serviceName: istio-ingressgateway
              servicePort: 80
            path: /*
`)
	th.writeK("/manifests/aws/istio-ingress/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- ingress.yaml
commonLabels:
  kustomize.component: istio-ingress
`)
}

func TestIstioIngressOverlaysOidc(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/aws/istio-ingress/overlays/oidc")
	writeIstioIngressOverlaysOidc(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../aws/istio-ingress/overlays/oidc"
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
