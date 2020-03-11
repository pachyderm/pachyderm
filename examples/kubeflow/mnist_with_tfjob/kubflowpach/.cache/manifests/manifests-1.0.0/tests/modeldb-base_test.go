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

func writeModeldbBase(th *KustTestHarness) {
	th.writeF("/manifests/modeldb/base/artifact-store-deployment.yaml", `

apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: modeldb
  name: modeldb-artifact-store
spec:
  selector:
    matchLabels:
      app: modeldb
      tier: artifact-store
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: modeldb
        tier: artifact-store
    spec:
      containers:
      - env:
        - name: VERTA_ARTIFACT_CONFIG
          value: /config/config.yaml
        image: vertaaiofficial/modeldb-artifact-store:kubeflow
        imagePullPolicy: Always
        name: modeldb-artifact-store
        ports:
        - containerPort: 8086
        volumeMounts:
        - mountPath: /config
          name: modeldb-artifact-store-config
          readOnly: true
      volumes:
      - configMap:
          name: modeldb-artifact-store-config
        name: modeldb-artifact-store-config
`)
	th.writeF("/manifests/modeldb/base/artifact-store-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: modeldb
  name: modeldb-artifact-store
spec:
  ports:
  - port: 8086
    targetPort: 8086
  selector:
    app: modeldb
    tier: artifact-store
  type: ClusterIP
`)
	th.writeF("/manifests/modeldb/base/backend-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: modeldb
  name: modeldb-backend
spec:
  selector:
    matchLabels:
      app: modeldb
      tier: backend
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: modeldb
        tier: backend
    spec:
      containers:
      - env:
        - name: VERTA_MODELDB_CONFIG
          value: /config-backend/config.yaml
        image: vertaaiofficial/modeldb-backend:kubeflow
        imagePullPolicy: Always
        name: modeldb-backend
        ports:
        - containerPort: 8085
        volumeMounts:
        - mountPath: /config-backend
          name: modeldb-backend-secret-volume
          readOnly: true
      volumes:
      - name: modeldb-backend-secret-volume
        secret:
          secretName: modeldb-backend-config-secret
`)
	th.writeF("/manifests/modeldb/base/backend-proxy-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: modeldb
  name: modeldb-backend-proxy
spec:
  ports:
  - port: 8080
    targetPort: 8080
  selector:
    app: modeldb
    tier: backend-proxy
  type: LoadBalancer
`)
	th.writeF("/manifests/modeldb/base/backend-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: modeldb
  name: modeldb-backend
spec:
  ports:
  - port: 8085
  selector:
    app: modeldb
    tier: backend
  type: LoadBalancer
`)
	th.writeF("/manifests/modeldb/base/configmap.yaml", `
apiVersion: v1
data:
  config.yaml: |-
    #ArtifactStore Properties
    artifactStore_grpcServer:
      port: 8086

    artifactStoreConfig:
      initializeBuckets: false
      storageTypeName: amazonS3 #amazonS3, googleCloudStorage, nfs
      #nfsRootPath: /path/to/my/nfs/storage/location
      bucket_names:
        - artifactstoredemo
kind: ConfigMap
metadata:
  name: modeldb-artifact-store-config
type: Opaque
`)
	th.writeF("/manifests/modeldb/base/mysql-backend-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: modeldb
  name: modeldb-mysql-backend
spec:
  selector:
    matchLabels:
      app: modeldb
      tier: mysql
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: modeldb
        tier: mysql
    spec:
      containers:
      - args:
        - --ignore-db-dir=lost+found
        env:
        - name: MYSQL_ROOT_PASSWORD
          value: root
        image: mysql:5.7
        imagePullPolicy: Always
        name: modeldb-mysql-backend
        ports:
        - containerPort: 3306
        volumeMounts:
        - mountPath: /var/lib/mysql
          name: modeldb-mysql-persistent-storage
      volumes:
      - name: modeldb-mysql-persistent-storage
        persistentVolumeClaim:
          claimName: modeldb-mysql-pv-claim
`)
	th.writeF("/manifests/modeldb/base/mysql-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: modeldb
  name: modeldb-mysql-backend
spec:
  ports:
  - port: 3306
    targetPort: 3306
  selector:
    app: modeldb
    tier: mysql
  type: ClusterIP
`)
	th.writeF("/manifests/modeldb/base/persistent-volume-claim.yaml", `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  labels:
    app: modeldb
  name: modeldb-mysql-pv-claim
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 20Gi
`)
	th.writeF("/manifests/modeldb/base/proxy-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: modeldb
  name: modeldb-backend-proxy
spec:
  selector:
    matchLabels:
      app: modeldb
      tier: backend-proxy
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: modeldb
        tier: backend-proxy
    spec:
      containers:
      - args:
        - -project_endpoint
        - modeldb-backend:8085
        - -experiment_endpoint
        - modeldb-backend:8085
        - -experiment_run_endpoint
        - modeldb-backend:8085
        command:
        - /go/bin/proxy
        image: vertaaiofficial/modeldb-backend-proxy:kubeflow
        imagePullPolicy: Always
        name: modeldb-backend-proxy
        ports:
        - containerPort: 8080
`)
	th.writeF("/manifests/modeldb/base/secret.yaml", `
apiVersion: v1
kind: Secret
metadata:
  name: modeldb-backend-config-secret
stringData:
  config.yaml: |-
    #ModelDB Properties
    grpcServer:
      port: 8085

    #Entity name list
    entities:
      projectEntity: Project
      experimentEntity: Experiment
      experimentRunEntity: ExperimentRun
      artifactStoreMappingEntity: ArtifactStoreMapping
      jobEntity: Job
      collaboratorEntity: Collaborator

    # Database settings (type mysql, mongodb, couchbasedb etc..)
    database:
      DBType: rdbms
      RdbConfiguration:
        RdbDatabaseName: modeldb
        RdbDriver: "com.mysql.cj.jdbc.Driver"
        RdbDialect: "org.hibernate.dialect.MySQL5Dialect"
        RdbUrl: "jdbc:mysql://modeldb-mysql-backend:3306"
        RdbUsername: root
        RdbPassword: root

    #ArtifactStore Properties
    artifactStore_grpcServer:
      host: artifact-store-backend
      port: 8086

    #AuthService Properties
    authService:
      host: #uacservice # Docker container name OR docker IP
      port: #50051
type: Opaque
`)
	th.writeF("/manifests/modeldb/base/webapp-deplyment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: modeldb
  name: modeldb-webapp
spec:
  selector:
    matchLabels:
      app: modeldb
      tier: webapp
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: modeldb
        tier: webapp
    spec:
      containers:
      - image: vertaaiofficial/modeldb-frontend:kubeflow
        imagePullPolicy: Always
        name: modeldb-webapp
        ports:
        - containerPort: 3000
`)
	th.writeF("/manifests/modeldb/base/webapp-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: modeldb
  name: modeldb-webapp
spec:
  ports:
  - port: 80
    targetPort: 3000
  selector:
    app: modeldb
    tier: webapp
  type: LoadBalancer
`)
	th.writeK("/manifests/modeldb/base", `
namePrefix: modeldb-

resources:
- artifact-store-deployment.yaml
- artifact-store-service.yaml
- backend-deployment.yaml
- backend-proxy-service.yaml
- backend-service.yaml
- configmap.yaml
- mysql-backend-deployment.yaml
- mysql-service.yaml
- persistent-volume-claim.yaml
- proxy-deployment.yaml
- secret.yaml
- webapp-deplyment.yaml
- webapp-service.yaml

commonLabels:
  kustomize.component: modeldb
images:
- name: vertaaiofficial/modeldb-frontend
  newName: vertaaiofficial/modeldb-frontend
  newTag: kubeflow
- name: vertaaiofficial/modeldb-backend
  newName: vertaaiofficial/modeldb-backend
  newTag: kubeflow
- name: vertaaiofficial/modeldb-artifact-store
  newName: vertaaiofficial/modeldb-artifact-store
  newTag: kubeflow
- name: mysql
  newName: mysql
  newTag: '5.7'
- name: vertaaiofficial/modeldb-backend-proxy
  newName: vertaaiofficial/modeldb-backend-proxy
  newTag: kubeflow
`)
}

func TestModeldbBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/modeldb/base")
	writeModeldbBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../modeldb/base"
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
