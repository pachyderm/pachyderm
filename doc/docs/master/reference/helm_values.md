# Helm Chart Values

This document discusses each of the fields present in the `values.yaml` that can be used to deploy with Helm.
To see how to use a helm values files to customize your deployment, refer to our [Helm Deployment Documentation](../../deploy-manage/deploy/helm_install/) section.

!!! Note
    Values that are unchanged from the defaults can be omitted from the values file you supply at installation.

## Values.yaml
Find the complete list of fields available here. Each section is detailed further in its own sub-chapter.
```yaml
imagePullSecret: ""

deployTarget: ""

global:
  postgresql:
    postgresqlUsername: "pachyderm"
    postgresqlPassword: "elephantastic"
    postgresqlDatabase: "pachyderm"
    postgresqlHost: "postgres"
    postgresqlPort: "5432"
    postgresqlSSL: "disable"
  imagePullSecrets: []

console:
  enabled: false
  image:
    repository: "pachyderm/haberdashery"
    pullPolicy: "IfNotPresent"
    tag: "41f09332d40f1e897314f0529fd5bbda37fc069e"
  podLabels: {}
  resources: {}
  config:
    issuerURI: ""
    reactAppRuntimeIssuerURI: ""
    oauthRedirectURI: ""
    oauthClientID: ""
    oauthClientSecret: ""
    graphqlPort: 4000
    oauthPachdClientID: ""
    pachdAddress: "pachd-peer.default.svc.cluster.local:30653"

  service:
    labels: {}
    type: ClusterIP

etcd:
  affinity: {}
  dynamicNodes: 1
  image:
    repository: "pachyderm/etcd"
    tag: "v3.3.5"
    pullPolicy: "IfNotPresent"
  podLabels: {}
  resources: {}
  storageClass: ""
  storageSize: ""
  service:
    annotations: {}
    labels: {}
    type: ClusterIP

enterpriseServer:
  enabled: false
  affinity: {}
  service:
    type: ClusterIP
  tls:
    enabled: false
    secretName: ""
    newSecret:
      create: false
      crt: ""
      key: ""
  resources: {}
  podLabels: {}
  clusterDeploymentID: ""
  image:
    repository: "pachyderm/pachd"
    pullPolicy: "IfNotPresent"
    tag: ""

ingress:
  enabled: false
  annotations: {}
  host: ""
  tls:
    enabled: false
    secretName: ""
    newSecret:
      create: false
      crt: ""
      key: ""

pachd:
  enabled: true
  affinity: {}
  clusterDeploymentID: ""
  goMaxProcs: 0
  image:
    repository: "pachyderm/pachd"
    pullPolicy: "IfNotPresent"
    tag: ""
  logLevel: "info"
  lokiLogging: false
  metrics:
    enabled: true
    endpoint: ""
  podLabels: {}
  resources: {}

  requireCriticalServersOnly: false

  externalService:
    enabled: false
    loadBalancerIP: ""
    apiGRPCPort: 30650
    s3GatewayPort: 30600
  service:
    labels: {}
    type: "ClusterIP"

  enterpriseLicenseKey: ""
  rootToken: ""
  enterpriseSecret: ""
  oauthClientId: "pachd"
  oauthClientSecret: ""
  oauthIssuer: ""
  oauthRedirectURI: ""
  userAccessibleOauthIssuerHost: ""
  upstreamIDPs: []
  mockIDP: false
  serviceAccount:
    create: true
    additionalAnnotations: {}
    name: "pachyderm"
  storage:
    backend: ""
    amazon:
      bucket: ""
      cloudFrontDistribution: ""
      customEndpoint: ""
      disableSSL: false
      id: ""
      logOptions: ""
      maxUploadParts: 10000
      verifySSL: true
      partSize: "5242880"
      region: ""
      retries: 10
      reverse: true
      secret: ""
      timeout: "5m"
      token: ""
      uploadACL: "bucket-owner-full-control"
    google:
      bucket: ""
      cred: ""
    local:
      hostPath: ""
      requireRoot: true
    microsoft:
      container: ""
      id: ""
      secret: ""
    minio:
      bucket: ""
      endpoint: ""
      id: ""
      secret: ""
      secure: ""
      signature: ""
    putFileConcurrencyLimit: 100
    uploadConcurrencyLimit: 100
  ppsWorkerGRPCPort: 1080
  tls:
    enabled: false
    secretName: ""
    newSecret:
      create: false
      crt: ""
      key: ""
  worker:
    image:
      repository: "pachyderm/worker"
      pullPolicy: "IfNotPresent"
    serviceAccount:
      create: true
      additionalAnnotations: {}
      name: "pachyderm-worker"
  rbac:
    create: true
    clusterRBAC: true

pgbouncer:
  service:
    type: ClusterIP
  resources: {}

postgresql:
  enabled: true
  image:
    tag: "13.3.0"
  initdbScripts:
    dex.sh: |
      #!/bin/bash
      set -e
      psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
        CREATE DATABASE dex;
        GRANT ALL PRIVILEGES ON DATABASE dex TO "$POSTGRES_USER";
      EOSQL
  fullnameOverride: postgres
  persistence:
    storageClass: ""
    size: 10Gi
    labels:
      suite: pachyderm

cloudsqlAuthProxy:
  connectionName: ""
  serviceAccount: ""
  port: 5432
  enabled: false
  image:
    repository: "gcr.io/cloudsql-docker/gce-proxy"
    pullPolicy: "IfNotPresent"
    tag: "1.23.0"
  podLabels: {}
  resources: {}

  service:
    labels: {}
    type: ClusterIP
```
!!! Example "In Pratice"
    You rarely need to specify all the fields.
    Most fields either come with sensible defaults or can be empty.
    The following text is an example of a minimum spec that includes the console:

    ```yaml
    deployTarget: LOCAL

    console:
      enabled: true
      config:
        issuerURI: "http://localhost:30658/"
        oauthRedirectURI: "http://localhost:4000/oauth/callback/?inline=true"
        oauthClientID: "console"
        oauthClientSecret: "abc"
        graphqlPort: 4000
        oauthPachdClientID: "pachd"
        pachdAddress: "pachd-peer.default.svc.cluster.local:30653"
    ```

