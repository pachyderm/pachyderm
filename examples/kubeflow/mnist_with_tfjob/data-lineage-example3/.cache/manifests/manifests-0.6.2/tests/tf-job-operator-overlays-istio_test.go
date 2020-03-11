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

func writeTfJobOperatorOverlaysIstio(th *KustTestHarness) {
	th.writeF("/manifests/tf-training/tf-job-operator/overlays/istio/virtual-service.yaml", `
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: tf-job-dashboard
spec:
  gateways:
  - kubeflow-gateway
  hosts:
  - '*'
  http:
  - match:
    - uri:
        prefix: /tfjobs/
    rewrite:
      uri: /tfjobs/
    route:
    - destination:
        host: tf-job-dashboard.$(namespace).svc.$(clusterDomain)
        port:
          number: 80
`)
	th.writeF("/manifests/tf-training/tf-job-operator/overlays/istio/params.yaml", `
varReference:
- path: spec/http/route/destination/host
  kind: VirtualService
`)
	th.writeK("/manifests/tf-training/tf-job-operator/overlays/istio", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- virtual-service.yaml
configurations:
- params.yaml
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/cluster-role-binding.yaml", `
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    app: tf-job-dashboard
  name: tf-job-dashboard
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: tf-job-dashboard
subjects:
- kind: ServiceAccount
  name: tf-job-dashboard
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    app: tf-job-operator
  name: tf-job-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: tf-job-operator
subjects:
- kind: ServiceAccount
  name: tf-job-operator
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/cluster-role.yaml", `
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    app: tf-job-dashboard
  name: tf-job-dashboard
rules:
- apiGroups:
  - tensorflow.org
  - kubeflow.org
  resources:
  - tfjobs
  - tfjobs/status
  verbs:
  - '*'
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - '*'
- apiGroups:
  - storage.k8s.io
  resources:
  - storageclasses
  verbs:
  - '*'
- apiGroups:
  - batch
  resources:
  - jobs
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - configmaps
  - pods
  - services
  - endpoints
  - persistentvolumeclaims
  - events
  - pods/log
  - namespaces
  verbs:
  - '*'
- apiGroups:
  - apps
  - extensions
  resources:
  - deployments
  verbs:
  - '*'
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    app: tf-job-operator
  name: tf-job-operator
rules:
- apiGroups:
  - tensorflow.org
  - kubeflow.org
  resources:
  - tfjobs
  - tfjobs/status
  verbs:
  - '*'
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - '*'
- apiGroups:
  - storage.k8s.io
  resources:
  - storageclasses
  verbs:
  - '*'
- apiGroups:
  - batch
  resources:
  - jobs
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - configmaps
  - pods
  - services
  - endpoints
  - persistentvolumeclaims
  - events
  verbs:
  - '*'
- apiGroups:
  - apps
  - extensions
  resources:
  - deployments
  verbs:
  - '*'
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/config-map.yaml", `
apiVersion: v1
data:
  controller_config_file.yaml: |-
    {
        "grpcServerFilePath": "/opt/mlkube/grpc_tensorflow_server/grpc_tensorflow_server.py"
    }
kind: ConfigMap
metadata:
  name: tf-job-operator-config
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: tfjobs.kubeflow.org
spec:
  additionalPrinterColumns:
  - JSONPath: .status.conditions[-1:].type
    name: State
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
  group: kubeflow.org
  names:
    kind: TFJob
    plural: tfjobs
    singular: tfjob
  scope: Namespaced
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      properties:
        spec:
          properties:
            tfReplicaSpecs:
              properties:
                Chief:
                  properties:
                    replicas:
                      maximum: 1
                      minimum: 1
                      type: integer
                PS:
                  properties:
                    replicas:
                      minimum: 1
                      type: integer
                Worker:
                  properties:
                    replicas:
                      minimum: 1
                      type: integer
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
  - name: v1beta2
    served: true
    storage: false
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: tf-job-dashboard
spec:
  template:
    metadata:
      labels:
        name: tf-job-dashboard
    spec:
      containers:
      - command:
        - /opt/tensorflow_k8s/dashboard/backend
        env:
        - name: KUBEFLOW_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: gcr.io/kubeflow-images-public/tf_operator:v0.6.0.rc0
        name: tf-job-dashboard
        ports:
        - containerPort: 8080
      serviceAccountName: tf-job-dashboard
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: tf-job-operator
spec:
  replicas: 1
  template:
    metadata:
      labels:
        name: tf-job-operator
    spec:
      containers:
      - command:
        - /opt/kubeflow/tf-operator.v1
        - --alsologtostderr
        - -v=1
        - --monitoring-port=8443
        env:
        - name: MY_POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: MY_POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        image: gcr.io/kubeflow-images-public/tf_operator:v0.6.0.rc0
        name: tf-job-operator
        volumeMounts:
        - mountPath: /etc/config
          name: config-volume
      serviceAccountName: tf-job-operator
      volumes:
      - configMap:
          name: tf-job-operator-config
        name: config-volume
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/service-account.yaml", `
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: tf-job-dashboard
  name: tf-job-dashboard
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: tf-job-operator
  name: tf-job-operator
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  annotations:
    getambassador.io/config: |-
      ---
      apiVersion: ambassador/v0
      kind:  Mapping
      name: tfjobs-ui-mapping
      prefix: /tfjobs/
      rewrite: /tfjobs/
      service: tf-job-dashboard.$(namespace)
  name: tf-job-dashboard
spec:
  ports:
  - port: 80
    targetPort: 8080
  selector:
    name: tf-job-dashboard
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    prometheus.io/path: /metrics
    prometheus.io/scrape: "true"
    prometheus.io/port: "8443"
  labels:
    app: tf-job-operator
  name: tf-job-operator
spec:
  ports:
  - name: monitoring-port
    port: 8443
    targetPort: 8443
  selector:
    name: tf-job-operator
  type: ClusterIP
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/params.yaml", `
varReference:
- path: metadata/annotations/getambassador.io\/config
  kind: Service
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/params.env", `
namespace=
clusterDomain=cluster.local
`)
	th.writeK("/manifests/tf-training/tf-job-operator/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kubeflow
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- config-map.yaml
- crd.yaml
- deployment.yaml
- service-account.yaml
- service.yaml
commonLabels:
  kustomize.component: tf-job-operator
configMapGenerator:
- name: parameters
  env: params.env
vars:
- name: namespace
  objref:
    kind: Service
    name: tf-job-dashboard
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
`)
}

func TestTfJobOperatorOverlaysIstio(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/tf-training/tf-job-operator/overlays/istio")
	writeTfJobOperatorOverlaysIstio(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../tf-training/tf-job-operator/overlays/istio"
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
