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

func writeIstioCrds131Base(th *KustTestHarness) {
	th.writeF("/manifests/istio-1-3-1/istio-crds-1-3-1/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: virtualservices.networking.istio.io
  labels:
    app: istio-pilot
spec:
  group: networking.istio.io
  names:
    kind: VirtualService
    listKind: VirtualServiceList
    plural: virtualservices
    singular: virtualservice
    shortNames:
    - vs
    categories:
    - istio-io
    - networking-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha3
      served: true
      storage: true
  additionalPrinterColumns:
  - JSONPath: .spec.gateways
    description: The names of gateways and sidecars that should apply these routes
    name: Gateways
    type: string
  - JSONPath: .spec.hosts
    description: The destination hosts to which traffic is being sent
    name: Hosts
    type: string
  - JSONPath: .metadata.creationTimestamp
    description: |-
      CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.

      Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
    name: Age
    type: date
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: destinationrules.networking.istio.io
  labels:
    app: istio-pilot
spec:
  group: networking.istio.io
  names:
    kind: DestinationRule
    listKind: DestinationRuleList
    plural: destinationrules
    singular: destinationrule
    shortNames:
    - dr
    categories:
    - istio-io
    - networking-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha3
      served: true
      storage: true
  additionalPrinterColumns:
  - JSONPath: .spec.host
    description: The name of a service from the service registry
    name: Host
    type: string
  - JSONPath: .metadata.creationTimestamp
    description: |-
      CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.

      Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
    name: Age
    type: date
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: serviceentries.networking.istio.io
  labels:
    app: istio-pilot
spec:
  group: networking.istio.io
  names:
    kind: ServiceEntry
    listKind: ServiceEntryList
    plural: serviceentries
    singular: serviceentry
    shortNames:
    - se
    categories:
    - istio-io
    - networking-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha3
      served: true
      storage: true
  additionalPrinterColumns:
  - JSONPath: .spec.hosts
    description: The hosts associated with the ServiceEntry
    name: Hosts
    type: string
  - JSONPath: .spec.location
    description: Whether the service is external to the mesh or part of the mesh (MESH_EXTERNAL or MESH_INTERNAL)
    name: Location
    type: string
  - JSONPath: .spec.resolution
    description: Service discovery mode for the hosts (NONE, STATIC, or DNS)
    name: Resolution
    type: string
  - JSONPath: .metadata.creationTimestamp
    description: |-
      CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.

      Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
    name: Age
    type: date
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: gateways.networking.istio.io
  labels:
    app: istio-pilot
spec:
  group: networking.istio.io
  names:
    kind: Gateway
    plural: gateways
    singular: gateway
    shortNames:
    - gw
    categories:
    - istio-io
    - networking-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha3
      served: true
      storage: true
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: envoyfilters.networking.istio.io
  labels:
    app: istio-pilot
spec:
  group: networking.istio.io
  names:
    kind: EnvoyFilter
    plural: envoyfilters
    singular: envoyfilter
    categories:
    - istio-io
    - networking-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha3
      served: true
      storage: true
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: clusterrbacconfigs.rbac.istio.io
  labels:
    app: istio-pilot
    istio: rbac
  annotations:
    "helm.sh/resource-policy": keep
spec:
  group: rbac.istio.io
  names:
    kind: ClusterRbacConfig
    plural: clusterrbacconfigs
    singular: clusterrbacconfig
    categories:
    - istio-io
    - rbac-istio-io
  scope: Cluster
  versions:
    - name: v1alpha1
      served: true
      storage: true
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: policies.authentication.istio.io
  labels:
    app: istio-citadel
spec:
  group: authentication.istio.io
  names:
    kind: Policy
    plural: policies
    singular: policy
    categories:
    - istio-io
    - authentication-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha1
      served: true
      storage: true
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: meshpolicies.authentication.istio.io
  labels:
    app: istio-citadel
spec:
  group: authentication.istio.io
  names:
    kind: MeshPolicy
    listKind: MeshPolicyList
    plural: meshpolicies
    singular: meshpolicy
    categories:
    - istio-io
    - authentication-istio-io
  scope: Cluster
  versions:
    - name: v1alpha1
      served: true
      storage: true
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: httpapispecbindings.config.istio.io
  labels:
    app: istio-mixer
spec:
  group: config.istio.io
  names:
    kind: HTTPAPISpecBinding
    plural: httpapispecbindings
    singular: httpapispecbinding
    categories:
    - istio-io
    - apim-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha2
      served: true
      storage: true
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: httpapispecs.config.istio.io
  labels:
    app: istio-mixer
spec:
  group: config.istio.io
  names:
    kind: HTTPAPISpec
    plural: httpapispecs
    singular: httpapispec
    categories:
    - istio-io
    - apim-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha2
      served: true
      storage: true
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: quotaspecbindings.config.istio.io
  labels:
    app: istio-mixer
spec:
  group: config.istio.io
  names:
    kind: QuotaSpecBinding
    plural: quotaspecbindings
    singular: quotaspecbinding
    categories:
    - istio-io
    - apim-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha2
      served: true
      storage: true
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: quotaspecs.config.istio.io
  labels:
    app: istio-mixer
spec:
  group: config.istio.io
  names:
    kind: QuotaSpec
    plural: quotaspecs
    singular: quotaspec
    categories:
    - istio-io
    - apim-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha2
      served: true
      storage: true
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: rules.config.istio.io
  labels:
    app: mixer
    package: istio.io.mixer
    istio: core
spec:
  group: config.istio.io
  names:
    kind: rule
    plural: rules
    singular: rule
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha2
      served: true
      storage: true
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: attributemanifests.config.istio.io
  labels:
    app: mixer
    package: istio.io.mixer
    istio: core
spec:
  group: config.istio.io
  names:
    kind: attributemanifest
    plural: attributemanifests
    singular: attributemanifest
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha2
      served: true
      storage: true
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: rbacconfigs.rbac.istio.io
  labels:
    app: mixer
    package: istio.io.mixer
    istio: rbac
spec:
  group: rbac.istio.io
  names:
    kind: RbacConfig
    plural: rbacconfigs
    singular: rbacconfig
    categories:
    - istio-io
    - rbac-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha1
      served: true
      storage: true
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: serviceroles.rbac.istio.io
  labels:
    app: mixer
    package: istio.io.mixer
    istio: rbac
spec:
  group: rbac.istio.io
  names:
    kind: ServiceRole
    plural: serviceroles
    singular: servicerole
    categories:
    - istio-io
    - rbac-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha1
      served: true
      storage: true
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: servicerolebindings.rbac.istio.io
  labels:
    app: mixer
    package: istio.io.mixer
    istio: rbac
spec:
  group: rbac.istio.io
  names:
    kind: ServiceRoleBinding
    plural: servicerolebindings
    singular: servicerolebinding
    categories:
    - istio-io
    - rbac-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha1
      served: true
      storage: true
  additionalPrinterColumns:
  - JSONPath: .spec.roleRef.name
    description: The name of the ServiceRole object being referenced
    name: Reference
    type: string
  - JSONPath: .metadata.creationTimestamp
    description: |-
      CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.

      Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
    name: Age
    type: date
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: adapters.config.istio.io
  labels:
    app: mixer
    package: adapter
    istio: mixer-adapter
spec:
  group: config.istio.io
  names:
    kind: adapter
    plural: adapters
    singular: adapter
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha2
      served: true
      storage: true
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: instances.config.istio.io
  labels:
    app: mixer
    package: instance
    istio: mixer-instance
spec:
  group: config.istio.io
  names:
    kind: instance
    plural: instances
    singular: instance
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha2
      served: true
      storage: true
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: templates.config.istio.io
  labels:
    app: mixer
    package: template
    istio: mixer-template
spec:
  group: config.istio.io
  names:
    kind: template
    plural: templates
    singular: template
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha2
      served: true
      storage: true
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: handlers.config.istio.io
  labels:
    app: mixer
    package: handler
    istio: mixer-handler
spec:
  group: config.istio.io
  names:
    kind: handler
    plural: handlers
    singular: handler
    categories:
    - istio-io
    - policy-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha2
      served: true
      storage: true
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: sidecars.networking.istio.io
  labels:
    app: istio-pilot
spec:
  group: networking.istio.io
  names:
    kind: Sidecar
    plural: sidecars
    singular: sidecar
    categories:
    - istio-io
    - networking-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha3
      served: true
      storage: true
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: authorizationpolicies.rbac.istio.io
  labels:
    app: istio-pilot
    istio: rbac
spec:
  group: rbac.istio.io
  names:
    kind: AuthorizationPolicy
    plural: authorizationpolicies
    singular: authorizationpolicy
    categories:
      - istio-io
      - rbac-istio-io
  scope: Namespaced
  versions:
    - name: v1alpha1
      served: true
      storage: true
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: clusterissuers.certmanager.k8s.io
  labels:
    app: certmanager
spec:
  group: certmanager.k8s.io
  versions:
    - name: v1alpha1
      served: true
      storage: true
  names:
    kind: ClusterIssuer
    plural: clusterissuers
  scope: Cluster
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: issuers.certmanager.k8s.io
  labels:
    app: certmanager
spec:
  group: certmanager.k8s.io
  versions:
    - name: v1alpha1
      served: true
      storage: true
  names:
    kind: Issuer
    plural: issuers
  scope: Namespaced
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: certificates.certmanager.k8s.io
  labels:
    app: certmanager
spec:
  additionalPrinterColumns:
    - JSONPath: .status.conditions[?(@.type=="Ready")].status
      name: Ready
      type: string
    - JSONPath: .spec.secretName
      name: Secret
      type: string
    - JSONPath: .spec.issuerRef.name
      name: Issuer
      type: string
      priority: 1
    - JSONPath: .status.conditions[?(@.type=="Ready")].message
      name: Status
      type: string
      priority: 1
    - JSONPath: .metadata.creationTimestamp
      description: |-
        CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.

        Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
      name: Age
      type: date
  group: certmanager.k8s.io
  versions:
    - name: v1alpha1
      served: true
      storage: true
  scope: Namespaced
  names:
    kind: Certificate
    plural: certificates
    shortNames:
      - cert
      - certs
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: orders.certmanager.k8s.io
  labels:
    app: certmanager
spec:
  additionalPrinterColumns:
    - JSONPath: .status.state
      name: State
      type: string
    - JSONPath: .spec.issuerRef.name
      name: Issuer
      type: string
      priority: 1
    - JSONPath: .status.reason
      name: Reason
      type: string
      priority: 1
    - JSONPath: .metadata.creationTimestamp
      description: |-
        CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.

        Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
      name: Age
      type: date
  group: certmanager.k8s.io
  versions:
    - name: v1alpha1
      served: true
      storage: true
  names:
    kind: Order
    plural: orders
  scope: Namespaced
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: challenges.certmanager.k8s.io
  labels:
    app: certmanager
spec:
  additionalPrinterColumns:
    - JSONPath: .status.state
      name: State
      type: string
    - JSONPath: .spec.dnsName
      name: Domain
      type: string
    - JSONPath: .status.reason
      name: Reason
      type: string
    - JSONPath: .metadata.creationTimestamp
      description: |-
        CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.

        Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
      name: Age
      type: date
  group: certmanager.k8s.io
  versions:
    - name: v1alpha1
      served: true
      storage: true
  names:
    kind: Challenge
    plural: challenges
  scope: Namespaced
---
`)
	th.writeK("/manifests/istio-1-3-1/istio-crds-1-3-1/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- crd.yaml

commonLabels:
  kustomize.component: istio-crds
`)
}

func TestIstioCrds131Base(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/istio-1-3-1/istio-crds-1-3-1/base")
	writeIstioCrds131Base(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../istio-1-3-1/istio-crds-1-3-1/base"
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
