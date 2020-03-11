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

func writeJupyterWebAppBase(th *KustTestHarness) {
	th.writeF("/manifests/jupyter/jupyter-web-app/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cluster-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-role
subjects:
- kind: ServiceAccount
  name: service-account
`)
	th.writeF("/manifests/jupyter/jupyter-web-app/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-role
rules:
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - get
  - list
  - create
  - delete
- apiGroups:
  - kubeflow.org
  resources:
  - notebooks
  - poddefaults  
  verbs:
  - get
  - list
  - create
  - delete
- apiGroups:
  - ""
  resources:
  - persistentvolumeclaims
  verbs:
  - create
  - delete
  - get
  - list
- apiGroups:
  - storage.k8s.io
  resources:
  - storageclasses
  verbs:
  - get
  - list
  - watch
`)
	th.writeF("/manifests/jupyter/jupyter-web-app/base/config-map.yaml", `
apiVersion: v1
data:
  spawner_ui_config.yaml: |
    # Configuration file for the Jupyter UI.
    #
    # Each Jupyter UI option is configured by two keys: 'value' and 'readOnly'
    # - The 'value' key contains the default value
    # - The 'readOnly' key determines if the option will be available to users
    #
    # If the 'readOnly' key is present and set to 'true', the respective option
    # will be disabled for users and only set by the admin. Also when a
    # Notebook is POSTED to the API if a necessary field is not present then
    # the value from the config will be used.
    #
    # If the 'readOnly' key is missing (defaults to 'false'), the respective option
    # will be available for users to edit.
    #
    # Note that some values can be templated. Such values are the names of the
    # Volumes as well as their StorageClass
    spawnerFormDefaults:
      image:
        # The container Image for the user's Jupyter Notebook
        # If readonly, this value must be a member of the list below
        value: gcr.io/kubeflow-images-public/tensorflow-1.13.1-notebook-cpu:v0.5.0
        # The list of available standard container Images
        options:
          - gcr.io/kubeflow-images-public/tensorflow-1.5.1-notebook-cpu:v0.5.0
          - gcr.io/kubeflow-images-public/tensorflow-1.5.1-notebook-gpu:v0.5.0
          - gcr.io/kubeflow-images-public/tensorflow-1.6.0-notebook-cpu:v0.5.0
          - gcr.io/kubeflow-images-public/tensorflow-1.6.0-notebook-gpu:v0.5.0
          - gcr.io/kubeflow-images-public/tensorflow-1.7.0-notebook-cpu:v0.5.0
          - gcr.io/kubeflow-images-public/tensorflow-1.7.0-notebook-gpu:v0.5.0
          - gcr.io/kubeflow-images-public/tensorflow-1.8.0-notebook-cpu:v0.5.0
          - gcr.io/kubeflow-images-public/tensorflow-1.8.0-notebook-gpu:v0.5.0
          - gcr.io/kubeflow-images-public/tensorflow-1.9.0-notebook-cpu:v0.5.0
          - gcr.io/kubeflow-images-public/tensorflow-1.9.0-notebook-gpu:v0.5.0
          - gcr.io/kubeflow-images-public/tensorflow-1.10.1-notebook-cpu:v0.5.0
          - gcr.io/kubeflow-images-public/tensorflow-1.10.1-notebook-gpu:v0.5.0
          - gcr.io/kubeflow-images-public/tensorflow-1.11.0-notebook-cpu:v0.5.0
          - gcr.io/kubeflow-images-public/tensorflow-1.11.0-notebook-gpu:v0.5.0
          - gcr.io/kubeflow-images-public/tensorflow-1.12.0-notebook-cpu:v0.5.0
          - gcr.io/kubeflow-images-public/tensorflow-1.12.0-notebook-gpu:v0.5.0
          - gcr.io/kubeflow-images-public/tensorflow-1.13.1-notebook-cpu:v0.5.0
          - gcr.io/kubeflow-images-public/tensorflow-1.13.1-notebook-gpu:v0.5.0
          - gcr.io/kubeflow-images-public/tensorflow-2.0.0a-notebook-cpu:v0.5.0
          - gcr.io/kubeflow-images-public/tensorflow-2.0.0a-notebook-gpu:v0.5.0
        # By default, custom container Images are allowed
        # Uncomment the following line to only enable standard container Images
        readOnly: false
      cpu:
        # CPU for user's Notebook
        value: '0.5'
        readOnly: false
      memory:
        # Memory for user's Notebook
        value: 1.0Gi
        readOnly: false
      workspaceVolume:
        # Workspace Volume to be attached to user's Notebook
        # Each Workspace Volume is declared with the following attributes:
        # Type, Name, Size, MountPath and Access Mode
        value:
          type:
            # The Type of the Workspace Volume
            # Supported values: 'New', 'Existing'
            value: New
          name:
            # The Name of the Workspace Volume
            # Note that this is a templated value. Special values:
            # {notebook-name}: Replaced with the name of the Notebook. The frontend
            #                  will replace this value as the user types the name
            value: 'workspace-{notebook-name}'
          size:
            # The Size of the Workspace Volume (in Gi)
            value: '10Gi'
          mountPath:
            # The Path that the Workspace Volume will be mounted
            value: /home/jovyan
          accessModes:
            # The Access Mode of the Workspace Volume
            # Supported values: 'ReadWriteOnce', 'ReadWriteMany', 'ReadOnlyMany'
            value: ReadWriteOnce
          class:
            # The StrageClass the PVC will use if type is New. Special values are:
            # {none}: default StorageClass
            # {empty}: empty string ""
            value: '{none}'
        readOnly: false
      dataVolumes:
        # List of additional Data Volumes to be attached to the user's Notebook
        value: []
        # Each Data Volume is declared with the following attributes:
        # Type, Name, Size, MountPath and Access Mode
        #
        # For example, a list with 2 Data Volumes:
        # value:
        #   - value:
        #       type:
        #         value: New
        #       name:
        #         value: '{notebook-name}-vol-1'
        #       size:
        #         value: '10Gi'
        #       class:
        #         value: standard
        #       mountPath:
        #         value: /home/jovyan/vol-1
        #       accessModes:
        #         value: ReadWriteOnce
        #       class:
        #         value: {none}
        #   - value:
        #       type:
        #         value: New
        #       name:
        #         value: '{notebook-name}-vol-2'
        #       size:
        #         value: '10Gi'
        #       mountPath:
        #         value: /home/jovyan/vol-2
        #       accessModes:
        #         value: ReadWriteMany
        #       class:
        #         value: {none}
        readOnly: false
      extraResources:
        # Extra Resource Limits for user's Notebook
        # e.x. "{'nvidia.com/gpu': 2}"
        value: "{}"
        readOnly: false
      shm:
        value: true
        readOnly: false
      configurations:
        # List of labels to be selected, these are the labels from PodDefaults
        # value:
        #   - add-gcp-secret
        #   - default-editor
        value: []
        readOnly: false
kind: ConfigMap
metadata:
  name: config
`)
	th.writeF("/manifests/jupyter/jupyter-web-app/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment
spec:
  replicas: 1
  template:
    spec:
      containers:
      - env:
        - name: ROK_SECRET_NAME
          valueFrom:
            configMapKeyRef:
              name: parameters
              key: ROK_SECRET_NAME
        - name: UI
          valueFrom:
            configMapKeyRef:
              name: parameters
              key: UI
        - name: USERID_HEADER
          value: $(userid-header)
        - name: USERID_PREFIX
          value: $(userid-prefix)
        image: gcr.io/kubeflow-images-public/jupyter-web-app:v0.5.0
        imagePullPolicy: $(policy)
        name: jupyter-web-app
        ports:
        - containerPort: 5000
        volumeMounts:
        - mountPath: /etc/config
          name: config-volume
      serviceAccountName: service-account
      volumes:
      - configMap:
          name: config
        name: config-volume
`)
	th.writeF("/manifests/jupyter/jupyter-web-app/base/role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: jupyter-notebook-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: jupyter-notebook-role
subjects:
- kind: ServiceAccount
  name: jupyter-notebook
`)
	th.writeF("/manifests/jupyter/jupyter-web-app/base/role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  name: jupyter-notebook-role
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - pods/log
  - secrets
  - services
  verbs:
  - '*'
- apiGroups:
  - ""
  - apps
  - extensions
  resources:
  - deployments
  - replicasets
  verbs:
  - '*'
- apiGroups:
  - kubeflow.org
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - batch
  resources:
  - jobs
  verbs:
  - '*'
`)
	th.writeF("/manifests/jupyter/jupyter-web-app/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: service-account
`)
	th.writeF("/manifests/jupyter/jupyter-web-app/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  annotations:
    getambassador.io/config: |-
      ---
      apiVersion: ambassador/v0
      kind:  Mapping
      name: webapp_mapping
      prefix: /$(prefix)/
      service: jupyter-web-app-service.$(namespace)
      add_request_headers:
        x-forwarded-prefix: /jupyter
  labels:
    run: jupyter-web-app
  name: service
spec:
  ports:
  - name: http
    port: 80
    protocol: TCP
    targetPort: 5000
  type: ClusterIP
`)
	th.writeF("/manifests/jupyter/jupyter-web-app/base/params.yaml", `
varReference:
- path: spec/template/spec/containers/imagePullPolicy
  kind: Deployment
- path: metadata/annotations/getambassador.io\/config
  kind: Service
- path: spec/template/spec/containers/0/env/2/value
  kind: Deployment
- path: spec/template/spec/containers/0/env/3/value
  kind: Deployment
`)
	th.writeF("/manifests/jupyter/jupyter-web-app/base/params.env", `
UI=default
ROK_SECRET_NAME=secret-rok-{username}
policy=Always
prefix=jupyter
clusterDomain=cluster.local
userid-header=
userid-prefix=
`)
	th.writeK("/manifests/jupyter/jupyter-web-app/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- config-map.yaml
- deployment.yaml
- role-binding.yaml
- role.yaml
- service-account.yaml
- service.yaml
namePrefix: jupyter-web-app-
namespace: kubeflow
commonLabels:
  app: jupyter-web-app
  kustomize.component: jupyter-web-app
images:
  - name: gcr.io/kubeflow-images-public/jupyter-web-app
    newName: gcr.io/kubeflow-images-public/jupyter-web-app
    newTag: 9419d4d
configMapGenerator:
- name: parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: policy
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.policy
- name: prefix
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.prefix
- name: clusterDomain
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.clusterDomain
- name: namespace
  objref:
    kind: Service
    name: service
    apiVersion: v1
  fieldref:
    fieldpath: metadata.namespace
- name: userid-header
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.userid-header
- name: userid-prefix
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.userid-prefix
configurations:
- params.yaml
`)
}

func TestJupyterWebAppBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/jupyter/jupyter-web-app/base")
	writeJupyterWebAppBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../jupyter/jupyter-web-app/base"
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
