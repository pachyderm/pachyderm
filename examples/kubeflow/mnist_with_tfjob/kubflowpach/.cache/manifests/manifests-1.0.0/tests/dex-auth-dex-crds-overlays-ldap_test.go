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

func writeDexCrdsOverlaysLdap(th *KustTestHarness) {
	th.writeF("/manifests/dex-auth/dex-crds/overlays/ldap/config-map.yaml", `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: dex
data:
  config.yaml: |
    issuer: $(issuer)
    storage:
      type: kubernetes
      config:
        inCluster: true
    web:
      https: 0.0.0.0:5556
      tlsCert: /etc/dex/tls/tls.crt
      tlsKey: /etc/dex/tls/tls.key
      # For HTTP configuration remove tls configs and add
      #http: 0.0.0.0:5556
    logger:
      level: "debug"
      format: text
    connectors:
      - type: ldap
        # Required field for connector id.
        id: ldap
        # Required field for connector name.
        name: LDAP
        config:
          # Host and optional port of the LDAP server in the form "host:port".
          # If the port is not supplied, it will be guessed based on "insecureNoSSL",
          # and "startTLS" flags. 389 for insecure or StartTLS connections, 636
          # otherwise.
          host: $(ldap_host)
          # Following field is required if the LDAP host is not using TLS (port 389).
          # Because this option inherently leaks passwords to anyone on the same network
          # as dex, THIS OPTION MAY BE REMOVED WITHOUT WARNING IN A FUTURE RELEASE.
          #
          insecureNoSSL: true
          # If a custom certificate isn't provide, this option can be used to turn on
          # TLS certificate checks. As noted, it is insecure and shouldn't be used outside
          # of explorative phases.
          #
          insecureSkipVerify: true
          # When connecting to the server, connect using the ldap:// protocol then issue
          # a StartTLS command. If unspecified, connections will use the ldaps:// protocol
          #
          # startTLS: true
          # Path to a trusted root certificate file. Default: use the host's root CA.
          #rootCA: /etc/dex/ldap.ca
          # A raw certificate file can also be provided inline.
          #rootCAData:
          # The DN and password for an application service account. The connector uses
          # these credentials to search for users and groups. Not required if the LDAP
          # server provides access for anonymous auth.
          # Please note that if the bind password contains a '$', it has to be saved in an
          # environment variable which should be given as the value to 'bindPW'.
          bindDN: $(ldap_bind_dn)
          bindPW: $(ldap_bind_pw)
          # User search maps a username and password entered by a user to a LDAP entry.
          userSearch:
            # BaseDN to start the search from. It will translate to the query
            # "(&(objectClass=person)(uid=<username>))".
            baseDN: $(ldap_user_base_dn)
            # Optional filter to apply when searching the directory.
            filter: "(objectClass=posixAccount)"
            # username attribute used for comparing user entries. This will be translated
            # and combine with the other filter as "(<attr>=<username>)".
            username: mail
            # The following three fields are direct mappings of attributes on the user entry.
            # String representation of the user.
            idAttr: uid
            # Required. Attribute to map to Email.
            emailAttr: mail
            # Maps to display name of users. No default value.
            nameAttr: uid

          # Group search queries for groups given a user entry.
          groupSearch:
            # BaseDN to start the search from. It will translate to the query
            # "(&(objectClass=group)(member=<user uid>))".
            baseDN: $(ldap_group_base_dn)
            # Optional filter to apply when searching the directory.
            filter: "(objectClass=posixGroup)"
            # Following two fields are used to match a user to a group. It adds an additional
            # requirement to the filter that an attribute in the group must match the user's
            # attribute value.
            userAttr: gidNumber
            groupAttr: gidNumber
            # Represents group name.
            nameAttr: cn
    oauth2:
      skipApprovalScreen: true
    staticClients:
    - id: $(client_id)
      redirectURIs: $(oidc_redirect_uris)
      name: 'Dex Login Application'
      secret: $(application_secret)
`)
	th.writeF("/manifests/dex-auth/dex-crds/overlays/ldap/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: dex
  name: dex
spec:
  template:
    spec:
      serviceAccountName: dex
      containers:
      - name: dex
        volumeMounts:
        - name: tls
          mountPath: /etc/dex/tls
      volumes:
      - name: tls
        secret:
          secretName: $(dex_domain).tls
`)
	th.writeF("/manifests/dex-auth/dex-crds/overlays/ldap/params.yaml", `
varReference:
- path: data/config.yaml
  kind: ConfigMap
`)
	th.writeF("/manifests/dex-auth/dex-crds/overlays/ldap/params.env", `
# Dex Server Parameters (some params are shared with client)
dex_domain=dex.example.com
issuer=https://dex.example.com:32000
ldap_host=ldap.auth.svc.cluster.local:389
ldap_bind_dn=cn=admin,dc=example,dc=org
ldap_bind_pw=admin
ldap_user_base_dn=ou=People,dc=example,dc=org
ldap_group_base_dn=ou=Groups,dc=example,dc=org
client_id=ldapdexapp
oidc_redirect_uris=['http://login.example.org:5555/callback/onprem-cluster']
application_secret=pUBnBOY80SnXgjibTYM9ZWNzY2xreNGQok
`)
	th.writeK("/manifests/dex-auth/dex-crds/overlays/ldap", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: auth
bases:
- ../../base

patchesStrategicMerge:
- config-map.yaml
- deployment.yaml

configMapGenerator:
- name: dex-parameters
  behavior: merge
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: ldap_host
  objref:
    kind: ConfigMap
    name: dex-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.ldap_host
- name: ldap_bind_dn
  objref:
    kind: ConfigMap
    name: dex-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.ldap_bind_dn
- name: ldap_bind_pw
  objref:
    kind: ConfigMap
    name: dex-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.ldap_bind_pw
- name: ldap_user_base_dn
  objref:
    kind: ConfigMap
    name: dex-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.ldap_user_base_dn
- name: ldap_group_base_dn
  objref:
    kind: ConfigMap
    name: dex-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.ldap_group_base_dn
configurations:
- params.yaml
`)
	th.writeF("/manifests/dex-auth/dex-crds/base/namespace.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  name: auth
`)
	th.writeF("/manifests/dex-auth/dex-crds/base/config-map.yaml", `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: dex
data:
  config.yaml: |
    issuer: $(issuer)
    storage:
      type: kubernetes
      config:
        inCluster: true
    web:
      http: 0.0.0.0:5556
    logger:
      level: "debug"
      format: text
    oauth2:
      skipApprovalScreen: true
    enablePasswordDB: true
    staticPasswords:
    - email: $(static_email)
      hash: $(static_password_hash)
      username: $(static_username)
      userID: $(static_user_id)
    staticClients:
    - id: $(client_id)
      redirectURIs: $(oidc_redirect_uris)
      name: 'Dex Login Application'
      secret: $(application_secret)
`)
	th.writeF("/manifests/dex-auth/dex-crds/base/crds.yaml", `
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: authcodes.dex.coreos.com
spec:
  group: dex.coreos.com
  names:
    kind: AuthCode
    listKind: AuthCodeList
    plural: authcodes
    singular: authcode
  scope: Namespaced
  version: v1
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: dex
rules:
- apiGroups: ["dex.coreos.com"] # API group created by dex
  resources: ["*"]
  verbs: ["*"]
- apiGroups: ["apiextensions.k8s.io"]
  resources: ["customresourcedefinitions"]
  verbs: ["create"] # To manage its own resources identity must be able to create customresourcedefinitions.
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: dex
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: dex
subjects:
- kind: ServiceAccount
  name: dex                 # Service account assigned to the dex pod.
  namespace: auth           # The namespace dex is running in.
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: dex
  namespace: auth
`)
	th.writeF("/manifests/dex-auth/dex-crds/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: dex
  name: dex
spec:
  replicas: 1
  selector:
    matchLabels:
        app: dex
  template:
    metadata:
      labels:
        app: dex
    spec:
      serviceAccountName: dex
      containers:
      - image: quay.io/coreos/dex:v2.9.0
        name: dex
        command: ["dex", "serve", "/etc/dex/cfg/config.yaml"]
        ports:
        - name: http
          containerPort: 5556
        volumeMounts:
        - name: config
          mountPath: /etc/dex/cfg
      volumes:
      - name: config
        configMap:
          name: dex
          items:
          - key: config.yaml
            path: config.yaml
`)
	th.writeF("/manifests/dex-auth/dex-crds/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: dex
spec:
  type: NodePort
  ports:
  - name: dex
    port: 5556
    protocol: TCP
    targetPort: 5556
    nodePort: 32000
  selector:
    app: dex
`)
	th.writeF("/manifests/dex-auth/dex-crds/base/params.yaml", `
varReference:
- path: spec/template/spec/volumes/secret/secretName
  kind: Deployment
- path: data/config.yaml
  kind: ConfigMap
`)
	th.writeF("/manifests/dex-auth/dex-crds/base/params.env", `
# Dex Server Parameters (some params are shared with client)
dex_domain=dex.example.com
# Set issuer to https if tls is enabled
issuer=http://dex.example.com:32000
static_email=admin@example.com
static_password_hash=$2a$10$2b2cU8CPhOTaGrs1HRQuAueS7JTT5ZHsHSzYiFPm1leZck7Mc8T4W
static_username=admin
static_user_id=08a8684b-db88-4b73-90a9-3cd1661f5466
client_id=ldapdexapp
oidc_redirect_uris=['http://login.example.org:5555/callback/onprem-cluster']
application_secret=pUBnBOY80SnXgjibTYM9ZWNzY2xreNGQok
`)
	th.writeK("/manifests/dex-auth/dex-crds/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: auth
resources:
- namespace.yaml
- config-map.yaml
- crds.yaml
- deployment.yaml
- service.yaml
configMapGenerator:
- name: dex-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: dex_domain
  objref:
    kind: ConfigMap
    name: dex-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.dex_domain
- name: issuer
  objref:
    kind: ConfigMap
    name: dex-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.issuer
- name: static_email
  objref:
    kind: ConfigMap
    name: dex-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.static_email
- name: static_password_hash
  objref:
    kind: ConfigMap
    name: dex-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.static_password_hash
- name: static_username
  objref:
    kind: ConfigMap
    name: dex-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.static_username
- name: static_user_id
  objref:
    kind: ConfigMap
    name: dex-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.static_user_id
- name: client_id
  objref:
    kind: ConfigMap
    name: dex-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.client_id
- name: oidc_redirect_uris
  objref:
    kind: ConfigMap
    name: dex-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.oidc_redirect_uris
- name: application_secret
  objref:
    kind: ConfigMap
    name: dex-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.application_secret
configurations:
- params.yaml
images:
- name: quay.io/coreos/dex
  newName: gcr.io/arrikto/dexidp/dex
  newTag: 4bede5eb80822fc3a7fc9edca0ed2605cd339d17
`)
}

func TestDexCrdsOverlaysLdap(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/dex-auth/dex-crds/overlays/ldap")
	writeDexCrdsOverlaysLdap(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../dex-auth/dex-crds/overlays/ldap"
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
