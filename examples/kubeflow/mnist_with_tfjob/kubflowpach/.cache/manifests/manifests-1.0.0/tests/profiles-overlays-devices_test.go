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

func writeProfilesOverlaysDevices(th *KustTestHarness) {
	th.writeF("/manifests/profiles/overlays/devices/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment
spec:
  template:
    spec:
      containers:
      - name: manager
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
`)
	th.writeK("/manifests/profiles/overlays/devices", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
patchesStrategicMerge:
- deployment.yaml
`)
	th.writeF("/manifests/profiles/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cluster-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: controller-service-account
`)
	th.writeF("/manifests/profiles/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  creationTimestamp: null
  name: profiles.kubeflow.org
spec:
  conversion:
    strategy: None
  group: kubeflow.org
  names:
    kind: Profile
    plural: profiles
  scope: Cluster
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: Profile is the Schema for the profiles API
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          description: ProfileSpec defines the desired state of Profile
          properties:
            owner:
              description: The profile owner
              properties:
                apiGroup:
                  description: APIGroup holds the API group of the referenced subject.
                    Defaults to "" for ServiceAccount subjects. Defaults to "rbac.authorization.k8s.io"
                    for User and Group subjects.
                  type: string
                kind:
                  description: Kind of object being referenced. Values defined by
                    this API group are "User", "Group", and "ServiceAccount". If the
                    Authorizer does not recognized the kind value, the Authorizer
                    should report an error.
                  type: string
                name:
                  description: Name of the object being referenced.
                  type: string
              required:
                - kind
                - name
              type: object
            plugins:
              items:
                description: Plugin is for customize actions on different platform.
                properties:
                  apiVersion:
                    description: 'APIVersion defines the versioned schema of this
                      representation of an object. Servers should convert recognized
                      schemas to the latest internal value, and may reject unrecognized
                      values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                    type: string
                  kind:
                    description: 'Kind is a string value representing the REST resource
                      this object represents. Servers may infer this from the endpoint
                      the client submits requests to. Cannot be updated. In CamelCase.
                      More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  spec:
                    type: object
                type: object
              type: array
            resourceQuotaSpec:
              description: Resourcequota that will be applied to target namespace
              properties:
                hard:
                  additionalProperties:
                    type: string
                  description: 'hard is the set of desired hard limits for each named
                    resource. More info: https://kubernetes.io/docs/concepts/policy/resource-quotas/'
                  type: object
                scopeSelector:
                  description: scopeSelector is also a collection of filters like
                    scopes that must match each object tracked by a quota but expressed
                    using ScopeSelectorOperator in combination with possible values.
                    For a resource to match, both scopes AND scopeSelector (if specified
                    in spec), must be matched.
                  properties:
                    matchExpressions:
                      description: A list of scope selector requirements by scope
                        of the resources.
                      items:
                        description: A scoped-resource selector requirement is a selector
                          that contains values, a scope name, and an operator that
                          relates the scope name and values.
                        properties:
                          operator:
                            description: Represents a scope's relationship to a set
                              of values. Valid operators are In, NotIn, Exists, DoesNotExist.
                            type: string
                          scopeName:
                            description: The name of the scope that the selector applies
                              to.
                            type: string
                          values:
                            description: An array of string values. If the operator
                              is In or NotIn, the values array must be non-empty.
                              If the operator is Exists or DoesNotExist, the values
                              array must be empty. This array is replaced during a
                              strategic merge patch.
                            items:
                              type: string
                            type: array
                        required:
                          - operator
                          - scopeName
                        type: object
                      type: array
                  type: object
                scopes:
                  description: A collection of filters that must match each object
                    tracked by a quota. If not specified, the quota matches all objects.
                  items:
                    description: A ResourceQuotaScope defines a filter that must match
                      each object tracked by a quota
                    type: string
                  type: array
              type: object
          type: object
        status:
          description: ProfileStatus defines the observed state of Profile
          properties:
            conditions:
              items:
                properties:
                  message:
                    type: string
                  status:
                    type: string
                  type:
                    type: string
                type: object
              type: array
          type: object
      type: object
  version: v1
  versions:
    - name: v1
      served: true
      storage: true
    - name: v1beta1
      served: true
      storage: false
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []`)
	th.writeF("/manifests/profiles/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment
spec:
  replicas: 1
  template:
    spec:
      containers:
      - command:
        - /manager
        args:
        - "-userid-header"
        - $(userid-header)
        - "-userid-prefix"
        - $(userid-prefix)
        - "-workload-identity"
        - $(gcp-sa)
        image: gcr.io/kubeflow-images-public/profile-controller:v20190619-v0-219-gbd3daa8c-dirty-1ced0e
        imagePullPolicy: Always
        name: manager
        livenessProbe:
          httpGet:
            path: /metrics
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 30
      - command:
        - /access-management
        args:
        - "-cluster-admin"
        - $(admin)
        - "-userid-header"
        - $(userid-header)
        - "-userid-prefix"
        - $(userid-prefix)
        image: gcr.io/kubeflow-images-public/kfam:v20190612-v0-170-ga06cdb79-dirty-a33ee4
        imagePullPolicy: Always
        name: kfam
        livenessProbe:
          httpGet:
            path: /metrics
            port: 8081
          initialDelaySeconds: 30
          periodSeconds: 30
      serviceAccountName: controller-service-account
`)
	th.writeF("/manifests/profiles/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: kfam
spec:
  ports:
    - port: 8081`)
	th.writeF("/manifests/profiles/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: controller-service-account
`)
	th.writeF("/manifests/profiles/base/params.yaml", `
varReference:
- path: spec/template/spec/containers/0/args/1
  kind: Deployment
- path: spec/template/spec/containers/0/args/3
  kind: Deployment
- path: spec/template/spec/containers/0/args/5
  kind: Deployment
- path: spec/template/spec/containers/1/args/1
  kind: Deployment
- path: spec/template/spec/containers/1/args/3
  kind: Deployment
- path: spec/template/spec/containers/1/args/5
  kind: Deployment
`)
	th.writeF("/manifests/profiles/base/params.env", `
admin=anonymous
gcp-sa=
userid-header=
userid-prefix=
`)
	th.writeK("/manifests/profiles/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cluster-role-binding.yaml
- crd.yaml
- deployment.yaml
- service.yaml
- service-account.yaml
namePrefix: profiles-
namespace: kubeflow
commonLabels:
  kustomize.component: profiles
configMapGenerator:
- envs:
  - params.env
  name: profiles-parameters
images:
- name: gcr.io/kubeflow-images-public/kfam
  newName: gcr.io/kubeflow-images-public/kfam
  newTag: v1.0.0-gf3e09203
- name: gcr.io/kubeflow-images-public/profile-controller
  newName: gcr.io/kubeflow-images-public/profile-controller
  newTag: v1.0.0-g34aa47c2
vars:
- fieldref:
    fieldPath: data.admin
  name: admin
  objref:
    apiVersion: v1
    kind: ConfigMap
    name: profiles-parameters
- fieldref:
    fieldPath: data.gcp-sa
  name: gcp-sa
  objref:
    apiVersion: v1
    kind: ConfigMap
    name: profiles-parameters
- fieldref:
    fieldPath: data.userid-header
  name: userid-header
  objref:
    apiVersion: v1
    kind: ConfigMap
    name: profiles-parameters
- fieldref:
    fieldPath: data.userid-prefix
  name: userid-prefix
  objref:
    apiVersion: v1
    kind: ConfigMap
    name: profiles-parameters
- fieldref:
    fieldPath: metadata.namespace
  name: namespace
  objref:
    apiVersion: v1
    kind: Service
    name: kfam
configurations:
- params.yaml
`)
}

func TestProfilesOverlaysDevices(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/profiles/overlays/devices")
	writeProfilesOverlaysDevices(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../profiles/overlays/devices"
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
