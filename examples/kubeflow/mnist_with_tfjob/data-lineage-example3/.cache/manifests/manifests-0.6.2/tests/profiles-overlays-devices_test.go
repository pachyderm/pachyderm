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
	th.writeF("/manifests/profiles/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  creationTimestamp: null
  labels:
    controller-tools.k8s.io: "1.0"
  name: profiles.kubeflow.org
spec:
  group: kubeflow.org
  names:
    kind: Profile
    plural: profiles
  scope: Cluster
  validation:
    openAPIV3Schema:
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
          properties:
            owner:
              description: The profile owner
              type: object
          type: object
        status:
          properties:
            message:
              type: string
            status:
              description: 'INSERT ADDITIONAL STATUS FIELD - define observed state
                of cluster Important: Run "make" to regenerate code after modifying
                this file'
              type: string
          type: object
  version: v1alpha1
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`)
	th.writeF("/manifests/profiles/base/service-account.yaml", `
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: controller-service-account
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: default-service-account
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
	th.writeF("/manifests/profiles/base/role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: default-role
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - profiles
  verbs:
  - create
  - watch
  - list
`)
	th.writeF("/manifests/profiles/base/role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: default-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: default-role
subjects:
- kind: ServiceAccount
  name: default-service-account
`)
	th.writeF("/manifests/profiles/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: kfam
spec:
  ports:
    - port: 8081
`)
	th.writeF("/manifests/profiles/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment
spec:
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
        image: gcr.io/kubeflow-images-public/profile-controller:v20190619-v0-219-gbd3daa8c-dirty-1ced0e
        imagePullPolicy: Always
        name: manager
      - command:
        - /opt/kubeflow/access-management
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
      serviceAccountName: controller-service-account
`)
	th.writeF("/manifests/profiles/base/params.yaml", `
varReference:
- path: spec/template/spec/containers/0/args/1
  kind: Deployment
- path: spec/template/spec/containers/0/args/3
  kind: Deployment
- path: spec/template/spec/containers/1/args/1
  kind: Deployment
- path: spec/template/spec/containers/1/args/3
  kind: Deployment
- path: spec/template/spec/containers/1/args/5
  kind: Deployment
`)
	th.writeF("/manifests/profiles/base/params.env", `
admin=
userid-header=
userid-prefix=
`)
	th.writeK("/manifests/profiles/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- crd.yaml
- service-account.yaml
- cluster-role-binding.yaml
- role.yaml
- role-binding.yaml
- service.yaml
- deployment.yaml
namePrefix: profiles-
namespace: kubeflow
commonLabels:
  kustomize.component: profiles
configMapGenerator:
  - name: profiles-parameters
    env: params.env
images:
  - name: gcr.io/kubeflow-images-public/profile-controller
    newName: gcr.io/kubeflow-images-public/profile-controller
    newTag: v20190619-v0-219-gbd3daa8c-dirty-1ced0e
vars:
  - name: admin
    objref:
      kind: ConfigMap
      name: profiles-parameters
      apiVersion: v1
    fieldref:
      fieldpath: data.admin
  - name: userid-header
    objref:
      kind: ConfigMap
      name: profiles-parameters
      apiVersion: v1
    fieldref:
      fieldpath: data.userid-header
  - name: userid-prefix
    objref:
      kind: ConfigMap
      name: profiles-parameters
      apiVersion: v1
    fieldref:
      fieldpath: data.userid-prefix
  - name: namespace
    objref:
      kind: Service
      name: kfam
      apiVersion: v1
    fieldref:
      fieldpath: metadata.namespace
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
	targetPath := "../profiles/overlays/devices"
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
