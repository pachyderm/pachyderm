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

func writeAmbassadorBase(th *KustTestHarness) {
	th.writeF("/manifests/common/ambassador/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: ambassador
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ambassador
subjects:
- kind: ServiceAccount
  name: ambassador
`)
	th.writeF("/manifests/common/ambassador/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: ambassador
rules:
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - create
  - update
  - patch
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - list
  - watch
`)
	th.writeF("/manifests/common/ambassador/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ambassador
spec:
  replicas: 3
  template:
    metadata:
      labels:
        service: ambassador
    spec:
      containers:
      - env:
        - name: AMBASSADOR_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: quay.io/datawire/ambassador:0.37.0
        livenessProbe:
          httpGet:
            path: /ambassador/v0/check_alive
            port: 8877
          initialDelaySeconds: 30
          periodSeconds: 30
        name: ambassador
        readinessProbe:
          httpGet:
            path: /ambassador/v0/check_ready
            port: 8877
          initialDelaySeconds: 30
          periodSeconds: 30
        resources:
          limits:
            cpu: 1
            memory: 400Mi
          requests:
            cpu: 200m
            memory: 100Mi
      restartPolicy: Always
      serviceAccountName: ambassador
`)
	th.writeF("/manifests/common/ambassador/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ambassador
`)
	th.writeF("/manifests/common/ambassador/base/service.yaml", `
---
apiVersion: v1
kind: Service
metadata:
  labels:
    service: ambassador-admin
  name: ambassador-admin
spec:
  ports:
  - name: ambassador-admin
    port: 8877
    targetPort: 8877
  selector:
    service: ambassador
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    # Ambassador is only used on GCP with basic auth.
    beta.cloud.google.com/backend-config: '{"ports": {"ambassador":"basicauth-backendconfig"}}'
  labels:
    service: ambassador
  name: ambassador
spec:
  ports:
  - name: ambassador
    port: 80
    targetPort: 80
  selector:
    service: ambassador
  type: $(ambassadorServiceType)
`)
	th.writeF("/manifests/common/ambassador/base/params.yaml", `
varReference:
- path: spec/type
  kind: Service
`)
	th.writeF("/manifests/common/ambassador/base/deployment-ambassador-patch.yaml", `
- op: replace
  path: /spec/replicas
  value: 1
`)
	th.writeF("/manifests/common/ambassador/base/params.env", `
ambassadorServiceType=ClusterIP
`)
	th.writeK("/manifests/common/ambassador/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- deployment.yaml
- service-account.yaml
- service.yaml
namespace: istio-system
commonLabels:
  kustomize.component: ambassador
images:
- name: quay.io/datawire/ambassador
  newName: quay.io/datawire/ambassador
  newTag: 0.37.0
configMapGenerator:
- name: ambassador-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
patchesJson6902:
- target:
    group: apps
    version: v1
    kind: Deployment
    name: ambassador
  path: deployment-ambassador-patch.yaml
vars:
- name: ambassadorServiceType
  objref:
    kind: ConfigMap
    name: ambassador-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.ambassadorServiceType
configurations:
- params.yaml
`)
}

func TestAmbassadorBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/common/ambassador/base")
	writeAmbassadorBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../common/ambassador/base"
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
