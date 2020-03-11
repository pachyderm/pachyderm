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

func writeIapIngressBase(th *KustTestHarness) {
	th.writeF("/manifests/gcp/iap-ingress/base/backend-config.yaml", `
apiVersion: cloud.google.com/v1beta1
kind: BackendConfig
metadata:
  name: iap-backendconfig
spec:
  # Jupyter uses websockets so we want to increase the timeout.
  timeoutSec: 3600
  iap:
    enabled: true
    oauthclientCredentials:
      secretName: $(oauthSecretName)
`)
	th.writeF("/manifests/gcp/iap-ingress/base/cloud-endpoint.yaml", `
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
	th.writeF("/manifests/gcp/iap-ingress/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: kf-admin-iap
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kf-admin-iap
subjects:
- kind: ServiceAccount
  name: kf-admin
`)
	th.writeF("/manifests/gcp/iap-ingress/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: kf-admin-iap
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
- apiGroups:
  - authentication.istio.io
  resources:
  - policies
  verbs:
  - '*'
- apiGroups:
  - networking.istio.io
  resources:
  - gateways
  - virtualservices
  verbs:
  - '*'
`)
	th.writeF("/manifests/gcp/iap-ingress/base/config-map.yaml", `
---
apiVersion: v1
data:
  healthcheck_route.yaml: |
    apiVersion: networking.istio.io/v1alpha3
    kind: VirtualService
    metadata:
      name: default-routes
      namespace: $(namespace)
    spec:
      hosts:
      - "*"
      gateways:
      - kubeflow-gateway
      http:
      - match:
        - uri:
            exact: /healthz
        route:
        - destination:
            port:
              number: 80
            host: whoami-app.kubeflow.svc.cluster.local
      - match:
        - uri:
            exact: /whoami
        route:
        - destination:
            port:
              number: 80
            host: whoami-app.kubeflow.svc.cluster.local
    ---
    apiVersion: networking.istio.io/v1alpha3
    kind: Gateway
    metadata:
      name: kubeflow-gateway
      namespace: $(namespace)
    spec:
      selector:
        istio: ingressgateway
      servers:
      - port:
          number: 80
          name: http
          protocol: HTTP
        hosts:
        - "*"
  setup_backend.sh: |
    #!/usr/bin/env bash
    #
    # A simple shell script to configure the JWT audience used with ISTIO
    set -x
    [ -z ${NAMESPACE} ] && echo Error NAMESPACE must be set && exit 1
    [ -z ${SERVICE} ] && echo Error SERVICE must be set && exit 1
    [ -z ${INGRESS_NAME} ] && echo Error INGRESS_NAME must be set && exit 1

    PROJECT=$(curl -s -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/project/project-id)
    if [ -z ${PROJECT} ]; then
      echo Error unable to fetch PROJECT from compute metadata
      exit 1
    fi

    PROJECT_NUM=$(curl -s -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/project/numeric-project-id)
    if [ -z ${PROJECT_NUM} ]; then
      echo Error unable to fetch PROJECT_NUM from compute metadata
      exit 1
    fi

    # Activate the service account
    if [ ! -z "${GOOGLE_APPLICATION_CREDENTIALS}" ]; then
      # As of 0.7.0 we should be using workload identity and never setting GOOGLE_APPLICATION_CREDENTIALS.
      # But we kept this for backwards compatibility but can remove later.
      gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS}
    fi

    # Print out the config for debugging
    gcloud config list
    gcloud auth list

    set_jwt_policy () {
      NODE_PORT=$(kubectl --namespace=${NAMESPACE} get svc ${SERVICE} -o jsonpath='{.spec.ports[?(@.name=="http2")].nodePort}')
      echo "node port is ${NODE_PORT}"

      BACKEND_NAME=""
      while [[ -z ${BACKEND_NAME} ]]; do
        BACKENDS=$(kubectl --namespace=${NAMESPACE} get ingress ${INGRESS_NAME} -o jsonpath='{.metadata.annotations.ingress\.kubernetes\.io/backends}')
        echo "fetching backends info with ${INGRESS_NAME}: ${BACKENDS}"
        BACKEND_NAME=$(echo $BACKENDS | grep -o "k8s-be-${NODE_PORT}--[0-9a-z]\+")
        echo "backend name is ${BACKEND_NAME}"
        sleep 2
      done

      BACKEND_ID=""
      while [[ -z ${BACKEND_ID} ]]; do
        BACKEND_ID=$(gcloud compute --project=${PROJECT} backend-services list --filter=name~${BACKEND_NAME} --format='value(id)')
        echo "Waiting for backend id PROJECT=${PROJECT} NAMESPACE=${NAMESPACE} SERVICE=${SERVICE} filter=name~${BACKEND_NAME}"
        sleep 2
      done
      echo BACKEND_ID=${BACKEND_ID}

      JWT_AUDIENCE="/projects/${PROJECT_NUM}/global/backendServices/${BACKEND_ID}"
      
      # Use kubectl patch.
      echo patch JWT audience: ${JWT_AUDIENCE}
      kubectl -n ${NAMESPACE} patch policy ingress-jwt --type json -p '[{"op": "replace", "path": "/spec/origins/0/jwt/audiences/0", "value": "'${JWT_AUDIENCE}'"}]'

      echo "Clearing lock on service annotation"
      kubectl patch svc "${SERVICE}" -p "{\"metadata\": { \"annotations\": {\"backendlock\": \"\" }}}"
    }

    while true; do
      set_jwt_policy
      # Every 5 minutes recheck the JWT policy and reset it if the backend has changed for some reason.
      # This follows Kubernetes level based design.
      # We have at least one report see 
      # https://github.com/kubeflow/kubeflow/issues/4342#issuecomment-544653657
      # of the backend id changing over time.
      sleep 300
    done
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

    if [[ ! -z "${GOOGLE_APPLICATION_CREDENTIALS}" ]]; then
      # TODO(jlewi): As of 0.7 we should always be using workload identity. We can remove it post 0.7.0 once we have workload identity
      # fully working
      # Activate the service account, allow 5 retries
      for i in {1..5}; do gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS} && break || sleep 10; done
    fi      

    set_health_check () {
      NODE_PORT=$(kubectl --namespace=${NAMESPACE} get svc ${SERVICE} -o jsonpath='{.spec.ports[?(@.name=="http2")].nodePort}')
      echo node port is ${NODE_PORT}

      while [[ -z ${BACKEND_NAME} ]]; do
        BACKENDS=$(kubectl --namespace=${NAMESPACE} get ingress ${INGRESS_NAME} -o jsonpath='{.metadata.annotations.ingress\.kubernetes\.io/backends}')
        echo "fetching backends info with ${INGRESS_NAME}: ${BACKENDS}"
        BACKEND_NAME=$(echo $BACKENDS | grep -o "k8s-be-${NODE_PORT}--[0-9a-z]\+")
        echo "backend name is ${BACKEND_NAME}"
        sleep 2
      done

      while [[ -z ${BACKEND_SERVICE} ]];
      do BACKEND_SERVICE=$(gcloud --project=${PROJECT} compute backend-services list --filter=name~${BACKEND_NAME} --uri);
      echo "Waiting for the backend-services resource PROJECT=${PROJECT} BACKEND_NAME=${BACKEND_NAME} SERVICE=${SERVICE}...";
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
        # This is basic auth
        echo Running health checks update ${HEALTH_CHECK_URI} with ${HEALTHCHECK_PATH}
        gcloud --project=${PROJECT} compute health-checks update http ${HEALTH_CHECK_URI} --request-path=${HEALTHCHECK_PATH}
      else
        # /healthz/ready is the health check path for istio-ingressgateway
        echo Running health checks update ${HEALTH_CHECK_URI} with /healthz/ready
        gcloud --project=${PROJECT} compute health-checks update http ${HEALTH_CHECK_URI} --request-path=/healthz/ready
        # We need the nodeport for istio-ingressgateway status-port
        STATUS_NODE_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="status-port")].nodePort}')
        gcloud --project=${PROJECT} compute health-checks update http ${HEALTH_CHECK_URI} --port=${STATUS_NODE_PORT}
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
  name: ingress-bootstrap-config
