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

func writeKatibControllerOverlaysIstio(th *KustTestHarness) {
	th.writeF("/manifests/katib/katib-controller/overlays/istio/katib-ui-virtual-service.yaml", `
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: katib-ui
spec:
  gateways:
  - kubeflow-gateway
  hosts:
  - '*'
  http:
  - match:
    - uri:
        prefix: /katib/
    rewrite:
      uri: /katib/
    route:
    - destination:
        host: katib-ui.$(namespace).svc.$(clusterDomain)
        port:
          number: 80
`)
	th.writeF("/manifests/katib/katib-controller/overlays/istio/params.yaml", `
varReference:
- path: spec/http/route/destination/host
  kind: VirtualService
`)
	th.writeK("/manifests/katib/katib-controller/overlays/istio", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- katib-ui-virtual-service.yaml
configurations:
- params.yaml
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-configmap.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: katib-config
data:
  metrics-collector-sidecar: |-
    {
      "StdOut": {
        "image": "gcr.io/kubeflow-images-public/katib/v1alpha3/file-metrics-collector:v0.8.0"
      },
      "File": {
        "image": "gcr.io/kubeflow-images-public/katib/v1alpha3/file-metrics-collector:v0.8.0"
      },
      "TensorFlowEvent": {
        "image": "gcr.io/kubeflow-images-public/katib/v1alpha3/tfevent-metrics-collector:v0.8.0",
        "resources": {
          "limits": {
            "memory": "1Gi"
          }
        }
      }
    }
  suggestion: |-
    {
      "random": {
        "image": "gcr.io/kubeflow-images-public/katib/v1alpha3/suggestion-hyperopt:v0.8.0"
      },
      "grid": {
        "image": "gcr.io/kubeflow-images-public/katib/v1alpha3/suggestion-chocolate:v0.8.0"
      },
      "hyperband": {
        "image": "gcr.io/kubeflow-images-public/katib/v1alpha3/suggestion-hyperband:v0.8.0"
      },
      "bayesianoptimization": {
        "image": "gcr.io/kubeflow-images-public/katib/v1alpha3/suggestion-skopt:v0.8.0"
      },
      "tpe": {
        "image": "gcr.io/kubeflow-images-public/katib/v1alpha3/suggestion-hyperopt:v0.8.0"
      },
      "nasrl": {
        "image": "gcr.io/kubeflow-images-public/katib/v1alpha3/suggestion-nasrl:v0.8.0",
        "imagePullPolicy": "Always",
        "resources": {
          "limits": {
            "memory": "200Mi"
          }
        }
      }
    }
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-controller-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: katib-controller
  labels:
    app: katib-controller
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib-controller
  template:
    metadata:
      labels:
        app: katib-controller
      annotations:
        sidecar.istio.io/inject: "false"
        prometheus.io/scrape: "true"
    spec:
      serviceAccountName: katib-controller
      containers:
      - name: katib-controller
        image: gcr.io/kubeflow-images-public/katib/v1alpha3/katib-controller
        imagePullPolicy: IfNotPresent
        command: ["./katib-controller"]
        args:
          - '--webhook-port=8443'
        ports:
        - containerPort: 8443
          name: webhook
          protocol: TCP
        - containerPort: 8080
          name: metrics
          protocol: TCP
        env:
        - name: KATIB_CORE_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        volumeMounts:
        - mountPath: /tmp/cert
          name: cert
          readOnly: true
      volumes:
      - name: cert
        secret:
          defaultMode: 420
          secretName: katib-controller
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-controller-rbac.yaml", `
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: katib-controller
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  - serviceaccounts
  - services
  - secrets
  - events
  - namespaces
  verbs:
  - "*"
- apiGroups:
  - ""
  resources:
  - pods
  - pods/log
  - pods/status
  verbs:
  - "*"
- apiGroups:
  - apps
  resources:
  - deployments
  verbs:
  - "*"
- apiGroups:
  - batch
  resources:
  - jobs
  - cronjobs
  verbs:
  - "*"
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - create
  - get
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - validatingwebhookconfigurations
  - mutatingwebhookconfigurations
  verbs:
  - '*'
- apiGroups:
  - kubeflow.org
  resources:
  - experiments
  - experiments/status
  - trials
  - trials/status
  - suggestions
  - suggestions/status
  verbs:
  - "*"
- apiGroups:
  - kubeflow.org
  resources:
  - tfjobs
  - pytorchjobs
  verbs:
  - "*"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: katib-controller
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: katib-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: katib-controller
subjects:
- kind: ServiceAccount
  name: katib-controller

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-katib-admin
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-admin: "true"
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.authorization.kubeflow.org/aggregate-to-kubeflow-katib-admin: "true"
rules: []

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-katib-edit
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-edit: "true"
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-katib-admin: "true"
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - experiments
  - trials
  - suggestions
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
  name: kubeflow-katib-view
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-view: "true"
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - experiments
  - trials
  - suggestions
  verbs:
  - get
  - list
  - watch
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-controller-secret.yaml", `
apiVersion: v1
kind: Secret
metadata:
  name: katib-controller
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-controller-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-controller
  annotations:
    prometheus.io/port: "8080"
    prometheus.io/scheme: http
    prometheus.io/scrape: "true"
spec:
  ports:
  - port: 443
    protocol: TCP
    targetPort: 8443
    name: webhook
  - name: metrics
    port: 8080
    targetPort: 8080
  selector:
    app: katib-controller
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-mysql-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: katib-mysql
  labels:
    app: katib
    component: mysql
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib
      component: mysql
  template:
    metadata:
      name: katib-mysql
      labels:
        app: katib
        component: mysql
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      containers:
      - name: katib-mysql
        image: mysql:8
        args:
        - --datadir
        - /var/lib/mysql/datadir
        env:
          - name: MYSQL_ROOT_PASSWORD
            valueFrom:
              secretKeyRef:
                name: katib-mysql-secrets
                key: MYSQL_ROOT_PASSWORD
          - name: MYSQL_ALLOW_EMPTY_PASSWORD
            value: "true"
          - name: MYSQL_DATABASE
            value: "katib"
        ports:
        - name: dbapi
          containerPort: 3306
        readinessProbe:
          exec:
            command:
            - "/bin/bash"
            - "-c"
            - "mysql -D ${MYSQL_DATABASE} -u root -p${MYSQL_ROOT_PASSWORD} -e 'SELECT 1'"
          initialDelaySeconds: 5
          periodSeconds: 10
          timeoutSeconds: 1
        livenessProbe:
          exec:
            command:
            - "/bin/bash"
            - "-c"
            - "mysqladmin ping -u root -p${MYSQL_ROOT_PASSWORD}"
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
        volumeMounts:
        - name: katib-mysql
          mountPath: /var/lib/mysql
      volumes:
      - name: katib-mysql
        persistentVolumeClaim:
          claimName: katib-mysql
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-mysql-pvc.yaml", `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: katib-mysql
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-mysql-secret.yaml", `
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: katib-mysql-secrets
data:
  MYSQL_ROOT_PASSWORD: dGVzdA== # "test"
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-mysql-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-mysql
  labels:
    app: katib
    component: mysql
spec:
  type: ClusterIP
  ports:
    - port: 3306
      protocol: TCP
      name: dbapi
  selector:
    app: katib
    component: mysql
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-db-manager-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: katib-db-manager
  labels:
    app: katib
    component: db-manager
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib
      component: db-manager
  template:
    metadata:
      name: katib-db-manager
      labels:
        app: katib
        component: db-manager
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      containers:
      - name: katib-db-manager
        image: gcr.io/kubeflow-images-public/katib/v1alpha3/katib-db-manager
        imagePullPolicy: IfNotPresent
        env:
          - name : DB_NAME
            value: "mysql"
          - name: DB_PASSWORD
            valueFrom:
              secretKeyRef:
                name: katib-mysql-secrets
                key: MYSQL_ROOT_PASSWORD
        command:
          - './katib-db-manager'
        ports:
        - name: api
          containerPort: 6789
        readinessProbe:
          exec:
            command: ["/bin/grpc_health_probe", "-addr=:6789"]
          initialDelaySeconds: 5
        livenessProbe:
          exec:
            command: ["/bin/grpc_health_probe", "-addr=:6789"]
          initialDelaySeconds: 10
          periodSeconds: 60
          failureThreshold: 5
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-db-manager-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-db-manager
  labels:
    app: katib
    component: db-manager
spec:
  type: ClusterIP
  ports:
    - port: 6789
      protocol: TCP
      name: api
  selector:
    app: katib
    component: db-manager
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-ui-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: katib-ui
  labels:
    app: katib
    component: ui
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib
      component: ui
  template:
    metadata:
      name: katib-ui
      labels:
        app: katib
        component: ui
    spec:
      containers:
      - name: katib-ui
        image: gcr.io/kubeflow-images-public/katib/v1alpha3/katib-ui
        imagePullPolicy: IfNotPresent
        command:
          - './katib-ui'
        args:
          - '--port=8080'
        env:
        - name: KATIB_CORE_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        ports:
        - name: ui
          containerPort: 8080
      serviceAccountName: katib-ui
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-ui-rbac.yaml", `
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: katib-ui
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  - namespaces
  verbs:
  - "*"
- apiGroups:
  - kubeflow.org
  resources:
  - experiments
  - trials
  verbs:
  - "*"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: katib-ui
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: katib-ui
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: katib-ui
subjects:
- kind: ServiceAccount
  name: katib-ui
`)
	th.writeF("/manifests/katib/katib-controller/base/katib-ui-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-ui
  labels:
    app: katib
    component: ui
spec:
  type: ClusterIP
  ports:
    - port: 80
      protocol: TCP
      name: ui
      targetPort: 8080
  selector:
    app: katib
    component: ui
`)
	th.writeF("/manifests/katib/katib-controller/base/trial-template-configmap.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: trial-template
data:
  defaultTrialTemplate.yaml : |-
    apiVersion: batch/v1
    kind: Job
    metadata:
      name: {{.Trial}}
      namespace: {{.NameSpace}}
    spec:
      template:
        spec:
          containers:
          - name: {{.Trial}}
            image: docker.io/kubeflowkatib/mxnet-mnist
            command:
            - "python3"
            - "/opt/mxnet-mnist/mnist.py"
            - "--batch-size=64"
            {{- with .HyperParameters}}
            {{- range .}}
            - "{{.Name}}={{.Value}}"
            {{- end}}
            {{- end}}
          restartPolicy: Never
`)
	th.writeF("/manifests/katib/katib-controller/base/params.yaml", `
varReference:
- path: data/config
  kind: ConfigMap
- path: data/config
  kind: Deployment
- path: metadata/annotations/getambassador.io\/config
  kind: Service
`)
	th.writeF("/manifests/katib/katib-controller/base/params.env", `
clusterDomain=cluster.local
`)
	th.writeK("/manifests/katib/katib-controller/base", `
namespace: kubeflow
resources:
- katib-configmap.yaml
- katib-controller-deployment.yaml
- katib-controller-rbac.yaml
- katib-controller-secret.yaml
- katib-controller-service.yaml
- katib-mysql-deployment.yaml
- katib-mysql-pvc.yaml
- katib-mysql-secret.yaml
- katib-mysql-service.yaml
- katib-db-manager-deployment.yaml
- katib-db-manager-service.yaml
- katib-ui-deployment.yaml
- katib-ui-rbac.yaml
- katib-ui-service.yaml
- trial-template-configmap.yaml
configMapGenerator:
- name: katib-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
images:
- name: gcr.io/kubeflow-images-public/katib/v1alpha3/katib-controller
  newTag: v0.8.0
  newName: gcr.io/kubeflow-images-public/katib/v1alpha3/katib-controller
- name: gcr.io/kubeflow-images-public/katib/v1alpha3/katib-db-manager
  newTag: v0.8.0
  newName: gcr.io/kubeflow-images-public/katib/v1alpha3/katib-db-manager
- name: gcr.io/kubeflow-images-public/katib/v1alpha3/katib-ui
  newTag: v0.8.0
  newName: gcr.io/kubeflow-images-public/katib/v1alpha3/katib-ui
- name: mysql
  newTag: "8"
  newName: mysql
vars:
- name: clusterDomain
  objref:
    kind: ConfigMap
    name: katib-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.clusterDomain
- name: namespace
  objref:
    kind: Service
    name: katib-ui
    apiVersion: v1
  fieldref:
    fieldpath: metadata.namespace
configurations:
- params.yaml
`)
}

func TestKatibControllerOverlaysIstio(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib/katib-controller/overlays/istio")
	writeKatibControllerOverlaysIstio(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib/katib-controller/overlays/istio"
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
