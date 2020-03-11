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

func writeBasicAuthOverlaysIstio(th *KustTestHarness) {
	th.writeF("/manifests/common/basic-auth/overlays/istio/kflogin-virtual-service.yaml", `
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: basic-auth-login
spec:
  gateways:
  - kubeflow-gateway
  hosts:
  - '*'
  http:
  - match:
    - uri:
        prefix: /kflogin
    rewrite:
      uri: /kflogin
    route:
    - destination:
        host: basic-auth-login.$(service-namespace).svc.$(clusterDomain)
        port:
          number: 8085
`)
	th.writeF("/manifests/common/basic-auth/overlays/istio/params.yaml", `
varReference:
- path: spec/http/route/destination/host
  kind: VirtualService
`)
	th.writeK("/manifests/common/basic-auth/overlays/istio", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- kflogin-virtual-service.yaml
configurations:
- params.yaml
`)
	th.writeF("/manifests/common/basic-auth/base/kflogin-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: basic-auth-login
spec:
  selector:
    matchLabels:
      app: basic-auth-login
  replicas: 1
  strategy:
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: basic-auth-login
    spec:
      containers:
      - name: app
        image: gcr.io/kubeflow-images-public/kflogin-ui:v0.5.0
        ports:
        - containerPort: 5000
`)
	th.writeF("/manifests/common/basic-auth/base/gatekeeper-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: basic-auth
spec:
  selector:
    matchLabels:
      app: basic-auth
  replicas: 1
  strategy:
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: basic-auth
    spec:
      containers:
      - name: app
        args:
        - --username=$(USERNAME)
        - --pwhash=$(PASSWORDHASH)
        command:
        - /opt/kubeflow/gatekeeper
        env:
        - name: USERNAME
          valueFrom:
            secretKeyRef:
              key: username
              name: $(authSecretName)
        - name: PASSWORDHASH
          valueFrom:
            secretKeyRef:
              key: passwordhash
              name: $(authSecretName)
        image: gcr.io/kubeflow-images-public/gatekeeper:v0.5.0
        ports:
        - containerPort: 8085
        workingDir: /opt/kubeflow
`)
	th.writeF("/manifests/common/basic-auth/base/gatekeeper-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  annotations:
    getambassador.io/config: |-
      ---
      apiVersion: ambassador/v0
      kind:  AuthService
      name: basic-auth
      auth_service: basic-auth.$(service-namespace):8085
      allowed_headers:
      - "x-from-login"
  labels:
    app: basic-auth
  name: basic-auth
spec:
  ports:
  - port: 8085
    targetPort: 8085
  selector:
    app: basic-auth
  type: ClusterIP
`)
	th.writeF("/manifests/common/basic-auth/base/kflogin-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  annotations:
    getambassador.io/config: |-
      ---
      apiVersion: ambassador/v0
      kind:  Mapping
      name: kflogin-mapping
      prefix: /kflogin
      rewrite: /kflogin
      timeout_ms: 300000
      service: basic-auth-login.$(service-namespace)
      use_websocket: true
  labels:
    app: basic-auth-login
  name: basic-auth-login
spec:
  ports:
  - port: 80
    targetPort: 5000
  selector:
    app: basic-auth-login
  type: ClusterIP
`)
	th.writeF("/manifests/common/basic-auth/base/params.yaml", `
varReference:
- path: metadata/annotations/getambassador.io\/config
  kind: Service
- path: spec/template/spec/containers/env/valueFrom/secretKeyRef/name
  kind: Deployment
`)
	th.writeF("/manifests/common/basic-auth/base/params.env", `
authSecretName=kubeflow-login
clusterDomain=cluster.local
`)
	th.writeK("/manifests/common/basic-auth/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- kflogin-deployment.yaml
- gatekeeper-deployment.yaml
- gatekeeper-service.yaml
- kflogin-service.yaml
commonLabels:
  kustomize.component: basic-auth
namespace: kubeflow
images:
  - name: gcr.io/kubeflow-images-public/kflogin-ui
    newName: gcr.io/kubeflow-images-public/kflogin-ui
    newTag: v0.5.0
  - name: gcr.io/kubeflow-images-public/gatekeeper
    newName: gcr.io/kubeflow-images-public/gatekeeper
    newTag: v0.5.0
generatorOptions:
  disableNameSuffixHash: true
configMapGenerator:
- name: basic-auth-parameters
  env: params.env
vars:
- name: service-namespace
  objref:
    kind: Service
    name: basic-auth-login
    apiVersion: v1
  fieldref:
    fieldpath: metadata.namespace
- name: authSecretName
  objref:
    kind: ConfigMap
    name: basic-auth-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.authSecretName
- name: clusterDomain
  objref:
    kind: ConfigMap
    name: basic-auth-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.clusterDomain
configurations:
- params.yaml
`)
}

func TestBasicAuthOverlaysIstio(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/common/basic-auth/overlays/istio")
	writeBasicAuthOverlaysIstio(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../common/basic-auth/overlays/istio"
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
