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

func writeBootstrapBase(th *KustTestHarness) {
	th.writeF("/manifests/admission-webhook/bootstrap/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cluster-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-role
subjects:
- kind: ServiceAccount
  name: service-account
`)
	th.writeF("/manifests/admission-webhook/bootstrap/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cluster-role
rules:
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - mutatingwebhookconfigurations
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - list
  - delete  
    
`)
	th.writeF("/manifests/admission-webhook/bootstrap/base/config-map.yaml", `
apiVersion: v1
data:
  create_ca.sh: |
    #!/bin/bash

    set -e

    usage() {
        cat <<EOF
    Generate certificate suitable for use with an sidecar-injector webhook service.
    This script uses k8s' CertificateSigningRequest API to a generate a
    certificate signed by k8s CA suitable for use with sidecar-injector webhook
    services. This requires permissions to create and approve CSR. See
    https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster for
    detailed explantion and additional instructions.
    The server key/cert k8s CA cert are stored in a k8s secret.
    usage: ${0} [OPTIONS]
    The following flags are required.
           --service          Service name of webhook.
           --namespace        Namespace where webhook service and secret reside.
           --secret           Secret name for CA certificate and server certificate/key pair.
    EOF
        exit 1
    }

    while [[ $# -gt 0 ]]; do
        case ${1} in
            --service)
                service="$2"
                shift
                ;;
            --secret)
                secret="$2"
                shift
                ;;
            --namespace)
                namespace="$2"
                shift
                ;;
            *)
                usage
                ;;
        esac
        shift
    done

    [ -z ${service} ] && service=$(webhookNamePrefix)service
    [ -z ${secret} ] && secret=webhook-certs
    [ -z ${namespace} ] && namespace=$(namespace)
    [ -z ${namespace} ] && namespace=default

    webhookDeploymentName=$(webhookNamePrefix)deployment
    mutatingWebhookConfigName=$(webhookNamePrefix)mutating-webhook-configuration
    echo ${service}
    echo ${namespace}
    echo ${secret}
    echo ${webhookDeploymentName}
    echo ${mutatingWebhookconfigName}
    if [ ! -x "$(command -v openssl)" ]; then
        echo "openssl not found"
        exit 1
    fi
    csrName=${service}.${namespace}
    tmpdir=$(mktemp -d)
    echo "creating certs in tmpdir ${tmpdir} "

    # x509 outputs a self signed certificate instead of certificate request, later used as self signed root CA
    openssl req -x509 -newkey rsa:2048 -keyout ${tmpdir}/self_ca.key -out ${tmpdir}/self_ca.crt -days 365 -nodes -subj /C=/ST=/L=/O=/OU=/CN=test-certificate-authority

    cat <<EOF >> ${tmpdir}/csr.conf
    [req]
    req_extensions = v3_req
    distinguished_name = req_distinguished_name
    [req_distinguished_name]
    [ v3_req ]
    basicConstraints = CA:FALSE
    keyUsage = nonRepudiation, digitalSignature, keyEncipherment
    extendedKeyUsage = serverAuth
    subjectAltName = @alt_names
    [alt_names]
    DNS.1 = ${service}
    DNS.2 = ${service}.${namespace}
    DNS.3 = ${service}.${namespace}.svc
    EOF

    openssl genrsa -out ${tmpdir}/server-key.pem 2048
    openssl req -new -key ${tmpdir}/server-key.pem -subj "/CN=${service}.${namespace}.svc" -out ${tmpdir}/server.csr -config ${tmpdir}/csr.conf

    # Self sign
    openssl x509 -req -days 365 -in ${tmpdir}/server.csr -CA ${tmpdir}/self_ca.crt -CAkey ${tmpdir}/self_ca.key -CAcreateserial -out ${tmpdir}/server-cert.pem

    # create the secret with CA cert and server cert/key
    kubectl create secret generic ${secret} \
            --from-file=key.pem=${tmpdir}/server-key.pem \
            --from-file=cert.pem=${tmpdir}/server-cert.pem \
            --dry-run -o yaml |
        kubectl -n ${namespace} apply -f -

    # Webhook pod needs to be restarted so that the service reload the secret
    # http://github.com/kueflow/kubeflow/issues/3227
    webhookPod=$(kubectl get pods -n ${namespace} |grep ${webhookDeploymentName} |awk '{print $1;}')
    # ignore error if webhook pod does not exist
    kubectl delete pod ${webhookPod} 2>/dev/null || true
    echo "webhook ${webhookPod} is restarted to utilize the new secret"
    
    cat ${tmpdir}/self_ca.crt
    
    # -a means base64 encode
    caBundle=$(cat ${tmpdir}/self_ca.crt | openssl enc -a -A)
    echo ${caBundle}

    patchString='[{"op": "replace", "path": "/webhooks/0/clientConfig/caBundle", "value":"{{CA_BUNDLE}}"}]'
    patchString=$(echo ${patchString} | sed "s|{{CA_BUNDLE}}|${caBundle}|g")
    echo ${patchString}

    checkWebhookConfig() {
      currentBundle=$(kubectl get mutatingwebhookconfigurations -n ${namespace} ${mutatingWebhookConfigName} -o jsonpath='{.webhooks[0].clientConfig.caBundle}')
      [[ "$currentBundle" == "$caBundle" ]]
    }

    while true; do
      if ! checkWebhookConfig; then
        echo "patching ca bundle for webhook configuration..."
        kubectl patch mutatingwebhookconfiguration ${mutatingWebhookConfigName} \
            --type='json' -p="${patchString}"
      fi
      sleep 10
    done
kind: ConfigMap
metadata:
  name: config-map
`)
	th.writeF("/manifests/admission-webhook/bootstrap/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: service-account
`)
	th.writeF("/manifests/admission-webhook/bootstrap/base/stateful-set.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: stateful-set
spec:
  replicas: 1
  serviceName: service
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      containers:
      - command:
        - sh
        - /var/webhook-config/create_ca.sh
        image: gcr.io/kubeflow-images-public/ingress-setup:latest
        name: bootstrap
        volumeMounts:
        - mountPath: /var/webhook-config/
          name: admission-webhook-config
      restartPolicy: Always
      serviceAccountName: service-account
      volumes:
      - configMap:
          name: config-map
        name: admission-webhook-config
  # Workaround for https://github.com/kubernetes-sigs/kustomize/issues/677
  volumeClaimTemplates: []
`)
	th.writeF("/manifests/admission-webhook/bootstrap/base/params.yaml", `
varReference:
- path: data/create_ca.sh
  kind: ConfigMap
`)
	th.writeF("/manifests/admission-webhook/bootstrap/base/params.env", `
namespace=kubeflow
webhookNamePrefix=admission-webhook-
`)
	th.writeK("/manifests/admission-webhook/bootstrap/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- config-map.yaml
- service-account.yaml
- stateful-set.yaml
commonLabels:
  kustomize.component: admission-webhook-bootstrap
namePrefix: admission-webhook-bootstrap-
images:
- name: gcr.io/kubeflow-images-public/ingress-setup
  newName: gcr.io/kubeflow-images-public/ingress-setup
  newTag: latest
generatorOptions:
  disableNameSuffixHash: true
configurations:
- params.yaml
namespace: kubeflow
configMapGenerator:
- name: config-map
  behavior: merge
  env: params.env
vars:
- name: webhookNamePrefix
  objref:
    kind: ConfigMap
    name: config-map
    apiVersion: v1
  fieldref:
    fieldpath: data.webhookNamePrefix
- name: namespace
  objref:
    kind: ConfigMap
    name: config-map
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace
`)
}

func TestBootstrapBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/admission-webhook/bootstrap/base")
	writeBootstrapBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../admission-webhook/bootstrap/base"
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
