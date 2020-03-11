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

func writeIstioBase(th *KustTestHarness) {
	th.writeF("/manifests/istio/istio/base/kf-istio-resources.yaml", `
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: kubeflow-gateway
spec:
  selector:
    istio: $(gatewaySelector)
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "*"
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: grafana-vs
spec:
  hosts:
  - "*"
  gateways:
  - "kubeflow-gateway"
  http:
  - match:
    - uri:
        prefix: "/istio/grafana/"
      method:
        exact: "GET"
    rewrite:
      uri: "/"
    route:
    - destination:
        host: "grafana.istio-system.svc.cluster.local"
        port:
          number: 3000
---
apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: google-api-entry
spec:
  hosts:
  - www.googleapis.com
  ports:
  - number: 443
    name: https
    protocol: HTTPS
  resolution: DNS
  location: MESH_EXTERNAL
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: google-api-vs
spec:
  hosts:
  - www.googleapis.com
  tls:
  - match:
    - port: 443
      sni_hosts:
      - www.googleapis.com
    route:
    - destination:
        host: www.googleapis.com
        port:
          number: 443
      weight: 100
---
apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: google-storage-api-entry
spec:
  hosts:
  - storage.googleapis.com
  ports:
  - number: 443
    name: https
    protocol: HTTPS
  resolution: DNS
  location: MESH_EXTERNAL
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: google-storage-api-vs
spec:
  hosts:
  - storage.googleapis.com
  tls:
  - match:
    - port: 443
      sni_hosts:
      - storage.googleapis.com
    route:
    - destination:
        host: storage.googleapis.com
        port:
          number: 443
      weight: 100
---
apiVersion: rbac.istio.io/v1alpha1
kind: ClusterRbacConfig
metadata:
  name: default
spec:
  mode: $(clusterRbacConfig)
`)
	th.writeF("/manifests/istio/istio/base/cluster-roles.yaml", `
---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-istio-admin
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-admin: "true"
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.authorization.kubeflow.org/aggregate-to-kubeflow-istio-admin: "true"
rules: []

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-istio-edit
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-edit: "true"
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-istio-admin: "true"
rules:
- apiGroups: 
  - istio.io
  - networking.istio.io
  resources: ["*"]
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
  name: kubeflow-istio-view
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-view: "true"
rules:
- apiGroups:
  - istio.io
  - networking.istio.io
  resources: ["*"]
  verbs:
  - get
  - list
  - watch
`)
	th.writeF("/manifests/istio/istio/base/params.yaml", `
varReference:
- path: spec/mode
  kind: ClusterRbacConfig
- path: spec/selector
  kind: Gateway`)
	th.writeF("/manifests/istio/istio/base/params.env", `
clusterRbacConfig=ON
gatewaySelector=ingressgateway`)
	th.writeK("/manifests/istio/istio/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- kf-istio-resources.yaml
- cluster-roles.yaml
namespace: kubeflow
configMapGenerator:
- name: istio-parameters
  env: params.env
vars:
- name: clusterRbacConfig
  objref:
    kind: ConfigMap
    name: istio-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.clusterRbacConfig
- name: gatewaySelector
  objref:
    kind: ConfigMap
    name: istio-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.gatewaySelector
configurations:
- params.yaml
`)
}

func TestIstioBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/istio/istio/base")
	writeIstioBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../istio/istio/base"
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
