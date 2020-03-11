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

func writeKatibDbBase(th *KustTestHarness) {
	th.writeF("/manifests/katib-v1alpha2/katib-db/base/katib-db-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: katib-db
  labels:
    app: katib
    component: db
spec:
  replicas: 1
  template:
    metadata:
      name: katib-db
      labels:
        app: katib
        component: db
    spec:
      containers:
      - name: katib-db
        image: mysql:8.0.3
        args:
        - --datadir
        - /var/lib/mysql/datadir
        env:
          - name: MYSQL_ROOT_PASSWORD
            valueFrom:
              secretKeyRef:
                name: katib-db-secrets
                key: MYSQL_ROOT_PASSWORD
          - name: MYSQL_ALLOW_EMPTY_PASSWORD
            value: "true"
          - name: MYSQL_DATABASE
            value: "katib"
        ports:
        - name: dbapi
          containerPort: 3306
        readinessProbe:
          exec:
            command:
            - "/bin/bash"
            - "-c"
            - "mysql -D $$MYSQL_DATABASE -p$$MYSQL_ROOT_PASSWORD -e 'SELECT 1'"
          initialDelaySeconds: 5
          periodSeconds: 2
          timeoutSeconds: 1
        volumeMounts:
        - name: katib-mysql
          mountPath: /var/lib/mysql
      volumes:
      - name: katib-mysql
        persistentVolumeClaim:
          claimName: katib-mysql
`)
	th.writeF("/manifests/katib-v1alpha2/katib-db/base/katib-db-pvc.yaml", `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: katib-mysql
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
`)
	th.writeF("/manifests/katib-v1alpha2/katib-db/base/katib-db-secret.yaml", `
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: katib-db-secrets
data:
  MYSQL_ROOT_PASSWORD: dGVzdA== # "test"
`)
	th.writeF("/manifests/katib-v1alpha2/katib-db/base/katib-db-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-db
  labels:
    app: katib
    component: db
spec:
  type: ClusterIP
  ports:
    - port: 3306
      protocol: TCP
      name: dbapi
  selector:
    app: katib
    component: db
`)
	th.writeK("/manifests/katib-v1alpha2/katib-db/base", `
namespace: kubeflow
resources:
- katib-db-deployment.yaml
- katib-db-pvc.yaml
- katib-db-secret.yaml
- katib-db-service.yaml
generatorOptions:
  disableNameSuffixHash: true
images:
  - name: mysql
    newTag: 8.0.3
`)
}

func TestKatibDbBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib-v1alpha2/katib-db/base")
	writeKatibDbBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib-v1alpha2/katib-db/base"
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
