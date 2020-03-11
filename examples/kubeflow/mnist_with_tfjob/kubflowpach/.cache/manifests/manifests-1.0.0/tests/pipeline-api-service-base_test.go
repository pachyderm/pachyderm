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

func writeApiServiceBase(th *KustTestHarness) {
	th.writeF("/manifests/pipeline/api-service/base/config-map.yaml", `
# The configuration for the ML pipelines APIServer
# Based on https://github.com/kubeflow/pipelines/blob/master/backend/src/apiserver/config/config.json
apiVersion: v1
data:
  # apiserver assumes the config is named config.json
  config.json: |
    {
      "DBConfig": {
        "DriverName": "mysql",
        "DataSourceName": "",
        "DBName": "mlpipeline"
      },
      "ObjectStoreConfig":{
        "AccessKey": "minio",
        "SecretAccessKey": "minio123",
        "BucketName": "mlpipeline"
      },
      "InitConnectionTimeout": "6m",
      "DefaultPipelineRunnerServiceAccount": "pipeline-runner",
      "ML_PIPELINE_VISUALIZATIONSERVER_SERVICE_HOST": "ml-pipeline-ml-pipeline-visualizationserver",
      "ML_PIPELINE_VISUALIZATIONSERVER_SERVICE_PORT": 8888
    }
kind: ConfigMap
metadata:
  name: ml-pipeline-config
`)
	th.writeF("/manifests/pipeline/api-service/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ml-pipeline
spec:
  template:
    spec:
      containers:
      - name: ml-pipeline-api-server
        env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: gcr.io/ml-pipeline/api-server
        imagePullPolicy: IfNotPresent
        command:
          - apiserver 
          - --config=/etc/ml-pipeline-config
          - --sampleconfig=/config/sample_config.json 
          - -logtostderr=true
        ports:
        - containerPort: 8888
        - containerPort: 8887
        volumeMounts:
        - name: config-volume
          mountPath: /etc/ml-pipeline-config
      serviceAccountName: ml-pipeline      
      volumes:
        - name: config-volume
          configMap:
            name: ml-pipeline-config
`)
	th.writeF("/manifests/pipeline/api-service/base/role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: ml-pipeline
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: ml-pipeline
subjects:
- kind: ServiceAccount
  name: ml-pipeline
`)
	th.writeF("/manifests/pipeline/api-service/base/role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
# TODO: Does this need to be changed to a clusterrole?
# see  manifests in kubeflow/pipelines
kind: Role
metadata:
  name: ml-pipeline
rules:
- apiGroups:
  - argoproj.io
  resources:
  - workflows
  verbs:
  - create
  - get
  - list
  - watch
  - update
  - patch
  - delete
- apiGroups:
  - kubeflow.org
  resources:
  - scheduledworkflows
  verbs:
  - create
  - get
  - list
  - update
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - delete

`)
	th.writeF("/manifests/pipeline/api-service/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ml-pipeline
`)
	th.writeF("/manifests/pipeline/api-service/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: ml-pipeline
spec:
  ports:
  - name: http
    port: 8888
    protocol: TCP
    targetPort: 8888
  - name: grpc
    port: 8887
    protocol: TCP
    targetPort: 8887
`)
	th.writeK("/manifests/pipeline/api-service/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
commonLabels:
  app: ml-pipeline
resources:
- config-map.yaml
- deployment.yaml
- role-binding.yaml
- role.yaml
- service-account.yaml
- service.yaml
images:
- name: gcr.io/ml-pipeline/api-server
  newTag: 0.2.0
  newName: gcr.io/ml-pipeline/api-server
`)
}

func TestApiServiceBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pipeline/api-service/base")
	writeApiServiceBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pipeline/api-service/base"
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
