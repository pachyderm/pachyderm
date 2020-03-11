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

func writeKfservingInstallBase(th *KustTestHarness) {
	th.writeF("/manifests/kfserving/kfserving-install/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kfserving-proxy-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kfserving-proxy-role
subjects:
- kind: ServiceAccount
  name: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  creationTimestamp: null
  name: manager-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: manager-role
subjects:
- kind: ServiceAccount
  name: default
---
`)
	th.writeF("/manifests/kfserving/kfserving-install/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kfserving-proxy-role
rules:
- apiGroups:
  - authentication.k8s.io
  resources:
  - tokenreviews
  verbs:
  - create
- apiGroups:
  - authorization.k8s.io
  resources:
  - subjectaccessreviews
  verbs:
  - create
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  creationTimestamp: null
  name: manager-role
rules:
- apiGroups:
  - serving.knative.dev
  resources:
  - services
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - serving.knative.dev
  resources:
  - services/status
  verbs:
  - get
  - update
  - patch
- apiGroups:
  - networking.istio.io
  resources:
  - virtualservices
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - networking.istio.io
  resources:
  - virtualservices/status
  verbs:
  - get
  - update
  - patch
- apiGroups:
  - serving.kubeflow.org
  resources:
  - inferenceservices
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - serving.kubeflow.org
  resources:
  - inferenceservices/status
  verbs:
  - get
  - update
  - patch
- apiGroups:
  - ""
  resources:
  - serviceaccounts
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
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - mutatingwebhookconfigurations
  - validatingwebhookconfigurations
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - get
  - list
  - watch

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-kfserving-admin
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-admin: "true"
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.authorization.kubeflow.org/aggregate-to-kubeflow-kfserving-admin: "true"
rules: []

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-kfserving-edit
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-edit: "true"
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-kfserving-admin: "true"
rules:
- apiGroups:
  - serving.kubeflow.org
  resources:
  - inferenceservices
  verbs:
  - get
  - list
  - watch
  - create
  - delete
  - deletecollection
  - patch
  - update

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-kfserving-view
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-view: "true"
rules:
- apiGroups:
  - serving.kubeflow.org
  resources:
  - inferenceservices
  verbs:
  - get
  - list
  - watch
`)
	th.writeF("/manifests/kfserving/kfserving-install/base/config-map.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: inferenceservice-config
data:
  predictors: |-
    {
        "tensorflow": {
            "image": "tensorflow/serving",
            "defaultImageVersion": "1.14.0",
            "defaultGpuImageVersion": "1.14.0-gpu",
            "allowedImageVersions": [
               "1.11.0",
               "1.11.0-gpu",
               "1.12.0",
               "1.12.0-gpu",
               "1.13.0",
               "1.13.0-gpu",
               "1.14.0",
               "1.14.0-gpu"
            ]
        },
        "onnx": {
            "image": "mcr.microsoft.com/onnxruntime/server",
            "defaultImageVersion": "v0.5.1",
            "allowedImageVersions": [
               "v0.5.1"
            ]
        },
        "sklearn": {
            "image": "gcr.io/kfserving/sklearnserver",
            "defaultImageVersion": "0.2.2",
            "allowedImageVersions": [
               "0.2.2"
            ]
        },
        "xgboost": {
            "image": "gcr.io/kfserving/xgbserver",
            "defaultImageVersion": "0.2.2",
            "allowedImageVersions": [
               "0.2.2"
            ]
        },
        "pytorch": {
            "image": "gcr.io/kfserving/pytorchserver",
            "defaultImageVersion": "0.2.2",
            "allowedImageVersions": [
               "0.2.2"
            ]
        },
        "tensorrt": {
            "image": "nvcr.io/nvidia/tensorrtserver",
            "defaultImageVersion": "19.05-py3",
            "allowedImageVersions": [
               "19.05-py3"
            ]
        }
    }
  transformers: |-
    {
    }
  explainers: |-
    {
        "alibi": {
            "image" : "gcr.io/kfserving/alibi-explainer",
            "defaultImageVersion": "0.2.2",
            "allowedImageVersions": [
               "0.2.2"
            ]
        }
    }
  storageInitializer: |-
    {
        "image" : "gcr.io/kfserving/storage-initializer:0.2.2",
        "memoryRequest": "100Mi",
        "memoryLimit": "1Gi",
        "cpuRequest": "100m",
        "cpuLimit": "1"
    }
  credentials: |-
    {
       "gcs": {
           "gcsCredentialFileName": "gcloud-application-credentials.json"
       },
       "s3": {
           "s3AccessKeyIDName": "awsAccessKeyID",
           "s3SecretAccessKeyName": "awsSecretAccessKey"
       }
    }
  ingress: |-
    {
        "ingressGateway" : "kubeflow-gateway.kubeflow",
        "ingressService" : "istio-ingressgateway.istio-system.svc.cluster.local"
    }
  logger: |-
    {
        "image" : "gcr.io/kfserving/logger:0.2.2",
        "memoryRequest": "100Mi",
        "memoryLimit": "1Gi",
        "cpuRequest": "100m",
        "cpuLimit": "1"
    }`)
	th.writeF("/manifests/kfserving/kfserving-install/base/secret.yaml", `
apiVersion: v1
kind: Secret
metadata:
  name: kfserving-webhook-server-secret
`)
	th.writeF("/manifests/kfserving/kfserving-install/base/statefulset.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    control-plane: kfserving-controller-manager
    controller-tools.k8s.io: "1.0"
  name: kfserving-controller-manager
spec:
  selector:
    matchLabels:
      control-plane: kfserving-controller-manager
      controller-tools.k8s.io: "1.0"
  serviceName: controller-manager-service
  template:
    metadata:
      labels:
        control-plane: kfserving-controller-manager
        controller-tools.k8s.io: "1.0"
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      containers:
      - args:
        - --secure-listen-address=0.0.0.0:8443
        - --upstream=http://127.0.0.1:8080/
        - --logtostderr=true
        - --v=10
        image: gcr.io/kubebuilder/kube-rbac-proxy:v0.4.0
        name: kube-rbac-proxy
        ports:
        - containerPort: 8443
          name: https
      - args:
        - --metrics-addr=127.0.0.1:8080
        command:
        - /manager
        env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: SECRET_NAME
          value: kfserving-webhook-server-secret
        - name: ENABLE_WEBHOOK_NAMESPACE_SELECTOR
          value: enabled
        image: $(registry)/kfserving-controller:0.2.2
        imagePullPolicy: Always
        name: manager
        ports:
        - containerPort: 9876
          name: webhook-server
          protocol: TCP
        resources:
          limits:
            cpu: 100m
            memory: 300Mi
          requests:
            cpu: 100m
            memory: 200Mi
        volumeMounts:
        - mountPath: /tmp/cert
          name: cert
          readOnly: true
      terminationGracePeriodSeconds: 10
      volumes:
      - name: cert
        secret:
          defaultMode: 420
          secretName: kfserving-webhook-server-secret
  volumeClaimTemplates: []
`)
	th.writeF("/manifests/kfserving/kfserving-install/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  annotations:
    prometheus.io/port: "8443"
    prometheus.io/scheme: https
    prometheus.io/scrape: "true"
  labels:
    control-plane: controller-manager
    controller-tools.k8s.io: "1.0"
  name: kfserving-controller-manager-metrics-service
spec:
  ports:
  - name: https
    port: 8443
    targetPort: https
  selector:
    control-plane: controller-manager
    controller-tools.k8s.io: "1.0"
---
apiVersion: v1
kind: Service
metadata:
  labels:
    control-plane: kfserving-controller-manager
    controller-tools.k8s.io: "1.0"
  name: kfserving-controller-manager-service
spec:
  ports:
  - port: 443
  selector:
    control-plane: kfserving-controller-manager
    controller-tools.k8s.io: "1.0"
---
`)
	th.writeF("/manifests/kfserving/kfserving-install/base/params.yaml", `
varReference:
- path: spec/template/spec/containers/image
  kind: StatefulSet
- path: data/frameworks
  kind: ConfigMap
`)
	th.writeF("/manifests/kfserving/kfserving-install/base/params.env", `
registry=gcr.io/kfserving
`)
	th.writeK("/manifests/kfserving/kfserving-install/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kubeflow
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- config-map.yaml
- secret.yaml
- statefulset.yaml
- service.yaml
commonLabels:
  kustomize.component: kfserving
configMapGenerator:
- name: kfserving-parameters
  env: params.env
vars:
- name: registry
  objref:
    kind: ConfigMap
    name: kfserving-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.registry
configurations:
- params.yaml
images:
- name: gcr.io/kubebuilder/kube-rbac-proxy
  newName: gcr.io/kubebuilder/kube-rbac-proxy
  newTag: v0.4.0
- name: $(registry)/kfserving-controller
  newName: $(registry)/kfserving-controller
  newTag: 0.2.2
`)
}

func TestKfservingInstallBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/kfserving/kfserving-install/base")
	writeKfservingInstallBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../kfserving/kfserving-install/base"
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
