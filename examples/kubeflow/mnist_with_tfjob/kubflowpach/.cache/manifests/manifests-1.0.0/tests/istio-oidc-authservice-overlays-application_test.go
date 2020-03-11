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

func writeOidcAuthserviceOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/istio/oidc-authservice/overlays/application/application.yaml", `
 
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: oidc-authservice
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: oidc-authservice
      app.kubernetes.io/instance: oidc-authservice-v1.0.0
      app.kubernetes.io/managed-by: kfctl
      app.kubernetes.io/component: oidc-authservice
      app.kubernetes.io/part-of: kubeflow
      app.kubernetes.io/version: v1.0.0
  componentKinds:
  - group: apps
    kind: StatefulSet
  - group: core
    kind: Service
  - group: core
    kind: PersistentVolumeClaim
  - group: networking.istio.io
    kind: EnvoyFilter
  descriptor:
    type: oidc-authservice
    version: v1beta1
    description: Provides OIDC-based authentication for Kubeflow Applications, at the Istio Gateway.
    maintainers:
    - name: Yannis Zarkadas
      email: yanniszark@arrikto.com
    owners:
    - name: Yannis Zarkadas
      email: yanniszark@arrikto.com
    keywords:
     - oidc
     - authservice
     - authentication  
    links:
    - description: About
      url: https://github.com/kubeflow/kubeflow/tree/master/components/oidc-authservice
    - description: Docs
      url: https://www.kubeflow.org/docs/started/k8s/kfctl-existing-arrikto
  addOwnerRef: true
`)
	th.writeK("/manifests/istio/oidc-authservice/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
commonLabels:
  app.kubernetes.io/name: oidc-authservice
  app.kubernetes.io/instance: oidc-authservice-v1.0.0
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: oidc-authservice
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v1.0.0
`)
	th.writeF("/manifests/istio/oidc-authservice/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: authservice
spec:
  type: ClusterIP
  selector:
    app: authservice
  ports:
  - port: 8080
    name: http-authservice
    targetPort: http-api
  publishNotReadyAddresses: true`)
	th.writeF("/manifests/istio/oidc-authservice/base/statefulset.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: authservice
spec:
  replicas: 1
  selector:
    matchLabels:
      app: authservice
  serviceName: authservice
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
      labels:
        app: authservice
    spec:
      containers:
      - name: authservice
        image: gcr.io/arrikto/kubeflow/oidc-authservice:6ac9400
        imagePullPolicy: Always
        ports:
        - name: http-api
          containerPort: 8080
        env:
          - name: USERID_HEADER
            value: $(userid-header)
          - name: USERID_PREFIX
            value: $(userid-prefix)
          - name: USERID_CLAIM
            value: email
          - name: OIDC_PROVIDER
            value: $(oidc_provider)
          - name: OIDC_AUTH_URL
            value: $(oidc_auth_url)
          - name: OIDC_SCOPES
            value: "profile email groups"
          - name: REDIRECT_URL
            value: $(oidc_redirect_uri)
          - name: SKIP_AUTH_URI
            value: $(skip_auth_uri)
          - name: PORT
            value: "8080"
          - name: CLIENT_ID
            value: $(client_id)
          - name: CLIENT_SECRET
            value: $(application_secret)
          - name: STORE_PATH
            value: /var/lib/authservice/data.db
        volumeMounts:
          - name: data
            mountPath: /var/lib/authservice
        readinessProbe:
            httpGet:
              path: /
              port: 8081
      securityContext:
        fsGroup: 111
      volumes:
        - name: data
          persistentVolumeClaim:
              claimName: authservice-pvc          
`)
	th.writeF("/manifests/istio/oidc-authservice/base/envoy-filter.yaml", `
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: authn-filter
spec:
  workloadLabels:
    istio: ingressgateway
  filters:
  - filterConfig:
      httpService:
        serverUri:
          uri: http://authservice.$(namespace).svc.cluster.local
          cluster: outbound|8080||authservice.$(namespace).svc.cluster.local
          failureModeAllow: false
          timeout: 10s
        authorizationRequest:
          allowedHeaders:
            patterns:
            - exact: "cookie"
            - exact: "X-Auth-Token"
        authorizationResponse:
          allowedUpstreamHeaders:
            patterns:
            - exact: "kubeflow-userid"
      statusOnError:
        code: GatewayTimeout
    filterName: envoy.ext_authz
    filterType: HTTP
    insertPosition:
      index: FIRST
    listenerMatch:
      listenerType: GATEWAY
`)
	th.writeF("/manifests/istio/oidc-authservice/base/pvc.yaml", `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: authservice-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi`)
	th.writeF("/manifests/istio/oidc-authservice/base/params.yaml", `
varReference:
- path: spec/template/spec/containers/env/value
  kind: StatefulSet
- path: spec/filters/filterConfig/httpService/serverUri/uri
  kind: EnvoyFilter
- path: spec/filters/filterConfig/httpService/serverUri/cluster
  kind: EnvoyFilter`)
	th.writeF("/manifests/istio/oidc-authservice/base/params.env", `
client_id=ldapdexapp
oidc_provider=
oidc_redirect_uri=
oidc_auth_url=
application_secret=pUBnBOY80SnXgjibTYM9ZWNzY2xreNGQok
skip_auth_uri=
userid-header=
userid-prefix=
namespace=istio-system`)
	th.writeK("/manifests/istio/oidc-authservice/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- service.yaml
- statefulset.yaml
- envoy-filter.yaml
- pvc.yaml

namespace: istio-system

configMapGenerator:
- name: oidc-authservice-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true

vars:
- name: client_id
  objref:
    kind: ConfigMap
    name: oidc-authservice-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.client_id
- name: oidc_provider
  objref:
    kind: ConfigMap
    name: oidc-authservice-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.oidc_provider
- name: oidc_redirect_uri
  objref:
    kind: ConfigMap
    name: oidc-authservice-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.oidc_redirect_uri
- name: oidc_auth_url
  objref:
    kind: ConfigMap
    name: oidc-authservice-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.oidc_auth_url
- name: application_secret
  objref:
    kind: ConfigMap
    name: oidc-authservice-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.application_secret
- name: skip_auth_uri
  objref:
    kind: ConfigMap
    name: oidc-authservice-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.skip_auth_uri
- name: userid-header
  objref:
    kind: ConfigMap
    name: oidc-authservice-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.userid-header
- name: userid-prefix
  objref:
    kind: ConfigMap
    name: oidc-authservice-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.userid-prefix
- name: namespace
  objref:
    kind: ConfigMap
    name: oidc-authservice-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace
configurations:
- params.yaml
images:
- name: gcr.io/arrikto/kubeflow/oidc-authservice
  newName: gcr.io/arrikto/kubeflow/oidc-authservice
  newTag: 28c59ef
`)
}

func TestOidcAuthserviceOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/istio/oidc-authservice/overlays/application")
	writeOidcAuthserviceOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../istio/oidc-authservice/overlays/application"
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
