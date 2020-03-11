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

func writeWebhookOverlaysCertManager(th *KustTestHarness) {
	th.writeF("/manifests/admission-webhook/webhook/overlays/cert-manager/certificate.yaml", `
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  name: admission-webhook-cert
spec:
  isCA: true
  commonName: $(serviceName).$(namespace).svc
  dnsNames:
  - $(serviceName).$(namespace).svc
  - $(serviceName).$(namespace).svc.cluster.local
  issuerRef:
    kind: ClusterIssuer
    name: $(issuer)
  secretName: webhook-certs`)
	th.writeF("/manifests/admission-webhook/webhook/overlays/cert-manager/mutating-webhook-configuration.yaml", `
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: mutating-webhook-configuration
  annotations:
    cert-manager.io/inject-ca-from: $(namespace)/$(cert_name)
  `)
	th.writeF("/manifests/admission-webhook/webhook/overlays/cert-manager/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment
spec:
  template:
    spec:
      containers:
      - name: admission-webhook
        args:
        - --tlsCertFile=/etc/webhook/certs/tls.crt
        - --tlsKeyFile=/etc/webhook/certs/tls.key
`)
	th.writeF("/manifests/admission-webhook/webhook/overlays/cert-manager/params.yaml", `
varReference:
- path: spec/commonName
  kind: Certificate
- path: spec/dnsNames
  kind: Certificate
- path: spec/issuerRef/name
  kind: Certificate
- path: metadata/annotations
  kind: MutatingWebhookConfiguration
`)
	th.writeF("/manifests/admission-webhook/webhook/overlays/cert-manager/params.env", `
issuer=kubeflow-self-signing-issuer`)
	th.writeK("/manifests/admission-webhook/webhook/overlays/cert-manager", `
bases:
- ../../base

resources:
- certificate.yaml

patchesStrategicMerge:
- mutating-webhook-configuration.yaml
- deployment.yaml

configMapGenerator:
- name: admission-webhook-parameters
  behavior: merge
  env: params.env
generatorOptions:
  disableNameSuffixHash: true

vars:
- name: issuer
  objref:
    kind: ConfigMap
    name: admission-webhook-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.issuer
- name: cert_name
  objref:
      kind: Certificate
      group: cert-manager.io
      version: v1alpha2
      name: admission-webhook-cert
  fieldref:
    fieldpath: metadata.name

configurations:
- params.yaml`)
	th.writeF("/manifests/admission-webhook/webhook/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cluster-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-role
subjects:
- kind: ServiceAccount
  name: service-account
`)
	th.writeF("/manifests/admission-webhook/webhook/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-role
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - poddefaults
  verbs:
  - get
  - watch
  - list
  - update
  - create
  - patch
  - delete

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-poddefaults-admin
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-admin: "true"
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.authorization.kubeflow.org/aggregate-to-kubeflow-poddefaults-admin: "true"
rules: []

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-poddefaults-edit
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-edit: "true"
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.authorization.kubeflow.org/aggregate-to-kubeflow-poddefaults-edit: "true"
rules: []

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-poddefaults-view
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-poddefaults-admin: "true"
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-poddefaults-edit: "true"
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-view: "true"
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - poddefaults
  verbs:
  - get
  - list
  - watch
`)
	th.writeF("/manifests/admission-webhook/webhook/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment
spec:
  template:
    spec:
      containers:
      - image: gcr.io/kubeflow-images-public/admission-webhook:v20190520-v0-139-gcee39dbc-dirty-0d8f4c
        name: admission-webhook
        volumeMounts:
        - mountPath: /etc/webhook/certs
          name: webhook-cert
          readOnly: true
      volumes:
      - name: webhook-cert
        secret:
          secretName: webhook-certs
      serviceAccountName: service-account    
`)
	th.writeF("/manifests/admission-webhook/webhook/base/mutating-webhook-configuration.yaml", `
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: mutating-webhook-configuration
webhooks:
- clientConfig:
    caBundle: ""
    service:
      name: $(serviceName)
      namespace: $(namespace)
      path: /apply-poddefault
  name: $(deploymentName).kubeflow.org
  rules:
  - apiGroups:
    - ""
    apiVersions:
    - v1
    operations:
    - CREATE
    resources:
    - pods
`)
	th.writeF("/manifests/admission-webhook/webhook/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: service-account
`)
	th.writeF("/manifests/admission-webhook/webhook/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: service
spec:
  ports:
  - port: 443
    targetPort: 443
`)
	th.writeF("/manifests/admission-webhook/webhook/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: poddefaults.kubeflow.org
spec:
  group: kubeflow.org
  names:
    kind: PodDefault
    plural: poddefaults
    singular: poddefault
  scope: Namespaced
  version: v1alpha1
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          type: string
        kind:
          type: string
        metadata:
          type: object
        spec:
          properties:
            desc:
              type: string
            serviceAccountName:
              type: string
            env:
              items:
                type: object
              type: array
            envFrom:
              items:
                type: object
              type: array
            selector:
              type: object
            volumeMounts:
              items:
                type: object
              type: array
            volumes:
              items:
                type: object
              type: array
          required:
          - selector
          type: object
        status:
          type: object
      type: object
`)
	th.writeF("/manifests/admission-webhook/webhook/base/params.yaml", `
varReference:
- path: webhooks/clientConfig/service/namespace
  kind: MutatingWebhookConfiguration
- path: webhooks/clientConfig/service/name
  kind: MutatingWebhookConfiguration
- path: webhooks/name
  kind: MutatingWebhookConfiguration
`)
	th.writeF("/manifests/admission-webhook/webhook/base/params.env", `
namespace=kubeflow
`)
	th.writeK("/manifests/admission-webhook/webhook/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- deployment.yaml
- mutating-webhook-configuration.yaml
- service-account.yaml
- service.yaml
- crd.yaml
commonLabels:
  app: admission-webhook
  kustomize.component: admission-webhook
namePrefix: admission-webhook-
images:
- name: gcr.io/kubeflow-images-public/admission-webhook
  newName: gcr.io/kubeflow-images-public/admission-webhook
  newTag: v1.0.0-gaf96e4e3
namespace: kubeflow
configMapGenerator:
- envs:
  - params.env
  name: admission-webhook-parameters
generatorOptions:
  disableNameSuffixHash: true
vars:
- fieldref:
    fieldPath: data.namespace
  name: namespace
  objref:
    apiVersion: v1
    kind: ConfigMap
    name: admission-webhook-parameters
- fieldref:
    fieldPath: metadata.name
  name: serviceName
  objref:
    apiVersion: v1
    kind: Service
    name: service
- fieldref:
    fieldPath: metadata.name
  name: deploymentName
  objref:
    apiVersion: apps/v1
    kind: Deployment
    name: deployment
configurations:
- params.yaml
`)
}

func TestWebhookOverlaysCertManager(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/admission-webhook/webhook/overlays/cert-manager")
	writeWebhookOverlaysCertManager(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../admission-webhook/webhook/overlays/cert-manager"
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
