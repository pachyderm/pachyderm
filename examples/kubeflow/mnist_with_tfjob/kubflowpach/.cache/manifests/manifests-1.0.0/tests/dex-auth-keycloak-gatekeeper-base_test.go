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

func writeKeycloakGatekeeperBase(th *KustTestHarness) {
	th.writeF("/manifests/dex-auth/keycloak-gatekeeper/base/config-map.yaml", `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: keycloak-gatekeeper-page-templates
data:
  forbidden-page: |
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <title>Forbidden</title>
        <style>
        body { margin:0; padding:0; }
        .fullscreen { position:fixed; margin:0; padding:0; height:100vh; width:100vw; text-align:center; }
        .bg { z-index:1; }
        .fg { z-index:2; }
        .textwall { font-size:35vw; color: lavender; line-height:100vh; vertical-align:middle; }
        .textblock { position:absolute; width: 100vw; top:50%; transform:translate(0,-50%); }
        .textblock h1,p { padding: 1em 0 1em 0; margin:0; background-color: rgba(255,255,255,0.6); }
        </style>
    </head>
    <body>

    <div class="fullscreen bg textwall">(°◇°)</div>
    <div class="fullscreen fg">
      <div class="textblock">
        <h1>Access forbidden</h1>
        <p>You do not have sufficient privileges to access this resource.</p>
      </div>
    </div>

    </body>
    </html>
  login-page: |
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <title>Redirecting to SSO login page...</title>
        <style>
        body { margin:0; padding:0; }
        .fullscreen { position:fixed; margin:0; padding:0; height:100vh; width:100vw; text-align:center; }
        .bg { z-index:1; }
        .fg { z-index:2; }
        .textwall { font-size:35vw; color: lavender; line-height:100vh; vertical-align:middle; }
        .textblock { position:absolute; width: 100vw; top:50%; transform:translate(0,-50%); }
        .textblock h1,p { padding: 1em 0 1em 0; margin:0; background-color: rgba(255,255,255,0.6); }
        </style>
        <script>
        function redirectToLoginPage() {
          let loginPageURL = "{{ .redirect }}"
          window.location.replace(loginPageURL);
        }
        setTimeout(redirectToLoginPage,3000);
        </script>
    </head>
    <body>

    <div class="fullscreen bg textwall">¯\_(ツ)_/¯</div>
    <div class="fullscreen fg">
      <div class="textblock">
        <h1>Access token expired</h1>
        <p>You will be automatically redirected to your SSO provider's Sign In page for this app.</p>
        <p>If not, <a href="{{ .redirect }}">click here to sign in</a>.</p>
      </div>
    </div>

    </body>
    </html>
`)
	th.writeF("/manifests/dex-auth/keycloak-gatekeeper/base/namespace.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  name: auth
`)
	th.writeF("/manifests/dex-auth/keycloak-gatekeeper/base/deployment.yaml", `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: keycloak-gatekeeper
spec:
  replicas: 1
  revisionHistoryLimit: 0
  selector:
    matchLabels:
      app: keycloak-gatekeeper
  template:
    metadata:
      labels:
        app: keycloak-gatekeeper
      annotations:
        checksum/config: 485074e1c0607eca69f97a813313e55bce27515a65f57b11036c8dd074ea3a30
    spec:
      securityContext:
        fsGroup: 1000
        runAsNonRoot: true
        runAsUser: 1000
      containers:
      - name: main
        image: keycloak/keycloak-gatekeeper:5.0.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          containerPort: 3000
          protocol: TCP
        args:
        - --listen=:3000
        - --client-id=$(client_id)
        - --client-secret=$(client_secret)
        - --secure-cookie=$(secure_cookie)
        - --discovery-url=$(discovery_url)
        - --upstream-url=$(upstream_url)
        - --redirection-url=$(redirection_url)
        - --scopes=groups
        - --sign-in-page=/opt/templates/sign_in.html.tmpl
        - --forbidden-page=/opt/templates/forbidden.html.tmpl
        - --enable-refresh-tokens=true
        - --http-only-cookie=true
        - --preserve-host=true
        - --enable-encrypted-token=true
        - --encryption-key=$(encryption_key)
        - --enable-authorization-header
        - --resources=uri=/*
        volumeMounts:
        - name: page-templates
          mountPath: /opt/templates/forbidden.html.tmpl
          subPath: forbidden-page
        - name: page-templates
          mountPath: /opt/templates/sign_in.html.tmpl
          subPath: login-page
        securityContext:
            readOnlyRootFilesystem: true
      volumes:
      - name: page-templates
        configMap:
          name: keycloak-gatekeeper-page-templates
`)
	th.writeF("/manifests/dex-auth/keycloak-gatekeeper/base/service.yaml", `
---
apiVersion: v1
kind: Service
metadata:
  name: keycloak-gatekeeper
spec:
  type: NodePort
  ports:
    - port: 5554
      protocol: TCP
      name: http
      targetPort: http
      nodePort: 32004
  selector:
    app: keycloak-gatekeeper
`)
	th.writeF("/manifests/dex-auth/keycloak-gatekeeper/base/virtualservice.yaml", `
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: keycloak-gatekeeper
spec:
  gateways:
  - kubeflow/kubeflow-gateway
  hosts:
  - '*'
  http:
  - match:
    - port: 5554
      uri:
        prefix: /
    rewrite:
      uri: /
    route:
    - destination:
        host: keycloak-gatekeeper.auth.svc.cluster.local
        port:
          number: 5554
`)
	th.writeF("/manifests/dex-auth/keycloak-gatekeeper/base/params.yaml", `
varReference:
- path: spec/template/spec/containers/args
  kind: Deployment
`)
	th.writeF("/manifests/dex-auth/keycloak-gatekeeper/base/params.env", `
client_id=ldapdexapp
client_secret=pUBnBOY80SnXgjibTYM9ZWNzY2xreNGQok
secure_cookie=false
discovery_url=http://dex.example.com:31200
upstream_url=http://kubeflow.centraldashboard.com:31380
redirection_url=http://keycloak-gatekeeper.example.com:31204
encryption_key=nm6xjpPXPJFInLYo
`)
	th.writeK("/manifests/dex-auth/keycloak-gatekeeper/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: auth

resources:
- config-map.yaml
- namespace.yaml
- deployment.yaml
- service.yaml
- virtualservice.yaml

configMapGenerator:
- name: keycloak-gatekeeper-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true

vars:
- name: client_id
  objref:
    kind: ConfigMap
    name: keycloak-gatekeeper-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.client_id
- name: client_secret
  objref:
    kind: ConfigMap
    name: keycloak-gatekeeper-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.client_secret
- name: secure_cookie
  objref:
    kind: ConfigMap
    name: keycloak-gatekeeper-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.secure_cookie
- name: discovery_url
  objref:
    kind: ConfigMap
    name: keycloak-gatekeeper-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.discovery_url
- name: upstream_url
  objref:
    kind: ConfigMap
    name: keycloak-gatekeeper-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.upstream_url
- name: redirection_url
  objref:
    kind: ConfigMap
    name: keycloak-gatekeeper-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.redirection_url
- name: encryption_key
  objref:
    kind: ConfigMap
    name: keycloak-gatekeeper-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.encryption_key
configurations:
- params.yaml
images:
- name: keycloak/keycloak-gatekeeper
  newName: keycloak/keycloak-gatekeeper
  newTag: 5.0.0
`)
}

func TestKeycloakGatekeeperBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/dex-auth/keycloak-gatekeeper/base")
	writeKeycloakGatekeeperBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../dex-auth/keycloak-gatekeeper/base"
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
