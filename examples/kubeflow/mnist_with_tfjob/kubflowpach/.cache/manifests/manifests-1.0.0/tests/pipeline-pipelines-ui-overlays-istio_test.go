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

func writePipelinesUiOverlaysIstio(th *KustTestHarness) {
	th.writeF("/manifests/pipeline/pipelines-ui/overlays/istio/virtual-service.yaml", `
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: ml-pipeline-tensorboard-ui
spec:
  gateways:
  - kubeflow-gateway
  hosts:
  - '*'
  http:
  - match:
    - uri:
        prefix: /data
    rewrite:
      uri: /data
    route:
    - destination:
        host: $(tensorboard-service).$(ui-namespace).svc.$(ui-clusterDomain)
        port:
          number: 80
    timeout: 300s
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: ml-pipeline-ui
spec:
  gateways:
  - kubeflow-gateway
  hosts:
  - '*'
  http:
  - match:
    - uri:
        prefix: /pipeline
    rewrite:
      uri: /pipeline
    route:
    - destination:
        host: $(service).$(ui-namespace).svc.$(ui-clusterDomain)
        port:
          number: 80
    timeout: 300s
`)
	th.writeF("/manifests/pipeline/pipelines-ui/overlays/istio/params.yaml", `
varReference:
- path: spec/http/route/destination/host
  kind: VirtualService
`)
	th.writeK("/manifests/pipeline/pipelines-ui/overlays/istio", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- virtual-service.yaml
configurations:
- params.yaml
`)
	th.writeF("/manifests/pipeline/pipelines-ui/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: ml-pipeline-ui
  name: ml-pipeline-ui
spec:
  selector:
    matchLabels:
      app: ml-pipeline-ui
  template:
    metadata:
      labels:
        app: ml-pipeline-ui
    spec:
      containers:
      - name: ml-pipeline-ui
        image: gcr.io/ml-pipeline/frontend
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 3000
      serviceAccountName: ml-pipeline-ui
`)
	th.writeF("/manifests/pipeline/pipelines-ui/base/role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  labels:
    app: ml-pipeline-ui
  name: ml-pipeline-ui
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: ml-pipeline-ui
subjects:
- kind: ServiceAccount
  name: ml-pipeline-ui
`)
	th.writeF("/manifests/pipeline/pipelines-ui/base/role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  labels:
    app: ml-pipeline-ui
  name: ml-pipeline-ui
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - pods/log
  verbs:
  - create
  - get
  - list
- apiGroups:
  - "kubeflow.org"
  resources:
  - viewers
  verbs:
  - create
  - get
  - list
  - watch
  - delete
`)
	th.writeF("/manifests/pipeline/pipelines-ui/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ml-pipeline-ui
`)
	th.writeF("/manifests/pipeline/pipelines-ui/base/service.yaml", `
---
apiVersion: v1
kind: Service
metadata:
  name: ml-pipeline-ui
  annotations:
    getambassador.io/config: |-
      ---
      apiVersion: ambassador/v0
      kind:  Mapping
      name: pipelineui-mapping
      prefix: /pipeline
      rewrite: /pipeline
      timeout_ms: 300000
      service: $(service).$(ui-namespace)
      use_websocket: true
  labels:
    app: ml-pipeline-ui
spec:
  ports:
  - port: 80
    targetPort: 3000
  selector:
    app: ml-pipeline-ui
---
apiVersion: v1
kind: Service
metadata:
  name: ml-pipeline-tensorboard-ui
  annotations:
    getambassador.io/config: |-
      ---
      apiVersion: ambassador/v0
      kind:  Mapping
      name: pipeline-tensorboard-ui-mapping
      prefix: /data
      rewrite: /data
      timeout_ms: 300000
      service: $(service).$(ui-namespace)
      use_websocket: true
  labels:
    app: ml-pipeline-tensorboard-ui
spec:
  ports:
  - port: 80
    targetPort: 3000
  selector:
    app: ml-pipeline-tensorboard-ui
`)
	th.writeF("/manifests/pipeline/pipelines-ui/base/params.yaml", `
varReference:
- path: metadata/annotations/getambassador.io\/config
  kind: Service
`)
	th.writeF("/manifests/pipeline/pipelines-ui/base/params.env", `
uiClusterDomain=cluster.local
`)
	th.writeK("/manifests/pipeline/pipelines-ui/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kubeflow
resources:
- deployment.yaml
- role-binding.yaml
- role.yaml
- service-account.yaml
- service.yaml
configMapGenerator:
- name: ui-parameters
  env: params.env
images:
- name: gcr.io/ml-pipeline/frontend
  newTag: 0.2.0
  newName: gcr.io/ml-pipeline/frontend
vars:
- name: ui-namespace
  objref:
    kind: Service
    name: ml-pipeline-ui
    apiVersion: v1
  fieldref:
    fieldpath: metadata.namespace
- name: ui-clusterDomain
  objref:
    kind: ConfigMap
    name: ui-parameters
    version: v1
  fieldref:
    fieldpath: data.uiClusterDomain
- name: service
  objref:
    kind: Service
    name: ml-pipeline-ui
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
- name: tensorboard-service
  objref:
    kind: Service
    name: ml-pipeline-tensorboard-ui
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
configurations:
- params.yaml
`)
}

func TestPipelinesUiOverlaysIstio(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pipeline/pipelines-ui/overlays/istio")
	writePipelinesUiOverlaysIstio(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pipeline/pipelines-ui/overlays/istio"
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
