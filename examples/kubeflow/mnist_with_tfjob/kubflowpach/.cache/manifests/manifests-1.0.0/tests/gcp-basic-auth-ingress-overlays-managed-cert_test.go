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

func writeBasicAuthIngressOverlaysManagedCert(th *KustTestHarness) {
	th.writeF("/manifests/gcp/basic-auth-ingress/overlays/managed-cert/cert.yaml", `
apiVersion: networking.gke.io/v1beta1
kind: ManagedCertificate
metadata:
  name: gke-certificate
spec:
  domains:
  - $(hostname)`)
	th.writeK("/manifests/gcp/basic-auth-ingress/overlays/managed-cert", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- cert.yaml
namespace: kubeflow
commonLabels:
  kustomize.component: basic-auth-ingress
`)
	th.writeF("/manifests/gcp/basic-auth-ingress/base/backend-config.yaml", `
apiVersion: cloud.google.com/v1beta1
kind: BackendConfig
metadata:
  name: basicauth-backendconfig
spec:
  # Jupyter uses websockets so we want to increase the timeout.
  timeoutSec: 3600`)
	th.writeF("/manifests/gcp/basic-auth-ingress/base/cloud-endpoint.yaml", `
apiVersion: ctl.isla.solutions/v1
kind: CloudEndpoint
metadata:
  name: $(appName)
spec:
  project: $(project)
  targetIngress:
    name: $(ingressName)
    namespace: $(istioNamespace)
`)
	th.writeF("/manifests/gcp/basic-auth-ingress/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: kf-admin-basic-auth
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kf-admin-basic-auth
subjects:
- kind: ServiceAccount
  name: kf-admin
  namespace: $(istioNamespace)
`)
	th.writeF("/manifests/gcp/basic-auth-ingress/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: kf-admin-basic-auth
rules:
- apiGroups:
  - ""
  resources:
  - services
  - configmaps
  - secrets
  verbs:
  - get
  - list
  - patch
  - update
- apiGroups:
  - extensions
  - networking.k8s.io
  resources:
  - ingresses
  verbs:
  - get
  - list
  - update
  - patch
`)
	th.writeF("/manifests/gcp/basic-auth-ingress/base/config-map.yaml", `
apiVersion: v1
data:
  update_backend.sh: |
    #!/bin/bash
    #
    # A simple shell script to configure the health checks by using gcloud.
    set -x 
    [ -z ${NAMESPACE} ] && echo Error NAMESPACE must be set && exit 1
    [ -z ${SERVICE} ] && echo Error SERVICE must be set && exit 1
    [ -z ${INGRESS_NAME} ] && echo Error INGRESS_NAME must be set && exit 1

    PROJECT=$(curl -s -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/project/project-id)
    if [ -z ${PROJECT} ]; then
      echo Error unable to fetch PROJECT from compute metadata
      exit 1
    fi

    set_health_check() {
      # Activate the service account, allow 5 retries
      if [[ ! -z "${GOOGLE_APPLICATION_CREDENTIALS}" ]]; then
        # TODO(jlewi): As of 0.7 we should always be using workload identity. We can remove it post 0.7.0 once we have workload identity
        # fully working
        # Activate the service account, allow 5 retries
        for i in {1..5}; do gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS} && break || sleep 10; done
      fi      

      # For debugging print out what account we are using
      gcloud auth list

      NODE_PORT=$(kubectl --namespace=${NAMESPACE} get svc ${SERVICE} -o jsonpath='{.spec.ports[0].nodePort}')
      echo node port is ${NODE_PORT}

      while [[ -z ${BACKEND_NAME} ]]; do
        BACKENDS=$(kubectl --namespace=${NAMESPACE} get ingress ${INGRESS_NAME} -o jsonpath='{.metadata.annotations.ingress\.kubernetes\.io/backends}')
        echo "fetching backends info with ${INGRESS_NAME}: ${BACKENDS}"
        BACKEND_NAME=$(echo $BACKENDS | grep -o "k8s-be-${NODE_PORT}--[0-9a-z]\+")
        echo "backend name is ${BACKEND_NAME}"
        sleep 2
      done

      while [[ -z ${BACKEND_SERVICE} ]];
      do BACKEND_SERVICE=$(gcloud --project=${PROJECT} compute backend-services list --filter=name~k8s-be-${NODE_PORT}- --uri);
      echo "Waiting for the backend-services resource PROJECT=${PROJECT} NODEPORT=${NODE_PORT} SERVICE=${SERVICE}...";
      sleep 2;
      done

      while [[ -z ${HEALTH_CHECK_URI} ]];
      do HEALTH_CHECK_URI=$(gcloud compute --project=${PROJECT} health-checks list --filter=name~${BACKEND_NAME} --uri);
      echo "Waiting for the healthcheck resource PROJECT=${PROJECT} NODEPORT=${NODE_PORT} SERVICE=${SERVICE}...";
      sleep 2;
      done

      echo health check URI is ${HEALTH_CHECK_URI}

      # Since we create the envoy-ingress ingress object before creating the envoy
      # deployment object, healthcheck will not be configured correctly in the GCP
      # load balancer. It will default the healthcheck request path to a value of
      # / instead of the intended /healthz.
      # Manually update the healthcheck request path to /healthz
      if [[ ${HEALTHCHECK_PATH} ]]; then
        echo Running health checks update ${HEALTH_CHECK_URI} with ${HEALTHCHECK_PATH}
        gcloud --project=${PROJECT} compute health-checks update http ${HEALTH_CHECK_URI} --request-path=${HEALTHCHECK_PATH}
      else
        echo Running health checks update ${HEALTH_CHECK_URI} with /healthz
        gcloud --project=${PROJECT} compute health-checks update http ${HEALTH_CHECK_URI} --request-path=/healthz
      fi

      if [[ ${USE_ISTIO} ]]; then
        # Create the route so healthcheck can pass
        kubectl apply -f /var/envoy-config/healthcheck_route.yaml
      fi
    }

    while true; do
      set_health_check
      echo "Backend updated successfully. Waiting 1 hour before updating again."
      sleep 3600
    done
kind: ConfigMap
metadata:
  name: envoy-config
---
apiVersion: v1
data:
  ingress_bootstrap.sh: |
    #!/usr/bin/env bash

    set -x
    set -e

    # This is a workaround until this is resolved: https://github.com/kubernetes/ingress-gce/pull/388
    # The long-term solution is to use a managed SSL certificate on GKE once the feature is GA.

    # The ingress is initially created without a tls spec.
    # Wait until cert-manager generates the certificate using the http-01 challenge on the GCLB ingress.
    # After the certificate is obtained, patch the ingress with the tls spec to enable SSL on the GCLB.

    # Wait for certificate.
    until kubectl -n ${NAMESPACE} get secret ${TLS_SECRET_NAME} 2>/dev/null; do
      echo "Waiting for certificate..."
      sleep 2
    done

    kubectl -n ${NAMESPACE} patch ingress ${INGRESS_NAME} --type='json' -p '[{"op": "add", "path": "/spec/tls", "value": [{"secretName": "'${TLS_SECRET_NAME}'", "hosts":["'${TLS_HOST_NAME}'"]}]}]'

    echo "Done"
kind: ConfigMap
metadata:
  labels:
    ksonnet.io/component: basic-auth-ingress
  name: ingress-bootstrap-config
---
`)
	th.writeF("/manifests/gcp/basic-auth-ingress/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: whoami-app
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: whoami
    spec:
      containers:
      - env:
        - name: PORT
          value: "8081"
        image: gcr.io/cloud-solutions-group/esp-sample-app:1.0.0
        name: app
        ports:
        - containerPort: 8081
        readinessProbe:
          failureThreshold: 2
          httpGet:
            path: /healthz
            port: 8081
            scheme: HTTP
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 5
`)
	th.writeF("/manifests/gcp/basic-auth-ingress/base/ingress.yaml", `
apiVersion: extensions/v1beta1 # networking.k8s.io/v1beta1
kind: Ingress
metadata:
  annotations:
    ingress.kubernetes.io/ssl-redirect: "true"
    kubernetes.io/ingress.global-static-ip-name: $(ipName)
    networking.gke.io/managed-certificates: gke-certificate
  name: $(ingressName)
spec:
  rules:
  - host: $(hostname)
    http:
      paths:
      - backend:
          serviceName: ambassador
          servicePort: 80
        path: /*
`)
	th.writeF("/manifests/gcp/basic-auth-ingress/base/istio-mapping-svc.yaml", `
apiVersion: v1
kind: Service
metadata:
  annotations:
    getambassador.io/config: |-
      ---
      apiVersion: ambassador/v0
      kind:  Mapping
      name: istio-mapping
      prefix_regex: true
      prefix: /(?!whoami|kflogin).*
      rewrite: ""
      service: istio-ingressgateway.istio-system
      precedence: 1
      use_websocket: true
  labels:
    app: istioMappingSvc
    ksonnet.io/component: basic-auth-ingress
  name: istio-mapping-service
  namespace: istio-system
spec:
  ports:
    - port: 80
      targetPort: 8081
  selector:
    app: istioMappingSvc
  type: ClusterIP
`)
	th.writeF("/manifests/gcp/basic-auth-ingress/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kf-admin
`)
	th.writeF("/manifests/gcp/basic-auth-ingress/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  annotations:
    getambassador.io/config: |-
      ---
      apiVersion: ambassador/v0
      kind:  Mapping
      name: whoami-mapping
      prefix: /whoami
      rewrite: /whoami
      service: whoami-app.$(namespace)
  labels:
    app: whoami
  name: whoami-app
spec:
  ports:
  - port: 80
    targetPort: 8081
  selector:
    app: whoami
  type: ClusterIP
`)
	th.writeF("/manifests/gcp/basic-auth-ingress/base/stateful-set.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    service: backend-updater
  name: backend-updater
spec:
  selector:
    matchLabels:
      service: backend-updater
  serviceName: "backend-updater"
  template:
    metadata:
      labels:
        service: backend-updater
    spec:
      containers:
      - command:
        - bash
        - /var/envoy-config/update_backend.sh
        env:
        - name: NAMESPACE
          value: $(namespace)
        - name: SERVICE
          value: ambassador
        - name: HEALTHCHECK_PATH
          value: /whoami
        - name: INGRESS_NAME
          value: $(ingressName)
        image: gcr.io/kubeflow-images-public/ingress-setup:latest
        name: backend-updater
        volumeMounts:
        - mountPath: /var/envoy-config/
          name: config-volume
      serviceAccountName: kf-admin
      volumes:
      - configMap:
          name: envoy-config
        name: config-volume
  # Workaround for https://github.com/kubernetes-sigs/kustomize/issues/677
  volumeClaimTemplates: []
`)
	th.writeF("/manifests/gcp/basic-auth-ingress/base/params.yaml", `
varReference:
- path: metadata/name
  kind: Certificate
- path: metadata/annotations/getambassador.io\/config
  kind: Service
- path: spec/dnsNames
  kind: Certificate
- path: spec/issuerRef/name
  kind: Certificate
- path: metadata/annotations/kubernetes.io\/ingress.global-static-ip-name
  kind: Ingress
- path: spec/commonName
  kind: Certificate
- path: spec/secretName
  kind: Certificate
- path: spec/acme/config/domains
  kind: Certificate
- path: spec/acme/config/http01/ingress
  kind: Certificate
- path: metadata/name
  kind: Ingress
- path: metadata/annotations/certmanager.k8s.io\/issuer
  kind: Ingress
- path: metadata/name
  kind: CloudEndpoint
- path: spec/project
  kind: CloudEndpoint
- path: spec/targetIngress/name
  kind: CloudEndpoint
- path: spec/targetIngress/namespace
  kind: CloudEndpoint
- path: spec/domains
  kind: ManagedCertificate
- path: subjects/namespace
  kind: ClusterRoleBinding
`)
	th.writeF("/manifests/gcp/basic-auth-ingress/base/params.env", `
appName=kubeflow
namespace=kubeflow
hostname=
project=
ipName=
secretName=envoy-ingress-tls
privateGKECluster=false
ingressName=envoy-ingress
istioNamespace=istio-system
`)
	th.writeK("/manifests/gcp/basic-auth-ingress/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- backend-config.yaml
- cloud-endpoint.yaml
- cluster-role-binding.yaml
- cluster-role.yaml
- config-map.yaml
- deployment.yaml
- ingress.yaml
- istio-mapping-svc.yaml
- service-account.yaml
- service.yaml
- stateful-set.yaml
namespace: kubeflow
commonLabels:
  kustomize.component: basic-auth-ingress
images:
- name: gcr.io/kubeflow-images-public/ingress-setup
  newName: gcr.io/kubeflow-images-public/ingress-setup
  newTag: latest
- name: gcr.io/cloud-solutions-group/esp-sample-app
  newName: gcr.io/cloud-solutions-group/esp-sample-app
  newTag: 1.0.0
configMapGenerator:
- name: basic-auth-ingress-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: secretName
  objref:
    kind: ConfigMap
    name: basic-auth-ingress-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.secretName
- name: appName
  objref:
    kind: ConfigMap
    name: basic-auth-ingress-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.appName
- name: namespace
  objref:
    kind: ConfigMap
    name: basic-auth-ingress-parameters
    apiVersion: v1
  fieldref:
    fieldpath: metadata.namespace
- name: hostname
  objref:
    kind: ConfigMap
    name: basic-auth-ingress-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.hostname
- name: project
  objref:
    kind: ConfigMap
    name: basic-auth-ingress-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.project
- name: ipName
  objref:
    kind: ConfigMap
    name: basic-auth-ingress-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.ipName
- name: ingressName
  objref:
    kind: ConfigMap
    name: basic-auth-ingress-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.ingressName
- name: istioNamespace
  objref:
    kind: ConfigMap
    name: basic-auth-ingress-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.istioNamespace
configurations:
- params.yaml
`)
}

func TestBasicAuthIngressOverlaysManagedCert(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/gcp/basic-auth-ingress/overlays/managed-cert")
	writeBasicAuthIngressOverlaysManagedCert(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../gcp/basic-auth-ingress/overlays/managed-cert"
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
