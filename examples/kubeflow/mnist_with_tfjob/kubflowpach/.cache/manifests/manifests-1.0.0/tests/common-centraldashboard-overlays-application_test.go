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

func writeCentraldashboardOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/common/centraldashboard/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: centraldashboard
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: centraldashboard
      app.kubernetes.io/instance: centraldashboard-v1.0.0
      app.kubernetes.io/managed-by: kfctl
      app.kubernetes.io/component: centraldashboard
      app.kubernetes.io/part-of: kubeflow
      app.kubernetes.io/version: v1.0.0
  componentKinds:
  - group: core
    kind: ConfigMap
  - group: apps
    kind: Deployment
  - group: rbac.authorization.k8s.io
    kind: RoleBinding
  - group: rbac.authorization.k8s.io
    kind: Role
  - group: core
    kind: ServiceAccount
  - group: core
    kind: Service
  - group: networking.istio.io
    kind: VirtualService
  descriptor:
    type: centraldashboard
    version: v1beta1
    description: Provides a Dashboard UI for kubeflow
    maintainers:
    - name: Jason Prodonovich
      email: prodonjs@gmail.com
    - name: Apoorv Verma
      email: apverma@google.com
    - name: Adhita Selvaraj
      email: adhita94@gmail.com
    owners:
    - name: Jason Prodonovich
      email: prodonjs@gmail.com
    - name: Apoorv Verma
      email: apverma@google.com
    - name: Adhita Selvaraj
      email: adhita94@gmail.com
    keywords:
     - centraldashboard
     - kubeflow
    links:
    - description: About
      url: https://github.com/kubeflow/kubeflow/tree/master/components/centraldashboard
  addOwnerRef: true

`)
	th.writeK("/manifests/common/centraldashboard/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
commonLabels:
  app.kubernetes.io/name: centraldashboard
  app.kubernetes.io/instance: centraldashboard-v1.0.0
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: centraldashboard
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v1.0.0
`)
	th.writeF("/manifests/common/centraldashboard/base/clusterrole-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: centraldashboard
  name: centraldashboard
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: centraldashboard
subjects:
- kind: ServiceAccount
  name: centraldashboard
  namespace: $(namespace)
`)
	th.writeF("/manifests/common/centraldashboard/base/clusterrole.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app: centraldashboard
  name: centraldashboard
rules:
- apiGroups:
  - ""
  resources:
  - events
  - namespaces
  - nodes
  verbs:
  - get
  - list
  - watch
`)
	th.writeF("/manifests/common/centraldashboard/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: centraldashboard
  name: centraldashboard
spec:
  replicas: 1
  selector:
    matchLabels:
      app: centraldashboard
  template:
    metadata:
      labels:
        app: centraldashboard
    spec:
      containers:
      - image: gcr.io/kubeflow-images-public/centraldashboard
        imagePullPolicy: IfNotPresent
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8082
          initialDelaySeconds: 30
          periodSeconds: 30
        name: centraldashboard
        ports:
        - containerPort: 8082
          protocol: TCP
        env:
        - name: USERID_HEADER
          value: $(userid-header)
        - name: USERID_PREFIX
          value: $(userid-prefix)
        - name: PROFILES_KFAM_SERVICE_HOST
          value: profiles-kfam.kubeflow
      serviceAccountName: centraldashboard
`)
	th.writeF("/manifests/common/centraldashboard/base/role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app: centraldashboard
  name: centraldashboard
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: centraldashboard
subjects:
- kind: ServiceAccount
  name: centraldashboard
  namespace: $(namespace)
`)
	th.writeF("/manifests/common/centraldashboard/base/role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  labels:
    app: centraldashboard
  name: centraldashboard
rules:
- apiGroups:
  - ""
  - "app.k8s.io"
  resources:
  - applications
  - pods
  - pods/exec
  - pods/log
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
`)
	th.writeF("/manifests/common/centraldashboard/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: centraldashboard
`)
	th.writeF("/manifests/common/centraldashboard/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  annotations:
    getambassador.io/config: |-
      ---
      apiVersion: ambassador/v0
      kind:  Mapping
      name: centralui-mapping
      prefix: /
      rewrite: /
      service: centraldashboard.$(namespace)
  labels:
    app: centraldashboard
  name: centraldashboard
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 8082
  selector:
    app: centraldashboard
  sessionAffinity: None
  type: ClusterIP
`)
	th.writeF("/manifests/common/centraldashboard/base/params.yaml", `
varReference:
- path: metadata/annotations/getambassador.io\/config
  kind: Service
- path: spec/http/route/destination/host
  kind: VirtualService
- path: spec/template/spec/containers/0/env/0/value
  kind: Deployment
- path: spec/template/spec/containers/0/env/1/value
  kind: Deployment`)
	th.writeF("/manifests/common/centraldashboard/base/params.env", `
clusterDomain=cluster.local
userid-header=
userid-prefix=`)
	th.writeK("/manifests/common/centraldashboard/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- clusterrole-binding.yaml
- clusterrole.yaml
- deployment.yaml
- role-binding.yaml
- role.yaml
- service-account.yaml
- service.yaml
namespace: kubeflow
commonLabels:
  kustomize.component: centraldashboard
images:
- name: gcr.io/kubeflow-images-public/centraldashboard
  newName: gcr.io/kubeflow-images-public/centraldashboard
  newTag: v1.0.0-g3ec0de71
configMapGenerator:
- envs:
  - params.env
  name: parameters
generatorOptions:
  disableNameSuffixHash: true
vars:
- fieldref:
    fieldPath: metadata.namespace
  name: namespace
  objref:
    apiVersion: v1
    kind: Service
    name: centraldashboard
- fieldref:
    fieldPath: data.clusterDomain
  name: clusterDomain
  objref:
    apiVersion: v1
    kind: ConfigMap
    name: parameters
- fieldref:
    fieldPath: data.userid-header
  name: userid-header
  objref:
    apiVersion: v1
    kind: ConfigMap
    name: parameters
- fieldref:
    fieldPath: data.userid-prefix
  name: userid-prefix
  objref:
    apiVersion: v1
    kind: ConfigMap
    name: parameters
configurations:
- params.yaml
`)
}

func TestCentraldashboardOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/common/centraldashboard/overlays/application")
	writeCentraldashboardOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../common/centraldashboard/overlays/application"
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
