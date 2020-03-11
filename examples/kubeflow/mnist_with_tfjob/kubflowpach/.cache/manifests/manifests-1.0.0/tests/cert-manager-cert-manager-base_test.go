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

func writeCertManagerBase(th *KustTestHarness) {
	th.writeF("/manifests/cert-manager/cert-manager/base/namespace.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  name: $(namespace)
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/api-service.yaml", `
apiVersion: apiregistration.k8s.io/v1beta1
kind: APIService
metadata:
  name: v1beta1.webhook.cert-manager.io
  labels:
    app: webhook
  annotations:
    cert-manager.io/inject-ca-from-secret: "cert-manager/cert-manager-webhook-tls"
spec:
  group: webhook.cert-manager.io
  groupPriorityMinimum: 1000
  versionPriority: 15
  service:
    name: cert-manager-webhook
    namespace: $(namespace)
  version: v1beta1
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-controller-issuers
  labels:
    app: cert-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager-controller-issuers
subjects:
- name: cert-manager
  namespace: $(namespace)
  kind: ServiceAccount

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-controller-clusterissuers
  labels:
    app: cert-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager-controller-clusterissuers
subjects:
- name: cert-manager
  namespace: $(namespace)
  kind: ServiceAccount

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-controller-certificates
  labels:
    app: cert-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager-controller-certificates
subjects:
- name: cert-manager
  namespace: $(namespace)
  kind: ServiceAccount

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-controller-orders
  labels:
    app: cert-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager-controller-orders
subjects:
- name: cert-manager
  namespace: $(namespace)
  kind: ServiceAccount

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-controller-challenges
  labels:
    app: cert-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager-controller-challenges
subjects:
- name: cert-manager
  namespace: $(namespace)
  kind: ServiceAccount

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-controller-ingress-shim
  labels:
    app: cert-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager-controller-ingress-shim
subjects:
- name: cert-manager
  namespace: $(namespace)
  kind: ServiceAccount

---
# apiserver gets the auth-delegator role to delegate auth decisions to
# the core apiserver
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-webhook:auth-delegator
  labels:
    app: webhook
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
subjects:
- apiGroup: ""
  kind: ServiceAccount
  name: cert-manager-webhook
  namespace: $(namespace)

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-cainjector
  labels:
    app: cainjector
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager-cainjector
subjects:
- name: cert-manager-cainjector
  namespace: $(namespace)
  kind: ServiceAccount
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cert-manager-cainjector
  labels:
    app: cainjector
rules:
  - apiGroups: ["cert-manager.io"]
    resources: ["certificates"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["get", "create", "update", "patch"]
  - apiGroups: ["admissionregistration.k8s.io"]
    resources: ["validatingwebhookconfigurations", "mutatingwebhookconfigurations"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["apiregistration.k8s.io"]
    resources: ["apiservices"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["apiextensions.k8s.io"]
    resources: ["customresourcedefinitions"]
    verbs: ["get", "list", "watch", "update"]

---

# Issuer controller role
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cert-manager-controller-issuers
  labels:
    app: cert-manager
rules:
  - apiGroups: ["cert-manager.io"]
    resources: ["issuers", "issuers/status"]
    verbs: ["update"]
  - apiGroups: ["cert-manager.io"]
    resources: ["issuers"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "patch"]

---

# ClusterIssuer controller role
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cert-manager-controller-clusterissuers
  labels:
    app: cert-manager
rules:
  - apiGroups: ["cert-manager.io"]
    resources: ["clusterissuers", "clusterissuers/status"]
    verbs: ["update"]
  - apiGroups: ["cert-manager.io"]
    resources: ["clusterissuers"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "patch"]

---

# Certificates controller role
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cert-manager-controller-certificates
  labels:
    app: cert-manager
rules:
  - apiGroups: ["cert-manager.io"]
    resources: ["certificates", "certificates/status", "certificaterequests", "certificaterequests/status"]
    verbs: ["update"]
  - apiGroups: ["cert-manager.io"]
    resources: ["certificates", "certificaterequests", "clusterissuers", "issuers"]
    verbs: ["get", "list", "watch"]
  # We require these rules to support users with the OwnerReferencesPermissionEnforcement
  # admission controller enabled:
  # https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#ownerreferencespermissionenforcement
  - apiGroups: ["cert-manager.io"]
    resources: ["certificates/finalizers"]
    verbs: ["update"]
  - apiGroups: ["acme.cert-manager.io"]
    resources: ["orders"]
    verbs: ["create", "delete", "get", "list", "watch"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "patch"]

---

# Orders controller role
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cert-manager-controller-orders
  labels:
    app: cert-manager
rules:
  - apiGroups: ["acme.cert-manager.io"]
    resources: ["orders", "orders/status"]
    verbs: ["update"]
  - apiGroups: ["acme.cert-manager.io"]
    resources: ["orders", "challenges"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["cert-manager.io"]
    resources: ["clusterissuers", "issuers"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["acme.cert-manager.io"]
    resources: ["challenges"]
    verbs: ["create", "delete"]
  # We require these rules to support users with the OwnerReferencesPermissionEnforcement
  # admission controller enabled:
  # https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#ownerreferencespermissionenforcement
  - apiGroups: ["acme.cert-manager.io"]
    resources: ["orders/finalizers"]
    verbs: ["update"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "patch"]

---

# Challenges controller role
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cert-manager-controller-challenges
  labels:
    app: cert-manager
rules:
  # Use to update challenge resource status
  - apiGroups: ["acme.cert-manager.io"]
    resources: ["challenges", "challenges/status"]
    verbs: ["update"]
  # Used to watch challenge resources
  - apiGroups: ["acme.cert-manager.io"]
    resources: ["challenges"]
    verbs: ["get", "list", "watch"]
  # Used to watch challenges, issuer and clusterissuer resources
  - apiGroups: ["cert-manager.io"]
    resources: ["issuers", "clusterissuers"]
    verbs: ["get", "list", "watch"]
  # Need to be able to retrieve ACME account private key to complete challenges
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch"]
  # Used to create events
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "patch"]
  # HTTP01 rules
  - apiGroups: [""]
    resources: ["pods", "services"]
    verbs: ["get", "list", "watch", "create", "delete"]
  - apiGroups: ["extensions", "networking.k8s.io/v1"]
    resources: ["ingresses"]
    verbs: ["get", "list", "watch", "create", "delete", "update"]
  # We require these rules to support users with the OwnerReferencesPermissionEnforcement
  # admission controller enabled:
  # https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#ownerreferencespermissionenforcement
  - apiGroups: ["acme.cert-manager.io"]
    resources: ["challenges/finalizers"]
    verbs: ["update"]
  # DNS01 rules (duplicated above)
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch"]

---

# ingress-shim controller role
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cert-manager-controller-ingress-shim
  labels:
    app: cert-manager
rules:
  - apiGroups: ["cert-manager.io"]
    resources: ["certificates", "certificaterequests"]
    verbs: ["create", "update", "delete"]
  - apiGroups: ["cert-manager.io"]
    resources: ["certificates", "certificaterequests", "issuers", "clusterissuers"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["networking.k8s.io/v1"]
    resources: ["ingresses"]
    verbs: ["get", "list", "watch"]
  # We require these rules to support users with the OwnerReferencesPermissionEnforcement
  # admission controller enabled:
  # https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#ownerreferencespermissionenforcement
  - apiGroups: ["networking.k8s.io/v1"]
    resources: ["ingresses/finalizers"]
    verbs: ["update"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "patch"]

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cert-manager-webhook:webhook-requester
  labels:
    app: webhook
rules:
- apiGroups:
  - admission.cert-manager.io
  resources:
  - certificates
  - certificaterequests
  - issuers
  - clusterissuers
  verbs:
  - create

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cert-manager-view
  labels:
    app: cert-manager
    rbac.authorization.k8s.io/aggregate-to-view: "true"
    rbac.authorization.k8s.io/aggregate-to-edit: "true"
    rbac.authorization.k8s.io/aggregate-to-admin: "true"
rules:
- apiGroups: ["cert-manager.io"]
  resources: ["certificates", "certificaterequests", "issuers"]
  verbs: ["get", "list", "watch"]

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cert-manager-edit
  labels:
    app: cert-manager
    rbac.authorization.k8s.io/aggregate-to-edit: "true"
    rbac.authorization.k8s.io/aggregate-to-admin: "true"
rules:
- apiGroups: ["cert-manager.io"]
  resources: ["certificates", "certificaterequests", "issuers"]
  verbs: ["create", "delete", "deletecollection", "patch", "update"]
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cert-manager-cainjector
  labels:
    app: cainjector
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cainjector
  template:
    metadata:
      labels:
        app: cainjector
      annotations:
    spec:
      serviceAccountName: cert-manager-cainjector
      containers:
        - name: cainjector
          image: "quay.io/jetstack/cert-manager-cainjector:v0.11.0"
          imagePullPolicy: IfNotPresent
          args:
          - --v=2
          - --leader-election-namespace=kube-system
          env:
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          resources:
            {}

---

apiVersion: apps/v1
kind: Deployment
metadata:
  name: cert-manager
  labels:
    app: cert-manager
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cert-manager
  template:
    metadata:
      labels:
        app: cert-manager
      annotations:
        prometheus.io/path: "/metrics"
        prometheus.io/scrape: 'true'
        prometheus.io/port: '9402'
    spec:
      serviceAccountName: cert-manager
      containers:
        - name: cert-manager
          image: "quay.io/jetstack/cert-manager-controller:v0.11.0"
          imagePullPolicy: IfNotPresent
          args:
          - --v=2
          - --cluster-resource-namespace=$(POD_NAMESPACE)
          - --leader-election-namespace=kube-system
          - --webhook-namespace=$(POD_NAMESPACE)
          - --webhook-ca-secret=cert-manager-webhook-ca
          - --webhook-serving-secret=cert-manager-webhook-tls
          - --webhook-dns-names=cert-manager-webhook,cert-manager-webhook.$(namespace),cert-manager-webhook.$(namespace).svc
          ports:
          - containerPort: 9402
          env:
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          resources:
            requests:
              cpu: 10m
              memory: 32Mi

---

apiVersion: apps/v1
kind: Deployment
metadata:
  name: cert-manager-webhook
  labels:
    app: webhook
spec:
  replicas: 1
  selector:
    matchLabels:
      app: webhook
  template:
    metadata:
      labels:
        app: webhook
      annotations:
    spec:
      serviceAccountName: cert-manager-webhook
      containers:
        - name: cert-manager
          image: "quay.io/jetstack/cert-manager-webhook:v0.11.0"
          imagePullPolicy: IfNotPresent
          args:
          - --v=2
          - --secure-port=6443
          - --tls-cert-file=/certs/tls.crt
          - --tls-private-key-file=/certs/tls.key
          env:
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          resources:
            {}

          volumeMounts:
          - name: certs
            mountPath: /certs
      volumes:
      - name: certs
        secret:
          secretName: cert-manager-webhook-tls
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/mutating-webhook-configuration.yaml", `
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: cert-manager-webhook
  labels:
    app: webhook
  annotations:
    cert-manager.io/inject-apiserver-ca: "true"
webhooks:
  - name: webhook.cert-manager.io
    rules:
      - apiGroups:
          - "cert-manager.io"
        apiVersions:
          - v1alpha2
        operations:
          - CREATE
          - UPDATE
        resources:
          - certificates
          - issuers
          - clusterissuers
          - orders
          - challenges
          - certificaterequests
    failurePolicy: Fail
    clientConfig:
      service:
        name: kubernetes
        namespace: default
        path: /apis/webhook.cert-manager.io/v1beta1/mutations
      caBundle: ""
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cert-manager-cainjector
  labels:
    app: cainjector

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: cert-manager
  annotations:
  labels:
    app: cert-manager

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: cert-manager-webhook
  labels:
    app: webhook
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: cert-manager
  labels:
    app: cert-manager
spec:
  type: ClusterIP
  ports:
    - protocol: TCP
      port: 9402
      targetPort: 9402
  selector:
    app: cert-manager

---
apiVersion: v1
kind: Service
metadata:
  name: cert-manager-webhook
  labels:
    app: webhook
spec:
  type: ClusterIP
  ports:
  - name: https
    port: 443
    targetPort: 6443
  selector:
    app: webhook
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/validating-webhook-configuration.yaml", `
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: cert-manager-webhook
  labels:
    app: webhook
  annotations:
    cert-manager.io/inject-apiserver-ca: "true"
webhooks:
  - name: webhook.certmanager.k8s.io
    rules:
      - apiGroups:
          - "cert-manager.io"
        apiVersions:
          - v1alpha2
        operations:
          - CREATE
          - UPDATE
        resources:
          - certificates
          - issuers
          - clusterissuers
          - certificaterequests
    failurePolicy: Fail
    sideEffects: None
    clientConfig:
      service:
        name: kubernetes
        namespace: default
        path: /apis/webhook.cert-manager.io/v1beta1/validations
      caBundle: ""
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/params.yaml", `
varReference:
- path: subjects/namespace
  kind: ClusterRoleBinding
- path: spec/template/spec/containers/args
  kind: Deployment
- path: metadata/name
  kind: Namespace
- path: spec/service/namespace
  kind: APIService
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/params.env", `
namespace=cert-manager
`)
	th.writeK("/manifests/cert-manager/cert-manager/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: cert-manager
resources:
- namespace.yaml
- api-service.yaml
- cluster-role-binding.yaml
- cluster-role.yaml
- deployment.yaml
- mutating-webhook-configuration.yaml
- service-account.yaml
- service.yaml
- validating-webhook-configuration.yaml
commonLabels:
  kustomize.component: cert-manager
images:
- name: quay.io/jetstack/cert-manager-controller
  newName: quay.io/jetstack/cert-manager-controller
  newTag: v0.11.0
- name: quay.io/jetstack/cert-manager-webhook
  newName: quay.io/jetstack/cert-manager-webhook
  newTag: v0.11.0
- name: quay.io/jetstack/cert-manager-cainjector
  newName: quay.io/jetstack/cert-manager-cainjector
  newTag: v0.11.0
configMapGenerator:
- name: cert-manager-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: namespace
  objref:
    kind: ConfigMap
    name: cert-manager-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace
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
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../cert-manager/cert-manager/base"
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
