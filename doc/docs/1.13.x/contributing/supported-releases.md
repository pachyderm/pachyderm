# Pachyderm Supported Releases and Features

Pachyderm lists the status for each release and feature, so that you can understand expectations for support and stability.
 
## Supported Releases

Pachyderm supports the latest Generally Available (GA) release and the previous two major and minor GA releases. Releases three or more major and minor versions back are considered End of Life (EOL).

## Release Status by Version

| Version  | Release Status | Support |
| -------- | -------------- | ------- |
| 2.0.x    | alpha          | No      |
| 1.13.x   | GA             | Yes     |
| 1.12.x   | GA             | Yes     |
| 1.11.x   | GA             | Yes     |
| 1.10.x   | EOL            | No      |
| < 1.9.11 | EOL            | No      |

## Releases Under Development

A release under development may undergo several pre-release stages before becoming Generally Available (GA). These pre-releases enable the Pachyderm team to do development and testing in partnership with our users before a release is considered ready for a Generally Availability (GA).

`alpha > beta > Release Candidate (RC) > Generally Available (GA)`

## Release Status

### Alpha

`alpha` releases are a pre-release version of a product, intended for development and testing purposes only. `alpha` releases include many bugs and unfinished features, and are only suitable for early technical feedback. `alpha` releases should not be used in a production environment.

### Beta

`beta` releases are a pre-release version of a product, intended for development and testing purposes only, and include a wider range of users than an `alpha` release. `beta` releases should not be used in a production environment.

### Release Candidate (RC)

`Release Candidate` or `RC` releases are a pre-release version of a product, intended for users to prepare for a `GA` release. `RC` releases should not be used in a production environment.

### Generally Available (GA)

`Generally Available` or `GA` releases are considered stable and intended for production usage.

- Contain new features, fixed defects, and patched security vulnerabilities.
- Support is available from Pachyderm.

### End of Life (EOL)

`End of Life` or `EOL` indicates the release will no longer receive support.

- Documentation will be archived.
- Release artifacts will remain available. We keep release artifacts on [Github](https://github.com/pachyderm/pachyderm/releases) and [Docker Hub](https://hub.docker.com/u/pachyderm).
- Support is no longer available for End of Life (EOL) releases. Support can assist with upgrading to a newer version.

## Supported Features

### Stable

`stable` indicates that the Pachyderm team believes the feature is ready for use in a production environment.

- The feature's API is stable and unlikely to change.
- There are no major defects for the feature.
- The Pachyderm team believes there is a sufficient amount of testing, including automated tests, community testing, and user production environments.
- Support is available from Pachyderm.

### Experimental

`experimental` indicates that a feature has not met the Pachyderm team's criteria for production use. Therefore, these features should be used with caution in production environments. `experimental` features are likely to change, have outstanding defects, and/or missing documentation. Users considering using `experimental` features should contact Pachyderm for guidance.

- Production use is not recommended without guidance from Pachyderm.
- These features may have missing documentation, lack of examples, and lack of content.
- Support is available from Pachyderm, which may be limited in scope based on our guidance.

### Deprecated

`deprecated` indicates that a feature is no longer developed. Users of deprecated features are encouraged to upgrade or migrate to newer versions or compatible features. `deprecated` features become `End of Life` (EOL) features after 6 months.

- Users continuing to use deprecated features should contact support to migrate to features.
- Support is available from Pachyderm. 

### End of Life (EOL) Features

`End of Life` or `EOL` indicates that a feature is no longer supported.

- Documentation will be archived.
- Support is no longer available for End of Life (EOL) features. Support can assist upgrading to a newer version.

## Experimental Features

| Feature           | Version | Date       |
| ----------------- | --------| ---------- |
| Build Pipelines   | 1.11.0  | 2019-01-05 |
| Service Pipelines | 1.9.9   | 2019-11-06 |
| Git Inputs        | 1.4.0   | 2017-03-27 |

## Deprecated Features

| Feature             | Version |EOL Date   |
| ------------------- | --------| ---------- |
| Some [`pachctl deploy` flags](../../deploy-manage/deploy/helm_install/#pachctl-deploy-flag-deprecation)|  1.13.0 | 2021-06-15|
| Spouts: Named Pipes | 1.12.0  | 2021-07-05 |
| Vault Plugin        | 1.12.5  | 2021-09-12 |
| Build Pipelines     | 1.13.3  | 2021-07-25  |
| Git Inputs          | 1.13.3  | 2021-07-25  |
| `pachctl deploy`    | 1.13.3  | 2021-07-25  |


## End of Life (EOL) Features

| Feature           | Version | EOL Date   |
| ----------------- | --------| ---------- |
| S3v2 signatures   | 1.12.0  | 2021-01-05 |
| atom inputs       | 1.9.0   | 2019-06-12 |