---
`)
	th.writeF("/manifests/gcp/iap-ingress/base/deployment.yaml", `
---
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
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: iap-enabler
spec:
  replicas: 1
  template:
    metadata:
      labels:
        service: iap-enabler
    spec:
      containers:
      - command:
        - bash
        - /var/envoy-config/setup_backend.sh
        env:
        - name: NAMESPACE
          value: $(istioNamespace)
        - name: SERVICE
          value: istio-ingressgateway
        - name: INGRESS_NAME
          value: $(ingressName)
        - name: ENVOY_ADMIN
          value: http://localhost:8001
        - name: USE_ISTIO
          value: "true"
        image: gcr.io/kubeflow-images-public/ingress-setup:latest
        name: iap
        volumeMounts:
        - mountPath: /var/envoy-config/
          name: config-volume
      restartPolicy: Always
      serviceAccountName: kf-admin
      volumes:
      - configMap:
          name: envoy-config
        name: config-volume
`)
	th.writeF("/manifests/gcp/iap-ingress/base/ingress.yaml", `
apiVersion: extensions/v1beta1 # networking.k8s.io/v1beta1
kind: Ingress
metadata:
  annotations:
    ingress.kubernetes.io/ssl-redirect: "true"
    kubernetes.io/ingress.global-static-ip-name: $(ipName)
    networking.gke.io/managed-certificates: gke-certificate
  name: envoy-ingress
