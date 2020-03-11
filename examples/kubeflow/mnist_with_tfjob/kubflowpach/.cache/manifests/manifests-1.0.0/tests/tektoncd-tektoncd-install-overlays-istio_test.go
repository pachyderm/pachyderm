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

func writeTektoncdInstallOverlaysIstio(th *KustTestHarness) {
	th.writeF("/manifests/tektoncd/tektoncd-install/overlays/istio/virtual-service.yaml", `
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: tektoncd
spec:
  gateways:
  - kubeflow-gateway
  hosts:
  - '*'
  http:
  - match:
    - uri:
        prefix: /tektoncd/
    rewrite:
      uri: /tektoncd/
    route:
    - destination:
        host: tekton-pipelines-controller.$(namespace).svc.$(clusterDomain)
        port:
          number: 80
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/overlays/istio/params.yaml", `
varReference:
- path: spec/http/route/destination/host
  kind: VirtualService
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/overlays/istio/params.env", `
namespace=
clusterDomain=cluster.local
`)
	th.writeK("/manifests/tektoncd/tektoncd-install/overlays/istio", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- virtual-service.yaml
configMapGenerator:
- name: tektoncd-install-istio-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: clusterDomain
  objref:
    kind: ConfigMap
    name: tektoncd-install-istio-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.clusterDomain
- name: namespace
  objref:
    kind: ConfigMap
    name: tektoncd-install-istio-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace
configurations:
- params.yaml
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/base/namespace.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  name: tekton-pipelines
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/base/crds.yaml", `
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: clustertasks.tekton.dev
spec:
  group: tekton.dev
  names:
    categories:
    - all
    - tekton-pipelines
    kind: ClusterTask
    plural: clustertasks
  scope: Cluster
  subresources:
    status: {}
  version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: conditions.tekton.dev
spec:
  group: tekton.dev
  names:
    categories:
    - all
    - tekton-pipelines
    kind: Condition
    plural: conditions
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    knative.dev/crd-install: "true"
  name: images.caching.internal.knative.dev
spec:
  group: caching.internal.knative.dev
  names:
    categories:
    - knative-internal
    - caching
    kind: Image
    plural: images
    shortNames:
    - img
    singular: image
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: pipelines.tekton.dev
spec:
  group: tekton.dev
  names:
    categories:
    - all
    - tekton-pipelines
    kind: Pipeline
    plural: pipelines
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: pipelineruns.tekton.dev
spec:
  additionalPrinterColumns:
  - JSONPath: .status.conditions[?(@.type=="Succeeded")].status
    name: Succeeded
    type: string
  - JSONPath: .status.conditions[?(@.type=="Succeeded")].reason
    name: Reason
    type: string
  - JSONPath: .status.startTime
    name: StartTime
    type: date
  - JSONPath: .status.completionTime
    name: CompletionTime
    type: date
  group: tekton.dev
  names:
    categories:
    - all
    - tekton-pipelines
    kind: PipelineRun
    plural: pipelineruns
    shortNames:
    - pr
    - prs
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: pipelineresources.tekton.dev
spec:
  group: tekton.dev
  names:
    categories:
    - all
    - tekton-pipelines
    kind: PipelineResource
    plural: pipelineresources
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: tasks.tekton.dev
spec:
  group: tekton.dev
  names:
    categories:
    - all
    - tekton-pipelines
    kind: Task
    plural: tasks
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: taskruns.tekton.dev
spec:
  additionalPrinterColumns:
  - JSONPath: .status.conditions[?(@.type=="Succeeded")].status
    name: Succeeded
    type: string
  - JSONPath: .status.conditions[?(@.type=="Succeeded")].reason
    name: Reason
    type: string
  - JSONPath: .status.startTime
    name: StartTime
    type: date
  - JSONPath: .status.completionTime
    name: CompletionTime
    type: date
  group: tekton.dev
  names:
    categories:
    - all
    - tekton-pipelines
    kind: TaskRun
    plural: taskruns
    shortNames:
    - tr
    - trs
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha1
---
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: tekton-pipelines-controller-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: tekton-pipelines-admin
subjects:
- kind: ServiceAccount
  name: tekton-pipelines-controller
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/base/cluster-role.yaml", `
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: tekton-pipelines-admin
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - pods/log
  - namespaces
  - secrets
  - events
  - serviceaccounts
  - configmaps
  - persistentvolumeclaims
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - apps
  resources:
  - deployments
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - apps
  resources:
  - deployments/finalizers
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - mutatingwebhookconfigurations
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - tekton.dev
  resources:
  - tasks
  - clustertasks
  - taskruns
  - pipelines
  - pipelineruns
  - pipelineresources
  - conditions
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - tekton.dev
  resources:
  - taskruns/finalizers
  - pipelineruns/finalizers
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - tekton.dev
  resources:
  - tasks/status
  - clustertasks/status
  - taskruns/status
  - pipelines/status
  - pipelineruns/status
  - pipelineresources/status
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - policy
  resourceNames:
  - tekton-pipelines
  resources:
  - podsecuritypolicies
  verbs:
  - use
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    rbac.authorization.k8s.io/aggregate-to-admin: "true"
    rbac.authorization.k8s.io/aggregate-to-edit: "true"
  name: tekton-aggregate-edit
rules:
- apiGroups:
  - tekton.dev
  resources:
  - tasks
  - taskruns
  - pipelines
  - pipelineruns
  - pipelineresources
  - conditions
  verbs:
  - create
  - delete
  - deletecollection
  - get
  - list
  - patch
  - update
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    rbac.authorization.k8s.io/aggregate-to-view: "true"
  name: tekton-aggregate-view
rules:
- apiGroups:
  - tekton.dev
  resources:
  - tasks
  - taskruns
  - pipelines
  - pipelineruns
  - pipelineresources
  - conditions
  verbs:
  - get
  - list
  - watch
---
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/base/config-map.yaml", `
---
apiVersion: v1
data: null
kind: ConfigMap
metadata:
  name: config-artifact-bucket
---
apiVersion: v1
data: null
kind: ConfigMap
metadata:
  name: config-artifact-pvc
---
apiVersion: v1
data:
  _example: |
    ################################
    #                              #
    #    EXAMPLE CONFIGURATION     #
    #                              #
    ################################

    # This block is not actually functional configuration,
    # but serves to illustrate the available configuration
    # options and document them in a way that is accessible
    # to users that `+"`"+`kubectl edit`+"`"+` this config map.
    #
    # These sample configuration options may be copied out of
    # this example block and unindented to be in the data block
    # to actually change the configuration.

    # default-timeout-minutes contains the default number of
    # minutes to use for TaskRun and PipelineRun, if none is specified.
    default-timeout-minutes: "60"  # 60 minutes
kind: ConfigMap
metadata:
  name: config-defaults
---
apiVersion: v1
data:
  loglevel.controller: info
  loglevel.webhook: info
  zap-logger-config: |
    {
      "level": "info",
      "development": false,
      "sampling": {
        "initial": 100,
        "thereafter": 100
      },
      "outputPaths": ["stdout"],
      "errorOutputPaths": ["stderr"],
      "encoding": "json",
      "encoderConfig": {
        "timeKey": "",
        "levelKey": "level",
        "nameKey": "logger",
        "callerKey": "caller",
        "messageKey": "msg",
        "stacktraceKey": "stacktrace",
        "lineEnding": "",
        "levelEncoder": "",
        "timeEncoder": "",
        "durationEncoder": "",
        "callerEncoder": ""
      }
    }
kind: ConfigMap
metadata:
  name: config-logging
---
apiVersion: v1
data:
  _example: |
    ################################
    #                              #
    #    EXAMPLE CONFIGURATION     #
    #                              #
    ################################

    # This block is not actually functional configuration,
    # but serves to illustrate the available configuration
    # options and document them in a way that is accessible
    # to users that `+"`"+`kubectl edit`+"`"+` this config map.
    #
    # These sample configuration options may be copied out of
    # this example block and unindented to be in the data block
    # to actually change the configuration.

    # metrics.backend-destination field specifies the system metrics destination.
    # It supports either prometheus (the default) or stackdriver.
    # Note: Using Stackdriver will incur additional charges.
    metrics.backend-destination: prometheus

    # metrics.stackdriver-project-id field specifies the Stackdriver project ID. This
    # field is optional. When running on GCE, application default credentials will be
    # used and metrics will be sent to the cluster's project if this field is
    # not provided.
    metrics.stackdriver-project-id: "<your stackdriver project id>"

    # metrics.allow-stackdriver-custom-metrics indicates whether it is allowed
    # to send metrics to Stackdriver using "global" resource type and custom
    # metric type. Setting this flag to "true" could cause extra Stackdriver
    # charge.  If metrics.backend-destination is not Stackdriver, this is
    # ignored.
    metrics.allow-stackdriver-custom-metrics: "false"
kind: ConfigMap
metadata:
  name: config-observability
---
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/base/pod-security-policy.yaml", `
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: tekton-pipelines
spec:
  allowPrivilegeEscalation: false
  fsGroup:
    ranges:
    - max: 65535
      min: 1
    rule: MustRunAs
  hostIPC: false
  hostNetwork: false
  hostPID: false
  privileged: false
  runAsUser:
    rule: RunAsAny
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    ranges:
    - max: 65535
      min: 1
    rule: MustRunAs
  volumes:
  - emptyDir
  - configMap
  - secret
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tekton-pipelines-controller
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/base/service.yaml", `
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: tekton-pipelines-controller
  name: tekton-pipelines-controller
spec:
  ports:
  - name: metrics
    port: 9090
    protocol: TCP
    targetPort: 9090
  selector:
    app: tekton-pipelines-controller
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: tekton-pipelines-webhook
  name: tekton-pipelines-webhook
spec:
  ports:
  - port: 443
    targetPort: 8443
  selector:
    app: tekton-pipelines-webhook
---
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/base/deployment.yaml", `
---
apiVersion: apps/v1
kind: Deployment
metadata:
#  labels:
#    app.kubernetes.io/component: controller
#    app.kubernetes.io/name: tekton-pipelines
  name: tekton-pipelines-controller
spec:
  replicas: 1
  selector:
    matchLabels:
      app: tekton-pipelines-controller
  template:
    metadata:
      annotations:
        cluster-autoscaler.kubernetes.io/safe-to-evict: "false"
      labels:
        app: tekton-pipelines-controller
#        app.kubernetes.io/component: controller
#        app.kubernetes.io/name: tekton-pipelines
    spec:
      containers:
      - args:
        - -logtostderr
        - -stderrthreshold
        - INFO
        - -kubeconfig-writer-image
        - $(registry)/$(kubeconfigwriter)
        - -creds-image
        - $(registry)/$(creds-init)
        - -git-image
        - $(registry)/$(git-init)
        - -nop-image
        - $(registry)/$(nop)
        - -bash-noop-image
        - $(registry)/$(bash)
        - -gsutil-image
        - $(registry)/$(gsutil)
        - -entrypoint-image
        - $(registry)/$(entrypoint)
        - -imagedigest-exporter-image
        - $(registry)/$(imagedigestexporter)
        - -pr-image
        - $(registry)/$(pullrequest-init)
        - -build-gcs-fetcher-image
        - $(registry)/$(gcs-fetcher)
        env:
        - name: SYSTEM_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: CONFIG_LOGGING_NAME
          value: config-logging
        - name: CONFIG_OBSERVABILITY_NAME
          value: config-observability
        - name: METRICS_DOMAIN
          value: tekton.dev/pipeline
        image: $(registry)/$(controller)
        name: tekton-pipelines-controller
        volumeMounts:
        - mountPath: /etc/config-logging
          name: config-logging
      serviceAccountName: tekton-pipelines-controller
      volumes:
      - configMap:
          name: config-logging
        name: config-logging
---
apiVersion: apps/v1
kind: Deployment
metadata:
#  labels:
#    app.kubernetes.io/component: webhook-controller
#    app.kubernetes.io/name: tekton-pipelines
  name: tekton-pipelines-webhook
spec:
  replicas: 1
  selector:
    matchLabels:
      app: tekton-pipelines-webhook
  template:
    metadata:
      annotations:
        cluster-autoscaler.kubernetes.io/safe-to-evict: "false"
      labels:
        app: tekton-pipelines-webhook
#        app.kubernetes.io/component: webhook-controller
#        app.kubernetes.io/name: tekton-pipelines
    spec:
      containers:
      - env:
        - name: SYSTEM_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: $(registry)/$(webhook)
        name: webhook
        volumeMounts:
        - mountPath: /etc/config-logging
          name: config-logging
      serviceAccountName: tekton-pipelines-controller
      volumes:
      - configMap:
          name: config-logging
        name: config-logging
---
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/base/params.yaml", `
varReference:
- path: spec/template/spec/containers/image
  kind: Deployment
`)
	th.writeF("/manifests/tektoncd/tektoncd-install/base/params.env", `
registry=gcr.io/tekton-releases
webhook=github.com/tektoncd/pipeline/cmd/webhook@sha256:7215a25a58c074bbe30a50db93e6a47d2eb5672f9af7570a4e4ab75e50329131
nop=github.com/tektoncd/pipeline/cmd/nop@sha256:b372d0cb991cb960854880957c93c460d35e75339016ca6472b5ea2955f08dcb
entrypoint=github.com/tektoncd/pipeline/cmd/entrypoint@sha256:ac46866bd14ac38960c6aa100ee7468e707a955324ea4c88ce8d39b8cdfee11e
gsutil=github.com/tektoncd/pipeline/cmd/gsutil@sha256:c404edde7ec5ccf550784f2d71ea4b184ec1378329bdad316e26bce81d5f466c
gcs-fetcher=github.com/tektoncd/pipeline/vendor/github.com/googlecloudplatform/cloud-builders/gcs-fetcher/cmd/gcs-fetcher@sha256:7741f416742ac14744e8c8d0c1a628ce93d801dadfdb1ff9da8a4b9df4d6573c
bash=github.com/tektoncd/pipeline/cmd/bash@sha256:d101b69175e60cf43956ba850ec62c2db8eead17d3aa9cfb40ad7f7f3f6a3f53
creds-init=github.com/tektoncd/pipeline/cmd/creds-init@sha256:beff30d239273c4986b2e8f9d26a23cc84cc4ffda074e4e83f1cc50905c2d3da
git-init=github.com/tektoncd/pipeline/cmd/git-init@sha256:b0e6fb4f8fdd6728c6ff5bd63be30e04f88f103b9a1e972e204125aeb6a04d33
pullrequest-init=github.com/tektoncd/pipeline/cmd/pullrequest-init@sha256:c7e2a8178bc3e87405303212290836de9f781409fd60cee25cac1383aaa76f1b
imagedigestexporter=github.com/tektoncd/pipeline/cmd/imagedigestexporter@sha256:04e1eda72b3db4e4b12cc4caa2c01f33384ba80702a4dd8c41a1a940e0d69296
kubeconfigwriter=github.com/tektoncd/pipeline/cmd/kubeconfigwriter@sha256:8f8aee782bb47d7436c40e5b10a19966b21d00e1d35d2f3cd8713e206ce24841
controller=github.com/tektoncd/pipeline/cmd/controller@sha256:ebc6f768038aa3e31f3d7acda4bc26bf1380b5f2a132f0618181cacc30e295fa
`)
	th.writeK("/manifests/tektoncd/tektoncd-install/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- namespace.yaml
- crds.yaml
- cluster-role-binding.yaml
- cluster-role.yaml
- config-map.yaml
- pod-security-policy.yaml
- service-account.yaml
- service.yaml
- deployment.yaml
namespace: tekton-pipelines
configMapGenerator:
- name: tektoncd-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: registry
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.registry
- name: entrypoint
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.entrypoint
- name: nop
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.nop
- name: webhook
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.webhook
- name: gcs-fetcher
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.gcs-fetcher
- name: gsutil
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.gsutil
- name: bash
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.bash
- name: git-init
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.git-init
- name: creds-init
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.creds-init
- name: pullrequest-init
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.pullrequest-init
- name: imagedigestexporter
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.imagedigestexporter
- name: kubeconfigwriter
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.kubeconfigwriter
- name: controller
  objref:
    kind: ConfigMap
    name: tektoncd-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.controller
configurations:
- params.yaml
images:
- name: $(registry)/$(controller)
  newName: $(registry)/$(controller)
  newTag: latest
- name: $(registry)/$(webhook)
  newName: $(registry)/$(webhook)
  newTag: latest
`)
}

func TestTektoncdInstallOverlaysIstio(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/tektoncd/tektoncd-install/overlays/istio")
	writeTektoncdInstallOverlaysIstio(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../tektoncd/tektoncd-install/overlays/istio"
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
