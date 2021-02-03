# Pachyderm Supported Releases and Features

Pachyderm lists the status for each release and feature, so that you can understand expectations for support and stability.
 
## Supported Releases

Pachyderm supports the latest Generally Available (GA) release and the previous two major and minor GA releases. Releases three or more major and minor versions back are considered End of Life (EOL).

## Release Status by Version

<table>
<colgroup>
<col style="width: 25%" />
<col style="width: 25%" />
<col style="width: 25%" />
</colgroup>
<thead>
<tr class="header">
<th>Version</th>
<th>Release Status</th>
<th>Support</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td>2.0.0</td>
<td>beta</td>
<td>No</td>
</tr>
<tr class="even">
<td>1.12.1</td>
<td>GA</td>
<td>Yes</td>
</tr>
<tr class="odd">
<td>1.11.9</td>
<td>GA</td>
<td>Yes</td>
</tr>
<tr class="even">
<td>1.10.5</td>
<td>GA</td>
<td>Yes</td>
</tr>
<tr class="odd">
<td> < 1.9.11</td>
<td>EOL</td>
<td>No</td>
</tr>
</tbody>
</table>

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

### Unstable

`unstable` indicates that a feature has not met the Pachyderm team's critera for production use. Therefore, these features should be used with caution in production environments. `unstable` features are likely to change, have outstanding defects, and/or missing documentation. Users considering using `unstable` features should contact Pachyderm for guidance.

- Production use is not recommended without guidance from Pachyderm.
- These features may have missing documentation, lack of examples, and lack of content.
- Support is available from Pachyderm, which may be limited in scope based on our guidance.

### Deprecated

`deprecated` indicates that a feature is no longer developed. Users of deprecated features are encouraged to upgrade or migrate to newer versions or compatible features. `deprecated` features become `End of Life` (EOL) features after 6 months.

- Releases are only provided for critical defects or security vulnerabilities.
- Users continuing to use deprecated features should contact support to migrate to features.
- Support is available from Pachyderm. 

### End of Life (EOL) Features

`End of Life` or `EOL` inidicates that a feature is no longer supported.

- Documentation will be archived.
- Support is no longer available for End of Life (EOL) features. Support can assist upgrading to a newer version.

## Unstable Features

<table>
<colgroup>
<col style="width: 25%" />
<col style="width: 25%" />
<col style="width: 25%" />
<col style="width: 25%" />
</colgroup>
<thead>
<tr class="header">
<th>Feature</th>
<th>Version</th>
<th>Date</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td>Build Pipelines</td>
<td>1.11.0</td>
<td>2019-01-05</td>
</tr>
<tr class="even">
<td>Service Pipelines</td>
<td>1.9.9</td>
<td>2019-11-06</td>
</tr>
<tr class="odd">
<td>Git Inputs</td>
<td>1.4.0</td>
<td>2017-03-27</td>
</tr>
</tbody>
</table>

## Deprecated Features

<table>
<colgroup>
<col style="width: 25%" />
<col style="width: 25%" />
<col style="width: 25%" />
<col style="width: 25%" />
</colgroup>
<thead>
<tr class="header">
<th>Feature</th>
<th>Deprecated Version</th>
<th>EOL Date</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td>Spouts using named pipes</td>
<td>1.12.0</td>
<td>2021-07-05</td>
</tr>
</tbody>
</table>

## End of Life (EOL) Features

<table>
<colgroup>
<col style="width: 25%" />
<col style="width: 25%" />
<col style="width: 25%" />
<col style="width: 25%" />
</colgroup>
<thead>
<tr class="header">
<th>Feature</th>
<th>EOL Version</th>
<th>EOL Date</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td>S3V2 signatures</td>
<td>1.12.0</td>
<td>2021-01-05</td>
</tr>
<tr class="even">
<td>atom inputs</td>
<td>1.9.0</td>
<td>2019-06-12</td>
</tr>
</tbody>
</table>