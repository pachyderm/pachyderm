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

func writeDexAuthenticatorBase(th *KustTestHarness) {
	th.writeF("/manifests/dex-auth/dex-authenticator/base/namespace.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  name: auth
`)
	th.writeF("/manifests/dex-auth/dex-authenticator/base/config-map.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: dex-authenticator-cm
  labels:
    app: dex-authenticator
data:
  config.yaml: |-
    clusters:
      # Specify 1 or more clusters
      - name: $(cluster_name)

        # Descriptions used in the WebUI
        short_description: "Dex Cluster"
        description: "Dex Server for Kubeflow"

        # Redirect Url pointing to dex-k8s-authenticator callback for this cluster
        # This should be configured in Dex as part of the staticClients
        # redirectURIs option
        redirect_uri: $(client_redirect_uri)

        # Client Secret - should match value in Dex
        client_secret: $(application_secret)

        # Client ID - should match value in Dex
        client_id: $(client_id)

        # Dex Issuer - Must be resolvable
        issuer: $(issuer)

        # Url to k8s API endpoint - used in WebUI instructions for generating
        # kubeconfig
        k8s_master_uri: $(k8s_master_uri)

        # don't use username for context
        static_context_name: false

        # CA for your k8s cluster - used in WebUI instructions for generating
        # kubeconfig
        # Both k8s_ca_uri and k8s_ca_pem are optional - you typically specifiy
        # one or the other if required
        #
        # Provides a link to the CA from a hosted site
        # k8s_ca_uri: http://url-to-your-ca.crt
        #
        # Provides abililty to specify CA inline
        # k8s_ca_pem: |
        #   -----BEGIN CERTIFICATE-----
        #   ...
        #   -----END CERTIFICATE-----
        k8s_ca_pem_file: /app/k8s_ca.pem

    # Specify multiple extra root CA files to be loaded
    # trusted_root_ca:
    #   -|
    #     -----BEGIN CERTIFICATE-----
    #     ...
    #     -----END CERTIFICATE-----
    trusted_root_ca_file: /app/idp_ca.pem

    # Specify path to tls_cert and tls_key - if enabled, set liten to use https
    # tls_cert: /path/to/dex-client.crt
    # tls_key: /path/to/dex-client.key

    # CA for your IDP - used in WebUI instructions for generating
    # kubeconfig
    # Both idp_ca_uri and idp_ca_pem are optional - you typically specifiy
    # one or the other if required
    #
    # Provides a link to the CA from a hosted site
    # idp_ca_uri: http://url-to-your-ca.crt
    #
    # Provides abililty to specify CA inline
    # idp_ca_pem: |
    #   -----BEGIN CERTIFICATE-----
    #   ...
    #   -----END CERTIFICATE-----
    idp_ca_pem_file: /app/idp_ca.pem

    # Which address to listen on (set to https if tls configured)
    listen: $(client_listen_addr)

    # A path-prefix from which to serve requests and assets
    web_path_prefix: /

    # Optional kubectl version which provides a download link to the the binary
    kubectl_version: v1.11.2

    # Optional Url to display a logo image
    # logo_uri: http://<path-to-your-logo.png>

    # Enable more debug
    debug: false
`)
	th.writeF("/manifests/dex-auth/dex-authenticator/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dex-authenticator
  labels:
    app: dex-authenticator
    env: dev
spec:
  replicas: 1
  selector:
    matchLabels:
      app: dex-authenticator
  template:
    metadata:
      labels:
        app: dex-authenticator
    spec:
      containers:
      - name: dex-k8s-authenticator
        image: "mintel/dex-k8s-authenticator:1.2.0"
        imagePullPolicy: Always
        args: [ "--config", "config.yaml" ]
        ports:
        - name: http
          containerPort: 5555
          protocol: TCP
        livenessProbe:
          httpGet:
            path: /healthz
            port: http
        readinessProbe:
          httpGet:
            path: /healthz
            port: http
        volumeMounts:
        - name: config
          subPath: config.yaml
          mountPath: /app/config.yaml
        - name: idp-ca
          subPath: ca.pem
          mountPath: /app/idp_ca.pem
        - name: k8s-ca
          subPath: k8s_ca.pem
          mountPath: /app/k8s_ca.pem
        resources:
          {}

      volumes:
      - name: config
        configMap:
          name: dex-authenticator-cm
      - name: idp-ca
        configMap:
          name: ca
      - name: k8s-ca
        configMap:
          name: k8s-ca
`)
	th.writeF("/manifests/dex-auth/dex-authenticator/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: dex-authenticator
  labels:
    app: dex-authenticator
spec:
  type: NodePort
  ports:
  - port: 5555
    targetPort: 5555
    nodePort: 32002
    protocol: TCP
    name: http
  selector:
    app: dex-authenticator
`)
	th.writeF("/manifests/dex-auth/dex-authenticator/base/params.yaml", `
varReference:
- path: data/config.yaml
  kind: ConfigMap
`)
	th.writeF("/manifests/dex-auth/dex-authenticator/base/params.env", `
# Dex Server Parameters (some params are shared with client)
# Set issuer to https if tls is enabled
issuer=http://dex.example.com:32000
client_id=ldapdexapp
application_secret=pUBnBOY80SnXgjibTYM9ZWNzY2xreNGQok
cluster_name=onprem-cluster
client_redirect_uri=http://login.example.org:5555/callback/onprem-cluster
k8s_master_uri=https://k8s.example.com:443
client_listen_addr=http://127.0.0.1:5555 # Set to HTTPS if TLS is configured
`)
	th.writeK("/manifests/dex-auth/dex-authenticator/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: auth
resources:
- namespace.yaml
- config-map.yaml
- deployment.yaml
- service.yaml
configMapGenerator:
- name: dex-authn-parameters
  env: params.env
vars:
- name: issuer
  objref:
    kind: ConfigMap
    name: dex-authn-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.issuer
- name: client_id
  objref:
    kind: ConfigMap
    name: dex-authn-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.client_id
- name: application_secret
  objref:
    kind: ConfigMap
    name: dex-authn-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.application_secret
- name: cluster_name
  objref:
    kind: ConfigMap
    name: dex-authn-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.cluster_name
- name: k8s_master_uri
  objref:
    kind: ConfigMap
    name: dex-authn-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.k8s_master_uri
- name: client_redirect_uri
  objref:
    kind: ConfigMap
    name: dex-authn-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.client_redirect_uri
- name: client_listen_addr
  objref:
    kind: ConfigMap
    name: dex-authn-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.client_listen_addr
configurations:
- params.yaml
images:
- name: mintel/dex-k8s-authenticator
  newName: mintel/dex-k8s-authenticator
  newTag: 1.2.0
`)
}

func TestDexAuthenticatorBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/dex-auth/dex-authenticator/base")
	writeDexAuthenticatorBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../dex-auth/dex-authenticator/base"
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