### deployTarget

`deployTarget` is where you're deploying pachyderm. It configures the storage backend to use and cloud provider settings.
It must be one of:

- `GOOGLE`
- `AMAZON`
- `MINIO`
- `MICROSOFT`
- `CUSTOM`
- `LOCAL`

### global

#### global.postgreSQL

This section is to configure the connection to the postgresql database. By default, it uses the included postgres service.

- `postgresqlUsername` is the username to access the pachyderm and dex databases
- `postgresqlPassword` to access the postgresql database. If blank, a value will be generated by the postgres subchart
When using autogenerated value for the initial install, it must be pulled from the postgres secret and added to values.yaml for future helm upgrades.
- `postgresqlDatabase` is the database name where pachyderm data will be stored
- `postgresqlHost` is the postgresql database host to connect to.
- `postgresqlPort` is the postgresql database port to connect to.`postgresqlSSL` is the SSL mode to use for connecting to Postgres, for the default local postgres it is disabled

#### global.imagePullSecrets

`imagePullSecrets` allow you to pull images from private repositories, these will also be added to pipeline workers https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/

Example:
```yaml
  imagePullSecrets:
    - regcred
```
### console

This section is to configure the Pachyderm UI (`console`) which requires an enterprise license key. It is disabled by default.

- `console.enabled` turns on the deployment of the UI.

- `console.image` sets the image to use for the console. This can be left at the defaults unless instructed.

- `console.podLables` specifies lables to add to the console pod.

- `console.resources` specifies resources and limits in standard kubernetes format. It is left unset by default.

- `console.service.labels` specifies labels to add to the console service.

- `console.service.type` specifies the Kubernetes type of the console service. The default is `ClusterIP`.

#### console.config

This is where the primary configuration settings for the console are configured, including authentication.

- `config.issuerURI` is the pachd oauth address accessible to console within the kubernetes cluster. The default is generally fine here.

- `config.reactAppRuntimeIssuerURI` this is the pachd oauth address thats accesible to clients outside of the cluster itself. When running local with `kubectl port-forward` this would be set to localhost (`"http://localhost:30658/"`). Otherwiswe this has to be an address acessible to clients.

