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

func writeDexLdapBase(th *KustTestHarness) {
	th.writeF("/manifests/dex-auth/dex-ldap/base/namespace.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  name: auth
`)
	th.writeF("/manifests/dex-auth/dex-ldap/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ldap
  labels:
    app: ldap
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ldap
  template:
    metadata:
      labels:
        app: ldap
    spec:
      containers:
      - name: openldap
        image: osixia/openldap
        ports:
          - containerPort: 389
          - containerPort: 636
      - name: phpldapadmin
        image: osixia/phpldapadmin
        ports:
          - containerPort: 80
        env:
        - name: PHPLDAPADMIN_HTTPS
          value: "false"
        - name: PHPLDAPADMIN_LDAP_HOSTS
          value: localhost
`)
	th.writeF("/manifests/dex-auth/dex-ldap/base/service.yaml", `
---

apiVersion: v1
kind: Service
metadata:
  name: ldap
spec:
  ports:
  - name: ldap
    port: 389
    targetPort: 389
  - name: ldap-ssl
    port: 636
    targetPort: 636
  selector:
    app: ldap

---

apiVersion: v1
kind: Service
metadata:
  name: ldap-admin
spec:
  type: NodePort
  ports:
  - port: 80
    targetPort: 80
    nodePort: 32006
  selector:
    app: ldap
`)
	th.writeK("/manifests/dex-auth/dex-ldap/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: auth

resources:
- namespace.yaml
- deployment.yaml
- service.yaml
images:
- name: osixia/openldap
  newName: osixia/openldap
  newTag: latest
- name: osixia/phpldapadmin
  newName: osixia/phpldapadmin
  newTag: latest
`)
}

func TestDexLdapBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/dex-auth/dex-ldap/base")
	writeDexLdapBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../dex-auth/dex-ldap/base"
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
