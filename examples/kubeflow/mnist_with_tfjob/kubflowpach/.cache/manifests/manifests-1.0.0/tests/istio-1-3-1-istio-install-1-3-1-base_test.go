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

func writeIstioInstall131Base(th *KustTestHarness) {
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/namespace.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  name: $(namespace)
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/attribute-manifest.yaml", `
apiVersion: config.istio.io/v1alpha2
kind: attributemanifest
metadata:
  labels:
    app: mixer
  name: istioproxy
spec:
  attributes:
    api.operation:
      valueType: STRING
    api.protocol:
      valueType: STRING
    api.service:
      valueType: STRING
    api.version:
      valueType: STRING
    check.cache_hit:
      valueType: BOOL
    check.error_code:
      valueType: INT64
    check.error_message:
      valueType: STRING
    connection.duration:
      valueType: DURATION
    connection.event:
      valueType: STRING
    connection.id:
      valueType: STRING
    connection.mtls:
      valueType: BOOL
    connection.received.bytes:
      valueType: INT64
    connection.received.bytes_total:
      valueType: INT64
    connection.requested_server_name:
      valueType: STRING
    connection.sent.bytes:
      valueType: INT64
    connection.sent.bytes_total:
      valueType: INT64
    context.protocol:
      valueType: STRING
    context.proxy_error_code:
      valueType: STRING
    context.proxy_version:
      valueType: STRING
    context.reporter.kind:
      valueType: STRING
    context.reporter.local:
      valueType: BOOL
    context.reporter.uid:
      valueType: STRING
    context.time:
      valueType: TIMESTAMP
    context.timestamp:
      valueType: TIMESTAMP
    destination.port:
      valueType: INT64
    destination.principal:
      valueType: STRING
    destination.uid:
      valueType: STRING
    origin.ip:
      valueType: IP_ADDRESS
    origin.uid:
      valueType: STRING
    origin.user:
      valueType: STRING
    quota.cache_hit:
      valueType: BOOL
    rbac.permissive.effective_policy_id:
      valueType: STRING
    rbac.permissive.response_code:
      valueType: STRING
    request.api_key:
      valueType: STRING
    request.auth.audiences:
      valueType: STRING
    request.auth.claims:
      valueType: STRING_MAP
    request.auth.presenter:
      valueType: STRING
    request.auth.principal:
      valueType: STRING
    request.auth.raw_claims:
      valueType: STRING
    request.headers:
      valueType: STRING_MAP
    request.host:
      valueType: STRING
    request.id:
      valueType: STRING
    request.method:
      valueType: STRING
    request.path:
      valueType: STRING
    request.query_params:
      valueType: STRING_MAP
    request.reason:
      valueType: STRING
    request.referer:
      valueType: STRING
    request.scheme:
      valueType: STRING
    request.size:
      valueType: INT64
    request.time:
      valueType: TIMESTAMP
    request.total_size:
      valueType: INT64
    request.url_path:
      valueType: STRING
    request.useragent:
      valueType: STRING
    response.code:
      valueType: INT64
    response.duration:
      valueType: DURATION
    response.grpc_message:
      valueType: STRING
    response.grpc_status:
      valueType: STRING
    response.headers:
      valueType: STRING_MAP
    response.size:
      valueType: INT64
    response.time:
      valueType: TIMESTAMP
    response.total_size:
      valueType: INT64
    source.principal:
      valueType: STRING
    source.uid:
      valueType: STRING
    source.user:
      valueType: STRING

---

apiVersion: config.istio.io/v1alpha2
kind: attributemanifest
metadata:
  labels:
    app: mixer
  name: kubernetes
spec:
  attributes:
    destination.container.name:
      valueType: STRING
    destination.ip:
      valueType: IP_ADDRESS
    destination.labels:
      valueType: STRING_MAP
    destination.metadata:
      valueType: STRING_MAP
    destination.name:
      valueType: STRING
    destination.namespace:
      valueType: STRING
    destination.owner:
      valueType: STRING
    destination.service.host:
      valueType: STRING
    destination.service.name:
      valueType: STRING
    destination.service.namespace:
      valueType: STRING
    destination.service.uid:
      valueType: STRING
    destination.serviceAccount:
      valueType: STRING
    destination.workload.name:
      valueType: STRING
    destination.workload.namespace:
      valueType: STRING
    destination.workload.uid:
      valueType: STRING
    source.ip:
      valueType: IP_ADDRESS
    source.labels:
      valueType: STRING_MAP
    source.metadata:
      valueType: STRING_MAP
    source.name:
      valueType: STRING
    source.namespace:
      valueType: STRING
    source.owner:
      valueType: STRING
    source.serviceAccount:
      valueType: STRING
    source.services:
      valueType: STRING
    source.workload.name:
      valueType: STRING
    source.workload.namespace:
      valueType: STRING
    source.workload.uid:
      valueType: STRING
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/config-map.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio
  labels:
    app: istio
data:
  mesh: |-
    # Set the following variable to true to disable policy checks by the Mixer.
    # Note that metrics will still be reported to the Mixer.
    disablePolicyChecks: true
    # reportBatchMaxEntries is the number of requests that are batched before telemetry data is sent to the mixer server
    reportBatchMaxEntries: 100
    # reportBatchMaxTime is the max waiting time before the telemetry data of a request is sent to the mixer server
    reportBatchMaxTime: 1s

    # Set enableTracing to false to disable request tracing.
    enableTracing: true

    # Set accessLogFile to empty string to disable access log.
    accessLogFile: ""

    # If accessLogEncoding is TEXT, value will be used directly as the log format
    # example: "[%START_TIME%] %REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\n"
    # If AccessLogEncoding is JSON, value will be parsed as map[string]string
    # example: '{"start_time": "%START_TIME%", "req_method": "%REQ(:METHOD)%"}'
    # Leave empty to use default log format
    accessLogFormat: ""

    # Set accessLogEncoding to JSON or TEXT to configure sidecar access log
    accessLogEncoding: 'TEXT'

    enableEnvoyAccessLogService: false
    mixerCheckServer: istio-policy.$(namespace).svc.cluster.local:15004
    mixerReportServer: istio-telemetry.$(namespace).svc.cluster.local:15004
    # policyCheckFailOpen allows traffic in cases when the mixer policy service cannot be reached.
    # Default is false which means the traffic is denied when the client is unable to connect to Mixer.
    policyCheckFailOpen: false
    # Let Pilot give ingresses the public IP of the Istio ingressgateway
    ingressService: istio-ingressgateway

    # Default connect timeout for dynamic clusters generated by Pilot and returned via XDS
    connectTimeout: 10s

    # Automatic protocol detection uses a set of heuristics to
    # determine whether the connection is using TLS or not (on the
    # server side), as well as the application protocol being used
    # (e.g., http vs tcp). These heuristics rely on the client sending
    # the first bits of data. For server first protocols like MySQL,
    # MongoDB, etc., Envoy will timeout on the protocol detection after
    # the specified period, defaulting to non mTLS plain TCP
    # traffic. Set this field to tweak the period that Envoy will wait
    # for the client to send the first bits of data. (MUST BE >=1ms)
    protocolDetectionTimeout: 100ms

    # DNS refresh rate for Envoy clusters of type STRICT_DNS
    dnsRefreshRate: 300s

    # Unix Domain Socket through which envoy communicates with NodeAgent SDS to get
    # key/cert for mTLS. Use secret-mount files instead of SDS if set to empty.
    sdsUdsPath: "unix:/var/run/sds/uds_path"

    # The trust domain corresponds to the trust root of a system.
    # Refer to https://github.com/spiffe/spiffe/blob/master/standards/SPIFFE-ID.md#21-trust-domain
    trustDomain: ""

    # Set the default behavior of the sidecar for handling outbound traffic from the application:
    # ALLOW_ANY - outbound traffic to unknown destinations will be allowed, in case there are no
    #   services or ServiceEntries for the destination port
    # REGISTRY_ONLY - restrict outbound traffic to services defined in the service registry as well
    #   as those defined through ServiceEntries
    outboundTrafficPolicy:
      mode: ALLOW_ANY
    localityLbSetting:
      enabled: true
    # The namespace to treat as the administrative root namespace for istio
    # configuration.
    rootNamespace: $(namespace)
    configSources:
    - address: istio-galley.$(namespace).svc:9901
      tlsSettings:
        mode: ISTIO_MUTUAL

    defaultConfig:
      #
      # TCP connection timeout between Envoy & the application, and between Envoys.  Used for static clusters
      # defined in Envoy's configuration file
      connectTimeout: 10s
      #
      ### ADVANCED SETTINGS #############
      # Where should envoy's configuration be stored in the istio-proxy container
      configPath: "/etc/istio/proxy"
      binaryPath: "/usr/local/bin/envoy"
      # The pseudo service name used for Envoy.
      serviceCluster: istio-proxy
      # These settings that determine how long an old Envoy
      # process should be kept alive after an occasional reload.
      drainDuration: 45s
      parentShutdownDuration: 1m0s
      #
      # The mode used to redirect inbound connections to Envoy. This setting
      # has no effect on outbound traffic: iptables REDIRECT is always used for
      # outbound connections.
      # If "REDIRECT", use iptables REDIRECT to NAT and redirect to Envoy.
      # The "REDIRECT" mode loses source addresses during redirection.
      # If "TPROXY", use iptables TPROXY to redirect to Envoy.
      # The "TPROXY" mode preserves both the source and destination IP
      # addresses and ports, so that they can be used for advanced filtering
      # and manipulation.
      # The "TPROXY" mode also configures the sidecar to run with the
      # CAP_NET_ADMIN capability, which is required to use TPROXY.
      #interceptionMode: REDIRECT
      #
      # Port where Envoy listens (on local host) for admin commands
      # You can exec into the istio-proxy container in a pod and
      # curl the admin port (curl http://localhost:15000/) to obtain
      # diagnostic information from Envoy. See
      # https://lyft.github.io/envoy/docs/operations/admin.html
      # for more details
      proxyAdminPort: 15000
      #
      # Set concurrency to a specific number to control the number of Proxy worker threads.
      # If set to 0 (default), then start worker thread for each CPU thread/core.
      concurrency: 2
      #
      tracing:
        zipkin:
          # Address of the Zipkin collector
          address: zipkin.$(namespace):9411
      #
      # Mutual TLS authentication between sidecars and istio control plane.
      controlPlaneAuthPolicy: MUTUAL_TLS
      #
      # Address where istio Pilot service is running
      discoveryAddress: istio-pilot.$(namespace):15011

  # Configuration file for the mesh networks to be used by the Split Horizon EDS.
  meshNetworks: |-
    networks: {}

---
# Source: istio/templates/sidecar-injector-configmap.yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-sidecar-injector
  labels:
    app: istio
    istio: sidecar-injector
data:
  values: |-
    {"certmanager":{"enabled":false},"galley":{"enabled":true,"global":{"arch":{"amd64":2,"ppc64le":2,"s390x":2},"configValidation":true,"controlPlaneSecurityEnabled":true,"defaultNodeSelector":{},"defaultPodDisruptionBudget":{"enabled":true},"defaultResources":{"requests":{"cpu":"10m"}},"defaultTolerations":[],"disablePolicyChecks":true,"enableHelmTest":false,"enableTracing":true,"hub":"gcr.io/istio-release","imagePullPolicy":"IfNotPresent","imagePullSecrets":[],"k8sIngress":{"enableHttps":false,"enabled":false,"gatewayName":"ingressgateway"},"localityLbSetting":{"enabled":true},"logging":{"level":"default:info"},"meshExpansion":{"enabled":false,"useILB":false},"meshID":"","meshNetworks":{},"monitoringPort":15014,"mtls":{"enabled":false},"multiCluster":{"clusterName":"","enabled":false},"oneNamespace":false,"outboundTrafficPolicy":{"mode":"ALLOW_ANY"},"policyCheckFailOpen":false,"priorityClassName":"","proxy":{"accessLogEncoding":"TEXT","accessLogFile":"","accessLogFormat":"","autoInject":"enabled","clusterDomain":"cluster.local","componentLogLevel":"","concurrency":2,"dnsRefreshRate":"300s","enableCoreDump":false,"enableCoreDumpImage":"ubuntu:xenial","envoyAccessLogService":{"enabled":false,"host":null,"port":null,"tcpKeepalive":{"interval":"10s","probes":3,"time":"10s"},"tlsSettings":{"caCertificates":null,"clientCertificate":null,"mode":"DISABLE","privateKey":null,"sni":null,"subjectAltNames":[]}},"envoyMetricsService":{"enabled":false,"host":null,"port":null},"envoyStatsd":{"enabled":false,"host":null,"port":null},"excludeIPRanges":"","excludeInboundPorts":"","excludeOutboundPorts":"","image":"proxyv2","includeIPRanges":"*","includeInboundPorts":"*","init":{"resources":{"limits":{"cpu":"100m","memory":"50Mi"},"requests":{"cpu":"10m","memory":"10Mi"}}},"kubevirtInterfaces":"","logLevel":"","privileged":false,"protocolDetectionTimeout":"100ms","readinessFailureThreshold":30,"readinessInitialDelaySeconds":1,"readinessPeriodSeconds":2,"resources":{"limits":{"cpu":"2000m","memory":"1024Mi"},"requests":{"cpu":"100m","memory":"128Mi"}},"statusPort":15020,"tracer":"zipkin"},"proxy_init":{"image":"proxy_init"},"sds":{"enabled":true,"token":{"aud":"istio-ca"},"udsPath":"unix:/var/run/sds/uds_path"},"tag":"release-1.3-latest-daily","tracer":{"datadog":{"address":"$(HOST_IP):8126"},"lightstep":{"accessToken":"","address":"","cacertPath":"","secure":true},"zipkin":{"address":""}},"trustDomain":"","useMCP":true},"image":"galley","nodeSelector":{},"podAntiAffinityLabelSelector":[],"podAntiAffinityTermLabelSelector":[],"replicaCount":1,"rollingMaxSurge":"100%","rollingMaxUnavailable":"25%","tolerations":[]},"gateways":{"enabled":true,"global":{"arch":{"amd64":2,"ppc64le":2,"s390x":2},"configValidation":true,"controlPlaneSecurityEnabled":true,"defaultNodeSelector":{},"defaultPodDisruptionBudget":{"enabled":true},"defaultResources":{"requests":{"cpu":"10m"}},"defaultTolerations":[],"disablePolicyChecks":true,"enableHelmTest":false,"enableTracing":true,"hub":"gcr.io/istio-release","imagePullPolicy":"IfNotPresent","imagePullSecrets":[],"k8sIngress":{"enableHttps":false,"enabled":false,"gatewayName":"ingressgateway"},"localityLbSetting":{"enabled":true},"logging":{"level":"default:info"},"meshExpansion":{"enabled":false,"useILB":false},"meshID":"","meshNetworks":{},"monitoringPort":15014,"mtls":{"enabled":false},"multiCluster":{"clusterName":"","enabled":false},"oneNamespace":false,"outboundTrafficPolicy":{"mode":"ALLOW_ANY"},"policyCheckFailOpen":false,"priorityClassName":"","proxy":{"accessLogEncoding":"TEXT","accessLogFile":"","accessLogFormat":"","autoInject":"enabled","clusterDomain":"cluster.local","componentLogLevel":"","concurrency":2,"dnsRefreshRate":"300s","enableCoreDump":false,"enableCoreDumpImage":"ubuntu:xenial","envoyAccessLogService":{"enabled":false,"host":null,"port":null,"tcpKeepalive":{"interval":"10s","probes":3,"time":"10s"},"tlsSettings":{"caCertificates":null,"clientCertificate":null,"mode":"DISABLE","privateKey":null,"sni":null,"subjectAltNames":[]}},"envoyMetricsService":{"enabled":false,"host":null,"port":null},"envoyStatsd":{"enabled":false,"host":null,"port":null},"excludeIPRanges":"","excludeInboundPorts":"","excludeOutboundPorts":"","image":"proxyv2","includeIPRanges":"*","includeInboundPorts":"*","init":{"resources":{"limits":{"cpu":"100m","memory":"50Mi"},"requests":{"cpu":"10m","memory":"10Mi"}}},"kubevirtInterfaces":"","logLevel":"","privileged":false,"protocolDetectionTimeout":"100ms","readinessFailureThreshold":30,"readinessInitialDelaySeconds":1,"readinessPeriodSeconds":2,"resources":{"limits":{"cpu":"2000m","memory":"1024Mi"},"requests":{"cpu":"100m","memory":"128Mi"}},"statusPort":15020,"tracer":"zipkin"},"proxy_init":{"image":"proxy_init"},"sds":{"enabled":true,"token":{"aud":"istio-ca"},"udsPath":"unix:/var/run/sds/uds_path"},"tag":"release-1.3-latest-daily","tracer":{"datadog":{"address":"$(HOST_IP):8126"},"lightstep":{"accessToken":"","address":"","cacertPath":"","secure":true},"zipkin":{"address":""}},"trustDomain":"","useMCP":true},"istio-egressgateway":{"autoscaleEnabled":true,"autoscaleMax":5,"autoscaleMin":1,"cpu":{"targetAverageUtilization":80},"enabled":false,"env":{"ISTIO_META_ROUTER_MODE":"sni-dnat"},"labels":{"app":"istio-egressgateway","istio":"egressgateway"},"nodeSelector":{},"podAnnotations":{},"podAntiAffinityLabelSelector":[],"podAntiAffinityTermLabelSelector":[],"ports":[{"name":"http2","port":80},{"name":"https","port":443},{"name":"tls","port":15443,"targetPort":15443}],"resources":{"limits":{"cpu":"2000m","memory":"1024Mi"},"requests":{"cpu":"100m","memory":"128Mi"}},"rollingMaxSurge":"100%","rollingMaxUnavailable":"25%","secretVolumes":[{"mountPath":"/etc/istio/egressgateway-certs","name":"egressgateway-certs","secretName":"istio-egressgateway-certs"},{"mountPath":"/etc/istio/egressgateway-ca-certs","name":"egressgateway-ca-certs","secretName":"istio-egressgateway-ca-certs"}],"serviceAnnotations":{},"tolerations":[],"type":"ClusterIP"},"istio-ilbgateway":{"autoscaleEnabled":true,"autoscaleMax":5,"autoscaleMin":1,"cpu":{"targetAverageUtilization":80},"enabled":false,"labels":{"app":"istio-ilbgateway","istio":"ilbgateway"},"loadBalancerIP":"","nodeSelector":{},"podAnnotations":{},"ports":[{"name":"grpc-pilot-mtls","port":15011},{"name":"grpc-pilot","port":15010},{"name":"tcp-citadel-grpc-tls","port":8060,"targetPort":8060},{"name":"tcp-dns","port":5353}],"resources":{"requests":{"cpu":"800m","memory":"512Mi"}},"rollingMaxSurge":"100%","rollingMaxUnavailable":"25%","secretVolumes":[{"mountPath":"/etc/istio/ilbgateway-certs","name":"ilbgateway-certs","secretName":"istio-ilbgateway-certs"},{"mountPath":"/etc/istio/ilbgateway-ca-certs","name":"ilbgateway-ca-certs","secretName":"istio-ilbgateway-ca-certs"}],"serviceAnnotations":{"cloud.google.com/load-balancer-type":"internal"},"tolerations":[],"type":"LoadBalancer"},"istio-ingressgateway":{"applicationPorts":"","autoscaleEnabled":true,"autoscaleMax":5,"autoscaleMin":1,"cpu":{"targetAverageUtilization":80},"enabled":true,"env":{"ISTIO_META_ROUTER_MODE":"sni-dnat"},"externalIPs":[],"labels":{"app":"istio-ingressgateway","istio":"ingressgateway"},"loadBalancerIP":"","loadBalancerSourceRanges":[],"meshExpansionPorts":[{"name":"tcp-pilot-grpc-tls","port":15011,"targetPort":15011},{"name":"tcp-mixer-grpc-tls","port":15004,"targetPort":15004},{"name":"tcp-citadel-grpc-tls","port":8060,"targetPort":8060},{"name":"tcp-dns-tls","port":853,"targetPort":853}],"nodeSelector":{},"podAnnotations":{},"podAntiAffinityLabelSelector":[],"podAntiAffinityTermLabelSelector":[],"ports":[{"name":"status-port","port":15020,"targetPort":15020},{"name":"http2","nodePort":31380,"port":80,"targetPort":80},{"name":"https","nodePort":31390,"port":443},{"name":"tcp","nodePort":31400,"port":31400},{"name":"https-kiali","port":15029,"targetPort":15029},{"name":"https-prometheus","port":15030,"targetPort":15030},{"name":"https-grafana","port":15031,"targetPort":15031},{"name":"https-tracing","port":15032,"targetPort":15032},{"name":"tls","port":15443,"targetPort":15443}],"resources":{"limits":{"cpu":"2000m","memory":"1024Mi"},"requests":{"cpu":"100m","memory":"128Mi"}},"rollingMaxSurge":"100%","rollingMaxUnavailable":"25%","sds":{"enabled":false,"image":"node-agent-k8s","resources":{"limits":{"cpu":"2000m","memory":"1024Mi"},"requests":{"cpu":"100m","memory":"128Mi"}}},"secretVolumes":[{"mountPath":"/etc/istio/ingressgateway-certs","name":"ingressgateway-certs","secretName":"istio-ingressgateway-certs"},{"mountPath":"/etc/istio/ingressgateway-ca-certs","name":"ingressgateway-ca-certs","secretName":"istio-ingressgateway-ca-certs"}],"serviceAnnotations":{},"tolerations":[],"type":"LoadBalancer"}},"global":{"arch":{"amd64":2,"ppc64le":2,"s390x":2},"configValidation":true,"controlPlaneSecurityEnabled":true,"defaultNodeSelector":{},"defaultPodDisruptionBudget":{"enabled":true},"defaultResources":{"requests":{"cpu":"10m"}},"defaultTolerations":[],"disablePolicyChecks":true,"enableHelmTest":false,"enableTracing":true,"hub":"gcr.io/istio-release","imagePullPolicy":"IfNotPresent","imagePullSecrets":[],"k8sIngress":{"enableHttps":false,"enabled":false,"gatewayName":"ingressgateway"},"localityLbSetting":{"enabled":true},"logging":{"level":"default:info"},"meshExpansion":{"enabled":false,"useILB":false},"meshID":"","meshNetworks":{},"monitoringPort":15014,"mtls":{"enabled":false},"multiCluster":{"clusterName":"","enabled":false},"oneNamespace":false,"outboundTrafficPolicy":{"mode":"ALLOW_ANY"},"policyCheckFailOpen":false,"priorityClassName":"","proxy":{"accessLogEncoding":"TEXT","accessLogFile":"","accessLogFormat":"","autoInject":"enabled","clusterDomain":"cluster.local","componentLogLevel":"","concurrency":2,"dnsRefreshRate":"300s","enableCoreDump":false,"enableCoreDumpImage":"ubuntu:xenial","envoyAccessLogService":{"enabled":false,"host":null,"port":null,"tcpKeepalive":{"interval":"10s","probes":3,"time":"10s"},"tlsSettings":{"caCertificates":null,"clientCertificate":null,"mode":"DISABLE","privateKey":null,"sni":null,"subjectAltNames":[]}},"envoyMetricsService":{"enabled":false,"host":null,"port":null},"envoyStatsd":{"enabled":false,"host":null,"port":null},"excludeIPRanges":"","excludeInboundPorts":"","excludeOutboundPorts":"","image":"proxyv2","includeIPRanges":"*","includeInboundPorts":"*","init":{"resources":{"limits":{"cpu":"100m","memory":"50Mi"},"requests":{"cpu":"10m","memory":"10Mi"}}},"kubevirtInterfaces":"","logLevel":"","privileged":false,"protocolDetectionTimeout":"100ms","readinessFailureThreshold":30,"readinessInitialDelaySeconds":1,"readinessPeriodSeconds":2,"resources":{"limits":{"cpu":"2000m","memory":"1024Mi"},"requests":{"cpu":"100m","memory":"128Mi"}},"statusPort":15020,"tracer":"zipkin"},"proxy_init":{"image":"proxy_init"},"sds":{"enabled":true,"token":{"aud":"istio-ca"},"udsPath":"unix:/var/run/sds/uds_path"},"tag":"release-1.3-latest-daily","tracer":{"datadog":{"address":"$(HOST_IP):8126"},"lightstep":{"accessToken":"","address":"","cacertPath":"","secure":true},"zipkin":{"address":""}},"trustDomain":"","useMCP":true},"grafana":{"enabled":false},"istio_cni":{"enabled":false},"istiocoredns":{"enabled":false},"kiali":{"enabled":false},"mixer":{"adapters":{"kubernetesenv":{"enabled":true},"prometheus":{"enabled":true,"metricsExpiryDuration":"10m"},"stdio":{"enabled":false,"outputAsJson":true},"useAdapterCRDs":false},"env":{"GODEBUG":"gctrace=1","GOMAXPROCS":"6"},"global":{"arch":{"amd64":2,"ppc64le":2,"s390x":2},"configValidation":true,"controlPlaneSecurityEnabled":true,"defaultNodeSelector":{},"defaultPodDisruptionBudget":{"enabled":true},"defaultResources":{"requests":{"cpu":"10m"}},"defaultTolerations":[],"disablePolicyChecks":true,"enableHelmTest":false,"enableTracing":true,"hub":"gcr.io/istio-release","imagePullPolicy":"IfNotPresent","imagePullSecrets":[],"k8sIngress":{"enableHttps":false,"enabled":false,"gatewayName":"ingressgateway"},"localityLbSetting":{"enabled":true},"logging":{"level":"default:info"},"meshExpansion":{"enabled":false,"useILB":false},"meshID":"","meshNetworks":{},"monitoringPort":15014,"mtls":{"enabled":false},"multiCluster":{"clusterName":"","enabled":false},"oneNamespace":false,"outboundTrafficPolicy":{"mode":"ALLOW_ANY"},"policyCheckFailOpen":false,"priorityClassName":"","proxy":{"accessLogEncoding":"TEXT","accessLogFile":"","accessLogFormat":"","autoInject":"enabled","clusterDomain":"cluster.local","componentLogLevel":"","concurrency":2,"dnsRefreshRate":"300s","enableCoreDump":false,"enableCoreDumpImage":"ubuntu:xenial","envoyAccessLogService":{"enabled":false,"host":null,"port":null,"tcpKeepalive":{"interval":"10s","probes":3,"time":"10s"},"tlsSettings":{"caCertificates":null,"clientCertificate":null,"mode":"DISABLE","privateKey":null,"sni":null,"subjectAltNames":[]}},"envoyMetricsService":{"enabled":false,"host":null,"port":null},"envoyStatsd":{"enabled":false,"host":null,"port":null},"excludeIPRanges":"","excludeInboundPorts":"","excludeOutboundPorts":"","image":"proxyv2","includeIPRanges":"*","includeInboundPorts":"*","init":{"resources":{"limits":{"cpu":"100m","memory":"50Mi"},"requests":{"cpu":"10m","memory":"10Mi"}}},"kubevirtInterfaces":"","logLevel":"","privileged":false,"protocolDetectionTimeout":"100ms","readinessFailureThreshold":30,"readinessInitialDelaySeconds":1,"readinessPeriodSeconds":2,"resources":{"limits":{"cpu":"2000m","memory":"1024Mi"},"requests":{"cpu":"100m","memory":"128Mi"}},"statusPort":15020,"tracer":"zipkin"},"proxy_init":{"image":"proxy_init"},"sds":{"enabled":true,"token":{"aud":"istio-ca"},"udsPath":"unix:/var/run/sds/uds_path"},"tag":"release-1.3-latest-daily","tracer":{"datadog":{"address":"$(HOST_IP):8126"},"lightstep":{"accessToken":"","address":"","cacertPath":"","secure":true},"zipkin":{"address":""}},"trustDomain":"","useMCP":true},"image":"mixer","nodeSelector":{},"podAnnotations":{},"podAntiAffinityLabelSelector":[],"podAntiAffinityTermLabelSelector":[],"policy":{"autoscaleEnabled":true,"autoscaleMax":5,"autoscaleMin":1,"cpu":{"targetAverageUtilization":80},"enabled":true,"replicaCount":1,"rollingMaxSurge":"100%","rollingMaxUnavailable":"25%"},"telemetry":{"autoscaleEnabled":true,"autoscaleMax":5,"autoscaleMin":1,"cpu":{"targetAverageUtilization":80},"enabled":true,"loadshedding":{"latencyThreshold":"100ms","mode":"enforce"},"replicaCount":1,"reportBatchMaxEntries":100,"reportBatchMaxTime":"1s","resources":{"limits":{"cpu":"4800m","memory":"4G"},"requests":{"cpu":"1000m","memory":"1G"}},"rollingMaxSurge":"100%","rollingMaxUnavailable":"25%","sessionAffinityEnabled":false},"tolerations":[]},"nodeagent":{"enabled":true,"env":{"CA_ADDR":"istio-citadel:8060","CA_PROVIDER":"Citadel","PLUGINS":"","VALID_TOKEN":true},"global":{"arch":{"amd64":2,"ppc64le":2,"s390x":2},"configValidation":true,"controlPlaneSecurityEnabled":true,"defaultNodeSelector":{},"defaultPodDisruptionBudget":{"enabled":true},"defaultResources":{"requests":{"cpu":"10m"}},"defaultTolerations":[],"disablePolicyChecks":true,"enableHelmTest":false,"enableTracing":true,"hub":"gcr.io/istio-release","imagePullPolicy":"IfNotPresent","imagePullSecrets":[],"k8sIngress":{"enableHttps":false,"enabled":false,"gatewayName":"ingressgateway"},"localityLbSetting":{"enabled":true},"logging":{"level":"default:info"},"meshExpansion":{"enabled":false,"useILB":false},"meshID":"","meshNetworks":{},"monitoringPort":15014,"mtls":{"enabled":false},"multiCluster":{"clusterName":"","enabled":false},"oneNamespace":false,"outboundTrafficPolicy":{"mode":"ALLOW_ANY"},"policyCheckFailOpen":false,"priorityClassName":"","proxy":{"accessLogEncoding":"TEXT","accessLogFile":"","accessLogFormat":"","autoInject":"enabled","clusterDomain":"cluster.local","componentLogLevel":"","concurrency":2,"dnsRefreshRate":"300s","enableCoreDump":false,"enableCoreDumpImage":"ubuntu:xenial","envoyAccessLogService":{"enabled":false,"host":null,"port":null,"tcpKeepalive":{"interval":"10s","probes":3,"time":"10s"},"tlsSettings":{"caCertificates":null,"clientCertificate":null,"mode":"DISABLE","privateKey":null,"sni":null,"subjectAltNames":[]}},"envoyMetricsService":{"enabled":false,"host":null,"port":null},"envoyStatsd":{"enabled":false,"host":null,"port":null},"excludeIPRanges":"","excludeInboundPorts":"","excludeOutboundPorts":"","image":"proxyv2","includeIPRanges":"*","includeInboundPorts":"*","init":{"resources":{"limits":{"cpu":"100m","memory":"50Mi"},"requests":{"cpu":"10m","memory":"10Mi"}}},"kubevirtInterfaces":"","logLevel":"","privileged":false,"protocolDetectionTimeout":"100ms","readinessFailureThreshold":30,"readinessInitialDelaySeconds":1,"readinessPeriodSeconds":2,"resources":{"limits":{"cpu":"2000m","memory":"1024Mi"},"requests":{"cpu":"100m","memory":"128Mi"}},"statusPort":15020,"tracer":"zipkin"},"proxy_init":{"image":"proxy_init"},"sds":{"enabled":true,"token":{"aud":"istio-ca"},"udsPath":"unix:/var/run/sds/uds_path"},"tag":"release-1.3-latest-daily","tracer":{"datadog":{"address":"$(HOST_IP):8126"},"lightstep":{"accessToken":"","address":"","cacertPath":"","secure":true},"zipkin":{"address":""}},"trustDomain":"","useMCP":true},"image":"node-agent-k8s","nodeSelector":{},"podAntiAffinityLabelSelector":[],"podAntiAffinityTermLabelSelector":[],"tolerations":[]},"pilot":{"autoscaleEnabled":true,"autoscaleMax":5,"autoscaleMin":1,"cpu":{"targetAverageUtilization":80},"enableProtocolSniffingForInbound":false,"enableProtocolSniffingForOutbound":true,"enabled":true,"env":{"GODEBUG":"gctrace=1","PILOT_PUSH_THROTTLE":100},"global":{"arch":{"amd64":2,"ppc64le":2,"s390x":2},"configValidation":true,"controlPlaneSecurityEnabled":true,"defaultNodeSelector":{},"defaultPodDisruptionBudget":{"enabled":true},"defaultResources":{"requests":{"cpu":"10m"}},"defaultTolerations":[],"disablePolicyChecks":true,"enableHelmTest":false,"enableTracing":true,"hub":"gcr.io/istio-release","imagePullPolicy":"IfNotPresent","imagePullSecrets":[],"k8sIngress":{"enableHttps":false,"enabled":false,"gatewayName":"ingressgateway"},"localityLbSetting":{"enabled":true},"logging":{"level":"default:info"},"meshExpansion":{"enabled":false,"useILB":false},"meshID":"","meshNetworks":{},"monitoringPort":15014,"mtls":{"enabled":false},"multiCluster":{"clusterName":"","enabled":false},"oneNamespace":false,"outboundTrafficPolicy":{"mode":"ALLOW_ANY"},"policyCheckFailOpen":false,"priorityClassName":"","proxy":{"accessLogEncoding":"TEXT","accessLogFile":"","accessLogFormat":"","autoInject":"enabled","clusterDomain":"cluster.local","componentLogLevel":"","concurrency":2,"dnsRefreshRate":"300s","enableCoreDump":false,"enableCoreDumpImage":"ubuntu:xenial","envoyAccessLogService":{"enabled":false,"host":null,"port":null,"tcpKeepalive":{"interval":"10s","probes":3,"time":"10s"},"tlsSettings":{"caCertificates":null,"clientCertificate":null,"mode":"DISABLE","privateKey":null,"sni":null,"subjectAltNames":[]}},"envoyMetricsService":{"enabled":false,"host":null,"port":null},"envoyStatsd":{"enabled":false,"host":null,"port":null},"excludeIPRanges":"","excludeInboundPorts":"","excludeOutboundPorts":"","image":"proxyv2","includeIPRanges":"*","includeInboundPorts":"*","init":{"resources":{"limits":{"cpu":"100m","memory":"50Mi"},"requests":{"cpu":"10m","memory":"10Mi"}}},"kubevirtInterfaces":"","logLevel":"","privileged":false,"protocolDetectionTimeout":"100ms","readinessFailureThreshold":30,"readinessInitialDelaySeconds":1,"readinessPeriodSeconds":2,"resources":{"limits":{"cpu":"2000m","memory":"1024Mi"},"requests":{"cpu":"100m","memory":"128Mi"}},"statusPort":15020,"tracer":"zipkin"},"proxy_init":{"image":"proxy_init"},"sds":{"enabled":true,"token":{"aud":"istio-ca"},"udsPath":"unix:/var/run/sds/uds_path"},"tag":"release-1.3-latest-daily","tracer":{"datadog":{"address":"$(HOST_IP):8126"},"lightstep":{"accessToken":"","address":"","cacertPath":"","secure":true},"zipkin":{"address":""}},"trustDomain":"","useMCP":true},"image":"pilot","keepaliveMaxServerConnectionAge":"30m","nodeSelector":{},"podAntiAffinityLabelSelector":[],"podAntiAffinityTermLabelSelector":[],"resources":{"requests":{"cpu":"500m","memory":"2048Mi"}},"rollingMaxSurge":"100%","rollingMaxUnavailable":"25%","sidecar":true,"tolerations":[],"traceSampling":1},"prometheus":{"contextPath":"/prometheus","enabled":true,"global":{"arch":{"amd64":2,"ppc64le":2,"s390x":2},"configValidation":true,"controlPlaneSecurityEnabled":true,"defaultNodeSelector":{},"defaultPodDisruptionBudget":{"enabled":true},"defaultResources":{"requests":{"cpu":"10m"}},"defaultTolerations":[],"disablePolicyChecks":true,"enableHelmTest":false,"enableTracing":true,"hub":"gcr.io/istio-release","imagePullPolicy":"IfNotPresent","imagePullSecrets":[],"k8sIngress":{"enableHttps":false,"enabled":false,"gatewayName":"ingressgateway"},"localityLbSetting":{"enabled":true},"logging":{"level":"default:info"},"meshExpansion":{"enabled":false,"useILB":false},"meshID":"","meshNetworks":{},"monitoringPort":15014,"mtls":{"enabled":false},"multiCluster":{"clusterName":"","enabled":false},"oneNamespace":false,"outboundTrafficPolicy":{"mode":"ALLOW_ANY"},"policyCheckFailOpen":false,"priorityClassName":"","proxy":{"accessLogEncoding":"TEXT","accessLogFile":"","accessLogFormat":"","autoInject":"enabled","clusterDomain":"cluster.local","componentLogLevel":"","concurrency":2,"dnsRefreshRate":"300s","enableCoreDump":false,"enableCoreDumpImage":"ubuntu:xenial","envoyAccessLogService":{"enabled":false,"host":null,"port":null,"tcpKeepalive":{"interval":"10s","probes":3,"time":"10s"},"tlsSettings":{"caCertificates":null,"clientCertificate":null,"mode":"DISABLE","privateKey":null,"sni":null,"subjectAltNames":[]}},"envoyMetricsService":{"enabled":false,"host":null,"port":null},"envoyStatsd":{"enabled":false,"host":null,"port":null},"excludeIPRanges":"","excludeInboundPorts":"","excludeOutboundPorts":"","image":"proxyv2","includeIPRanges":"*","includeInboundPorts":"*","init":{"resources":{"limits":{"cpu":"100m","memory":"50Mi"},"requests":{"cpu":"10m","memory":"10Mi"}}},"kubevirtInterfaces":"","logLevel":"","privileged":false,"protocolDetectionTimeout":"100ms","readinessFailureThreshold":30,"readinessInitialDelaySeconds":1,"readinessPeriodSeconds":2,"resources":{"limits":{"cpu":"2000m","memory":"1024Mi"},"requests":{"cpu":"100m","memory":"128Mi"}},"statusPort":15020,"tracer":"zipkin"},"proxy_init":{"image":"proxy_init"},"sds":{"enabled":true,"token":{"aud":"istio-ca"},"udsPath":"unix:/var/run/sds/uds_path"},"tag":"release-1.3-latest-daily","tracer":{"datadog":{"address":"$(HOST_IP):8126"},"lightstep":{"accessToken":"","address":"","cacertPath":"","secure":true},"zipkin":{"address":""}},"trustDomain":"","useMCP":true},"hub":"docker.io/prom","image":"prometheus","ingress":{"annotations":null,"enabled":false,"hosts":["prometheus.local"],"tls":null},"nodeSelector":{},"podAntiAffinityLabelSelector":[],"podAntiAffinityTermLabelSelector":[],"replicaCount":1,"retention":"6h","scrapeInterval":"15s","security":{"enabled":true},"service":{"annotations":{},"nodePort":{"enabled":false,"port":32090}},"tag":"v2.8.0","tolerations":[]},"security":{"citadelHealthCheck":false,"createMeshPolicy":true,"enableNamespacesByDefault":true,"enabled":true,"global":{"arch":{"amd64":2,"ppc64le":2,"s390x":2},"configValidation":true,"controlPlaneSecurityEnabled":true,"defaultNodeSelector":{},"defaultPodDisruptionBudget":{"enabled":true},"defaultResources":{"requests":{"cpu":"10m"}},"defaultTolerations":[],"disablePolicyChecks":true,"enableHelmTest":false,"enableTracing":true,"hub":"gcr.io/istio-release","imagePullPolicy":"IfNotPresent","imagePullSecrets":[],"k8sIngress":{"enableHttps":false,"enabled":false,"gatewayName":"ingressgateway"},"localityLbSetting":{"enabled":true},"logging":{"level":"default:info"},"meshExpansion":{"enabled":false,"useILB":false},"meshID":"","meshNetworks":{},"monitoringPort":15014,"mtls":{"enabled":false},"multiCluster":{"clusterName":"","enabled":false},"oneNamespace":false,"outboundTrafficPolicy":{"mode":"ALLOW_ANY"},"policyCheckFailOpen":false,"priorityClassName":"","proxy":{"accessLogEncoding":"TEXT","accessLogFile":"","accessLogFormat":"","autoInject":"enabled","clusterDomain":"cluster.local","componentLogLevel":"","concurrency":2,"dnsRefreshRate":"300s","enableCoreDump":false,"enableCoreDumpImage":"ubuntu:xenial","envoyAccessLogService":{"enabled":false,"host":null,"port":null,"tcpKeepalive":{"interval":"10s","probes":3,"time":"10s"},"tlsSettings":{"caCertificates":null,"clientCertificate":null,"mode":"DISABLE","privateKey":null,"sni":null,"subjectAltNames":[]}},"envoyMetricsService":{"enabled":false,"host":null,"port":null},"envoyStatsd":{"enabled":false,"host":null,"port":null},"excludeIPRanges":"","excludeInboundPorts":"","excludeOutboundPorts":"","image":"proxyv2","includeIPRanges":"*","includeInboundPorts":"*","init":{"resources":{"limits":{"cpu":"100m","memory":"50Mi"},"requests":{"cpu":"10m","memory":"10Mi"}}},"kubevirtInterfaces":"","logLevel":"","privileged":false,"protocolDetectionTimeout":"100ms","readinessFailureThreshold":30,"readinessInitialDelaySeconds":1,"readinessPeriodSeconds":2,"resources":{"limits":{"cpu":"2000m","memory":"1024Mi"},"requests":{"cpu":"100m","memory":"128Mi"}},"statusPort":15020,"tracer":"zipkin"},"proxy_init":{"image":"proxy_init"},"sds":{"enabled":true,"token":{"aud":"istio-ca"},"udsPath":"unix:/var/run/sds/uds_path"},"tag":"release-1.3-latest-daily","tracer":{"datadog":{"address":"$(HOST_IP):8126"},"lightstep":{"accessToken":"","address":"","cacertPath":"","secure":true},"zipkin":{"address":""}},"trustDomain":"","useMCP":true},"image":"citadel","nodeSelector":{},"podAntiAffinityLabelSelector":[],"podAntiAffinityTermLabelSelector":[],"replicaCount":1,"rollingMaxSurge":"100%","rollingMaxUnavailable":"25%","selfSigned":true,"tolerations":[],"workloadCertTtl":"2160h"},"sidecarInjectorWebhook":{"alwaysInjectSelector":[],"enableNamespacesByDefault":false,"enabled":true,"global":{"arch":{"amd64":2,"ppc64le":2,"s390x":2},"configValidation":true,"controlPlaneSecurityEnabled":true,"defaultNodeSelector":{},"defaultPodDisruptionBudget":{"enabled":true},"defaultResources":{"requests":{"cpu":"10m"}},"defaultTolerations":[],"disablePolicyChecks":true,"enableHelmTest":false,"enableTracing":true,"hub":"gcr.io/istio-release","imagePullPolicy":"IfNotPresent","imagePullSecrets":[],"k8sIngress":{"enableHttps":false,"enabled":false,"gatewayName":"ingressgateway"},"localityLbSetting":{"enabled":true},"logging":{"level":"default:info"},"meshExpansion":{"enabled":false,"useILB":false},"meshID":"","meshNetworks":{},"monitoringPort":15014,"mtls":{"enabled":false},"multiCluster":{"clusterName":"","enabled":false},"oneNamespace":false,"outboundTrafficPolicy":{"mode":"ALLOW_ANY"},"policyCheckFailOpen":false,"priorityClassName":"","proxy":{"accessLogEncoding":"TEXT","accessLogFile":"","accessLogFormat":"","autoInject":"enabled","clusterDomain":"cluster.local","componentLogLevel":"","concurrency":2,"dnsRefreshRate":"300s","enableCoreDump":false,"enableCoreDumpImage":"ubuntu:xenial","envoyAccessLogService":{"enabled":false,"host":null,"port":null,"tcpKeepalive":{"interval":"10s","probes":3,"time":"10s"},"tlsSettings":{"caCertificates":null,"clientCertificate":null,"mode":"DISABLE","privateKey":null,"sni":null,"subjectAltNames":[]}},"envoyMetricsService":{"enabled":false,"host":null,"port":null},"envoyStatsd":{"enabled":false,"host":null,"port":null},"excludeIPRanges":"","excludeInboundPorts":"","excludeOutboundPorts":"","image":"proxyv2","includeIPRanges":"*","includeInboundPorts":"*","init":{"resources":{"limits":{"cpu":"100m","memory":"50Mi"},"requests":{"cpu":"10m","memory":"10Mi"}}},"kubevirtInterfaces":"","logLevel":"","privileged":false,"protocolDetectionTimeout":"100ms","readinessFailureThreshold":30,"readinessInitialDelaySeconds":1,"readinessPeriodSeconds":2,"resources":{"limits":{"cpu":"2000m","memory":"1024Mi"},"requests":{"cpu":"100m","memory":"128Mi"}},"statusPort":15020,"tracer":"zipkin"},"proxy_init":{"image":"proxy_init"},"sds":{"enabled":true,"token":{"aud":"istio-ca"},"udsPath":"unix:/var/run/sds/uds_path"},"tag":"release-1.3-latest-daily","tracer":{"datadog":{"address":"$(HOST_IP):8126"},"lightstep":{"accessToken":"","address":"","cacertPath":"","secure":true},"zipkin":{"address":""}},"trustDomain":"","useMCP":true},"image":"sidecar_injector","neverInjectSelector":[],"nodeSelector":{},"podAntiAffinityLabelSelector":[],"podAntiAffinityTermLabelSelector":[],"replicaCount":1,"rewriteAppHTTPProbe":false,"rollingMaxSurge":"100%","rollingMaxUnavailable":"25%","tolerations":[]},"tracing":{"enabled":false}}

  config: |-
    policy: enabled
    alwaysInjectSelector:
      []
    neverInjectSelector:
      []
    template: |-
      rewriteAppHTTPProbe: {{ valueOrDefault .Values.sidecarInjectorWebhook.rewriteAppHTTPProbe false }}
      {{- if or (not .Values.istio_cni.enabled) .Values.global.proxy.enableCoreDump }}
      initContainers:
      {{ if ne (annotation .ObjectMeta `+"`"+`sidecar.istio.io/interceptionMode`+"`"+` .ProxyConfig.InterceptionMode) `+"`"+`NONE`+"`"+` }}
      {{- if not .Values.istio_cni.enabled }}
      - name: istio-init
      {{- if contains "/" .Values.global.proxy_init.image }}
        image: "{{ .Values.global.proxy_init.image }}"
      {{- else }}
        image: "{{ .Values.global.hub }}/{{ .Values.global.proxy_init.image }}:{{ .Values.global.tag }}"
      {{- end }}
        args:
        - "-p"
        - "15001"
        - "-z"
        - "15006"
        - "-u"
        - 1337
        - "-m"
        - "{{ annotation .ObjectMeta `+"`"+`sidecar.istio.io/interceptionMode`+"`"+` .ProxyConfig.InterceptionMode }}"
        - "-i"
        - "{{ annotation .ObjectMeta `+"`"+`traffic.sidecar.istio.io/includeOutboundIPRanges`+"`"+` .Values.global.proxy.includeIPRanges }}"
        - "-x"
        - "{{ annotation .ObjectMeta `+"`"+`traffic.sidecar.istio.io/excludeOutboundIPRanges`+"`"+` .Values.global.proxy.excludeIPRanges }}"
        - "-b"
        - "{{ annotation .ObjectMeta `+"`"+`traffic.sidecar.istio.io/includeInboundPorts`+"`"+` `+"`"+`*`+"`"+` }}"
        - "-d"
        - "{{ excludeInboundPort (annotation .ObjectMeta `+"`"+`status.sidecar.istio.io/port`+"`"+` .Values.global.proxy.statusPort) (annotation .ObjectMeta `+"`"+`traffic.sidecar.istio.io/excludeInboundPorts`+"`"+` .Values.global.proxy.excludeInboundPorts) }}"
        {{ if or (isset .ObjectMeta.Annotations `+"`"+`traffic.sidecar.istio.io/excludeOutboundPorts`+"`"+`) (ne .Values.global.proxy.excludeOutboundPorts "") -}}
        - "-o"
        - "{{ annotation .ObjectMeta `+"`"+`traffic.sidecar.istio.io/excludeOutboundPorts`+"`"+` .Values.global.proxy.excludeOutboundPorts }}"
        {{ end -}}
        {{ if (isset .ObjectMeta.Annotations `+"`"+`traffic.sidecar.istio.io/kubevirtInterfaces`+"`"+`) -}}
        - "-k"
        - "{{ index .ObjectMeta.Annotations `+"`"+`traffic.sidecar.istio.io/kubevirtInterfaces`+"`"+` }}"
        {{ end -}}
        imagePullPolicy: "{{ .Values.global.imagePullPolicy }}"
      {{- if .Values.global.proxy.init.resources }}
        resources:
          {{ toYaml .Values.global.proxy.init.resources | indent 4 }}
      {{- else }}
        resources: {}
      {{- end }}
        securityContext:
          runAsUser: 0
          runAsNonRoot: false
          capabilities:
            add:
            - NET_ADMIN
          {{- if .Values.global.proxy.privileged }}
          privileged: true
          {{- end }}
        restartPolicy: Always
      {{- end }}
      {{  end -}}
      {{- if eq .Values.global.proxy.enableCoreDump true }}
      - name: enable-core-dump
        args:
        - -c
        - sysctl -w kernel.core_pattern=/var/lib/istio/core.proxy && ulimit -c unlimited
        command:
          - /bin/sh
        image: {{ $.Values.global.proxy.enableCoreDumpImage }}
        imagePullPolicy: IfNotPresent
        resources: {}
        securityContext:
          runAsUser: 0
          runAsNonRoot: false
          privileged: true
      {{ end }}
      {{- end }}
      containers:
      - name: istio-proxy
      {{- if contains "/" (annotation .ObjectMeta `+"`"+`sidecar.istio.io/proxyImage`+"`"+` .Values.global.proxy.image) }}
        image: "{{ annotation .ObjectMeta `+"`"+`sidecar.istio.io/proxyImage`+"`"+` .Values.global.proxy.image }}"
      {{- else }}
        image: "{{ annotation .ObjectMeta `+"`"+`sidecar.istio.io/proxyImage`+"`"+` .Values.global.hub }}/{{ .Values.global.proxy.image }}:{{ .Values.global.tag }}"
      {{- end }}
        ports:
        - containerPort: 15090
          protocol: TCP
          name: http-envoy-prom
        args:
        - proxy
        - sidecar
        - --domain
        - $(POD_NAMESPACE).svc.{{ .Values.global.proxy.clusterDomain }}
        - --configPath
        - "{{ .ProxyConfig.ConfigPath }}"
        - --binaryPath
        - "{{ .ProxyConfig.BinaryPath }}"
        - --serviceCluster
        {{ if ne "" (index .ObjectMeta.Labels "app") -}}
        - "{{ index .ObjectMeta.Labels `+"`"+`app`+"`"+` }}.$(POD_NAMESPACE)"
        {{ else -}}
        - "{{ valueOrDefault .DeploymentMeta.Name `+"`"+`istio-proxy`+"`"+` }}.{{ valueOrDefault .DeploymentMeta.Namespace `+"`"+`default`+"`"+` }}"
        {{ end -}}
        - --drainDuration
        - "{{ formatDuration .ProxyConfig.DrainDuration }}"
        - --parentShutdownDuration
        - "{{ formatDuration .ProxyConfig.ParentShutdownDuration }}"
        - --discoveryAddress
        - "{{ annotation .ObjectMeta `+"`"+`sidecar.istio.io/discoveryAddress`+"`"+` .ProxyConfig.DiscoveryAddress }}"
      {{- if eq .Values.global.proxy.tracer "lightstep" }}
        - --lightstepAddress
        - "{{ .ProxyConfig.GetTracing.GetLightstep.GetAddress }}"
        - --lightstepAccessToken
        - "{{ .ProxyConfig.GetTracing.GetLightstep.GetAccessToken }}"
        - --lightstepSecure={{ .ProxyConfig.GetTracing.GetLightstep.GetSecure }}
        - --lightstepCacertPath
        - "{{ .ProxyConfig.GetTracing.GetLightstep.GetCacertPath }}"
      {{- else if eq .Values.global.proxy.tracer "zipkin" }}
        - --zipkinAddress
        - "{{ .ProxyConfig.GetTracing.GetZipkin.GetAddress }}"
      {{- else if eq .Values.global.proxy.tracer "datadog" }}
        - --datadogAgentAddress
        - "{{ .ProxyConfig.GetTracing.GetDatadog.GetAddress }}"
      {{- end }}
      {{- if .Values.global.proxy.logLevel }}
        - --proxyLogLevel={{ .Values.global.proxy.logLevel }}
      {{- end}}
      {{- if .Values.global.proxy.componentLogLevel }}
        - --proxyComponentLogLevel={{ .Values.global.proxy.componentLogLevel }}
      {{- end}}
        - --dnsRefreshRate
        - {{ .Values.global.proxy.dnsRefreshRate }}
        - --connectTimeout
        - "{{ formatDuration .ProxyConfig.ConnectTimeout }}"
      {{- if .Values.global.proxy.envoyStatsd.enabled }}
        - --statsdUdpAddress
        - "{{ .ProxyConfig.StatsdUdpAddress }}"
      {{- end }}
      {{- if .Values.global.proxy.envoyMetricsService.enabled }}
        - --envoyMetricsServiceAddress
        - "{{ .ProxyConfig.GetEnvoyMetricsService.GetAddress }}"
      {{- end }}
      {{- if .Values.global.proxy.envoyAccessLogService.enabled }}
        - --envoyAccessLogService
        - '{{ structToJSON .ProxyConfig.EnvoyAccessLogService }}'
      {{- end }}
        - --proxyAdminPort
        - "{{ .ProxyConfig.ProxyAdminPort }}"
        {{ if gt .ProxyConfig.Concurrency 0 -}}
        - --concurrency
        - "{{ .ProxyConfig.Concurrency }}"
        {{ end -}}
        - --controlPlaneAuthPolicy
        - "{{ annotation .ObjectMeta `+"`"+`sidecar.istio.io/controlPlaneAuthPolicy`+"`"+` .ProxyConfig.ControlPlaneAuthPolicy }}"
      {{- if (ne (annotation .ObjectMeta "status.sidecar.istio.io/port" .Values.global.proxy.statusPort) "0") }}
        - --statusPort
        - "{{ annotation .ObjectMeta `+"`"+`status.sidecar.istio.io/port`+"`"+` .Values.global.proxy.statusPort }}"
        - --applicationPorts
        - "{{ annotation .ObjectMeta `+"`"+`readiness.status.sidecar.istio.io/applicationPorts`+"`"+` (applicationPorts .Spec.Containers) }}"
      {{- end }}
      {{- if .Values.global.trustDomain }}
        - --trust-domain={{ .Values.global.trustDomain }}
      {{- end }}
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: ISTIO_META_POD_PORTS
          value: |-
            [
            {{- range $index1, $c := .Spec.Containers }}
              {{- range $index2, $p := $c.Ports }}
                {{if or (ne $index1 0) (ne $index2 0)}},{{end}}{{ structToJSON $p }}
              {{- end}}
            {{- end}}
            ]
        - name: ISTIO_META_CLUSTER_ID
          value: "{{ valueOrDefault .Values.global.multicluster.clusterName `+"`"+`Kubernetes`+"`"+` }}"
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: INSTANCE_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        - name: SERVICE_ACCOUNT
          valueFrom:
            fieldRef:
              fieldPath: spec.serviceAccountName
      {{- if eq .Values.global.proxy.tracer "datadog" }}
        - name: HOST_IP
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
      {{- if isset .ObjectMeta.Annotations `+"`"+`apm.datadoghq.com/env`+"`"+` }}
      {{- range $key, $value := fromJSON (index .ObjectMeta.Annotations `+"`"+`apm.datadoghq.com/env`+"`"+`) }}
        - name: {{ $key }}
          value: "{{ $value }}"
      {{- end }}
      {{- end }}
      {{- end }}
        - name: ISTIO_META_POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: ISTIO_META_CONFIG_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: SDS_ENABLED
          value: {{ $.Values.global.sds.enabled }}
        - name: ISTIO_META_INTERCEPTION_MODE
          value: "{{ or (index .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/interceptionMode`+"`"+`) .ProxyConfig.InterceptionMode.String }}"
        - name: ISTIO_META_INCLUDE_INBOUND_PORTS
          value: "{{ annotation .ObjectMeta `+"`"+`traffic.sidecar.istio.io/includeInboundPorts`+"`"+` (applicationPorts .Spec.Containers) }}"
        {{- if .Values.global.network }}
        - name: ISTIO_META_NETWORK
          value: "{{ .Values.global.network }}"
        {{- end }}
        {{ if .ObjectMeta.Annotations }}
        - name: ISTIO_METAJSON_ANNOTATIONS
          value: |
                 {{ toJSON .ObjectMeta.Annotations }}
        {{ end }}
        {{ if .ObjectMeta.Labels }}
        - name: ISTIO_METAJSON_LABELS
          value: |
                 {{ toJSON .ObjectMeta.Labels }}
        {{ end }}
        {{- if .DeploymentMeta.Name }}
        - name: ISTIO_META_WORKLOAD_NAME
          value: {{ .DeploymentMeta.Name }}
        {{ end }}
        {{- if and .TypeMeta.APIVersion .DeploymentMeta.Name }}
        - name: ISTIO_META_OWNER
          value: kubernetes://api/{{ .TypeMeta.APIVersion }}/namespaces/{{ valueOrDefault .DeploymentMeta.Namespace `+"`"+`default`+"`"+` }}/{{ toLower .TypeMeta.Kind}}s/{{ .DeploymentMeta.Name }}
         {{- end}}
        {{- if (isset .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/bootstrapOverride`+"`"+`) }}
        - name: ISTIO_BOOTSTRAP_OVERRIDE
          value: "/etc/istio/custom-bootstrap/custom_bootstrap.json"
        {{- end }}
        {{- if .Values.global.sds.customTokenDirectory }}
        - name: ISTIO_META_SDS_TOKEN_PATH
          value: "{{ .Values.global.sds.customTokenDirectory -}}/sdstoken"
        {{- end }}
        {{- if .Values.global.meshID }}
        - name: ISTIO_META_MESH_ID
          value: "{{ .Values.global.meshID }}"
        {{- else if .Values.global.trustDomain }}
        - name: ISTIO_META_MESH_ID
          value: "{{ .Values.global.trustDomain }}"
        {{- end }}
        imagePullPolicy: {{ .Values.global.imagePullPolicy }}
        {{ if ne (annotation .ObjectMeta `+"`"+`status.sidecar.istio.io/port`+"`"+` .Values.global.proxy.statusPort) `+"`"+`0`+"`"+` }}
        readinessProbe:
          httpGet:
            path: /healthz/ready
            port: {{ annotation .ObjectMeta `+"`"+`status.sidecar.istio.io/port`+"`"+` .Values.global.proxy.statusPort }}
          initialDelaySeconds: {{ annotation .ObjectMeta `+"`"+`readiness.status.sidecar.istio.io/initialDelaySeconds`+"`"+` .Values.global.proxy.readinessInitialDelaySeconds }}
          periodSeconds: {{ annotation .ObjectMeta `+"`"+`readiness.status.sidecar.istio.io/periodSeconds`+"`"+` .Values.global.proxy.readinessPeriodSeconds }}
          failureThreshold: {{ annotation .ObjectMeta `+"`"+`readiness.status.sidecar.istio.io/failureThreshold`+"`"+` .Values.global.proxy.readinessFailureThreshold }}
        {{ end -}}
        securityContext:
          {{- if .Values.global.proxy.privileged }}
          privileged: true
          {{- end }}
          {{- if ne .Values.global.proxy.enableCoreDump true }}
          readOnlyRootFilesystem: true
          {{- end }}
          {{ if eq (annotation .ObjectMeta `+"`"+`sidecar.istio.io/interceptionMode`+"`"+` .ProxyConfig.InterceptionMode) `+"`"+`TPROXY`+"`"+` -}}
          capabilities:
            add:
            - NET_ADMIN
          runAsGroup: 1337
          {{ else -}}
          {{ if .Values.global.sds.enabled }}
          runAsGroup: 1337
          {{- end }}
          runAsUser: 1337
          {{- end }}
        resources:
          {{ if or (isset .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/proxyCPU`+"`"+`) (isset .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/proxyMemory`+"`"+`) -}}
          requests:
            {{ if (isset .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/proxyCPU`+"`"+`) -}}
            cpu: "{{ index .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/proxyCPU`+"`"+` }}"
            {{ end}}
            {{ if (isset .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/proxyMemory`+"`"+`) -}}
            memory: "{{ index .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/proxyMemory`+"`"+` }}"
            {{ end }}
        {{ else -}}
      {{- if .Values.global.proxy.resources }}
          {{ toYaml .Values.global.proxy.resources | indent 4 }}
      {{- end }}
        {{  end -}}
        volumeMounts:
        {{ if (isset .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/bootstrapOverride`+"`"+`) }}
        - mountPath: /etc/istio/custom-bootstrap
          name: custom-bootstrap-volume
        {{- end }}
        - mountPath: /etc/istio/proxy
          name: istio-envoy
        {{- if .Values.global.sds.enabled }}
        - mountPath: /var/run/sds
          name: sds-uds-path
          readOnly: true
        - mountPath: /var/run/secrets/tokens
          name: istio-token
        {{- if .Values.global.sds.customTokenDirectory }}
        - mountPath: "{{ .Values.global.sds.customTokenDirectory -}}"
          name: custom-sds-token
          readOnly: true
        {{- end }}
        {{- else }}
        - mountPath: /etc/certs/
          name: istio-certs
          readOnly: true
        {{- end }}
        {{- if and (eq .Values.global.proxy.tracer "lightstep") .Values.global.tracer.lightstep.cacertPath }}
        - mountPath: {{ directory .ProxyConfig.GetTracing.GetLightstep.GetCacertPath }}
          name: lightstep-certs
          readOnly: true
        {{- end }}
          {{- if isset .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/userVolumeMount`+"`"+` }}
          {{ range $index, $value := fromJSON (index .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/userVolumeMount`+"`"+`) }}
        - name: "{{  $index }}"
          {{ toYaml $value | indent 4 }}
          {{ end }}
          {{- end }}
      volumes:
      {{- if (isset .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/bootstrapOverride`+"`"+`) }}
      - name: custom-bootstrap-volume
        configMap:
          name: {{ annotation .ObjectMeta `+"`"+`sidecar.istio.io/bootstrapOverride`+"`"+` "" }}
      {{- end }}
      - emptyDir:
          medium: Memory
        name: istio-envoy
      {{- if .Values.global.sds.enabled }}
      - name: sds-uds-path
        hostPath:
          path: /var/run/sds
      - name: istio-token
        projected:
          sources:
            - serviceAccountToken:
                path: istio-token
                expirationSeconds: 43200
                audience: {{ .Values.global.sds.token.aud }}
      {{- if .Values.global.sds.customTokenDirectory }}
      - name: custom-sds-token
        secret:
          secretName: sdstokensecret
      {{- end }}
      {{- else }}
      - name: istio-certs
        secret:
          optional: true
          {{ if eq .Spec.ServiceAccountName "" }}
          secretName: istio.default
          {{ else -}}
          secretName: {{  printf "istio.%s" .Spec.ServiceAccountName }}
          {{  end -}}
        {{- if isset .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/userVolume`+"`"+` }}
        {{range $index, $value := fromJSON (index .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/userVolume`+"`"+`) }}
      - name: "{{ $index }}"
        {{ toYaml $value | indent 2 }}
        {{ end }}
        {{ end }}
      {{- end }}
      {{- if and (eq .Values.global.proxy.tracer "lightstep") .Values.global.tracer.lightstep.cacertPath }}
      - name: lightstep-certs
        secret:
          optional: true
          secretName: lightstep.cacert
      {{- end }}
      {{- if .Values.global.podDNSSearchNamespaces }}
      dnsConfig:
        searches:
          {{- range .Values.global.podDNSSearchNamespaces }}
          - {{ render . }}
          {{- end }}
      {{- end }}
      podRedirectAnnot:
         sidecar.istio.io/interceptionMode: "{{ annotation .ObjectMeta `+"`"+`sidecar.istio.io/interceptionMode`+"`"+` .ProxyConfig.InterceptionMode }}"
         traffic.sidecar.istio.io/includeOutboundIPRanges: "{{ annotation .ObjectMeta `+"`"+`traffic.sidecar.istio.io/includeOutboundIPRanges`+"`"+` .Values.global.proxy.includeIPRanges }}"
         traffic.sidecar.istio.io/excludeOutboundIPRanges: "{{ annotation .ObjectMeta `+"`"+`traffic.sidecar.istio.io/excludeOutboundIPRanges`+"`"+` .Values.global.proxy.excludeIPRanges }}"
         traffic.sidecar.istio.io/includeInboundPorts: "{{ annotation .ObjectMeta `+"`"+`traffic.sidecar.istio.io/includeInboundPorts`+"`"+` (includeInboundPorts .Spec.Containers) }}"
         traffic.sidecar.istio.io/excludeInboundPorts: "{{ excludeInboundPort (annotation .ObjectMeta `+"`"+`status.sidecar.istio.io/port`+"`"+` .Values.global.proxy.statusPort) (annotation .ObjectMeta `+"`"+`traffic.sidecar.istio.io/excludeInboundPorts`+"`"+` .Values.global.proxy.excludeInboundPorts) }}"
      {{ if or (isset .ObjectMeta.Annotations `+"`"+`traffic.sidecar.istio.io/excludeOutboundPorts`+"`"+`) (ne .Values.global.proxy.excludeOutboundPorts "") }}
         traffic.sidecar.istio.io/excludeOutboundPorts: "{{ annotation .ObjectMeta `+"`"+`traffic.sidecar.istio.io/excludeOutboundPorts`+"`"+` .Values.global.proxy.excludeOutboundPorts }}"
      {{- end }}
         traffic.sidecar.istio.io/kubevirtInterfaces: "{{ index .ObjectMeta.Annotations `+"`"+`traffic.sidecar.istio.io/kubevirtInterfaces`+"`"+` }}"

---

apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus
data:
  prometheus.yaml: |-
    global:
      scrape_interval: 15s
    scrape_configs:

    - job_name: 'istio-mesh'
      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - $(namespace)

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: istio-telemetry;prometheus

    # Scrape config for envoy stats
    - job_name: 'envoy-stats'
      metrics_path: /stats/prometheus
      kubernetes_sd_configs:
      - role: pod

      relabel_configs:
      - source_labels: [__meta_kubernetes_pod_container_port_name]
        action: keep
        regex: '.*-envoy-prom'
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:15090
        target_label: __address__
      - action: labelmap
        regex: __meta_kubernetes_pod_label_(.+)
      - source_labels: [__meta_kubernetes_namespace]
        action: replace
        target_label: namespace
      - source_labels: [__meta_kubernetes_pod_name]
        action: replace
        target_label: pod_name

    - job_name: 'istio-policy'
      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - $(namespace)


      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: istio-policy;http-monitoring

    - job_name: 'istio-telemetry'
      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - $(namespace)

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: istio-telemetry;http-monitoring

    - job_name: 'pilot'
      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - $(namespace)

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: istio-pilot;http-monitoring

    - job_name: 'galley'
      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - $(namespace)

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: istio-galley;http-monitoring

    - job_name: 'citadel'
      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - $(namespace)

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: istio-citadel;http-monitoring

    # scrape config for API servers
    - job_name: 'kubernetes-apiservers'
      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - default
      scheme: https
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
      bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: kubernetes;https

    # scrape config for nodes (kubelet)
    - job_name: 'kubernetes-nodes'
      scheme: https
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
      bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
      kubernetes_sd_configs:
      - role: node
      relabel_configs:
      - action: labelmap
        regex: __meta_kubernetes_node_label_(.+)
      - target_label: __address__
        replacement: kubernetes.default.svc:443
      - source_labels: [__meta_kubernetes_node_name]
        regex: (.+)
        target_label: __metrics_path__
        replacement: /api/v1/nodes/${1}/proxy/metrics

    # Scrape config for Kubelet cAdvisor.
    #
    # This is required for Kubernetes 1.7.3 and later, where cAdvisor metrics
    # (those whose names begin with 'container_') have been removed from the
    # Kubelet metrics endpoint.  This job scrapes the cAdvisor endpoint to
    # retrieve those metrics.
    #
    # In Kubernetes 1.7.0-1.7.2, these metrics are only exposed on the cAdvisor
    # HTTP endpoint; use "replacement: /api/v1/nodes/${1}:4194/proxy/metrics"
    # in that case (and ensure cAdvisor's HTTP server hasn't been disabled with
    # the --cadvisor-port=0 Kubelet flag).
    #
    # This job is not necessary and should be removed in Kubernetes 1.6 and
    # earlier versions, or it will cause the metrics to be scraped twice.
    - job_name: 'kubernetes-cadvisor'
      scheme: https
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
      bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
      kubernetes_sd_configs:
      - role: node
      relabel_configs:
      - action: labelmap
        regex: __meta_kubernetes_node_label_(.+)
      - target_label: __address__
        replacement: kubernetes.default.svc:443
      - source_labels: [__meta_kubernetes_node_name]
        regex: (.+)
        target_label: __metrics_path__
        replacement: /api/v1/nodes/${1}/proxy/metrics/cadvisor

    # scrape config for service endpoints.
    - job_name: 'kubernetes-service-endpoints'
      kubernetes_sd_configs:
      - role: endpoints
      relabel_configs:
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
        action: replace
        target_label: __scheme__
        regex: (https?)
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
        action: replace
        target_label: __address__
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
      - action: labelmap
        regex: __meta_kubernetes_service_label_(.+)
      - source_labels: [__meta_kubernetes_namespace]
        action: replace
        target_label: kubernetes_namespace
      - source_labels: [__meta_kubernetes_service_name]
        action: replace
        target_label: kubernetes_name

    - job_name: 'kubernetes-pods'
      kubernetes_sd_configs:
      - role: pod
      relabel_configs:  # If first two labels are present, pod should be scraped  by the istio-secure job.
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      # Keep target if there's no sidecar or if prometheus.io/scheme is explicitly set to "http"
      - source_labels: [__meta_kubernetes_pod_annotation_sidecar_istio_io_status, __meta_kubernetes_pod_annotation_prometheus_io_scheme]
        action: keep
        regex: ((;.*)|(.*;http))
      - source_labels: [__meta_kubernetes_pod_annotation_istio_mtls]
        action: drop
        regex: (true)
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__
      - action: labelmap
        regex: __meta_kubernetes_pod_label_(.+)
      - source_labels: [__meta_kubernetes_namespace]
        action: replace
        target_label: namespace
      - source_labels: [__meta_kubernetes_pod_name]
        action: replace
        target_label: pod_name

    - job_name: 'kubernetes-pods-istio-secure'
      scheme: https
      tls_config:
        ca_file: /etc/istio-certs/root-cert.pem
        cert_file: /etc/istio-certs/cert-chain.pem
        key_file: /etc/istio-certs/key.pem
        insecure_skip_verify: true  # prometheus does not support secure naming.
      kubernetes_sd_configs:
      - role: pod
      relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      # sidecar status annotation is added by sidecar injector and
      # istio_workload_mtls_ability can be specifically placed on a pod to indicate its ability to receive mtls traffic.
      - source_labels: [__meta_kubernetes_pod_annotation_sidecar_istio_io_status, __meta_kubernetes_pod_annotation_istio_mtls]
        action: keep
        regex: (([^;]+);([^;]*))|(([^;]*);(true))
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scheme]
        action: drop
        regex: (http)
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__]  # Only keep address that is host:port
        action: keep    # otherwise an extra target with ':443' is added for https scheme
        regex: ([^:]+):(\d+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__
      - action: labelmap
        regex: __meta_kubernetes_pod_label_(.+)
      - source_labels: [__meta_kubernetes_namespace]
        action: replace
        target_label: namespace
      - source_labels: [__meta_kubernetes_pod_name]
        action: replace
        target_label: pod_name

---

apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-galley-configuration
data:
  validating-webhook-configuration.yaml: |
    apiVersion: admissionregistration.k8s.io/v1beta1
    kind: ValidatingWebhookConfiguration
    metadata:
      name: istio-galley
      labels:
        app: galley
        istio: galley
    webhooks:
      - name: pilot.validation.istio.io
        clientConfig:
          service:
            name: istio-galley
            namespace: $(namespace)
            path: "/admitpilot"
          caBundle: ""
        rules:
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - config.istio.io
            apiVersions:
            - v1alpha2
            resources:
            - httpapispecs
            - httpapispecbindings
            - quotaspecs
            - quotaspecbindings
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - rbac.istio.io
            apiVersions:
            - "*"
            resources:
            - "*"
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - authentication.istio.io
            apiVersions:
            - "*"
            resources:
            - "*"
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - networking.istio.io
            apiVersions:
            - "*"
            resources:
            - destinationrules
            - envoyfilters
            - gateways
            - serviceentries
            - sidecars
            - virtualservices
        failurePolicy: Fail
        sideEffects: None
      - name: mixer.validation.istio.io
        clientConfig:
          service:
            name: istio-galley
            namespace: $(namespace)
            path: "/admitmixer"
          caBundle: ""
        rules:
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - config.istio.io
            apiVersions:
            - v1alpha2
            resources:
            - rules
            - attributemanifests
            - circonuses
            - deniers
            - fluentds
            - kubernetesenvs
            - listcheckers
            - memquotas
            - noops
            - opas
            - prometheuses
            - rbacs
            - solarwindses
            - stackdrivers
            - cloudwatches
            - dogstatsds
            - statsds
            - stdios
            - apikeys
            - authorizations
            - checknothings
            # - kuberneteses
            - listentries
            - logentries
            - metrics
            - quotas
            - reportnothings
            - tracespans
            - adapters
            - handlers
            - instances
            - templates
            - zipkins
        failurePolicy: Fail
        sideEffects: None

---

apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-security-custom-resources
data:
  istio-security-custom-resources.yaml: |
    # Authentication policy to enable permissive mode for all services (that have sidecar) in the mesh.
    apiVersion: "authentication.istio.io/v1alpha1"
    kind: "MeshPolicy"
    metadata:
      name: "default"
      labels:
        app: security
    spec:
      peers:
      - mtls:
          mode: PERMISSIVE
  istio-security-run.sh: |-
    #!/bin/sh

    set -x

    if [ "$#" -ne "1" ]; then
        echo "first argument should be path to custom resource yaml"
        exit 1
    fi

    pathToResourceYAML=${1}

    kubectl get validatingwebhookconfiguration istio-galley 2>/dev/null
    if [ "$?" -eq 0 ]; then
        echo "istio-galley validatingwebhookconfiguration found - waiting for istio-galley deployment to be ready"
        while true; do
            kubectl -n $(namespace) get deployment istio-galley 2>/dev/null
            if [ "$?" -eq 0 ]; then
                break
            fi
            sleep 1
        done
        kubectl -n $(namespace) rollout status deployment istio-galley
        if [ "$?" -ne 0 ]; then
            echo "istio-galley deployment rollout status check failed"
            exit 1
        fi
        echo "istio-galley deployment ready for configuration validation"
    fi
    sleep 5
    kubectl apply -f ${pathToResourceYAML}
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app: security
  name: istio-citadel-$(namespace)
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - create
  - get
  - update
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - create
  - get
  - watch
  - list
  - update
  - delete
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  - services
  - namespaces
  verbs:
  - get
  - watch
  - list
- apiGroups:
  - authentication.k8s.io
  resources:
  - tokenreviews
  verbs:
  - create

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app: galley
  name: istio-galley-$(namespace)
rules:
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - validatingwebhookconfigurations
  verbs:
  - '*'
- apiGroups:
  - config.istio.io
  resources:
  - '*'
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - networking.istio.io
  resources:
  - '*'
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - authentication.istio.io
  resources:
  - '*'
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - rbac.istio.io
  resources:
  - '*'
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - extensions
  - apps
  resourceNames:
  - istio-galley
  resources:
  - deployments
  verbs:
  - get
- apiGroups:
  - ""
  resources:
  - pods
  - nodes
  - services
  - endpoints
  - namespaces
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - extensions
  resources:
  - ingresses
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - extensions
  resourceNames:
  - istio-galley
  resources:
  - deployments/finalizers
  verbs:
  - update
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - get
  - list
  - watch

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app: mixer
  name: istio-mixer-$(namespace)
rules:
- apiGroups:
  - config.istio.io
  resources:
  - '*'
  verbs:
  - create
  - get
  - list
  - watch
  - patch
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - configmaps
  - endpoints
  - pods
  - services
  - namespaces
  - secrets
  - replicationcontrollers
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - extensions
  - apps
  resources:
  - replicasets
  verbs:
  - get
  - list
  - watch

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app: nodeagent
  name: istio-nodeagent-$(namespace)
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app: pilot
  name: istio-pilot-$(namespace)
rules:
- apiGroups:
  - config.istio.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - rbac.istio.io
  resources:
  - '*'
  verbs:
  - get
  - watch
  - list
- apiGroups:
  - networking.istio.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - authentication.istio.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - '*'
- apiGroups:
  - extensions
  resources:
  - ingresses
  - ingresses/status
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - create
  - get
  - list
  - watch
  - update
- apiGroups:
  - ""
  resources:
  - endpoints
  - pods
  - services
  - namespaces
  - nodes
  - secrets
  verbs:
  - get
  - list
  - watch

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: istio-reader
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  - pods
  - services
  - endpoints
  - replicationcontrollers
  verbs:
  - get
  - watch
  - list
- apiGroups:
  - extensions
  - apps
  resources:
  - replicasets
  verbs:
  - get
  - list
  - watch

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    app: security
  name: istio-security-post-install-$(namespace)
rules:
- apiGroups:
  - authentication.istio.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - networking.istio.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - validatingwebhookconfigurations
  verbs:
  - get
- apiGroups:
  - extensions
  - apps
  resources:
  - deployments
  - replicasets
  verbs:
  - get
  - list
  - watch

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app: sidecarInjectorWebhook
    istio: sidecar-injector
  name: istio-sidecar-injector-$(namespace)
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - mutatingwebhookconfigurations
  verbs:
  - get
  - list
  - watch
  - patch

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app: prometheus
  name: prometheus-$(namespace)
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  - services
  - endpoints
  - pods
  - nodes/proxy
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
- nonResourceURLs:
  - /metrics
  verbs:
  - get
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: security
  name: istio-citadel-$(namespace)
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-citadel-$(namespace)
subjects:
- kind: ServiceAccount
  name: istio-citadel-service-account
  namespace: $(namespace)

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: galley
  name: istio-galley-admin-role-binding-$(namespace)
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-galley-$(namespace)
subjects:
- kind: ServiceAccount
  name: istio-galley-service-account
  namespace: $(namespace)

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: mixer
  name: istio-mixer-admin-role-binding-$(namespace)
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-mixer-$(namespace)
subjects:
- kind: ServiceAccount
  name: istio-mixer-service-account
  namespace: $(namespace)

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: istio-multi
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-reader
subjects:
- kind: ServiceAccount
  name: istio-multi
  namespace: $(namespace)

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: nodeagent
  name: istio-nodeagent-$(namespace)
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-nodeagent-$(namespace)
subjects:
- kind: ServiceAccount
  name: istio-nodeagent-service-account
  namespace: $(namespace)

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: pilot
  name: istio-pilot-$(namespace)
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-pilot-$(namespace)
subjects:
- kind: ServiceAccount
  name: istio-pilot-service-account
  namespace: $(namespace)

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    app: security
  name: istio-security-post-install-role-binding-$(namespace)
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-security-post-install-$(namespace)
subjects:
- kind: ServiceAccount
  name: istio-security-post-install-account
  namespace: $(namespace)

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: sidecarInjectorWebhook
    istio: sidecar-injector
  name: istio-sidecar-injector-admin-role-binding-$(namespace)
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-sidecar-injector-$(namespace)
subjects:
- kind: ServiceAccount
  name: istio-sidecar-injector-service-account
  namespace: $(namespace)

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: prometheus
  name: prometheus-$(namespace)
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus-$(namespace)
subjects:
- kind: ServiceAccount
  name: prometheus
  namespace: $(namespace)
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/daemon-set.yaml", `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app: nodeagent
    istio: nodeagent
  name: istio-nodeagent
spec:
  selector:
    matchLabels:
      istio: nodeagent
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
      labels:
        app: nodeagent
        istio: nodeagent
    spec:
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x
            weight: 2
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
      containers:
      - env:
        - name: CA_ADDR
          value: istio-citadel:8060
        - name: CA_PROVIDER
          value: Citadel
        - name: PLUGINS
          value: ""
        - name: VALID_TOKEN
          value: "true"
        - name: TRUST_DOMAIN
          value: ""
        - name: NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: gcr.io/istio-release/node-agent-k8s:release-1.3-latest-daily
        imagePullPolicy: IfNotPresent
        name: nodeagent
        volumeMounts:
        - mountPath: /var/run/sds
          name: sdsudspath
      serviceAccountName: istio-nodeagent-service-account
      volumes:
      - hostPath:
          path: /var/run/sds
        name: sdsudspath
  updateStrategy:
    type: RollingUpdate
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: security
    istio: citadel
  name: istio-citadel
spec:
  replicas: 1
  selector:
    matchLabels:
      istio: citadel
  strategy:
    rollingUpdate:
      maxSurge: 100%
      maxUnavailable: 25%
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
      labels:
        app: security
        istio: citadel
    spec:
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x
            weight: 2
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
      containers:
      - args:
        - --sds-enabled=true
        - --append-dns-names=true
        - --grpc-port=8060
        - --citadel-storage-namespace=$(namespace)
        - --custom-dns-names=istio-pilot-service-account.$(namespace):istio-pilot.$(namespace)
        - --monitoring-port=15014
        - --self-signed-ca=true
        - --workload-cert-ttl=2160h
        env:
        - name: CITADEL_ENABLE_NAMESPACES_BY_DEFAULT
          value: "true"
        image: gcr.io/istio-release/citadel:release-1.3-latest-daily
        imagePullPolicy: IfNotPresent
        name: citadel
        resources:
          requests:
            cpu: 10m
      serviceAccountName: istio-citadel-service-account

---

apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: galley
    istio: galley
  name: istio-galley
spec:
  replicas: 1
  selector:
    matchLabels:
      istio: galley
  strategy:
    rollingUpdate:
      maxSurge: 100%
      maxUnavailable: 25%
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
      labels:
        app: galley
        istio: galley
    spec:
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x
            weight: 2
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
      containers:
      - command:
        - /usr/local/bin/galley
        - server
        - --meshConfigFile=/etc/mesh-config/mesh
        - --livenessProbeInterval=1s
        - --livenessProbePath=/healthliveness
        - --readinessProbePath=/healthready
        - --readinessProbeInterval=1s
        - --deployment-namespace=$(namespace)
        - --insecure=false
        - --validation-webhook-config-file
        - /etc/config/validating-webhook-configuration.yaml
        - --monitoringPort=15014
        - --log_output_level=default:info
        image: gcr.io/istio-release/galley:release-1.3-latest-daily
        imagePullPolicy: IfNotPresent
        livenessProbe:
          exec:
            command:
            - /usr/local/bin/galley
            - probe
            - --probe-path=/healthliveness
            - --interval=10s
          initialDelaySeconds: 5
          periodSeconds: 5
        name: galley
        ports:
        - containerPort: 443
        - containerPort: 15014
        - containerPort: 9901
        readinessProbe:
          exec:
            command:
            - /usr/local/bin/galley
            - probe
            - --probe-path=/healthready
            - --interval=10s
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            cpu: 10m
        volumeMounts:
        - mountPath: /etc/certs
          name: certs
          readOnly: true
        - mountPath: /etc/config
          name: config
          readOnly: true
        - mountPath: /etc/mesh-config
          name: mesh-config
          readOnly: true
      serviceAccountName: istio-galley-service-account
      volumes:
      - name: certs
        secret:
          secretName: istio.istio-galley-service-account
      - configMap:
          name: istio-galley-configuration
        name: config
      - configMap:
          name: istio
        name: mesh-config

---

apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: istio-ingressgateway
    istio: ingressgateway
  name: istio-ingressgateway
spec:
  selector:
    matchLabels:
      app: istio-ingressgateway
      istio: ingressgateway
  strategy:
    rollingUpdate:
      maxSurge: 100%
      maxUnavailable: 25%
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
      labels:
        app: istio-ingressgateway
        istio: ingressgateway
    spec:
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x
            weight: 2
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
      containers:
      - args:
        - proxy
        - router
        - --domain
        - $(POD_NAMESPACE).svc.cluster.local
        - --log_output_level=default:info
        - --drainDuration
        - 45s
        - --parentShutdownDuration
        - 1m0s
        - --connectTimeout
        - 10s
        - --serviceCluster
        - istio-ingressgateway
        - --zipkinAddress
        - zipkin:9411
        - --proxyAdminPort
        - "15000"
        - --statusPort
        - "15020"
        - --controlPlaneAuthPolicy
        - MUTUAL_TLS
        - --discoveryAddress
        - istio-pilot:15011
        env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: POD_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: INSTANCE_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.podIP
        - name: HOST_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.hostIP
        - name: SERVICE_ACCOUNT
          valueFrom:
            fieldRef:
              fieldPath: spec.serviceAccountName
        - name: ISTIO_META_POD_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
        - name: ISTIO_META_CONFIG_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: SDS_ENABLED
          value: "true"
        - name: ISTIO_META_WORKLOAD_NAME
          value: istio-ingressgateway
        - name: ISTIO_META_OWNER
          value: kubernetes://api/apps/v1/namespaces/$(namespace)/deployments/istio-ingressgateway
        - name: ISTIO_META_ROUTER_MODE
          value: sni-dnat
        image: gcr.io/istio-release/proxyv2:release-1.3-latest-daily
        imagePullPolicy: IfNotPresent
        name: istio-proxy
        ports:
        - containerPort: 15020
        - containerPort: 80
        - containerPort: 443
        - containerPort: 31400
        - containerPort: 15029
        - containerPort: 15030
        - containerPort: 15031
        - containerPort: 15032
        - containerPort: 15443
        - containerPort: 15090
          name: http-envoy-prom
          protocol: TCP
        readinessProbe:
          failureThreshold: 30
          httpGet:
            path: /healthz/ready
            port: 15020
            scheme: HTTP
          initialDelaySeconds: 1
          periodSeconds: 2
          successThreshold: 1
          timeoutSeconds: 1
        resources:
          limits:
            cpu: 2000m
            memory: 1024Mi
          requests:
            cpu: 100m
            memory: 128Mi
        volumeMounts:
        - mountPath: /var/run/sds
          name: sdsudspath
          readOnly: true
        - mountPath: /var/run/secrets/tokens
          name: istio-token
        - mountPath: /etc/certs
          name: istio-certs
          readOnly: true
        - mountPath: /etc/istio/ingressgateway-certs
          name: ingressgateway-certs
          readOnly: true
        - mountPath: /etc/istio/ingressgateway-ca-certs
          name: ingressgateway-ca-certs
          readOnly: true
      serviceAccountName: istio-ingressgateway-service-account
      volumes:
      - hostPath:
          path: /var/run/sds
        name: sdsudspath
      - name: istio-token
        projected:
          sources:
          - serviceAccountToken:
              audience: istio-ca
              expirationSeconds: 43200
              path: istio-token
      - name: istio-certs
        secret:
          optional: true
          secretName: istio.istio-ingressgateway-service-account
      - name: ingressgateway-certs
        secret:
          optional: true
          secretName: istio-ingressgateway-certs
      - name: ingressgateway-ca-certs
        secret:
          optional: true
          secretName: istio-ingressgateway-ca-certs

---

apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    checksum/config-volume: f8da08b6b8c170dde721efd680270b2901e750d4aa186ebb6c22bef5b78a43f9
  labels:
    app: pilot
    istio: pilot
  name: istio-pilot
spec:
  selector:
    matchLabels:
      istio: pilot
  strategy:
    rollingUpdate:
      maxSurge: 100%
      maxUnavailable: 25%
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
      labels:
        app: pilot
        istio: pilot
    spec:
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x
            weight: 2
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
      containers:
      - args:
        - discovery
        - --monitoringAddr=:15014
        - --log_output_level=default:info
        - --domain
        - cluster.local
        - --secureGrpcAddr
        - ""
        - --keepaliveMaxServerConnectionAge
        - 30m
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: GODEBUG
          value: gctrace=1
        - name: PILOT_PUSH_THROTTLE
          value: "100"
        - name: PILOT_TRACE_SAMPLING
          value: "1"
        - name: PILOT_ENABLE_PROTOCOL_SNIFFING_FOR_OUTBOUND
          value: "true"
        - name: PILOT_ENABLE_PROTOCOL_SNIFFING_FOR_INBOUND
          value: "false"
        image: gcr.io/istio-release/pilot:release-1.3-latest-daily
        imagePullPolicy: IfNotPresent
        name: discovery
        ports:
        - containerPort: 8080
        - containerPort: 15010
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 30
          timeoutSeconds: 5
        resources:
          requests:
            cpu: 500m
            memory: 2048Mi
        volumeMounts:
        - mountPath: /etc/istio/config
          name: config-volume
        - mountPath: /etc/certs
          name: istio-certs
          readOnly: true
      - args:
        - proxy
        - --domain
        - $(POD_NAMESPACE).svc.cluster.local
        - --serviceCluster
        - istio-pilot
        - --templateFile
        - /etc/istio/proxy/envoy_pilot.yaml.tmpl
        - --controlPlaneAuthPolicy
        - MUTUAL_TLS
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: INSTANCE_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.podIP
        - name: SDS_ENABLED
          value: "true"
        image: gcr.io/istio-release/proxyv2:release-1.3-latest-daily
        imagePullPolicy: IfNotPresent
        name: istio-proxy
        ports:
        - containerPort: 15003
        - containerPort: 15005
        - containerPort: 15007
        - containerPort: 15011
        resources:
          limits:
            cpu: 2000m
            memory: 1024Mi
          requests:
            cpu: 100m
            memory: 128Mi
        volumeMounts:
        - mountPath: /etc/certs
          name: istio-certs
          readOnly: true
        - mountPath: /var/run/sds
          name: sds-uds-path
          readOnly: true
        - mountPath: /var/run/secrets/tokens
          name: istio-token
      serviceAccountName: istio-pilot-service-account
      volumes:
      - hostPath:
          path: /var/run/sds
        name: sds-uds-path
      - name: istio-token
        projected:
          sources:
          - serviceAccountToken:
              audience: istio-ca
              expirationSeconds: 43200
              path: istio-token
      - configMap:
          name: istio
        name: config-volume
      - name: istio-certs
        secret:
          optional: true
          secretName: istio.istio-pilot-service-account

---

apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: istio-mixer
    istio: mixer
  name: istio-policy
spec:
  selector:
    matchLabels:
      istio: mixer
      istio-mixer-type: policy
  strategy:
    rollingUpdate:
      maxSurge: 100%
      maxUnavailable: 25%
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
      labels:
        app: policy
        istio: mixer
        istio-mixer-type: policy
    spec:
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x
            weight: 2
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
      containers:
      - args:
        - --monitoringPort=15014
        - --address
        - unix:///sock/mixer.socket
        - --log_output_level=default:info
        - --configStoreURL=mcps://istio-galley.$(namespace).svc:9901
        - --configDefaultNamespace=$(namespace)
        - --useAdapterCRDs=false
        - --useTemplateCRDs=false
        - --trace_zipkin_url=http://zipkin.$(namespace):9411/api/v1/spans
        env:
        - name: GODEBUG
          value: gctrace=1
        - name: GOMAXPROCS
          value: "6"
        image: gcr.io/istio-release/mixer:release-1.3-latest-daily
        imagePullPolicy: IfNotPresent
        livenessProbe:
          httpGet:
            path: /version
            port: 15014
          initialDelaySeconds: 5
          periodSeconds: 5
        name: mixer
        ports:
        - containerPort: 15014
        - containerPort: 42422
        resources:
          requests:
            cpu: 10m
        volumeMounts:
        - mountPath: /etc/certs
          name: istio-certs
          readOnly: true
        - mountPath: /sock
          name: uds-socket
      - args:
        - proxy
        - --domain
        - $(POD_NAMESPACE).svc.cluster.local
        - --serviceCluster
        - istio-policy
        - --templateFile
        - /etc/istio/proxy/envoy_policy.yaml.tmpl
        - --controlPlaneAuthPolicy
        - MUTUAL_TLS
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: INSTANCE_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.podIP
        - name: SDS_ENABLED
          value: "true"
        image: gcr.io/istio-release/proxyv2:release-1.3-latest-daily
        imagePullPolicy: IfNotPresent
        name: istio-proxy
        ports:
        - containerPort: 9091
        - containerPort: 15004
        - containerPort: 15090
          name: http-envoy-prom
          protocol: TCP
        resources:
          limits:
            cpu: 2000m
            memory: 1024Mi
          requests:
            cpu: 100m
            memory: 128Mi
        volumeMounts:
        - mountPath: /etc/certs
          name: istio-certs
          readOnly: true
        - mountPath: /var/run/sds
          name: sds-uds-path
          readOnly: true
        - mountPath: /var/run/secrets/tokens
          name: istio-token
        - mountPath: /sock
          name: uds-socket
        - mountPath: /var/run/secrets/istio.io/policy/adapter
          name: policy-adapter-secret
          readOnly: true
      serviceAccountName: istio-mixer-service-account
      volumes:
      - name: istio-certs
        secret:
          optional: true
          secretName: istio.istio-mixer-service-account
      - hostPath:
          path: /var/run/sds
        name: sds-uds-path
      - name: istio-token
        projected:
          sources:
          - serviceAccountToken:
              audience: istio-ca
              expirationSeconds: 43200
              path: istio-token
      - name: uds-socket
      - name: policy-adapter-secret
        secret:
          optional: true
          secretName: policy-adapter-secret

---

apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: sidecarInjectorWebhook
    istio: sidecar-injector
  name: istio-sidecar-injector
spec:
  replicas: 1
  selector:
    matchLabels:
      istio: sidecar-injector
  strategy:
    rollingUpdate:
      maxSurge: 100%
      maxUnavailable: 25%
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
      labels:
        app: sidecarInjectorWebhook
        istio: sidecar-injector
    spec:
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x
            weight: 2
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
      containers:
      - args:
        - --caCertFile=/etc/istio/certs/root-cert.pem
        - --tlsCertFile=/etc/istio/certs/cert-chain.pem
        - --tlsKeyFile=/etc/istio/certs/key.pem
        - --injectConfig=/etc/istio/inject/config
        - --meshConfig=/etc/istio/config/mesh
        - --healthCheckInterval=2s
        - --healthCheckFile=/health
        image: gcr.io/istio-release/sidecar_injector:release-1.3-latest-daily
        imagePullPolicy: IfNotPresent
        livenessProbe:
          exec:
            command:
            - /usr/local/bin/sidecar-injector
            - probe
            - --probe-path=/health
            - --interval=4s
          initialDelaySeconds: 4
          periodSeconds: 4
        name: sidecar-injector-webhook
        readinessProbe:
          exec:
            command:
            - /usr/local/bin/sidecar-injector
            - probe
            - --probe-path=/health
            - --interval=4s
          initialDelaySeconds: 4
          periodSeconds: 4
        resources:
          requests:
            cpu: 10m
        volumeMounts:
        - mountPath: /etc/istio/config
          name: config-volume
          readOnly: true
        - mountPath: /etc/istio/certs
          name: certs
          readOnly: true
        - mountPath: /etc/istio/inject
          name: inject-config
          readOnly: true
      serviceAccountName: istio-sidecar-injector-service-account
      volumes:
      - configMap:
          name: istio
        name: config-volume
      - name: certs
        secret:
          secretName: istio.istio-sidecar-injector-service-account
      - configMap:
          items:
          - key: config
            path: config
          - key: values
            path: values
          name: istio-sidecar-injector
        name: inject-config

---

apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: istio-mixer
    istio: mixer
  name: istio-telemetry
spec:
  selector:
    matchLabels:
      istio: mixer
      istio-mixer-type: telemetry
  strategy:
    rollingUpdate:
      maxSurge: 100%
      maxUnavailable: 25%
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
      labels:
        app: telemetry
        istio: mixer
        istio-mixer-type: telemetry
    spec:
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x
            weight: 2
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
      containers:
      - args:
        - --monitoringPort=15014
        - --address
        - unix:///sock/mixer.socket
        - --log_output_level=default:info
        - --configStoreURL=mcps://istio-galley.$(namespace).svc:9901
        - --certFile=/etc/certs/cert-chain.pem
        - --keyFile=/etc/certs/key.pem
        - --caCertFile=/etc/certs/root-cert.pem
        - --configDefaultNamespace=$(namespace)
        - --useAdapterCRDs=false
        - --trace_zipkin_url=http://zipkin.$(namespace):9411/api/v1/spans
        - --averageLatencyThreshold
        - 100ms
        - --loadsheddingMode
        - enforce
        env:
        - name: GODEBUG
          value: gctrace=1
        - name: GOMAXPROCS
          value: "6"
        image: gcr.io/istio-release/mixer:release-1.3-latest-daily
        imagePullPolicy: IfNotPresent
        livenessProbe:
          httpGet:
            path: /version
            port: 15014
          initialDelaySeconds: 5
          periodSeconds: 5
        name: mixer
        ports:
        - containerPort: 15014
        - containerPort: 42422
        resources:
          limits:
            cpu: 4800m
            memory: 4G
          requests:
            cpu: 1000m
            memory: 1G
        volumeMounts:
        - mountPath: /etc/certs
          name: istio-certs
          readOnly: true
        - mountPath: /var/run/secrets/istio.io/telemetry/adapter
          name: telemetry-adapter-secret
          readOnly: true
        - mountPath: /sock
          name: uds-socket
      - args:
        - proxy
        - --domain
        - $(POD_NAMESPACE).svc.cluster.local
        - --serviceCluster
        - istio-telemetry
        - --templateFile
        - /etc/istio/proxy/envoy_telemetry.yaml.tmpl
        - --controlPlaneAuthPolicy
        - MUTUAL_TLS
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: INSTANCE_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.podIP
        - name: SDS_ENABLED
          value: "true"
        image: gcr.io/istio-release/proxyv2:release-1.3-latest-daily
        imagePullPolicy: IfNotPresent
        name: istio-proxy
        ports:
        - containerPort: 9091
        - containerPort: 15004
        - containerPort: 15090
          name: http-envoy-prom
          protocol: TCP
        resources:
          limits:
            cpu: 2000m
            memory: 1024Mi
          requests:
            cpu: 100m
            memory: 128Mi
        volumeMounts:
        - mountPath: /etc/certs
          name: istio-certs
          readOnly: true
        - mountPath: /var/run/sds
          name: sds-uds-path
          readOnly: true
        - mountPath: /var/run/secrets/tokens
          name: istio-token
        - mountPath: /sock
          name: uds-socket
      serviceAccountName: istio-mixer-service-account
      volumes:
      - name: istio-certs
        secret:
          optional: true
          secretName: istio.istio-mixer-service-account
      - hostPath:
          path: /var/run/sds
        name: sds-uds-path
      - name: istio-token
        projected:
          sources:
          - serviceAccountToken:
              audience: istio-ca
              expirationSeconds: 43200
              path: istio-token
      - name: uds-socket
      - name: telemetry-adapter-secret
        secret:
          optional: true
          secretName: telemetry-adapter-secret

---

apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: prometheus
  name: prometheus
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
      labels:
        app: prometheus
    spec:
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x
            weight: 2
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
      containers:
      - args:
        - --storage.tsdb.retention=6h
        - --config.file=/etc/prometheus/prometheus.yaml
        image: docker.io/prom/prometheus:v2.8.0
        imagePullPolicy: IfNotPresent
        livenessProbe:
          httpGet:
            path: /-/healthy
            port: 9090
        name: prometheus
        ports:
        - containerPort: 9090
          name: http
        readinessProbe:
          httpGet:
            path: /-/ready
            port: 9090
        resources:
          requests:
            cpu: 10m
        volumeMounts:
        - mountPath: /etc/prometheus
          name: config-volume
        - mountPath: /etc/istio-certs
          name: istio-certs
      serviceAccountName: prometheus
      volumes:
      - configMap:
          name: prometheus
        name: config-volume
      - name: istio-certs
        secret:
          defaultMode: 420
          secretName: istio.default
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/handler.yaml", `
apiVersion: config.istio.io/v1alpha2
kind: handler
metadata:
  labels:
    app: mixer
  name: kubernetesenv
spec:
  compiledAdapter: kubernetesenv

---

apiVersion: config.istio.io/v1alpha2
kind: handler
metadata:
  labels:
    app: mixer
  name: prometheus
spec:
  compiledAdapter: prometheus
  params:
    metrics:
    - instance_name: requestcount.instance.$(namespace)
      kind: COUNTER
      label_names:
      - reporter
      - source_app
      - source_principal
      - source_workload
      - source_workload_namespace
      - source_version
      - destination_app
      - destination_principal
      - destination_workload
      - destination_workload_namespace
      - destination_version
      - destination_service
      - destination_service_name
      - destination_service_namespace
      - request_protocol
      - response_code
      - response_flags
      - permissive_response_code
      - permissive_response_policyid
      - connection_security_policy
      name: requests_total
    - buckets:
        explicit_buckets:
          bounds:
          - 0.005
          - 0.01
          - 0.025
          - 0.05
          - 0.1
          - 0.25
          - 0.5
          - 1
          - 2.5
          - 5
          - 10
      instance_name: requestduration.instance.$(namespace)
      kind: DISTRIBUTION
      label_names:
      - reporter
      - source_app
      - source_principal
      - source_workload
      - source_workload_namespace
      - source_version
      - destination_app
      - destination_principal
      - destination_workload
      - destination_workload_namespace
      - destination_version
      - destination_service
      - destination_service_name
      - destination_service_namespace
      - request_protocol
      - response_code
      - response_flags
      - permissive_response_code
      - permissive_response_policyid
      - connection_security_policy
      name: request_duration_seconds
    - buckets:
        exponentialBuckets:
          growthFactor: 10
          numFiniteBuckets: 8
          scale: 1
      instance_name: requestsize.instance.$(namespace)
      kind: DISTRIBUTION
      label_names:
      - reporter
      - source_app
      - source_principal
      - source_workload
      - source_workload_namespace
      - source_version
      - destination_app
      - destination_principal
      - destination_workload
      - destination_workload_namespace
      - destination_version
      - destination_service
      - destination_service_name
      - destination_service_namespace
      - request_protocol
      - response_code
      - response_flags
      - permissive_response_code
      - permissive_response_policyid
      - connection_security_policy
      name: request_bytes
    - buckets:
        exponentialBuckets:
          growthFactor: 10
          numFiniteBuckets: 8
          scale: 1
      instance_name: responsesize.instance.$(namespace)
      kind: DISTRIBUTION
      label_names:
      - reporter
      - source_app
      - source_principal
      - source_workload
      - source_workload_namespace
      - source_version
      - destination_app
      - destination_principal
      - destination_workload
      - destination_workload_namespace
      - destination_version
      - destination_service
      - destination_service_name
      - destination_service_namespace
      - request_protocol
      - response_code
      - response_flags
      - permissive_response_code
      - permissive_response_policyid
      - connection_security_policy
      name: response_bytes
    - instance_name: tcpbytesent.instance.$(namespace)
      kind: COUNTER
      label_names:
      - reporter
      - source_app
      - source_principal
      - source_workload
      - source_workload_namespace
      - source_version
      - destination_app
      - destination_principal
      - destination_workload
      - destination_workload_namespace
      - destination_version
      - destination_service
      - destination_service_name
      - destination_service_namespace
      - connection_security_policy
      - response_flags
      name: tcp_sent_bytes_total
    - instance_name: tcpbytereceived.instance.$(namespace)
      kind: COUNTER
      label_names:
      - reporter
      - source_app
      - source_principal
      - source_workload
      - source_workload_namespace
      - source_version
      - destination_app
      - destination_principal
      - destination_workload
      - destination_workload_namespace
      - destination_version
      - destination_service
      - destination_service_name
      - destination_service_namespace
      - connection_security_policy
      - response_flags
      name: tcp_received_bytes_total
    - instance_name: tcpconnectionsopened.instance.$(namespace)
      kind: COUNTER
      label_names:
      - reporter
      - source_app
      - source_principal
      - source_workload
      - source_workload_namespace
      - source_version
      - destination_app
      - destination_principal
      - destination_workload
      - destination_workload_namespace
      - destination_version
      - destination_service
      - destination_service_name
      - destination_service_namespace
      - connection_security_policy
      - response_flags
      name: tcp_connections_opened_total
    - instance_name: tcpconnectionsclosed.instance.$(namespace)
      kind: COUNTER
      label_names:
      - reporter
      - source_app
      - source_principal
      - source_workload
      - source_workload_namespace
      - source_version
      - destination_app
      - destination_principal
      - destination_workload
      - destination_workload_namespace
      - destination_version
      - destination_service
      - destination_service_name
      - destination_service_namespace
      - connection_security_policy
      - response_flags
      name: tcp_connections_closed_total
    metricsExpirationPolicy:
      metricsExpiryDuration: 10m
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/horizontal-pod-autoscaler.yaml", `
apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
  labels:
    app: istio-ingressgateway
    istio: ingressgateway
  name: istio-ingressgateway
spec:
  maxReplicas: 5
  metrics:
  - resource:
      name: cpu
      targetAverageUtilization: 80
    type: Resource
  minReplicas: 1
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: istio-ingressgateway

---

apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
  labels:
    app: pilot
  name: istio-pilot
spec:
  maxReplicas: 5
  metrics:
  - resource:
      name: cpu
      targetAverageUtilization: 80
    type: Resource
  minReplicas: 1
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: istio-pilot

---

apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
  labels:
    app: mixer
  name: istio-policy
spec:
  maxReplicas: 5
  metrics:
  - resource:
      name: cpu
      targetAverageUtilization: 80
    type: Resource
  minReplicas: 1
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: istio-policy

---

apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
  labels:
    app: mixer
  name: istio-telemetry
spec:
  maxReplicas: 5
  metrics:
  - resource:
      name: cpu
      targetAverageUtilization: 80
    type: Resource
  minReplicas: 1
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: istio-telemetry
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/instance.yaml", `
apiVersion: config.istio.io/v1alpha2
kind: instance
metadata:
  labels:
    app: mixer
  name: attributes
spec:
  attributeBindings:
    destination.container.name: $out.destination_container_name | "unknown"
    destination.ip: $out.destination_pod_ip | ip("0.0.0.0")
    destination.labels: $out.destination_labels | emptyStringMap()
    destination.name: $out.destination_pod_name | "unknown"
    destination.namespace: $out.destination_namespace | "default"
    destination.owner: $out.destination_owner | "unknown"
    destination.serviceAccount: $out.destination_service_account_name | "unknown"
    destination.uid: $out.destination_pod_uid | "unknown"
    destination.workload.name: $out.destination_workload_name | "unknown"
    destination.workload.namespace: $out.destination_workload_namespace | "unknown"
    destination.workload.uid: $out.destination_workload_uid | "unknown"
    source.ip: $out.source_pod_ip | ip("0.0.0.0")
    source.labels: $out.source_labels | emptyStringMap()
    source.name: $out.source_pod_name | "unknown"
    source.namespace: $out.source_namespace | "default"
    source.owner: $out.source_owner | "unknown"
    source.serviceAccount: $out.source_service_account_name | "unknown"
    source.uid: $out.source_pod_uid | "unknown"
    source.workload.name: $out.source_workload_name | "unknown"
    source.workload.namespace: $out.source_workload_namespace | "unknown"
    source.workload.uid: $out.source_workload_uid | "unknown"
  compiledTemplate: kubernetes
  params:
    destination_port: destination.port | 0
    destination_uid: destination.uid | ""
    source_ip: source.ip | ip("0.0.0.0")
    source_uid: source.uid | ""

---

apiVersion: config.istio.io/v1alpha2
kind: instance
metadata:
  labels:
    app: mixer
  name: requestcount
spec:
  compiledTemplate: metric
  params:
    dimensions:
      connection_security_policy: conditional((context.reporter.kind | "inbound")
        == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls",
        "none"))
      destination_app: destination.labels["app"] | "unknown"
      destination_principal: destination.principal | "unknown"
      destination_service: destination.service.host | request.host | "unknown"
      destination_service_name: destination.service.name | "unknown"
      destination_service_namespace: destination.service.namespace | "unknown"
      destination_version: destination.labels["version"] | "unknown"
      destination_workload: destination.workload.name | "unknown"
      destination_workload_namespace: destination.workload.namespace | "unknown"
      permissive_response_code: rbac.permissive.response_code | "none"
      permissive_response_policyid: rbac.permissive.effective_policy_id | "none"
      reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source",
        "destination")
      request_protocol: api.protocol | context.protocol | "unknown"
      response_code: response.code | 200
      response_flags: context.proxy_error_code | "-"
      source_app: source.labels["app"] | "unknown"
      source_principal: source.principal | "unknown"
      source_version: source.labels["version"] | "unknown"
      source_workload: source.workload.name | "unknown"
      source_workload_namespace: source.workload.namespace | "unknown"
    monitored_resource_type: '"UNSPECIFIED"'
    value: "1"

---

apiVersion: config.istio.io/v1alpha2
kind: instance
metadata:
  labels:
    app: mixer
  name: requestduration
spec:
  compiledTemplate: metric
  params:
    dimensions:
      connection_security_policy: conditional((context.reporter.kind | "inbound")
        == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls",
        "none"))
      destination_app: destination.labels["app"] | "unknown"
      destination_principal: destination.principal | "unknown"
      destination_service: destination.service.host | request.host | "unknown"
      destination_service_name: destination.service.name | "unknown"
      destination_service_namespace: destination.service.namespace | "unknown"
      destination_version: destination.labels["version"] | "unknown"
      destination_workload: destination.workload.name | "unknown"
      destination_workload_namespace: destination.workload.namespace | "unknown"
      permissive_response_code: rbac.permissive.response_code | "none"
      permissive_response_policyid: rbac.permissive.effective_policy_id | "none"
      reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source",
        "destination")
      request_protocol: api.protocol | context.protocol | "unknown"
      response_code: response.code | 200
      response_flags: context.proxy_error_code | "-"
      source_app: source.labels["app"] | "unknown"
      source_principal: source.principal | "unknown"
      source_version: source.labels["version"] | "unknown"
      source_workload: source.workload.name | "unknown"
      source_workload_namespace: source.workload.namespace | "unknown"
    monitored_resource_type: '"UNSPECIFIED"'
    value: response.duration | "0ms"

---

apiVersion: config.istio.io/v1alpha2
kind: instance
metadata:
  labels:
    app: mixer
  name: requestsize
spec:
  compiledTemplate: metric
  params:
    dimensions:
      connection_security_policy: conditional((context.reporter.kind | "inbound")
        == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls",
        "none"))
      destination_app: destination.labels["app"] | "unknown"
      destination_principal: destination.principal | "unknown"
      destination_service: destination.service.host | request.host | "unknown"
      destination_service_name: destination.service.name | "unknown"
      destination_service_namespace: destination.service.namespace | "unknown"
      destination_version: destination.labels["version"] | "unknown"
      destination_workload: destination.workload.name | "unknown"
      destination_workload_namespace: destination.workload.namespace | "unknown"
      permissive_response_code: rbac.permissive.response_code | "none"
      permissive_response_policyid: rbac.permissive.effective_policy_id | "none"
      reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source",
        "destination")
      request_protocol: api.protocol | context.protocol | "unknown"
      response_code: response.code | 200
      response_flags: context.proxy_error_code | "-"
      source_app: source.labels["app"] | "unknown"
      source_principal: source.principal | "unknown"
      source_version: source.labels["version"] | "unknown"
      source_workload: source.workload.name | "unknown"
      source_workload_namespace: source.workload.namespace | "unknown"
    monitored_resource_type: '"UNSPECIFIED"'
    value: request.size | 0

---

apiVersion: config.istio.io/v1alpha2
kind: instance
metadata:
  labels:
    app: mixer
  name: responsesize
spec:
  compiledTemplate: metric
  params:
    dimensions:
      connection_security_policy: conditional((context.reporter.kind | "inbound")
        == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls",
        "none"))
      destination_app: destination.labels["app"] | "unknown"
      destination_principal: destination.principal | "unknown"
      destination_service: destination.service.host | request.host | "unknown"
      destination_service_name: destination.service.name | "unknown"
      destination_service_namespace: destination.service.namespace | "unknown"
      destination_version: destination.labels["version"] | "unknown"
      destination_workload: destination.workload.name | "unknown"
      destination_workload_namespace: destination.workload.namespace | "unknown"
      permissive_response_code: rbac.permissive.response_code | "none"
      permissive_response_policyid: rbac.permissive.effective_policy_id | "none"
      reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source",
        "destination")
      request_protocol: api.protocol | context.protocol | "unknown"
      response_code: response.code | 200
      response_flags: context.proxy_error_code | "-"
      source_app: source.labels["app"] | "unknown"
      source_principal: source.principal | "unknown"
      source_version: source.labels["version"] | "unknown"
      source_workload: source.workload.name | "unknown"
      source_workload_namespace: source.workload.namespace | "unknown"
    monitored_resource_type: '"UNSPECIFIED"'
    value: response.size | 0

---

apiVersion: config.istio.io/v1alpha2
kind: instance
metadata:
  labels:
    app: mixer
  name: tcpbytereceived
spec:
  compiledTemplate: metric
  params:
    dimensions:
      connection_security_policy: conditional((context.reporter.kind | "inbound")
        == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls",
        "none"))
      destination_app: destination.labels["app"] | "unknown"
      destination_principal: destination.principal | "unknown"
      destination_service: destination.service.host | "unknown"
      destination_service_name: destination.service.name | "unknown"
      destination_service_namespace: destination.service.namespace | "unknown"
      destination_version: destination.labels["version"] | "unknown"
      destination_workload: destination.workload.name | "unknown"
      destination_workload_namespace: destination.workload.namespace | "unknown"
      reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source",
        "destination")
      response_flags: context.proxy_error_code | "-"
      source_app: source.labels["app"] | "unknown"
      source_principal: source.principal | "unknown"
      source_version: source.labels["version"] | "unknown"
      source_workload: source.workload.name | "unknown"
      source_workload_namespace: source.workload.namespace | "unknown"
    monitored_resource_type: '"UNSPECIFIED"'
    value: connection.received.bytes | 0

---

apiVersion: config.istio.io/v1alpha2
kind: instance
metadata:
  labels:
    app: mixer
  name: tcpbytesent
spec:
  compiledTemplate: metric
  params:
    dimensions:
      connection_security_policy: conditional((context.reporter.kind | "inbound")
        == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls",
        "none"))
      destination_app: destination.labels["app"] | "unknown"
      destination_principal: destination.principal | "unknown"
      destination_service: destination.service.host | "unknown"
      destination_service_name: destination.service.name | "unknown"
      destination_service_namespace: destination.service.namespace | "unknown"
      destination_version: destination.labels["version"] | "unknown"
      destination_workload: destination.workload.name | "unknown"
      destination_workload_namespace: destination.workload.namespace | "unknown"
      reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source",
        "destination")
      response_flags: context.proxy_error_code | "-"
      source_app: source.labels["app"] | "unknown"
      source_principal: source.principal | "unknown"
      source_version: source.labels["version"] | "unknown"
      source_workload: source.workload.name | "unknown"
      source_workload_namespace: source.workload.namespace | "unknown"
    monitored_resource_type: '"UNSPECIFIED"'
    value: connection.sent.bytes | 0

---

apiVersion: config.istio.io/v1alpha2
kind: instance
metadata:
  labels:
    app: mixer
  name: tcpconnectionsclosed
spec:
  compiledTemplate: metric
  params:
    dimensions:
      connection_security_policy: conditional((context.reporter.kind | "inbound")
        == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls",
        "none"))
      destination_app: destination.labels["app"] | "unknown"
      destination_principal: destination.principal | "unknown"
      destination_service: destination.service.host | "unknown"
      destination_service_name: destination.service.name | "unknown"
      destination_service_namespace: destination.service.namespace | "unknown"
      destination_version: destination.labels["version"] | "unknown"
      destination_workload: destination.workload.name | "unknown"
      destination_workload_namespace: destination.workload.namespace | "unknown"
      reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source",
        "destination")
      response_flags: context.proxy_error_code | "-"
      source_app: source.labels["app"] | "unknown"
      source_principal: source.principal | "unknown"
      source_version: source.labels["version"] | "unknown"
      source_workload: source.workload.name | "unknown"
      source_workload_namespace: source.workload.namespace | "unknown"
    monitored_resource_type: '"UNSPECIFIED"'
    value: "1"

---

apiVersion: config.istio.io/v1alpha2
kind: instance
metadata:
  labels:
    app: mixer
  name: tcpconnectionsopened
spec:
  compiledTemplate: metric
  params:
    dimensions:
      connection_security_policy: conditional((context.reporter.kind | "inbound")
        == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls",
        "none"))
      destination_app: destination.labels["app"] | "unknown"
      destination_principal: destination.principal | "unknown"
      destination_service: destination.service.host | "unknown"
      destination_service_name: destination.service.name | "unknown"
      destination_service_namespace: destination.service.namespace | "unknown"
      destination_version: destination.labels["version"] | "unknown"
      destination_workload: destination.workload.name | "unknown"
      destination_workload_namespace: destination.workload.namespace | "unknown"
      reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source",
        "destination")
      response_flags: context.proxy_error_code | "-"
      source_app: source.labels["app"] | "unknown"
      source_principal: source.principal | "unknown"
      source_version: source.labels["version"] | "unknown"
      source_workload: source.workload.name | "unknown"
      source_workload_namespace: source.workload.namespace | "unknown"
    monitored_resource_type: '"UNSPECIFIED"'
    value: "1"
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/job.yaml", `
apiVersion: batch/v1
kind: Job
metadata:
  labels:
    app: security
  name: istio-security-post-install-release-1.3-latest-daily
spec:
  template:
    metadata:
      labels:
        app: security
      name: istio-security-post-install
    spec:
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
            weight: 2
          - preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x
            weight: 2
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
      containers:
      - command:
        - /bin/bash
        - /tmp/security/istio-security-run.sh
        - /tmp/security/istio-security-custom-resources.yaml
        image: gcr.io/istio-release/kubectl:release-1.3-latest-daily
        imagePullPolicy: IfNotPresent
        name: kubectl
        volumeMounts:
        - mountPath: /tmp/security
          name: tmp-configmap-security
      restartPolicy: OnFailure
      serviceAccountName: istio-security-post-install-account
      volumes:
      - configMap:
          name: istio-security-custom-resources
        name: tmp-configmap-security
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/mutating-webhook-configuration.yaml", `
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  labels:
    app: sidecarInjectorWebhook
  name: istio-sidecar-injector
webhooks:
- clientConfig:
    caBundle: ""
    service:
      name: istio-sidecar-injector
      namespace: $(namespace)
      path: /inject
  failurePolicy: Fail
  name: sidecar-injector.istio.io
  namespaceSelector:
    matchLabels:
      istio-injection: enabled
  rules:
  - apiGroups:
    - ""
    apiVersions:
    - v1
    operations:
    - CREATE
    resources:
    - pods
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/pod-disruption-budget.yaml", `
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  labels:
    app: galley
    istio: galley
  name: istio-galley
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: galley
      istio: galley

---

apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  labels:
    app: istio-ingressgateway
    istio: ingressgateway
  name: istio-ingressgateway
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: istio-ingressgateway
      istio: ingressgateway

---

apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  labels:
    app: pilot
    istio: pilot
  name: istio-pilot
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: pilot
      istio: pilot

---

apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  labels:
    app: policy
    istio: mixer
    istio-mixer-type: policy
    version: 1.1.0
  name: istio-policy
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: policy
      istio: mixer
      istio-mixer-type: policy

---

apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  labels:
    app: sidecarInjectorWebhook
    istio: sidecar-injector
  name: istio-sidecar-injector
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: sidecarInjectorWebhook
      istio: sidecar-injector

---

apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  labels:
    app: telemetry
    istio: mixer
    istio-mixer-type: telemetry
    version: 1.1.0
  name: istio-telemetry
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: telemetry
      istio: mixer
      istio-mixer-type: telemetry
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: istio-ingressgateway-sds
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - watch
  - list
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: istio-ingressgateway-sds
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: istio-ingressgateway-sds
subjects:
- kind: ServiceAccount
  name: istio-ingressgateway-service-account
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/rule.yaml", `
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  labels:
    app: mixer
  name: istio-policy
spec:
  host: istio-policy.$(namespace).svc.cluster.local
  trafficPolicy:
    connectionPool:
      http:
        http2MaxRequests: 10000
        maxRequestsPerConnection: 10000
    portLevelSettings:
    - port:
        number: 15004
      tls:
        mode: ISTIO_MUTUAL

---

apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  labels:
    app: mixer
  name: istio-telemetry
spec:
  host: istio-telemetry.$(namespace).svc.cluster.local
  trafficPolicy:
    connectionPool:
      http:
        http2MaxRequests: 10000
        maxRequestsPerConnection: 10000
    portLevelSettings:
    - port:
        number: 15004
      tls:
        mode: ISTIO_MUTUAL

---

apiVersion: config.istio.io/v1alpha2
kind: rule
metadata:
  labels:
    app: mixer
  name: kubeattrgenrulerule
spec:
  actions:
  - handler: kubernetesenv
    instances:
    - attributes

---

apiVersion: config.istio.io/v1alpha2
kind: rule
metadata:
  labels:
    app: mixer
  name: promhttp
spec:
  actions:
  - handler: prometheus
    instances:
    - requestcount
    - requestduration
    - requestsize
    - responsesize
  match: (context.protocol == "http" || context.protocol == "grpc") && (match((request.useragent
    | "-"), "kube-probe*") == false) && (match((request.useragent | "-"), "Prometheus*")
    == false)

---

apiVersion: config.istio.io/v1alpha2
kind: rule
metadata:
  labels:
    app: mixer
  name: promtcp
spec:
  actions:
  - handler: prometheus
    instances:
    - tcpbytesent
    - tcpbytereceived
  match: context.protocol == "tcp"

---

apiVersion: config.istio.io/v1alpha2
kind: rule
metadata:
  labels:
    app: mixer
  name: promtcpconnectionclosed
spec:
  actions:
  - handler: prometheus
    instances:
    - tcpconnectionsclosed
  match: context.protocol == "tcp" && ((connection.event | "na") == "close")

---

apiVersion: config.istio.io/v1alpha2
kind: rule
metadata:
  labels:
    app: mixer
  name: promtcpconnectionopen
spec:
  actions:
  - handler: prometheus
    instances:
    - tcpconnectionsopened
  match: context.protocol == "tcp" && ((connection.event | "na") == "open")

---

apiVersion: config.istio.io/v1alpha2
kind: rule
metadata:
  labels:
    app: mixer
  name: tcpkubeattrgenrulerule
spec:
  actions:
  - handler: kubernetesenv
    instances:
    - attributes
  match: context.protocol == "tcp"
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: security
    istio: citadel
  name: istio-citadel
spec:
  ports:
  - name: grpc-citadel
    port: 8060
    protocol: TCP
    targetPort: 8060
  - name: http-monitoring
    port: 15014
  selector:
    istio: citadel

---

apiVersion: v1
kind: Service
metadata:
  labels:
    app: galley
    istio: galley
  name: istio-galley
spec:
  ports:
  - name: https-validation
    port: 443
  - name: http-monitoring
    port: 15014
  - name: grpc-mcp
    port: 9901
  selector:
    istio: galley

---

apiVersion: v1
kind: Service
metadata:
  name: istio-ingressgateway
  annotations:
    beta.cloud.google.com/backend-config: '{"ports": {"http2":"iap-backendconfig"}}'
  labels:
    app: istio-ingressgateway
    istio: ingressgateway
spec:
  type: NodePort
  selector:
    app: istio-ingressgateway
    istio: ingressgateway
  ports:
    - name: status-port
      port: 15020
      targetPort: 15020
    - name: http2
      nodePort: 31380
      port: 80
      targetPort: 80
    - name: https
      nodePort: 31390
      port: 443
    - name: tcp
      nodePort: 31400
      port: 31400
    - name: https-kiali
      port: 15029
      targetPort: 15029
    - name: https-prometheus
      port: 15030
      targetPort: 15030
    - name: https-grafana
      port: 15031
      targetPort: 15031
    - name: https-tracing
      port: 15032
      targetPort: 15032
    - name: tls
      port: 15443
      targetPort: 15443

---

apiVersion: v1
kind: Service
metadata:
  labels:
    app: pilot
    istio: pilot
  name: istio-pilot
spec:
  ports:
  - name: grpc-xds
    port: 15010
  - name: https-xds
    port: 15011
  - name: http-legacy-discovery
    port: 8080
  - name: http-monitoring
    port: 15014
  selector:
    istio: pilot

---

apiVersion: v1
kind: Service
metadata:
  annotations:
    networking.istio.io/exportTo: '*'
  labels:
    app: mixer
    istio: mixer
  name: istio-policy
spec:
  ports:
  - name: grpc-mixer
    port: 9091
  - name: grpc-mixer-mtls
    port: 15004
  - name: http-monitoring
    port: 15014
  selector:
    istio: mixer
    istio-mixer-type: policy

---

apiVersion: v1
kind: Service
metadata:
  labels:
    app: sidecarInjectorWebhook
    istio: sidecar-injector
  name: istio-sidecar-injector
spec:
  ports:
  - name: https-inject
    port: 443
  - name: http-monitoring
    port: 15014
  selector:
    istio: sidecar-injector

---

apiVersion: v1
kind: Service
metadata:
  annotations:
    networking.istio.io/exportTo: '*'
  labels:
    app: mixer
    istio: mixer
  name: istio-telemetry
spec:
  ports:
  - name: grpc-mixer
    port: 9091
  - name: grpc-mixer-mtls
    port: 15004
  - name: http-monitoring
    port: 15014
  - name: prometheus
    port: 42422
  selector:
    istio: mixer
    istio-mixer-type: telemetry

---

apiVersion: v1
kind: Service
metadata:
  annotations:
    prometheus.io/scrape: "true"
  labels:
    app: prometheus
  name: prometheus
spec:
  ports:
  - name: http-prometheus
    port: 9090
    protocol: TCP
  selector:
    app: prometheus
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: security
  name: istio-citadel-service-account

---

apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: galley
  name: istio-galley-service-account

---

apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: istio-ingressgateway
  name: istio-ingressgateway-service-account

---

apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: mixer
  name: istio-mixer-service-account

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-multi

---

apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: nodeagent
  name: istio-nodeagent-service-account

---

apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: pilot
  name: istio-pilot-service-account

---

apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: security
  name: istio-security-post-install-account

---

apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: sidecarInjectorWebhook
    istio: sidecar-injector
  name: istio-sidecar-injector-service-account

---

apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: prometheus
  name: prometheus
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/service-role.yaml", `
# Added to allow all requests to istio-ingressgateway
# Refer issue: https://github.com/istio/istio/issues/14885
apiVersion: "rbac.istio.io/v1alpha1"
kind: ServiceRole
metadata:
  name: istio-ingressgateway
spec:
  rules:
  - services:
    - istio-ingressgateway.$(namespace).svc.cluster.local
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/service-role-binding.yaml", `
# Added to allow all requests to istio-ingressgateway
# Refer issue: https://github.com/istio/istio/issues/14885
apiVersion: "rbac.istio.io/v1alpha1"
kind: ServiceRoleBinding
metadata:
  name: istio-ingressgateway
spec:
  subjects:
  - user: "*"

  roleRef:
    kind: ServiceRole
    name: "istio-ingressgateway"
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/params.yaml", `
varReference:
- path: metadata/name
  kind: Namespace
- path: data/validatingwebhookconfiguration.yaml
  kind: ConfigMap
- path: metadata/name
  kind: ClusterRole
- path: webhooks/clientConfig/service/namespace
  kind: MutatingWebhookConfiguration
- path: metadata/name
  kind: ClusterRoleBinding
- path: roleRef/name
  kind: ClusterRoleBinding
- path: data/istio-performance-dashboard.json
  kind: ConfigMap
- path: data/mesh
  kind: ConfigMap
- path: data/config.yaml
  kind: ConfigMap
- path: data/prometheus.yaml
  kind: ConfigMap
- path: spec/params/metrics/instance_name
  kind: handler
- path: spec/host
  kind: DestinationRule
- path: spec/rules/services
  kind: ServiceRole
- path: data/istio-security-custom-resources.yaml
  kind: ConfigMap
- path: data/istio-security-run.sh
  kind: ConfigMap
- path: data/validating-webhook-configuration.yaml
  kind: ConfigMap
- path: subjects/namespace
  kind: ClusterRoleBinding
`)
	th.writeF("/manifests/istio-1-3-1/istio-install-1-3-1/base/params.env", `
namespace=istio-system
`)
	th.writeK("/manifests/istio-1-3-1/istio-install-1-3-1/base", `
# Each entry in this list results in the creation of
# one ConfigMap resource (it's a generator of n maps).
configMapGenerator:
- name: istio-install-parameters
  env: params.env

# Images modify the tags for images without
# creating patches.
images:
- name: docker.io/prom/prometheus
  newTag: v2.8.0
- name: gcr.io/istio-release/citadel
  newTag: release-1.3-latest-daily
- name: gcr.io/istio-release/galley
  newTag: release-1.3-latest-daily
- name: gcr.io/istio-release/kubectl
  newTag: release-1.3-latest-daily
- name: gcr.io/istio-release/mixer
  newTag: release-1.3-latest-daily
- name: gcr.io/istio-release/node-agent-k8s
  newTag: release-1.3-latest-daily
- name: gcr.io/istio-release/pilot
  newTag: release-1.3-latest-daily
- name: gcr.io/istio-release/proxyv2
  newTag: release-1.3-latest-daily
- name: gcr.io/istio-release/sidecar_injector
  newTag: release-1.3-latest-daily

resources:
- namespace.yaml
- attribute-manifest.yaml
- config-map.yaml
- cluster-role.yaml
- cluster-role-binding.yaml
- daemon-set.yaml
- deployment.yaml
- handler.yaml
- horizontal-pod-autoscaler.yaml
- instance.yaml
- job.yaml
- mutating-webhook-configuration.yaml
- pod-disruption-budget.yaml
- role.yaml
- role-binding.yaml
- rule.yaml
- service.yaml
- service-account.yaml
- service-role.yaml
- service-role-binding.yaml

vars:
- name: namespace
  objref:
    kind: ConfigMap
    name: istio-install-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace

configurations:
- params.yaml
`)
}

func TestIstioInstall131Base(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/istio-1-3-1/istio-install-1-3-1/base")
	writeIstioInstall131Base(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../istio-1-3-1/istio-install-1-3-1/base"
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
