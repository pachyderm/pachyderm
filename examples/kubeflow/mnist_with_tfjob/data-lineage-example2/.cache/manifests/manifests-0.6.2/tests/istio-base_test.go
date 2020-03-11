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

func writeIstioBase(th *KustTestHarness) {
	th.writeF("/manifests/istio/istio/base/kf-istio-resources.yaml", `
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: kubeflow-gateway
spec:
  selector:
    istio: ingressgateway
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
	th.writeF("/manifests/istio/istio/base/params.env", `
clusterRbacConfig=ON
`)
	th.writeF("/manifests/istio/istio/base/params.yaml", `
varReference:
- path: spec/mode
  kind: ClusterRbacConfig
`)
	th.writeK("/manifests/istio/istio/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- kf-istio-resources.yaml
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
	targetPath := "../istio/istio/base"
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