- `config.oauthRedirectURI` this is the oauth callback address within console that the pachd oauth service would redirect to. It's the URL of console with `/oauth/callback/?inline=true` appended. Running locally its therefore `"http://localhost:4000/oauth/callback/?inline=true"`.

- `config.oauthClientID` the client identifier for the Console with pachd

- `config.oauthClientSecret` the secret configured for the client with pachd

- `config.graphqlPort` the http port that the console service will be accessible on.

- `config.oauthPachdClientID` the identifier for pachd's oauth client.

- `config.pachdAddress` the address that console can access pachd at. It must be set if you install pachyderm in a different namespace than default. The format is `"pachd-peer.<namespace>.svc.cluster.local:30653"`

### etcd

This section is to configure the etcd cluster in the deployment.

- `etcd.image` sets the image to use for the etcd. This can be left at the defaults unless instructed.

- `etcd.podLables` specifies lables to add to the etcd pods.

- `etcd.resources` specifies resources and limits in standard kubernetes format. It is left unset by default.

- `etcd.dynamicNodes` sets the number of nodes in the etcd StatefulSet. The default is 1.

- `etcd.storageClass` indicates the etcd should use an existing StorageClass for its storage. If left blank, a storageClass will be created.

- `etcd.storageSize` specifies the size of the volume to use for etcd. Etcd does not require much space. For storage that scales IOPs with size, the size must be set large enought to provide at least 1000 IOPs for performance. If you do not specify, it will default to 256Gi on Azure and 100Gi on GCP/AWS for that reason.

- `etcd.service.labels` specifies labels to add to the console service.
- `etcd.service.annotations` specifies annotations to add to the etcd service.
- `etcd.service.type` specifies the Kubernetes type of the etcd service. The default is `ClusterIP`.

### enterpriseServer

This section is to configure the Enterprise Server deployment (if desired).

- `enterpriseServer.enabled` turns on the deployment of the Enterprise Server. It is disabled by default.

- `enterpriseServer.service.type` specifies the Kubernetes type of the console service. The default is `ClusterIP`.

- `enterpriseServer.resources` specifies resources and limits in standard kubernetes format. It is left unset by default.

- `enterpriseServer.podLables` specifies lables to add to the enterpriseServer pod.

- `enterpriseServer.image` sets the image to use for the etcd. This can be left at the defaults unless instructed.

#### enterpriseServer.tls

There are three options for configuring TLS on the Enterprise Server under `enterpriseServer.tls`.

1. `disabled`. TLS is not used.
1. `enabled`, using an existing secret. You must set enabled to true and provide a secret name where the exiting cert and key are stored.
1. `enabled`, using a new secret. You must set enabled to true and `newSecret.create` to true and specify a secret name, and a cert and key in string format

### ingress

This section is to configure an ingress resource for an existing ingress controller.

- `ingress.enabled` turns on the creation of the ingress for the UI.

- `ingress.annotations` specifies annotations to add to the ingress resource.

#### ingress.tls

There are three options for configuring TLS on the ingress under `ingress.tls`.

1. `disabled`. TLS is not used.
1. `enabled`, using an existing secret. You must set enabled to true and provide a secret name where the exiting cert and key are stored.
1. `enabled`, using a new secret. You must set enabled to true and `newSecret.create` to true and specify a secret name, and a cert and key in string format

### pachd

This section is to configure the pachd deployment.

- `pachd.enabled` turns on the deployment of pachd.

- `pachd.image` sets the image to use for pachd. This can be left at the defaults unless instructed.

- `pachd.logLevel` sets the logging level. `info` is default.

- `pachd.lokiLogging` enables Loki logging if set.

- `pachd.podLables` specifies lables to add to the pachd pod.

- `pachd.resources` specifies resources and limits in standard kubernetes format. It is left unset by default.

- `pachd.requireCriticalServersOnly` only requires the critical pachd servers to startup and run without errors.

- `pachd.service.labels` specifies labels to add to the console service.

- `pachd.externalService.enabled` creates a kubernetes service of type `loadBalancer` that is safe to expose externally.

- `pachd.externalService.loadBalancerIP` optionally supply the existing IP address of the load balancer.

- `pachd.externalService.apiGRPCPort` is the desired api GRPC port (30650 is default).

