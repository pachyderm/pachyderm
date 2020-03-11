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
  name: cloud-endpoints-controller
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
apiVersion: apps/v1beta1
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
      - env:
        - name: GOOGLE_APPLICATION_CREDENTIALS
          value: /var/run/secrets/sa/admin-gcp-sa.json
        image: gcr.io/cloud-solutions-group/cloud-endpoints-controller:0.2.1
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
        volumeMounts:
        - mountPath: /var/run/secrets/sa
          name: sa-key
          readOnly: true
      serviceAccountName: cloud-endpoints-controller
      terminationGracePeriodSeconds: 5
      volumes:
      - name: sa-key
        secret:
          secretName: $(secretName)
`)
	th.writeF("/manifests/gcp/cloud-endpoints/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cloud-endpoints-controller
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
	targetPath := "../gcp/cloud-endpoints/base"
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
