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

func writePipelineVisualizationServiceOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/pipeline/pipeline-visualization-service/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: pipeline-visualization-service
spec:
  addOwnerRef: true
  componentKinds:
  - group: core
    kind: ConfigMap
  - group: apps
    kind: Deployment
  descriptor:
    description: ''
    keywords:
    - pipeline-visualization-service
    - kubeflow
    links:
    - description: About
      url: ''
    maintainers: []
    owners: []
    type: pipeline-visualization-service
    version: v1beta1
  selector:
    matchLabels:
      app.kubernetes.io/component: pipeline-visualization-service
      app.kubernetes.io/instance: pipeline-visualization-service-0.2.0
      app.kubernetes.io/managed-by: kfctl
      app.kubernetes.io/name: pipeline-visualization-service
      app.kubernetes.io/part-of: kubeflow
      app.kubernetes.io/version: 0.2.0
`)
	th.writeK("/manifests/pipeline/pipeline-visualization-service/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
bases:
- ../../base
commonLabels:
  app.kubernetes.io/component: pipeline-visualization-service
  app.kubernetes.io/instance: pipeline-visualization-service-0.2.0
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/name: pipeline-visualization-service
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: 0.2.0
kind: Kustomization
resources:
- application.yaml
`)
	th.writeF("/manifests/pipeline/pipeline-visualization-service/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: ml-pipeline-visualizationserver
  name: ml-pipeline-visualizationserver
spec:
  selector:
    matchLabels:
      app: ml-pipeline-visualizationserver
  template:
    metadata:
      labels:
        app: ml-pipeline-visualizationserver
    spec:
      containers:
      - image: gcr.io/ml-pipeline/visualization-server
        imagePullPolicy: IfNotPresent
        name: ml-pipeline-visualizationserver
        ports:
        - containerPort: 8888
`)
	th.writeF("/manifests/pipeline/pipeline-visualization-service/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: ml-pipeline-visualizationserver
spec:
  ports:
  - name: http
    port: 8888
    protocol: TCP
    targetPort: 8888
  selector:
    app: ml-pipeline-visualizationserver
`)
	th.writeK("/manifests/pipeline/pipeline-visualization-service/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
nameprefix: ml-pipeline-
commonLabels:
  app: ml-pipeline-visualizationserver
resources:
- deployment.yaml
- service.yaml
images:
- name: gcr.io/ml-pipeline/visualization-server
  newTag: 0.2.0
  newName: gcr.io/ml-pipeline/visualization-server
`)
}

func TestPipelineVisualizationServiceOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pipeline/pipeline-visualization-service/overlays/application")
	writePipelineVisualizationServiceOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pipeline/pipeline-visualization-service/overlays/application"
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
