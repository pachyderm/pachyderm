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

func writeTensorboardOverlaysIstio(th *KustTestHarness) {
	th.writeF("/manifests/tensorboard/overlays/istio/virtual-service.yaml", `
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: tensorboard
spec:
  gateways:
  - kubeflow-gateway
  hosts:
  - '*'
  http:
  - match:
    - uri:
        prefix: /tensorboard/tensorboard/
    rewrite:
      uri: /
    route:
    - destination:
        host: tensorboard.$(namespace).svc.$(clusterDomain)
        port:
          number: 9000
`)
	th.writeF("/manifests/tensorboard/overlays/istio/params.yaml", `
varReference:
- path: spec/http/route/destination/host
  kind: VirtualService
`)
	th.writeK("/manifests/tensorboard/overlays/istio", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- virtual-service.yaml
configurations:
- params.yaml
`)
	th.writeF("/manifests/tensorboard/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: tensorboard
  name: tensorboard
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: tensorboard
    spec:
      containers:
      - args:
        - --logdir=logs
        - --port=6006
        command:
        - /usr/local/bin/tensorboard
        image: tensorflow/tensorflow:1.8.0
        imagePullPolicy: IfNotPresent
        name: tensorboard
        ports:
        - containerPort: 6006
        resources:
          limits:
            cpu: "4"
            memory: 4Gi
          requests:
            cpu: "1"
            memory: 1Gi
`)
	th.writeF("/manifests/tensorboard/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  annotations:
    getambassador.io/config: |-
      ---
      apiVersion: ambassador/v0
      kind:  Mapping
      name: tb-mapping-tensorboard-get
      prefix: /tensorboard/ tensorboard/
      rewrite: /
      method: GET
      service: tensorboard.$(namespace):9000
  labels:
    app: tensorboard
  name: tensorboard
spec:
  ports:
  - name: tb
    port: 9000
    targetPort: 6006
  selector:
    app: tensorboard
  type: ClusterIP
`)
	th.writeF("/manifests/tensorboard/base/params.yaml", `
varReference:
- path: metadata/annotations/getambassador.io\/config
  kind: Service
`)
	th.writeF("/manifests/tensorboard/base/params.env", `
# GCP
# @optionalParam logDir string logs Name of the log directory holding the TF events file
# @optionalParam targetPort number 6006 Name of the targetPort
# @optionalParam servicePort number 9000 Name of the servicePort
# @optionalParam serviceType string ClusterIP The service type for Jupyterhub.
# @optionalParam defaultTbImage string tensorflow/tensorflow:1.8.0 default tensorboard image to use
# @optionalParam gcpCredentialSecretName string null Name of the k8s secrets containing gcp credentials
# AWS
# @optionalParam logDir string logs Name of the log directory holding the TF events file
# @optionalParam targetPort number 6006 Name of the targetPort
# @optionalParam servicePort number 9000 Name of the servicePort
# @optionalParam serviceType string ClusterIP The service type for tensorboard service
# @optionalParam defaultTbImage string tensorflow/tensorflow:1.8.0 default tensorboard image to use
# @optionalParam s3Enabled string false Whether or not to use S3
# @optionalParam s3SecretName string null Name of the k8s secrets containing S3 credentials
# @optionalParam s3SecretAccesskeyidKeyName string null Name of the key in the k8s secret containing AWS_ACCESS_KEY_ID
# @optionalParam s3SecretSecretaccesskeyKeyName string null Name of the key in the k8s secret containing AWS_SECRET_ACCESS_KEY
# @optionalParam s3AwsRegion string us-west-1 S3 region
# @optionalParam s3UseHttps string true Whether or not to use https
# @optionalParam s3VerifySsl string true Whether or not to verify https certificates for S3 connections
# @optionalParam s3Endpoint string s3.us-west-1.amazonaws.com URL for your s3-compatible endpoint
# @optionalParam efsEnabled string false Whether or not to use EFS
# @optionalParam efsPvcName string null Name of the Persistent Volume Claim used for EFS
# @optionalParam efsVolumeName string null Name of the Volume to mount to the pod
# @optionalParam efsMountPath string null Where to mount the EFS Volume
namespace=
clusterDomain=cluster.local
`)
	th.writeK("/manifests/tensorboard/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kubeflow
resources:
- deployment.yaml
- service.yaml
commonLabels:
  kustomize.component: tensorboard
configMapGenerator:
- name: parameters
  env: params.env
vars:
- name: namespace
  objref:
    kind: Service
    name: tensorboard
    apiVersion: v1
  fieldref:
    fieldpath: metadata.namespace
- name: clusterDomain
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.clusterDomain
configurations:
- params.yaml
images:
- name: tensorflow/tensorflow
  newName: tensorflow/tensorflow
  newTag: 1.8.0
`)
}

func TestTensorboardOverlaysIstio(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/tensorboard/overlays/istio")
	writeTensorboardOverlaysIstio(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../tensorboard/overlays/istio"
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