- `pachd.externalService.s3GatewayPort` is the desired s3 gateway port (30600 is default).

- `pachd.service.type` specifies the Kubernetes type of the pachd service. The default is `ClusterIP`.

- `pachd.enterpriseLicenseKey` specify the enterprise license key if you have one.

- `pachd.rootToken` is the auth token used to communicate with the cluster as the root user.

- `pachd.enterpriseSecret` specifies the enterprise cluster secret.

- `pachd.oauthClientId` specifies the Oauth client ID representing pachd.

- `pachd.oauthClientSecret` specifies the Oauth client secret.

- `pachd.oauthIssuer` specifies the Oauth Issuer.

- `pachd.oauthRedirectURI` specifies the Oauth redirect URI served by pachd.

- `pachd.userAccessibleOauthIssuerHost` specifies the Oauth issuer's address host that's used in the Oauth authorization redirect URI. This value is only necessary in local settings or anytime the registered Issuer address isn't accessible outside the cluster.

- `pachd.upstreamIDPs` specifies a list of Identity Providers to use for authentication.

- `pachd.mockIDP` when set to `true`, specifes to ignore `upstreamIDPs` in favor of a placeholder IDP with a preset username/password.

If any of `rootToken`,`enterpriseSecret`, or `oauthClientSecret` are blank, a value will be generated automatically. When using autogenerated value for the initial install, it must be pulled from the config secret and added to values.yaml for future helm upgrades.

- `pachd.serviceAccount.create` creates a kubernetes service account for pachd. Default is true.

- `pachd.rbac.create`  indicates whether RBAC resources should be created. Default is true.

- `pachd.rbac.clusterRBAC` indicates that ClusterRole and ClusterRoleBinding should be used rather than Role and RoleBinding. Default is true.

#### pachd.storage

This section of `pachd` configures the back end storage for pachyderm.

- `storage.backend` configures the storage backend to use. It must be one of GOOGLE, AMAZON, MINIO, MICROSOFT or LOCAL. This is set automatically if deployTarget is GOOGLE, AMAZON, MICROSOFT, or LOCAL.

- `storage.putFileConcurrencyLimit` sets the maximum number of files to upload or fetch from remote sources (HTTP, blob storage) using PutFile concurrently. 

- `storage.uploadConcurrencyLimit` sets the maximum number of concurrent object storage uploads per Pachd instance.
##### pachd.storage.amazon

If you're using Amazon S3 as your storage backend, configure it here.

- `storage.amazon.bucket` sets the S3 bucket to use.

- `storage.amazon.cloudFrontDistribution` sets the CloudFront distribution in the storage secrets.

- `storage.amazon.customEndpoint` sets a custom s3 endpoint.

- `storage.amazon.disableSSL` disables SSL.

- `storage.amazon.id` sets the Amazon access key ID to use.

- `storage.amazon.logOptions` sets various log options in Pachyderm’s internal S3 client.  Comma-separated list containing zero or more of: 'Debug', 'Signing', 'HTTPBody', 'RequestRetries','RequestErrors', 'EventStreamBody', or 'all'
 (case-insensitive).  See 'AWS SDK for Go' docs for details.

- `storage.amazon.maxUploadParts` sets the maximum number of upload parts. Default is `10000`.

- `storage.amazon.verifySSL` performs SSL certificate verification.

- `storage.amazon.partSize` sets the part size for object storage uploads. It has to be a string due to Helm and YAML parsing integers as floats.

- `storage.amazon.region` sets the AWS region to use.

- `storage.amazon.retries` sets the number of retries for object storage requests..

- `storage.amazon.reverse` reverses object storage paths.

- `storage.amazon.secret` sets the Amazon secret access key to use.

- `storage.amazon.timeout` sets the timeout for object storage requests.

- `storage.amazon.token` optionally sets the Amazon token to use.

- `storage.amazon.uploadACL` sets the upload ACL for object storage uploads.

##### pachd.storage.google

If you're using Google Storage Buckets as your storage backend, configure it here.

- `storage.google.bucket` sets the object bucket to use.

