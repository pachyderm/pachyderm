# Helm Chart Values 

This document discusses each of the fields present in the `values.yaml` that can be used to deploy with helm.
To see how to use a helm values files to customize your deployment, refer to the [pachctl
create pipeline](./pachctl/pachctl_create_pipeline.md) section.

Values that are unchanaged from the defaults can be omitted from the values file you supply at installation.

## Values.yaml


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

rbac:
  create: true
  clusterRBAC: true

worker:
  image:
    repository: "pachyderm/worker"
    pullPolicy: "IfNotPresent"
  serviceAccount:
    create: true
    additionalAnnotations: {}
    name: "pachyderm-worker"

pgbouncer:
  service:
    type: ClusterIP
  resources: {}


postgresql:
  enabled: true
  storageClass: ""
  service:
    type: "ClusterIP"
  resources: {}


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

 
In practice, you rarely need to specify all the fields.
Most fields either come with sensible defaults or can be empty.
The following text is an example of a minimum spec that includes the console:

```yaml
deployTarget: LOCAL

console:
  enabled: true
  config:
    issuerURI: "http://pachd:1658"
    oauthRedirectURI: "http://localhost:4000/oauth/callback/?inline=true"
    reactAppRuntimeIssuerURI: "http://localhost:30658/"
    oauthClientID: "dash"
    oauthClientSecret: "notsecret"
    graphqlPort: 4000
    oauthPachdClientID: "pachd"
    pachdAddress: "pachd-peer.default.svc.cluster.local:30653"
```

### deployTarget

`deployTarget` is where you're deploying pachyderm.  It configures the storage backend to use and cloud provider settings (storage classes, etc). 
It must be one of 
- `GOOGLE`
- `AMAZON`
- `MINIO`
- `MICROSOFT`
- `CUSTOM`
- `LOCAL`

### Global

#### PostgreSQL

This section is to configure the connection to the postgresql database. By default, it uses the included postgres service.


### Console

This section is to configure the Pachyderm UI (`console`) which requires an enterprise license key. It is disabled by default.

`console.enabled` turns on the deployment of the UI.

`console.image` sets the image to use for the console. This can be left at the defaults unless instructed.

`console.podLables` specifies lables to add to the console pod.

`console.resources` specifies resources and limits in standard kubernetes format. It is left unset by default.

`console.service.labels` specifies labels to add to the console service.

`console.service.type` specifies the Kubernetes type of the console service. The default is `ClusterIP`.

#### Config 

This is where the primary configuration settings for the console are configured, including authentication.

`config.issuerURI` is the pachd oauth address accessible to console within the kubernetes cluster. The default is generally fine here.
`config.reactAppRuntimeIssuerURI`  this is the pachd oauth address thats accesible to clients outside of the cluster itself. When running local with `kubectl port-forward` this would be set to localhost (`"http://localhost:30658/"`). Otherwiswe this has to be an address acessible to clients.
`config.oauthRedirectURI` this is the oauth callback address within console that the pachd oauth service would redirect to. It's the URL of console with `/oauth/callback/?inline=true` appended. Running locally its therefore `"http://localhost:4000/oauth/callback/?inline=true"`.
`config.oauthClientID` the client identifier for the Console with pachd
`config.oauthClientSecret` the secret configured for the client with pachd
`config.graphqlPort` the http port that the console service will be accessible on.
`config.oauthPachdClientID` the identifier for pachd's oauth client.
`config.pachdAddress` the address that console can access pachd at. It must be set if you install pachyderm in a different namespace than default. The format is `"pachd-peer.<namespace>.svc.cluster.local:30653"`

### Etcd

This section is to configure the etcd cluster in the deployment.

`etcd.image` sets the image to use for the etcd. This can be left at the defaults unless instructed.

`etcd.podLables` specifies lables to add to the etcd pods.

`etcd.resources` specifies resources and limits in standard kubernetes format. It is left unset by default.

`etcd.dynamicNodes` sets the number of nodes in the etcd StatefulSet. The default is 1.

`etcd.storageClass` indicates the etcd should use an existing StorageClass for its storage. If left blank, a storageClass will be created.

`etcd.storageSize` specifies the size of the volume to use for etcd. Etcd does not require much space. For storage that scales IOPs with size, the size must be set large enought to provide at least 1000 IOPs for performance. If you do not specify, it will default to 256Gi on Azure and 100Gi on GCP/AWS for that reason.

`etcd.service.labels` specifies labels to add to the console service.
`etcd.service.annotations` specifies annotations to add to the etcd service.
`etcd.service.type` specifies the Kubernetes type of the console service. The default is `ClusterIP`.

### Enterprise Server

This section is to configure the Enterprise Server deployment (if desired). 

`enterpriseServer.enabled` turns on the deployment of the Enterprise Server. It is disabled by default.

`enterpriseServer.service.type` specifies the Kubernetes type of the console service. The default is `ClusterIP`.

`enterpriseServer.resources` specifies resources and limits in standard kubernetes format. It is left unset by default.

`enterpriseServer.podLables` specifies lables to add to the enterpriseServer pod.

`enterpriseServer.image` sets the image to use for the etcd. This can be left at the defaults unless instructed.

#### TLS

There are three options for configuring TLS on the Enterprise Server under `enterpriseServer.tls`.

1. Disabled. TLS is not used.
1. Enabled, using an existing secret. You must set enabled to true and provide a secret name where the exiting cert and key are stored.
1. Enabled, using a new secret. You must set enabled to true and `newSecret.create` to true and specify a secret name, and a cert and key in string format



### Ingress

This section is to configure an ingress resource for an existing ingress controller.

`ingress.enabled` turns on the creation of the ingress for the UI.

`ingress.annotations` specifies annotations to add to the ingress resource.

#### TLS

There are three options for configuring TLS on the Ingress under `ingress.tls`.

1. Disabled. TLS is not used.
1. Enabled, using an existing secret. You must set enabled to true and provide a secret name where the exiting cert and key are stored.
1. Enabled, using a new secret. You must set enabled to true and `newSecret.create` to true and specify a secret name, and a cert and key in string format

### Pachd

This section is to configure the pachd deployment.

`pachd.enabled` turns on the deployment of pachd.

`pachd.image` sets the image to use for pachd. This can be left at the defaults unless instructed.

`pachd.podLables` specifies lables to add to the pachd pod.

`pachd.resources` specifies resources and limits in standard kubernetes format. It is left unset by default.

`pachd.service.labels` specifies labels to add to the console service.

`pachd.service.type` specifies the Kubernetes type of the console service. The default is `ClusterIP`.

#### Storage

This section of `pachd` configures the back end storage for pachyderm.

`backend` configures the storage backend to use.  It must be one of GOOGLE, AMAZON, MINIO, MICROSOFT or LOCAL. This is set automatically if deployTarget is GOOGLE, AMAZON, MICROSOFT, or LOCAL.

##### Amazon

If you're using Amazon S3 as your storage backend, configure it here.

`bucket` sets the S3 bucket to use.

`cloudFrontDistribution` sets the CloudFront distribution in the storage secrets.

