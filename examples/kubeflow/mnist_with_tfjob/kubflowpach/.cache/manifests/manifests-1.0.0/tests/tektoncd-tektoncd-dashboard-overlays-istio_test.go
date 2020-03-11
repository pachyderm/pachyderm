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

func writeTektoncdDashboardOverlaysIstio(th *KustTestHarness) {
	th.writeF("/manifests/tektoncd/tektoncd-dashboard/overlays/istio/virtual-service.yaml", `
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: tektoncd-dashboard
spec:
  gateways:
  - kubeflow-gateway
  hosts:
  - '*'
  http:
  - match:
    - uri:
        prefix: /tektoncd-dashboard
    rewrite:
      uri: /tektoncd-dashboard
    route:
    - destination:
        host: tekton-dashboard.$(namespace).svc.$(clusterDomain)
        port:
          number: 80
    timeout: 300s
`)
	th.writeF("/manifests/tektoncd/tektoncd-dashboard/overlays/istio/params.yaml", `
varReference:
- path: spec/http/route/destination/host
  kind: VirtualService
`)
	th.writeF("/manifests/tektoncd/tektoncd-dashboard/overlays/istio/params.env", `
namespace=
clusterDomain=cluster.local
`)
	th.writeK("/manifests/tektoncd/tektoncd-dashboard/overlays/istio", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- virtual-service.yaml
configMapGenerator:
- name: tektoncd-dashboard-parameters
  env: params.env
vars:
- name: namespace
  objref:
    kind: ConfigMap
    name: tektoncd-dashboard-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace
- name: clusterDomain
  objref:
    kind: ConfigMap
    name: tektoncd-dashboard-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.clusterDomain
configurations:
- params.yaml
`)
	th.writeF("/manifests/tektoncd/tektoncd-dashboard/base/crds.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: extensions.dashboard.tekton.dev
spec:
  group: dashboard.tekton.dev
  names:
    kind: Extension
    plural: extensions
    categories:
      - all
      - tekton-pipelines
  scope: Namespaced
  # Opt into the status subresource so metadata.generation
  # starts to increment
  subresources:
    status: {}
  version: v1alpha1
`)
	th.writeF("/manifests/tektoncd/tektoncd-dashboard/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tekton-dashboard
`)
	th.writeF("/manifests/tektoncd/tektoncd-dashboard/base/cluster-role.yaml", `
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: tekton-dashboard-minimal
rules:
  - apiGroups: ["security.openshift.io"]
    resources: ["securitycontextconstraints"]
    verbs: ["use"]
  - apiGroups: ["extensions", "apps"]
    resources: ["ingresses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["serviceaccounts"]
    verbs: ["get", "list", "update", "patch"]
  - apiGroups: [""]
    resources: ["pods", "services"]
    verbs: ["get", "list", "create", "update", "delete", "patch", "watch"]
  - apiGroups: [""]
    resources: ["pods/log", "namespaces", "events"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["secrets", "configmaps"]
    verbs: ["get", "list", "create", "update", "watch", "delete"]
  - apiGroups: ["extensions", "apps"]
    resources: ["deployments"]
    verbs: ["get", "list", "create", "update", "delete", "patch", "watch"]
  - apiGroups: ["tekton.dev"]
    resources: ["tasks", "clustertasks", "taskruns", "pipelines", "pipelineruns", "pipelineresources"]
    verbs: ["get", "list", "create", "update", "delete", "patch", "watch"]
  - apiGroups: ["tekton.dev"]
    resources: ["taskruns/finalizers", "pipelineruns/finalizers"]
    verbs: ["get", "list", "create", "update", "delete", "patch", "watch"]
  - apiGroups: ["tekton.dev"]
    resources: ["tasks/status", "clustertasks/status", "taskruns/status", "pipelines/status", "pipelineruns/status"]
    verbs: ["get", "list", "create", "update", "delete", "patch", "watch"]
  - apiGroups: ["dashboard.tekton.dev"]
    resources: ["extensions"]
    verbs: ["get", "list", "create", "update", "delete", "patch", "watch"]
`)
	th.writeF("/manifests/tektoncd/tektoncd-dashboard/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: tekton-dashboard-minimal
subjects:
  - kind: ServiceAccount
    name: tekton-dashboard
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: tekton-dashboard-minimal
`)
	th.writeF("/manifests/tektoncd/tektoncd-dashboard/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tekton-dashboard
spec:
  replicas: 1
  template:
    metadata:
      name: tekton-dashboard
    spec:
      containers:
      - name: tekton-dashboard
        image: gcr.io/tekton-nightly/dashboard@sha256:e3e63e7a5e11a14927008cf61f6e6a1bfc36e9e13608e9c044570c162198f01d
        ports:
        - containerPort: 9097
        livenessProbe:
          httpGet:
            path: /health
            port: 9097
        readinessProbe:
          httpGet:
            path: /readiness
            port: 9097
        resources:
        env:
        - name: PORT
          value: "9097"
        - name: WEB_RESOURCES_DIR
          value: /var/run/ko/web
        - name: PIPELINE_RUN_SERVICE_ACCOUNT
          value: ""
        - name: INSTALLED_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
      serviceAccountName: tekton-dashboard
`)
	th.writeF("/manifests/tektoncd/tektoncd-dashboard/base/task.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: Task
metadata:
  name: pipeline0-task
spec:
  inputs:
    resources:
    - name: git-source
      type: git
    params:
    - name: pathToResourceFiles
      description: The path to the resource files to apply
      default: /workspace/git-source
    - name: apply-directory
      description: The directory from which resources are to be applied
      default: "."
    - name: target-namespace
      description: The namespace in which to create the resources being imported
      default: tekton-pipelines
  steps:
  - name: kubectl-apply
    image: lachlanevenson/k8s-kubectl
    command:
    - kubectl
    args:
    - apply
    - -f
    - ${inputs.params.pathToResourceFiles}/${inputs.params.apply-directory} 
    - -n
    - ${inputs.params.target-namespace}
`)
	th.writeF("/manifests/tektoncd/tektoncd-dashboard/base/pipeline.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: Pipeline
metadata:
  name: pipeline0
spec:
  resources:
  - name: git-source
    type: git
  params:
  - name: pathToResourceFiles
    description: The path to the resource files to apply
    default: /workspace/git-source
  - name: apply-directory
    description: The directory from which resources are to be applied
    default: "."
  - name: target-namespace
    description: The namespace in which to create the resources being imported
    default: tekton-pipelines
  tasks:
  - name: pipeline0-task
    taskRef:
      name: pipeline0-task
    params:
    - name: pathToResourceFiles
      value: ${params.pathToResourceFiles}
    - name: apply-directory
      value: ${params.apply-directory}
    - name: target-namespace
      value: ${params.target-namespace}
    resources:
      inputs:
      - name: git-source
        resource: git-source
`)
	th.writeF("/manifests/tektoncd/tektoncd-dashboard/base/service.yaml", `
kind: Service
apiVersion: v1
metadata:
  name: tekton-dashboard
spec:
  ports:
    - name: http
      protocol: TCP
      port: 9097
      targetPort: 9097
`)
	th.writeK("/manifests/tektoncd/tektoncd-dashboard/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- crds.yaml
- service-account.yaml
- cluster-role.yaml
- cluster-role-binding.yaml
- deployment.yaml
- task.yaml
- pipeline.yaml
- service.yaml
namespace: tekton-pipelines
images:
- name: gcr.io/tekton-nightly/dashboard
  newName: gcr.io/tekton-nightly/dashboard
  digest: sha256:e3e63e7a5e11a14927008cf61f6e6a1bfc36e9e13608e9c044570c162198f01d
`)
}

func TestTektoncdDashboardOverlaysIstio(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/tektoncd/tektoncd-dashboard/overlays/istio")
	writeTektoncdDashboardOverlaysIstio(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../tektoncd/tektoncd-dashboard/overlays/istio"
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
