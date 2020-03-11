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

func writeKnativeServingCrdsOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/knative/knative-serving-crds/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: knative-serving-crds
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: knative-serving-crds
      app.kubernetes.io/instance: knative-serving-crds-v0.11.1
      app.kubernetes.io/managed-by: kfctl
      app.kubernetes.io/component: knative-serving-crds
      app.kubernetes.io/part-of: kubeflow
      app.kubernetes.io/version: v0.11.1
  componentKinds:
  - group: core
    kind: ConfigMap
  - group: apps
    kind: Deployment
  descriptor:
    type: knative-serving-crds
    version: v1beta1
    description: ""
    maintainers: []
    owners: []
    keywords:
     - knative-serving-crds
     - kubeflow
    links:
    - description: About
      url: ""
  addOwnerRef: true
`)
	th.writeK("/manifests/knative/knative-serving-crds/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
commonLabels:
  app.kubernetes.io/name: knative-serving-crds
  app.kubernetes.io/instance: knative-serving-crds-v0.11.1
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: knative-serving-crds
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.11.1
`)
	th.writeF("/manifests/knative/knative-serving-crds/base/namespace.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  labels:
    istio-injection: enabled
    serving.knative.dev/release: "v0.11.1"
  name: knative-serving


`)
	th.writeF("/manifests/knative/knative-serving-crds/base/crd.yaml", `
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    knative.dev/crd-install: "true"
    serving.knative.dev/release: "v0.11.1"
  name: certificates.networking.internal.knative.dev
spec:
  additionalPrinterColumns:
    - JSONPath: .status.conditions[?(@.type=="Ready")].status
      name: Ready
      type: string
    - JSONPath: .status.conditions[?(@.type=="Ready")].reason
      name: Reason
      type: string
  group: networking.internal.knative.dev
  names:
    categories:
      - knative-internal
      - networking
    kind: Certificate
    plural: certificates
    shortNames:
      - kcert
    singular: certificate
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha1

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    duck.knative.dev/podspecable: "true"
    knative.dev/crd-install: "true"
    serving.knative.dev/release: "v0.11.1"
  name: configurations.serving.knative.dev
spec:
  additionalPrinterColumns:
    - JSONPath: .status.latestCreatedRevisionName
      name: LatestCreated
      type: string
    - JSONPath: .status.latestReadyRevisionName
      name: LatestReady
      type: string
    - JSONPath: .status.conditions[?(@.type=='Ready')].status
      name: Ready
      type: string
    - JSONPath: .status.conditions[?(@.type=='Ready')].reason
      name: Reason
      type: string
  group: serving.knative.dev
  names:
    categories:
      - all
      - knative
      - serving
    kind: Configuration
    plural: configurations
    shortNames:
      - config
      - cfg
    singular: configuration
  scope: Namespaced
  subresources:
    status: {}
  versions:
    - name: v1alpha1
      served: true
      storage: true
    - name: v1beta1
      served: true
      storage: false
    - name: v1
      served: true
      storage: false

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
  labels:
    knative.dev/crd-install: "true"
    serving.knative.dev/release: "v0.11.1"
  name: ingresses.networking.internal.knative.dev
spec:
  additionalPrinterColumns:
    - JSONPath: .status.conditions[?(@.type=='Ready')].status
      name: Ready
      type: string
    - JSONPath: .status.conditions[?(@.type=='Ready')].reason
      name: Reason
      type: string
  group: networking.internal.knative.dev
  names:
    categories:
      - knative-internal
      - networking
    kind: Ingress
    plural: ingresses
    shortNames:
      - ing
    singular: ingress
  scope: Namespaced
  subresources:
    status: {}
  versions:
    - name: v1alpha1
      served: true
      storage: true

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    knative.dev/crd-install: "true"
    serving.knative.dev/release: "v0.11.1"
  name: metrics.autoscaling.internal.knative.dev
spec:
  additionalPrinterColumns:
    - JSONPath: .status.conditions[?(@.type=='Ready')].status
      name: Ready
      type: string
    - JSONPath: .status.conditions[?(@.type=='Ready')].reason
      name: Reason
      type: string
  group: autoscaling.internal.knative.dev
  names:
    categories:
      - knative-internal
      - autoscaling
    kind: Metric
    plural: metrics
    singular: metric
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
    serving.knative.dev/release: "v0.11.1"
  name: podautoscalers.autoscaling.internal.knative.dev
spec:
  additionalPrinterColumns:
    - JSONPath: .status.desiredScale
      name: DesiredScale
      type: integer
    - JSONPath: .status.actualScale
      name: ActualScale
      type: integer
    - JSONPath: .status.conditions[?(@.type=='Ready')].status
      name: Ready
      type: string
    - JSONPath: .status.conditions[?(@.type=='Ready')].reason
      name: Reason
      type: string
  group: autoscaling.internal.knative.dev
  names:
    categories:
      - knative-internal
      - autoscaling
    kind: PodAutoscaler
    plural: podautoscalers
    shortNames:
      - kpa
      - pa
    singular: podautoscaler
  scope: Namespaced
  subresources:
    status: {}
  versions:
    - name: v1alpha1
      served: true
      storage: true

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    knative.dev/crd-install: "true"
    serving.knative.dev/release: "v0.11.1"
  name: revisions.serving.knative.dev
spec:
  additionalPrinterColumns:
    - JSONPath: .metadata.labels['serving\.knative\.dev/configuration']
      name: Config Name
      type: string
    - JSONPath: .status.serviceName
      name: K8s Service Name
      type: string
    - JSONPath: .metadata.labels['serving\.knative\.dev/configurationGeneration']
      name: Generation
      type: string
    - JSONPath: .status.conditions[?(@.type=='Ready')].status
      name: Ready
      type: string
    - JSONPath: .status.conditions[?(@.type=='Ready')].reason
      name: Reason
      type: string
  group: serving.knative.dev
  names:
    categories:
      - all
      - knative
      - serving
    kind: Revision
    plural: revisions
    shortNames:
      - rev
    singular: revision
  scope: Namespaced
  subresources:
    status: {}
  versions:
    - name: v1alpha1
      served: true
      storage: true
    - name: v1beta1
      served: true
      storage: false
    - name: v1
      served: true
      storage: false

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    duck.knative.dev/addressable: "true"
    knative.dev/crd-install: "true"
    serving.knative.dev/release: "v0.11.1"
  name: routes.serving.knative.dev
spec:
  additionalPrinterColumns:
    - JSONPath: .status.url
      name: URL
      type: string
    - JSONPath: .status.conditions[?(@.type=='Ready')].status
      name: Ready
      type: string
    - JSONPath: .status.conditions[?(@.type=='Ready')].reason
      name: Reason
      type: string
  group: serving.knative.dev
  names:
    categories:
      - all
      - knative
      - serving
    kind: Route
    plural: routes
    shortNames:
      - rt
    singular: route
  scope: Namespaced
  subresources:
    status: {}
  versions:
    - name: v1alpha1
      served: true
      storage: true
    - name: v1beta1
      served: true
      storage: false
    - name: v1
      served: true
      storage: false

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    duck.knative.dev/addressable: "true"
    duck.knative.dev/podspecable: "true"
    knative.dev/crd-install: "true"
    serving.knative.dev/release: "v0.11.1"
  name: services.serving.knative.dev
spec:
  additionalPrinterColumns:
    - JSONPath: .status.url
      name: URL
      type: string
    - JSONPath: .status.latestCreatedRevisionName
      name: LatestCreated
      type: string
    - JSONPath: .status.latestReadyRevisionName
      name: LatestReady
      type: string
    - JSONPath: .status.conditions[?(@.type=='Ready')].status
      name: Ready
      type: string
    - JSONPath: .status.conditions[?(@.type=='Ready')].reason
      name: Reason
      type: string
  group: serving.knative.dev
  names:
    categories:
      - all
      - knative
      - serving
    kind: Service
    plural: services
    shortNames:
      - kservice
      - ksvc
    singular: service
  scope: Namespaced
  subresources:
    status: {}
  versions:
    - name: v1alpha1
      served: true
      storage: true
    - name: v1beta1
      served: true
      storage: false
    - name: v1
      served: true
      storage: false

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    knative.dev/crd-install: "true"
    serving.knative.dev/release: "v0.11.1"
  name: serverlessservices.networking.internal.knative.dev
spec:
  additionalPrinterColumns:
    - JSONPath: .spec.mode
      name: Mode
      type: string
    - JSONPath: .status.serviceName
      name: ServiceName
      type: string
    - JSONPath: .status.privateServiceName
      name: PrivateServiceName
      type: string
    - JSONPath: .status.conditions[?(@.type=='Ready')].status
      name: Ready
      type: string
    - JSONPath: .status.conditions[?(@.type=='Ready')].reason
      name: Reason
      type: string
  group: networking.internal.knative.dev
  names:
    categories:
      - knative-internal
      - networking
    kind: ServerlessService
    plural: serverlessservices
    shortNames:
      - sks
    singular: serverlessservice
  scope: Namespaced
  subresources:
    status: {}
  versions:
    - name: v1alpha1
      served: true
      storage: true
`)
	th.writeK("/manifests/knative/knative-serving-crds/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- namespace.yaml
- crd.yaml
`)
}

func TestKnativeServingCrdsOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/knative/knative-serving-crds/overlays/application")
	writeKnativeServingCrdsOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../knative/knative-serving-crds/overlays/application"
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
