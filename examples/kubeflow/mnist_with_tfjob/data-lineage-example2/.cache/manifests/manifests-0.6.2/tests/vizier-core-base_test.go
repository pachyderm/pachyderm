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

func writeVizierCoreBase(th *KustTestHarness) {
	th.writeF("/manifests/katib-v1alpha1/vizier-core/base/vizier-core-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: vizier-core
  labels:
    component: core
spec:
  replicas: 1
  template:
    metadata:
      name: vizier-core
      labels:
        component: core
    spec:
      serviceAccountName: vizier-core
      containers:
      - name: vizier-core
        image: gcr.io/kubeflow-images-public/katib/vizier-core:v0.1.2-alpha-156-g4ab3dbd
        env:
          - name: MYSQL_ROOT_PASSWORD
            valueFrom:
              secretKeyRef:
                name: vizier-db-secrets
                key: MYSQL_ROOT_PASSWORD
        command:
          - './vizier-manager'
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
`)
	th.writeF("/manifests/katib-v1alpha1/vizier-core/base/vizier-core-rbac.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: vizier-core
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: vizier-core
subjects:
- kind: ServiceAccount
  name: vizier-core
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: vizier-core
rules:
  - apiGroups: [""]
    resources: ["pods", "nodes", "nodes/*", "pods/log", "pods/status", "services", "persistentvolumes", "persistentvolumes/status","persistentvolumeclaims","persistentvolumeclaims/status"]
    verbs: ["*"]
  - apiGroups: ["batch"]
    resources: ["jobs", "jobs/status"]
    verbs: ["*"]
  - apiGroups: ["extensions"]
    verbs: ["*"]
    resources: ["ingresses","ingresses/status","deployments","deployments/status"]
  - apiGroups: [""]
    verbs: ["*"]
    resources: ["services"]
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: vizier-core
`)
	th.writeF("/manifests/katib-v1alpha1/vizier-core/base/vizier-core-rest-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: vizier-core-rest
  labels:
    component: core-rest
spec:
  replicas: 1
  template:
    metadata:
      name: vizier-core-rest
      labels:
        component: core-rest
    spec:
      containers:
      - name: vizier-core-rest
        image: gcr.io/kubeflow-images-public/katib/vizier-core-rest:v0.1.2-alpha-156-g4ab3dbd
        command:
          - './vizier-manager-rest'
        ports:
        - name: api
          containerPort: 80
`)
	th.writeF("/manifests/katib-v1alpha1/vizier-core/base/vizier-core-rest-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: vizier-core-rest
  labels:
    component: core-rest
spec:
  type: ClusterIP
  ports:
    - port: 80
      protocol: TCP
      name: api
  selector:
    component: core-rest
`)
	th.writeF("/manifests/katib-v1alpha1/vizier-core/base/vizier-core-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: vizier-core
  labels:
    component: core
spec:
  type: NodePort
  ports:
    - port: 6789
      protocol: TCP
      nodePort: 30678
      name: api
  selector:
    component: core
`)
	th.writeK("/manifests/katib-v1alpha1/vizier-core/base", `
namespace: kubeflow
resources:
- vizier-core-deployment.yaml
- vizier-core-rbac.yaml
- vizier-core-rest-deployment.yaml
- vizier-core-rest-service.yaml
- vizier-core-service.yaml
generatorOptions:
  disableNameSuffixHash: true
images:
  - name: gcr.io/kubeflow-images-public/katib/vizier-core
    newTag: v0.1.2-alpha-157-g3d4cd04
  - name: gcr.io/kubeflow-images-public/katib/vizier-core-rest
    newTag: v0.1.2-alpha-157-g3d4cd04
`)
}

func TestVizierCoreBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib-v1alpha1/vizier-core/base")
	writeVizierCoreBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib-v1alpha1/vizier-core/base"
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
