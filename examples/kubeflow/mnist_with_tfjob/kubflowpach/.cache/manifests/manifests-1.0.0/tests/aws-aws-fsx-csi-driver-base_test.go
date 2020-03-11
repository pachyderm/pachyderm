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

func writeAwsFsxCsiDriverBase(th *KustTestHarness) {
	th.writeF("/manifests/aws/aws-fsx-csi-driver/base/csi-controller-stateful-set.yaml", `
kind: StatefulSet
apiVersion: apps/v1
metadata:
  name: fsx-csi-controller
spec:
  serviceName: fsx-csi-controller
  replicas: 1
  selector:
    matchLabels:
      app: fsx-csi-controller
  template:
    metadata:
      labels:
        app: fsx-csi-controller
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      serviceAccount: fsx-csi-controller-sa
#      priorityClassName: system-cluster-critical
      tolerations:
        - key: CriticalAddonsOnly
          operator: Exists
      containers:
        - name: fsx-plugin
          image: amazon/aws-fsx-csi-driver:latest
          args :
            - --endpoint=$(CSI_ENDPOINT)
            - --logtostderr
            - --v=5
          env:
            - name: CSI_ENDPOINT
              value: unix:///var/lib/csi/sockets/pluginproxy/csi.sock
          volumeMounts:
            - name: socket-dir
              mountPath: /var/lib/csi/sockets/pluginproxy/
        - name: csi-provisioner
          image: quay.io/k8scsi/csi-provisioner:v0.4.2
          args:
            - --provisioner=fsx.csi.aws.com
            - --csi-address=$(ADDRESS)
            - --connection-timeout=5m
            - --v=5
          env:
            - name: ADDRESS
              value: /var/lib/csi/sockets/pluginproxy/csi.sock
          volumeMounts:
            - name: socket-dir
              mountPath: /var/lib/csi/sockets/pluginproxy/
        - name: csi-attacher
          image: quay.io/k8scsi/csi-attacher:v0.4.2
          args:
            - --csi-address=$(ADDRESS)
            - --v=5
          env:
            - name: ADDRESS
              value: /var/lib/csi/sockets/pluginproxy/csi.sock
          volumeMounts:
            - name: socket-dir
              mountPath: /var/lib/csi/sockets/pluginproxy/
      volumes:
        - name: socket-dir
          emptyDir: {}
`)
	th.writeF("/manifests/aws/aws-fsx-csi-driver/base/csi-attacher-cluster-role.yaml", `
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: fsx-csi-external-attacher-clusterrole
rules:
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments"]
    verbs: ["get", "list", "watch", "update"]`)
	th.writeF("/manifests/aws/aws-fsx-csi-driver/base/csi-attacher-cluster-role-binding.yaml", `
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: fsx-csi-external-attacher-clusterrole-binding
subjects:
  - kind: ServiceAccount
    name: fsx-csi-controller-sa
    namespace: kubeflow
roleRef:
  kind: ClusterRole
  name: fsx-csi-external-attacher-clusterrole
  apiGroup: rbac.authorization.k8s.io`)
	th.writeF("/manifests/aws/aws-fsx-csi-driver/base/csi-controller-cluster-role.yaml", `
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: external-provisioner-role
rules:
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "delete"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]`)
	th.writeF("/manifests/aws/aws-fsx-csi-driver/base/csi-controller-cluster-role-binding.yaml", `
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: csi-provisioner-binding
subjects:
  - kind: ServiceAccount
    name: csi-controller-sa
    namespace: kubeflow
roleRef:
  kind: ClusterRole
  name: external-provisioner-role
  apiGroup: rbac.authorization.k8s.io`)
	th.writeF("/manifests/aws/aws-fsx-csi-driver/base/csi-controller-sa.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: fsx-csi-controller-sa
`)
	th.writeF("/manifests/aws/aws-fsx-csi-driver/base/csi-node-cluster-role.yaml", `
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: fsx-csi-node-clusterrole
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "update"]
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["csi.storage.k8s.io"]
    resources: ["csinodeinfos"]
    verbs: ["get", "list", "watch", "update"]
`)
	th.writeF("/manifests/aws/aws-fsx-csi-driver/base/csi-node-cluster-role-binding.yaml", `
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: fsx-csi-node-clusterrole-binding
subjects:
  - kind: ServiceAccount
    name: fsx-csi-node-sa
    namespace: kubeflow
roleRef:
  kind: ClusterRole
  name: fsx-csi-node-clusterrole
  apiGroup: rbac.authorization.k8s.io`)
	th.writeF("/manifests/aws/aws-fsx-csi-driver/base/csi-node-daemonset.yaml", `
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: fsx-csi-node-ds
spec:
  selector:
    matchLabels:
      app: fsx-csi-node
  template:
    metadata:
      labels:
        app: fsx-csi-node
    spec:
      serviceAccount: fsx-csi-node-sa
      hostNetwork: true
      containers:
        - name: fsx-plugin
          securityContext:
            privileged: true
          image: amazon/aws-fsx-csi-driver:latest
          args:
            - --endpoint=$(CSI_ENDPOINT)
            - --logtostderr
            - --v=5
          env:
            - name: CSI_ENDPOINT
              value: unix:/csi/csi.sock
          volumeMounts:
            - name: kubelet-dir
              mountPath: /var/lib/kubelet
              mountPropagation: "Bidirectional"
            - name: plugin-dir
              mountPath: /csi
            - name: device-dir
              mountPath: /dev
        - name: csi-driver-registrar
          image: quay.io/k8scsi/driver-registrar:v0.4.2
          args:
            - --csi-address=$(ADDRESS)
            - --mode=node-register
            - --driver-requires-attachment=true
            - --pod-info-mount-version="v1"
            - --kubelet-registration-path=$(DRIVER_REG_SOCK_PATH)
            - --v=5
          env:
            - name: ADDRESS
              value: /csi/csi.sock
            - name: DRIVER_REG_SOCK_PATH
              value: /var/lib/kubelet/plugins/fsx.csi.aws.com/csi.sock
            - name: KUBE_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
          volumeMounts:
            - name: plugin-dir
              mountPath: /csi
            - name: registration-dir
              mountPath: /registration
      volumes:
        - name: kubelet-dir
          hostPath:
            path: /var/lib/kubelet
            type: Directory
        - name: plugin-dir
          hostPath:
            path: /var/lib/kubelet/plugins/fsx.csi.aws.com/
            type: DirectoryOrCreate
        - name: registration-dir
          hostPath:
            path: /var/lib/kubelet/plugins/
            type: Directory
        - name: device-dir
          hostPath:
            path: /dev
            type: Directory
`)
	th.writeF("/manifests/aws/aws-fsx-csi-driver/base/csi-node-sa.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: fsx-csi-node-sa
`)
	th.writeF("/manifests/aws/aws-fsx-csi-driver/base/csi-provisioner-cluster-role.yaml", `
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: fsx-external-provisioner-clusterrole
rules:
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "delete"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]`)
	th.writeF("/manifests/aws/aws-fsx-csi-driver/base/csi-provisioner-cluster-role-binding.yaml", `
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: fsx-csi-provisioner-clusterrole-binding
subjects:
  - kind: ServiceAccount
    name: fsx-csi-controller-sa
    namespace: kubeflow
roleRef:
  kind: ClusterRole
  name: fsx-external-provisioner-clusterrole
  apiGroup: rbac.authorization.k8s.io`)
	th.writeF("/manifests/aws/aws-fsx-csi-driver/base/csi-default-storage.yaml", `
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fsx-default
provisioner: fsx.csi.aws.com`)
	th.writeK("/manifests/aws/aws-fsx-csi-driver/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kubeflow
resources:
- csi-controller-stateful-set.yaml
- csi-attacher-cluster-role.yaml
- csi-attacher-cluster-role-binding.yaml
- csi-controller-cluster-role.yaml
- csi-controller-cluster-role-binding.yaml
- csi-controller-sa.yaml
- csi-node-cluster-role.yaml
- csi-node-cluster-role-binding.yaml
- csi-node-daemonset.yaml
- csi-node-sa.yaml
- csi-provisioner-cluster-role.yaml
- csi-provisioner-cluster-role-binding.yaml
- csi-default-storage.yaml
generatorOptions:
  disableNameSuffixHash: true
images:
- name: amazon/aws-fsx-csi-driver
  newName: amazon/aws-fsx-csi-driver
  newTag: latest
- name: quay.io/k8scsi/driver-registrar
  newName: quay.io/k8scsi/driver-registrar
  newTag: v0.4.2
- name: quay.io/k8scsi/csi-provisioner
  newName: quay.io/k8scsi/csi-provisioner
  newTag: v0.4.2
- name: quay.io/k8scsi/csi-attacher
  newName: quay.io/k8scsi/csi-attacher
  newTag: v0.4.2
`)
}

func TestAwsFsxCsiDriverBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/aws/aws-fsx-csi-driver/base")
	writeAwsFsxCsiDriverBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../aws/aws-fsx-csi-driver/base"
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
