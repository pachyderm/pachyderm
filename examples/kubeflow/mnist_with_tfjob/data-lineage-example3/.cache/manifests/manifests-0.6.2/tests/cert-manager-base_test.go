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

func writeCertManagerBase(th *KustTestHarness) {
	th.writeF("/manifests/cert-manager/cert-manager/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager
subjects:
- kind: ServiceAccount
  name: cert-manager
---
apiVersion: certmanager.k8s.io/v1alpha1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    email: $(acmeEmail)
    http01: {}
    privateKeySecretRef:
      name: letsencrypt-prod-secret
    server: $(acmeUrl)
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cert-manager
rules:
- apiGroups:
  - certmanager.k8s.io
  resources:
  - certificates
  - issuers
  - clusterissuers
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - secrets
  - events
  - endpoints
  - services
  - pods
  - configmaps
  verbs:
  - '*'
- apiGroups:
  - extensions
  resources:
  - ingresses
  verbs:
  - '*'
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/deployment.yaml", `
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  labels:
    app: cert-manager
  name: cert-manager
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: cert-manager
    spec:
      containers:
      - args:
        - --cluster-resource-namespace=kubeflow
        - --leader-election-namespace=kubeflow
        image: quay.io/jetstack/cert-manager-controller:v0.4.0
        imagePullPolicy: IfNotPresent
        name: cert-manager
      serviceAccountName: cert-manager
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cert-manager
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/params.yaml", `
varReference:
- path: spec/acme/email
  kind: ClusterIssuer
- path: spec/acme/server
  kind: ClusterIssuer
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/params.env", `
acmeEmail=
acmeUrl=https://acme-v02.api.letsencrypt.org/directory
`)
	th.writeK("/manifests/cert-manager/cert-manager/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- deployment.yaml
- service-account.yaml
commonLabels:
  kustomize.component: cert-manager
images:
  - name: quay.io/jetstack/cert-manager-controller
    newName: quay.io/jetstack/cert-manager-controller
    newTag: v0.4.0
configMapGenerator:
- name: cert-manager-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: acmeEmail
  objref:
    kind: ConfigMap
    name: cert-manager-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.acmeEmail
- name: acmeUrl
  objref:
    kind: ConfigMap
    name: cert-manager-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.acmeUrl
configurations:
- params.yaml
`)
}

func TestCertManagerBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/cert-manager/cert-manager/base")
	writeCertManagerBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../cert-manager/cert-manager/base"
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
