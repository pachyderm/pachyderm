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

func writeFluentdCloudWatchBase(th *KustTestHarness) {
	th.writeF("/manifests/aws/fluentd-cloud-watch/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: fluentd
rules:
  - apiGroups: [""]
    resources:
      - namespaces
      - pods
    verbs: ["get", "list", "watch"]`)
	th.writeF("/manifests/aws/fluentd-cloud-watch/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: fluentd
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: fluentd
subjects:
  - kind: ServiceAccount
    name: fluentd
    namespace: kube-system`)
	th.writeF("/manifests/aws/fluentd-cloud-watch/base/configmap.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
  labels:
    k8s-app: fluentd-cloudwatch
data:
  fluent.conf: |
    @include containers.conf
    @include systemd.conf
    <match fluent.**>
      @type null
    </match>
  containers.conf: |
    <source>
      @type tail
      @id in_tail_container_logs
      @label @containers
      path /var/log/containers/*.log
      pos_file /var/log/fluentd-containers.log.pos
      tag *
      read_from_head true
      <parse>
        @type json
        time_format %Y-%m-%dT%H:%M:%S.%NZ
      </parse>
    </source>
    <label @containers>
      <filter **>
        @type kubernetes_metadata
        @id filter_kube_metadata
      </filter>
      <filter **>
        @type record_transformer
        @id filter_containers_stream_transformer
        <record>
          stream_name ${tag_parts[3]}
        </record>
      </filter>
      <match **>
        @type cloudwatch_logs
        @id out_cloudwatch_logs_containers
        region "#{ENV.fetch('REGION')}"
        log_group_name "/eks/#{ENV.fetch('CLUSTER_NAME')}/containers"
        log_stream_name_key stream_name
        remove_log_stream_name_key true
        auto_create_stream true
        <buffer>
          flush_interval 5
          chunk_limit_size 2m
          queued_chunks_limit_size 32
          retry_forever true
        </buffer>
      </match>
    </label>
  systemd.conf: |
    <source>
      @type systemd
      @id in_systemd_kubelet
      @label @systemd
      filters [{ "_SYSTEMD_UNIT": "kubelet.service" }]
      <entry>
        field_map {"MESSAGE": "message", "_HOSTNAME": "hostname", "_SYSTEMD_UNIT": "systemd_unit"}
        field_map_strict true
      </entry>
      path /run/log/journal
      pos_file /var/log/fluentd-journald-kubelet.pos
      read_from_head true
      tag kubelet.service
    </source>
    <source>
      @type systemd
      @id in_systemd_kubeproxy
      @label @systemd
      filters [{ "_SYSTEMD_UNIT": "kubeproxy.service" }]
      <entry>
        field_map {"MESSAGE": "message", "_HOSTNAME": "hostname", "_SYSTEMD_UNIT": "systemd_unit"}
        field_map_strict true
      </entry>
      path /run/log/journal
      pos_file /var/log/fluentd-journald-kubeproxy.pos
      read_from_head true
      tag kubeproxy.service
    </source>
    <source>
      @type systemd
      @id in_systemd_docker
      @label @systemd
      filters [{ "_SYSTEMD_UNIT": "docker.service" }]
      <entry>
        field_map {"MESSAGE": "message", "_HOSTNAME": "hostname", "_SYSTEMD_UNIT": "systemd_unit"}
        field_map_strict true
      </entry>
      path /run/log/journal
      pos_file /var/log/fluentd-journald-docker.pos
      read_from_head true
      tag docker.service
    </source>
    <label @systemd>
      <filter **>
        @type record_transformer
        @id filter_systemd_stream_transformer
        <record>
          stream_name ${tag}-${record["hostname"]}
        </record>
      </filter>
      <match **>
        @type cloudwatch_logs
        @id out_cloudwatch_logs_systemd
        region "#{ENV.fetch('REGION')}"
        log_group_name "/eks/#{ENV.fetch('CLUSTER_NAME')}/systemd"
        log_stream_name_key stream_name
        auto_create_stream true
        remove_log_stream_name_key true
        <buffer>
          flush_interval 5
          chunk_limit_size 2m
          queued_chunks_limit_size 32
          retry_forever true
        </buffer>
      </match>
    </label>`)
	th.writeF("/manifests/aws/fluentd-cloud-watch/base/daemonset.yaml", `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd-cloudwatch
  labels:
    k8s-app: fluentd-cloudwatch
spec:
  template:
    metadata:
      labels:
        k8s-app: fluentd-cloudwatch
    spec:
      serviceAccountName: fluentd
      terminationGracePeriodSeconds: 30
      # Because the image's entrypoint requires to write on /fluentd/etc but we mount configmap there which is read-only,
      # this initContainers workaround or other is needed.
      # See https://github.com/fluent/fluentd-kubernetes-daemonset/issues/90
      initContainers:
        - name: copy-fluentd-config
          image: busybox
          command: ['sh', '-c', 'cp /config-volume/..data/* /fluentd/etc']
          volumeMounts:
            - name: config-volume
              mountPath: /config-volume
            - name: fluentdconf
              mountPath: /fluentd/etc
      containers:
        - name: fluentd-cloudwatch
          image: fluent/fluentd-kubernetes-daemonset:v1.1-debian-cloudwatch
          env:
            - name: REGION
              value: $(REGION)
            - name: CLUSTER_NAME
              value: $(CLUSTER_NAME)
          resources:
            limits:
              memory: 200Mi
            requests:
              cpu: 100m
              memory: 200Mi
          volumeMounts:
            - name: config-volume
              mountPath: /config-volume
            - name: fluentdconf
              mountPath: /fluentd/etc
            - name: varlog
              mountPath: /var/log
            - name: varlibdockercontainers
              mountPath: /var/lib/docker/containers
              readOnly: true
            - name: runlogjournal
              mountPath: /run/log/journal
              readOnly: true
      volumes:
        - name: config-volume
          configMap:
            name: fluentd-config
        - name: fluentdconf
          emptyDir: {}
        - name: varlog
          hostPath:
            path: /var/log
        - name: varlibdockercontainers
          hostPath:
            path: /var/lib/docker/containers
        - name: runlogjournal
          hostPath:
            path: /run/log/journal
`)
	th.writeF("/manifests/aws/fluentd-cloud-watch/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: fluentd`)
	th.writeF("/manifests/aws/fluentd-cloud-watch/base/params.env", `
region=us-west-2
clusterName=
`)
	th.writeK("/manifests/aws/fluentd-cloud-watch/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kube-system
resources:
- cluster-role.yaml
- cluster-role-binding.yaml
- configmap.yaml
- daemonset.yaml
- service-account.yaml
commonLabels:
  kustomize.component: fluentd-cloud-watch
generatorOptions:
  disableNameSuffixHash: true
images:
- name: fluent/fluentd-kubernetes-daemonset
  newName: fluent/fluentd-kubernetes-daemonset
  newTag: v1.1-debian-cloudwatch
configMapGenerator:
- name: fluentd-cloud-watch-parameters
  env: params.env
vars:
- name: CLUSTER_NAME
  objref:
    kind: ConfigMap
    name: fluentd-cloud-watch-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.clusterName
- name: REGION
  objref:
    kind: ConfigMap
    name: fluentd-cloud-watch-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.region
`)
}

func TestFluentdCloudWatchBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/aws/fluentd-cloud-watch/base")
	writeFluentdCloudWatchBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../aws/fluentd-cloud-watch/base"
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
