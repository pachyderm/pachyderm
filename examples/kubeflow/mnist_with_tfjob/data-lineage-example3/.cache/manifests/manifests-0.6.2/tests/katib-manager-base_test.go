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

func writeKatibManagerBase(th *KustTestHarness) {
	th.writeF("/manifests/katib-v1alpha2/katib-manager/base/katib-manager-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: katib-manager
  labels:
    app: katib
    component: manager
spec:
  replicas: 1
  template:
    metadata:
      name: katib-manager
      labels:
        app: katib
        component: manager
    spec:
      containers:
      - name: katib-manager
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-manager:v0.1.2-alpha-289-g14dad8b
        imagePullPolicy: IfNotPresent
        env:
          - name: MYSQL_ROOT_PASSWORD
            valueFrom:
              secretKeyRef:
                name: katib-db-secrets
                key: MYSQL_ROOT_PASSWORD
        command:
          - './katib-manager'
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
	th.writeF("/manifests/katib-v1alpha2/katib-manager/base/katib-manager-rest-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: katib-manager-rest
  labels:
    app: katib
    component: manager-rest
spec:
  replicas: 1
  template:
    metadata:
      name: katib-manager-rest
      labels:
        app: katib
        component: manager-rest
    spec:
      containers:
      - name: katib-manager-rest
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-manager-rest:v0.1.2-alpha-289-g14dad8b
        imagePullPolicy: IfNotPresent
        command:
          - './katib-manager-rest'
        ports:
        - name: api
          containerPort: 80
`)
	th.writeF("/manifests/katib-v1alpha2/katib-manager/base/katib-manager-rest-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-manager-rest
  labels:
    app: katib
    component: manager-rest
spec:
  type: ClusterIP
  ports:
    - port: 80
      protocol: TCP
      name: api
  selector:
    app: katib
    component: manager-rest
`)
	th.writeF("/manifests/katib-v1alpha2/katib-manager/base/katib-manager-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-manager
  labels:
    app: katib
    component: manager
spec:
  type: ClusterIP
  ports:
    - port: 6789
      protocol: TCP
      name: api
  selector:
    app: katib
    component: manager
`)
	th.writeK("/manifests/katib-v1alpha2/katib-manager/base", `
namespace: kubeflow
resources:
- katib-manager-deployment.yaml
- katib-manager-rest-deployment.yaml
- katib-manager-rest-service.yaml
- katib-manager-service.yaml
generatorOptions:
  disableNameSuffixHash: true
images:
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-manager
    newTag: v0.6.0-rc.0
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-manager-rest
    newTag: v0.6.0-rc.0
`)
}

func TestKatibManagerBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib-v1alpha2/katib-manager/base")
	writeKatibManagerBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib-v1alpha2/katib-manager/base"
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