spec:
  rules:
  - host: $(hostname)
    http:
      paths:
      - backend:
          serviceName: istio-ingressgateway
          servicePort: 80
        path: /*
`)
	th.writeF("/manifests/gcp/iap-ingress/base/policy.yaml", `
apiVersion: authentication.istio.io/v1alpha1
kind: Policy
metadata:
  name: ingress-jwt
spec:
  origins:
  - jwt:
      audiences:
      - TO_BE_PATCHED
      issuer: https://cloud.google.com/iap
      jwksUri: https://www.gstatic.com/iap/verify/public_key-jwk
      jwtHeaders:
      - x-goog-iap-jwt-assertion
      trigger_rules:
      - excluded_paths:
        - exact: /healthz/ready
        - prefix: /.well-known/acme-challenge
  principalBinding: USE_ORIGIN
  targets:
  - name: istio-ingressgateway
    ports:
    - number: 80
`)
	th.writeF("/manifests/gcp/iap-ingress/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kf-admin
`)
	th.writeF("/manifests/gcp/iap-ingress/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
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
	th.writeF("/manifests/gcp/iap-ingress/base/stateful-set.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    service: backend-updater
  name: backend-updater
spec:
  serviceName: backend-updater
  selector:
    matchLabels:
      service: backend-updater
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
          value: $(istioNamespace)
        - name: SERVICE
          value: istio-ingressgateway
        - name: INGRESS_NAME
          value: $(ingressName)
        - name: USE_ISTIO
          value: "true"
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
  volumeClaimTemplates: []
`)
	th.writeF("/manifests/gcp/iap-ingress/base/params.yaml", `
varReference:
- path: metadata/name
  kind: Certificate
- path: spec/origins/jwt/issuer
  kind: Policy
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
- path: spec/rules/host
  kind: Ingress
- path: metadata/annotations/certmanager.k8s.io\/issuer
  kind: Ingress
- path: spec/template/spec/volumes/secret/secretName
  kind: Deployment
- path: spec/template/spec/volumes/secret/secretName
  kind: StatefulSet
- path: metadata/name
  kind: CloudEndpoint
- path: spec/project
  kind: CloudEndpoint
- path: spec/targetIngress/name
  kind: CloudEndpoint
- path: spec/targetIngress/namespace
  kind: CloudEndpoint
- path: spec/iap/oauthclientCredentials/secretName
  kind: BackendConfig
- path: data/healthcheck_route.yaml
  kind: ConfigMap
- path: spec/domains
  kind: ManagedCertificate`)
	th.writeF("/manifests/gcp/iap-ingress/base/params.env", `
namespace=kubeflow
appName=kubeflow
hostname=
ingressName=envoy-ingress
ipName=
oauthSecretName=kubeflow-oauth
project=
adminSaSecretName=admin-gcp-sa
tlsSecretName=envoy-ingress-tls
istioNamespace=istio-system`)
	th.writeK("/manifests/gcp/iap-ingress/base", `
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
- policy.yaml
- service-account.yaml
- service.yaml
- stateful-set.yaml
namespace: kubeflow
commonLabels:
  kustomize.component: iap-ingress
images:
- name: gcr.io/kubeflow-images-public/envoy
  newName: gcr.io/kubeflow-images-public/envoy
  newTag: v20180309-0fb4886b463698702b6a08955045731903a18738
- name: gcr.io/kubeflow-images-public/ingress-setup
  newName: gcr.io/kubeflow-images-public/ingress-setup
  newTag: latest
- name: gcr.io/cloud-solutions-group/esp-sample-app
  newName: gcr.io/cloud-solutions-group/esp-sample-app
  newTag: 1.0.0
configMapGenerator:
- name: parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: namespace
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace
- name: appName
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.appName
- name: hostname
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.hostname
- name: ipName
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.ipName
- name: ingressName
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.ingressName
- name: oauthSecretName
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.oauthSecretName
- name: project
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.project
- name: adminSaSecretName
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.adminSaSecretName
- name: tlsSecretName
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.tlsSecretName
- name: istioNamespace
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.istioNamespace
configurations:
- params.yaml
`)
}

func TestIapIngressBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/gcp/iap-ingress/base")
	writeIapIngressBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../gcp/iap-ingress/base"
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