- `storage.google.cred` is a string containing a GCP service account private key, in object (JSON or YAML) form.  A simple way to pass this on the command line is with the set-file flag, e.g.:

  ```shell
  helm install pachd -f my-values.yaml --set-file storage.google.cred=creds.json pach/pachyderm
  ```

  Example:

  ```yaml
  cred: |
    {
      "type": "service_account",
      "project_id": "…",
      "private_key_id": "…",
      "private_key": "-----BEGIN PRIVATE KEY-----\n…\n-----END PRIVATE KEY-----\n",
      "client_email": "…@….iam.gserviceaccount.com",
      "client_id": "…",
      "auth_uri": "https://accounts.google.com/o/oauth2/auth",
      "token_uri": "https://oauth2.googleapis.com/token",
      "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
      "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/…%40….iam.gserviceaccount.com"
    }
  ```

- `storage.google.local.hostpath` indicates the path on the host where the PFS metadata will be stored.

##### pachd.storage.microsoft

If you're using Microsoft Blob Storage as your storage backend, configure it here.

- `storage.microsoft.container` sets the blob storage container.

- `storage.microsoft.id` sets the access key ID to use.

- `storage.microsoft.secret` sets the secret access key to use.

##### pachd.storage.minio

If you're using [MinIO](https://min.io/) as your storage backend, configure it here.

- `storage.minio.bucket` sets the bucket to use.

- `storage.minio.endpoint` sets the object endpoint.

- `storage.minio.id` sets the access key ID to use.

- `storage.minio.secret` sets the secret access key to use.

- `storage.minio.secure` set to true for a secure connection.

- `storage.minio.signature` sets the signature version to use.

#### pachd.tls

There are three options for configuring TLS on pachd under `pachd.tls`.

1. `disabled`. TLS is not used.
1. `enabled`, using an existing secret. You must set enabled to true and provide a secret name where the exiting cert and key are stored.
1. `enabled`, using a new secret. You must set enabled to true and `newSecret.create` to true and specify a secret name, and a cert and key in string format.

### pgbouncer

This section is to configure the PGBouncer Postgres connection pooler.

- `service.type` specifies the Kubernetes type of the pgbouncer service. The default is `ClusterIP`.

- `resources` specifies resources and limits in standard kubernetes format. It is left unset by default.

### postgresql

This section is to configure the PostgresQL Subchart, if used.

- `enabled`  controls whether to install postgres or not. If not using the built in Postgres, you must specify a Postgresql database server to connect to in `global.postgresql`. The enabled value is watched by the 'condition' set on the Postgresql dependency in Chart.yaml

- `image.tag` sets the postgres version. Leave at the default unless instructed otherwise.

- `initdbScripts` creates the inital `dex` database that's needed for pachyderm. Leave at the default unless instructed otherwise.

- `persistence.storageClass` specifies the storage class for the postgresql Persistent Volume (PV)

!!! Note "More"
    See notes in Bitnami chart values.yaml file for more information.
    More info for setting up storage classes on various cloud providers:

      - AWS: https://docs.aws.amazon.com/eks/latest/userguide/storage-classes.html
      - GCP: https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/ssd-pd
      - Azure: https://docs.microsoft.com/en-us/azure/aks/concepts-storage

- `storageSize` specifies the size of the volume to use for postgresql.

!!! Attention
    - Recommended Minimum Disk size for Microsoft/Azure: 256Gi  - 1,100 IOPS https://azure.microsoft.com/en-us/pricing/details/managed-disks/
    - Recommended Minimum Disk size for Google/GCP: 50Gi        - 1,500 IOPS https://cloud.google.com/compute/docs/disks/performance
    - Recommended Minimum Disk size for Amazon/AWS: 500Gi (GP2) - 1,500 IOPS https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-volume-types.html

### cloudsqlAuthProxy

This section is to configure the CloudSQL Auth Proxy for deploying Pachyderm on GCP with CloudSQL.

- `connectionName` may be found by running `gcloud sql instances describe INSTANCE_NAME --project PROJECT_ID`

- `serviceAccount` is the account to use to connect to the cloudSql instance.

- `enabled`  controls whether to deploy the cloudsqlAuthProxy. Default is false.

- `port` is the cloudql database port to expose. The default is `5432`

- `service.type` specifies the Kubernetes type of the cloudsqlAuthProxy service. The default is `ClusterIP`.
