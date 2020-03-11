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

func writeDefaultInstallBase(th *KustTestHarness) {
	th.writeF("/manifests/default-install/base/profile-instance.yaml", `
apiVersion: kubeflow.org/v1beta1
kind: Profile
metadata:
  name: $(profile-name)
spec:
  owner:
    kind: User
    name: $(user)
`)
	th.writeF("/manifests/default-install/base/params.yaml", `
varReference:
- path: spec/owner/name
  kind: Profile
- path: metadata/name
  kind: Profile
`)
	th.writeF("/manifests/default-install/base/params.env", `
user=anonymous
profile-name=anonymous
`)
	th.writeK("/manifests/default-install/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- profile-instance.yaml
configMapGenerator:
- name: default-install-parameters
  env: params.env
vars:
- name: user
  objref:
    kind: ConfigMap
    name: default-install-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.user
- name: profile-name
  objref:
    kind: ConfigMap
    name: default-install-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.profile-name
configurations:
- params.yaml
`)
}

func TestDefaultInstallBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/default-install/base")
	writeDefaultInstallBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../default-install/base"
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
