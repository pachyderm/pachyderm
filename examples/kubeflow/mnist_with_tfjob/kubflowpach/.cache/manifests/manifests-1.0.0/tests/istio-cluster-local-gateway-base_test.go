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

func writeClusterLocalGatewayBase(th *KustTestHarness) {
	th.writeF("/manifests/istio/cluster-local-gateway/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cluster-local-gateway-istio-system
  labels:
    app: cluster-local-gateway
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-local-gateway-istio-system
subjects:
- kind: ServiceAccount
  name: cluster-local-gateway-service-account
  namespace: $(namespace)
`)
	th.writeF("/manifests/istio/cluster-local-gateway/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-local-gateway-istio-system
  labels:
    app: cluster-local-gateway
rules:
- apiGroups: ["networking.istio.io"]
  resources: ["virtualservices", "destinationrules", "gateways"]
  verbs: ["get", "watch", "list", "update"]
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: istio-reader
rules:
  - apiGroups: ['']
    resources: ['nodes', 'pods', 'services', 'endpoints', "replicationcontrollers"]
    verbs: ['get', 'watch', 'list']
  - apiGroups: ["extensions", "apps"]
    resources: ["replicasets"]
    verbs: ["get", "list", "watch"]
`)
	th.writeF("/manifests/istio/cluster-local-gateway/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cluster-local-gateway
  labels:
    app: cluster-local-gateway
    istio: cluster-local-gateway
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: cluster-local-gateway
        istio: cluster-local-gateway
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      serviceAccountName: cluster-local-gateway-service-account
      containers:
        - name: istio-proxy
          image: "docker.io/istio/proxyv2:1.1.6"
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 80
            - containerPort: 443
            - containerPort: 31400
            - containerPort: 15011
            - containerPort: 8060
            - containerPort: 15029
            - containerPort: 15030
            - containerPort: 15031
            - containerPort: 15032
            - containerPort: 15090
              protocol: TCP
              name: http-envoy-prom
          args:
          - proxy
          - router
          - --domain
          - $(POD_NAMESPACE).svc.cluster.local
          - --log_output_level=default:info
          - --drainDuration
          - '45s' #drainDuration
          - --parentShutdownDuration
          - '1m0s' #parentShutdownDuration
          - --connectTimeout
          - '10s' #connectTimeout
          - --serviceCluster
          - cluster-local-gateway
          - --zipkinAddress
          - zipkin.$(namespace):9411
          - --proxyAdminPort
          - "15000"
          - --statusPort
          - "15020"
          - --controlPlaneAuthPolicy
          - NONE
          - --discoveryAddress
          - istio-pilot.$(namespace):15010
          readinessProbe:
            failureThreshold: 30
            httpGet:
              path: /healthz/ready
              port: 15020
              scheme: HTTP
            initialDelaySeconds: 1
            periodSeconds: 2
            successThreshold: 1
            timeoutSeconds: 1
          resources:
            requests:
              cpu: 10m
          env:
          - name: POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.namespace
          - name: INSTANCE_IP
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: status.podIP
          - name: HOST_IP
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: status.hostIP
          - name: ISTIO_META_POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: ISTIO_META_CONFIG_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          volumeMounts:
          - name: istio-certs
            mountPath: /etc/certs
            readOnly: true
          - name: clusterlocalgateway-certs
            mountPath: "/etc/istio/clusterlocalgateway-certs"
            readOnly: true
          - name: clusterlocalgateway-ca-certs
            mountPath: "/etc/istio/clusterlocalgateway-ca-certs"
            readOnly: true
      volumes:
      - name: istio-certs
        secret:
          secretName: istio.cluster-local-gateway-service-account
          optional: true
      - name: clusterlocalgateway-certs
        secret:
          secretName: "istio-clusterlocalgateway-certs"
          optional: true
      - name: clusterlocalgateway-ca-certs
        secret:
          secretName: "istio-clusterlocalgateway-ca-certs"
          optional: true
      affinity:      
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x
`)
	th.writeF("/manifests/istio/cluster-local-gateway/base/horizontal-pod-autoscaler.yaml", `
apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
  labels:
    app: cluster-local-gateway
    istio: cluster-local-gateway
  name: cluster-local-gateway
spec:
  maxReplicas: 5
  metrics:
  - resource:
      name: cpu
      targetAverageUtilization: 80
    type: Resource
  minReplicas: 1
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: cluster-local-gateway
`)
	th.writeF("/manifests/istio/cluster-local-gateway/base/namespace.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  name: $(namespace)
`)
	th.writeF("/manifests/istio/cluster-local-gateway/base/pod-disruption-budget.yaml", `
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: cluster-local-gateway
  labels:
    app: cluster-local-gateway
    istio: cluster-local-gateway
spec:

  minAvailable: 1
  selector:
    matchLabels:
      app: cluster-local-gateway
      istio: cluster-local-gateway
`)
	th.writeF("/manifests/istio/cluster-local-gateway/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cluster-local-gateway-service-account
  labels:
    app: cluster-local-gateway
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-multi
`)
	th.writeF("/manifests/istio/cluster-local-gateway/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: cluster-local-gateway
  annotations:
  labels:
    app: cluster-local-gateway
    istio: cluster-local-gateway
spec:
  type: ClusterIP
  selector:
    app: cluster-local-gateway
    istio: cluster-local-gateway
  ports:
    -
      name: http2
      port: 80
      targetPort: 80
    -
      name: https
      port: 443
    -
      name: tcp
      port: 31400
    -
      name: tcp-pilot-grpc-tls
      port: 15011
      targetPort: 15011
    -
      name: tcp-citadel-grpc-tls
      port: 8060
      targetPort: 8060
    -
      name: http2-kiali
      port: 15029
      targetPort: 15029
    -
      name: http2-prometheus
      port: 15030
      targetPort: 15030
    -
      name: http2-grafana
      port: 15031
      targetPort: 15031
    -
      name: http2-tracing
      port: 15032
      targetPort: 15032
`)
	th.writeF("/manifests/istio/cluster-local-gateway/base/params.yaml", `
varReference:
- path: metadata/name
  kind: Namespace
- path: subjects/namespace
  kind: ClusterRoleBinding
`)
	th.writeF("/manifests/istio/cluster-local-gateway/base/params.env", `
namespace=istio-system
`)
	th.writeK("/manifests/istio/cluster-local-gateway/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

configMapGenerator:
- name: cluster-local-gateway-parameters
  envs:
  - params.env

resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- deployment.yaml
- horizontal-pod-autoscaler.yaml
- namespace.yaml
- pod-disruption-budget.yaml
- service-account.yaml
- service.yaml

vars:
- name: namespace
  objref:
    kind: ConfigMap
    name: cluster-local-gateway-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace

commonLabels:
  kustomize.component: cluster-local-gateway

configurations:
- params.yaml
`)
}

func TestClusterLocalGatewayBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/istio/cluster-local-gateway/base")
	writeClusterLocalGatewayBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../istio/cluster-local-gateway/base"
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
