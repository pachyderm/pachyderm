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

func writeNvidiaDevicePluginOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/aws/nvidia-device-plugin/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: nvidia-device-plugin
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: nvidia-device-plugin
      app.kubernetes.io/instance: nvidia-device-plugin-1.0.0-beta
      app.kubernetes.io/managed-by: kfctl
      app.kubernetes.io/component: nvidia-device-plugin
      app.kubernetes.io/part-of: kubeflow
      app.kubernetes.io/version: v1.0.0-beta
  componentKinds:
  - group: core
    kind: ConfigMap
  - group: apps
    kind: DaemonSet
  descriptor:
    type: nvidia-device-plugin
    version: v1beta1
    description: Nvidia Device Plugin for aws
    maintainers: []
    owners: []
    keywords:
     - aws
     - nvidia
     - kubeflow
    links:
    - description: About
      url: https://github.com/kubernetes/kops/tree/master/hooks/nvidia-device-plugin
  addOwnerRef: true

`)
	th.writeK("/manifests/aws/nvidia-device-plugin/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
commonLabels:
  app.kubernetes.io/name: nvidia-device-plugin
  app.kubernetes.io/instance: nvidia-device-plugin-1.0.0-beta
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: nvidia-device-plugin
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v1.0.0-beta
`)
	th.writeF("/manifests/aws/nvidia-device-plugin/base/daemonset.yaml", `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: nvidia-device-plugin-daemonset
spec:
  updateStrategy:
    type: RollingUpdate
  template:
    metadata:
      # Mark this pod as a critical add-on; when enabled, the critical add-on scheduler
      # reserves resources for critical add-on pods so that they can be rescheduled after
      # a failure.  This annotation works in tandem with the toleration below.
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      tolerations:
        # Allow this pod to be rescheduled while the node is in "critical add-ons only" mode.
        # This, along with the annotation above marks this pod as a critical add-on.
        - key: CriticalAddonsOnly
          operator: Exists
        - key: nvidia.com/gpu
          operator: Exists
          effect: NoSchedule
      containers:
        - image: nvidia/k8s-device-plugin:1.0.0-beta
          name: nvidia-device-plugin-ctr
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop: ["ALL"]
          volumeMounts:
            - name: device-plugin
              mountPath: /var/lib/kubelet/device-plugins
      volumes:
        - name: device-plugin
          hostPath:
            path: /var/lib/kubelet/device-plugins
`)
	th.writeK("/manifests/aws/nvidia-device-plugin/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kube-system
resources:
- daemonset.yaml
commonLabels:
  kustomize.component: nvidia-device-plugin
images:
- name: nvidia/k8s-device-plugin
  newName: nvidia/k8s-device-plugin
  newTag: 1.0.0-beta
`)
}

func TestNvidiaDevicePluginOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/aws/nvidia-device-plugin/overlays/application")
	writeNvidiaDevicePluginOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../aws/nvidia-device-plugin/overlays/application"
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
