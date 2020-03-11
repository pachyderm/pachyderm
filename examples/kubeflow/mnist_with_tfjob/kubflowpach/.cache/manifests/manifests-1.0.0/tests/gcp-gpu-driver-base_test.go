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

func writeGpuDriverBase(th *KustTestHarness) {
	th.writeF("/manifests/gcp/gpu-driver/base/daemon-set.yaml", `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    k8s-app: nvidia-driver-installer
  name: nvidia-driver-installer
  namespace: kube-system
spec:
  template:
    metadata:
      labels:
        k8s-app: nvidia-driver-installer
        name: nvidia-driver-installer
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: cloud.google.com/gke-accelerator
                operator: Exists
      containers:
      - image: gcr.io/google-containers/pause:2.0
        name: pause
      hostNetwork: true
      hostPID: true
      initContainers:
      - env:
        - name: NVIDIA_INSTALL_DIR_HOST
          value: /home/kubernetes/bin/nvidia
        - name: NVIDIA_INSTALL_DIR_CONTAINER
          value: /usr/local/nvidia
        - name: ROOT_MOUNT_DIR
          value: /root
        image: cos-nvidia-installer:fixed
        imagePullPolicy: Never
        name: nvidia-driver-installer
        resources:
          requests:
            cpu: 0.15
        securityContext:
          privileged: true
        volumeMounts:
        - mountPath: /usr/local/nvidia
          name: nvidia-install-dir-host
        - mountPath: /dev
          name: dev
        - mountPath: /root
          name: root-mount
      tolerations:
      - operator: Exists
      volumes:
      - hostPath:
          path: /dev
        name: dev
      - hostPath:
          path: /home/kubernetes/bin/nvidia
        name: nvidia-install-dir-host
      - hostPath:
          path: /
        name: root-mount
`)
	th.writeK("/manifests/gcp/gpu-driver/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- daemon-set.yaml
commonLabels:
  kustomize.component: gpu-driver
images:
- name: gcr.io/google-containers/pause
  newName: gcr.io/google-containers/pause
  newTag: '2.0'
- name: cos-nvidia-installer
  newName: cos-nvidia-installer
  newTag: fixed
`)
}

func TestGpuDriverBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/gcp/gpu-driver/base")
	writeGpuDriverBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../gcp/gpu-driver/base"
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
