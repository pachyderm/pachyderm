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

func writeCloudEndpointsBase(th *KustTestHarness) {
	th.writeF("/manifests/gcp/cloud-endpoints/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cloud-endpoints-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cloud-endpoints-controller
subjects:
- kind: ServiceAccount
  name: kf-admin
`)
	th.writeF("/manifests/gcp/cloud-endpoints/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cloud-endpoints-controller
rules:
- apiGroups:
  - ""
  resources:
  - services
  - configmaps
  verbs:
  - get
  - list
- apiGroups:
  - extensions
  - networking.k8s.io
  resources:
  - ingresses
  verbs:
  - get
  - list
`)
	th.writeF("/manifests/gcp/cloud-endpoints/base/composite-controller.yaml", `
apiVersion: metacontroller.k8s.io/v1alpha1
kind: CompositeController
metadata:
  name: cloud-endpoints-controller
spec:
  childResources: []
  clientConfig:
    service:
      caBundle: '...'
      name: cloud-endpoints-controller
      namespace: $(namespace)
  generateSelector: true
  hooks:
    sync:
      webhook:
        url: http://cloud-endpoints-controller.$(namespace)/sync
  parentResource:
    apiVersion: ctl.isla.solutions/v1
    resource: cloudendpoints
  resyncPeriodSeconds: 2
`)
	th.writeF("/manifests/gcp/cloud-endpoints/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cloudendpoints.ctl.isla.solutions
spec:
  group: ctl.isla.solutions
  names:
    kind: CloudEndpoint
    plural: cloudendpoints
    shortNames:
    - cloudep
    - ce
    singular: cloudendpoint
  scope: Namespaced
  version: v1
`)
	th.writeF("/manifests/gcp/cloud-endpoints/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloud-endpoints-controller
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: cloud-endpoints-controller
    spec:
      containers:
      - image: gcr.io/cloud-solutions-group/cloud-endpoints-controller:0.2.1
        imagePullPolicy: Always
        name: cloud-endpoints-controller
        readinessProbe:
          failureThreshold: 2
          httpGet:
            path: /healthz
            port: 80
            scheme: HTTP
          periodSeconds: 5
          successThreshold: 1
          timeoutSeconds: 5
      serviceAccountName: kf-admin
      terminationGracePeriodSeconds: 5
`)
	th.writeF("/manifests/gcp/cloud-endpoints/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kf-admin
`)
	th.writeF("/manifests/gcp/cloud-endpoints/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: cloud-endpoints-controller
spec:
  ports:
  - name: http
    port: 80
  selector:
    app: cloud-endpoints-controller
  type: ClusterIP
`)
	th.writeF("/manifests/gcp/cloud-endpoints/base/params.yaml", `
varReference:
- path: spec/template/spec/volumes/secret/secretName
  kind: Deployment
- path: spec/clientConfig/service/namespace
  kind: CompositeController
- path: spec/hooks/sync/webhook/url
  kind: CompositeController
`)
	th.writeF("/manifests/gcp/cloud-endpoints/base/params.env", `
namespace=kubeflow
secretName=admin-gcp-sa
`)
	th.writeK("/manifests/gcp/cloud-endpoints/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- composite-controller.yaml
- crd.yaml
- deployment.yaml
- service-account.yaml
- service.yaml
commonLabels:
  app: cloud-endpoints-controller
  kustomize.component: cloud-endpoints
images:
- name: gcr.io/cloud-solutions-group/cloud-endpoints-controller
  newName: gcr.io/cloud-solutions-group/cloud-endpoints-controller
  newTag: 0.2.1
configMapGenerator:
- name: cloud-endpoints-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: secretName
  objref:
    kind: ConfigMap
    name: cloud-endpoints-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.secretName
- name: namespace
  objref:
    kind: ConfigMap
    name: cloud-endpoints-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace
configurations:
- params.yaml
`)
}

func TestCloudEndpointsBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/gcp/cloud-endpoints/base")
	writeCloudEndpointsBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../gcp/cloud-endpoints/base"
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
