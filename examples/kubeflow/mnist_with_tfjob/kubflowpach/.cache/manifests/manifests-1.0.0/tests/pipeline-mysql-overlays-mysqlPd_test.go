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

func writeMysqlOverlaysMysqlPd(th *KustTestHarness) {
	th.writeF("/manifests/pipeline/mysql/overlays/mysqlPd/persistent-volume.yaml", `
apiVersion: v1
kind: PersistentVolume
metadata:
  name: $(mysqlPvName)
spec:
  capacity:
    storage: 20Gi
  accessModes:
  - ReadWriteOnce
  gcePersistentDisk:
    pdName: $(mysqlPd)
    fsType: ext4
`)
	th.writeF("/manifests/pipeline/mysql/overlays/mysqlPd/persistent-volume-claim.yaml", `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: $(mysqlPvcName)
spec:
  storageClassName: ""
  volumeName: $(mysqlPvName)
`)
	th.writeF("/manifests/pipeline/mysql/overlays/mysqlPd/params.yaml", `
varReference:
- path: spec/gcePersistentDisk/pdName
  kind: PersistentVolume
- path: metadata/name
  kind: PersistentVolume
- path: metadata/name
  kind: PersistentVolumeClaim
- path: spec/volumeName
  kind: PersistentVolumeClaim
`)
	th.writeF("/manifests/pipeline/mysql/overlays/mysqlPd/params.env", `
mysqlPd=dls-kf-storage-metadata-store
mysqlPvName=
`)
	th.writeK("/manifests/pipeline/mysql/overlays/mysqlPd", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- persistent-volume.yaml
patchesStrategicMerge:
- persistent-volume-claim.yaml
configMapGenerator:
- name: pipeline-mysql-parameters
  behavior: merge
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: mysqlPd
  objref:
    kind: ConfigMap
    name: pipeline-mysql-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.mysqlPd
- name: mysqlPvName
  objref:
    kind: ConfigMap
    name: pipeline-mysql-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.mysqlPvName
configurations:
- params.yaml
`)
	th.writeF("/manifests/pipeline/mysql/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql
spec:
  strategy:
    type: Recreate
  template:
    spec:
      containers:
      - name: mysql
        env:
        - name: MYSQL_ALLOW_EMPTY_PASSWORD
          value: "true"
        image: mysql:5.6
        ports:
        - containerPort: 3306
          name: mysql
        volumeMounts:
        - mountPath: /var/lib/mysql
          name: mysql-persistent-storage
      volumes:
      - name: mysql-persistent-storage
        persistentVolumeClaim:
          claimName: $(mysqlPvcName)
`)
	th.writeF("/manifests/pipeline/mysql/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: mysql
spec:
  ports:
  - port: 3306
`)
	th.writeF("/manifests/pipeline/mysql/base/persistent-volume-claim.yaml", `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: $(mysqlPvcName)
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 20Gi`)
	th.writeF("/manifests/pipeline/mysql/base/params.yaml", `
varReference:
- path: spec/template/spec/volumes/persistentVolumeClaim/claimName
  kind: Deployment
- path: metadata/name
  kind: PersistentVolumeClaim`)
	th.writeF("/manifests/pipeline/mysql/base/params.env", `
mysqlPvcName=
`)
	th.writeK("/manifests/pipeline/mysql/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
commonLabels:
  app: mysql
resources:
- deployment.yaml
- service.yaml
- persistent-volume-claim.yaml
configMapGenerator:
- name: pipeline-mysql-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: mysqlPvcName
  objref:
    kind: ConfigMap
    name: pipeline-mysql-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.mysqlPvcName
images:
- name: mysql
  newTag: '5.6'
  newName: mysql
configurations:
- params.yaml
`)
}

func TestMysqlOverlaysMysqlPd(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pipeline/mysql/overlays/mysqlPd")
	writeMysqlOverlaysMysqlPd(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pipeline/mysql/overlays/mysqlPd"
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
