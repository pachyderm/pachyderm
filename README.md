<p align="center">
	<img src='./Pachyderm_Icon-01.svg' height='225' title='Pachyderm'>
</p>

[![GitHub release](https://img.shields.io/github/release/pachyderm/pachyderm.svg?style=flat-square)](https://github.com/pachyderm/pachyderm/releases)
[![GitHub license](https://img.shields.io/badge/license-Pachyderm-blue)](https://github.com/pachyderm/pachyderm/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/pachyderm/pachyderm?status.svg)](https://pkg.go.dev/github.com/pachyderm/pachyderm/v2/src/client)
[![Go Report Card](https://goreportcard.com/badge/github.com/pachyderm/pachyderm)](https://goreportcard.com/report/github.com/pachyderm/pachyderm)
[![Slack Status](https://img.shields.io/badge/slack-pachyderm-brightgreen.svg?logo=slack)](https://www.pachyderm.com/slack)
[![CLA assistant](https://cla-assistant.io/readme/badge/pachyderm/pachyderm)](https://cla-assistant.io/pachyderm/pachyderm)

# Pachyderm – Automate data transformations with data versioning and lineage


Pachyderm is cost-effective at scale, enabling data engineering teams to automate complex pipelines with sophisticated data transformations across any type of data. Our unique approach provides parallelized processing of multi-stage, language-agnostic pipelines with data versioning and data lineage tracking. Pachyderm delivers the ultimate CI/CD engine for data. 

## Features

- Data-driven pipelines automatically trigger based on detecting data changes.
- Immutable data lineage with data versioning of any data type. 
- Autoscaling and parallel processing built on Kubernetes for resource orchestration.
- Uses standard object stores for data storage with automatic deduplication.  
- Runs across all major cloud providers and on-premises installations.


## Getting Started
To start deploying your end-to-end version-controlled data pipelines, run Pachyderm [locally](https://docs.pachyderm.com/latest/set-up/local-deploy/) or you can also [deploy on AWS/GCE/Azure](https://docs.pachyderm.com/latest/set-up/cloud-deploy) in about 5 minutes. 

You can also refer to our complete [documentation](https://docs.pachyderm.com) to see tutorials, check out example projects, and learn about advanced features of Pachyderm.

If you'd like to see some examples and learn about core use cases for Pachyderm:
- [Examples](https://github.com/pachyderm/examples)
- [Use Cases](https://www.pachyderm.com/use-cases/)
- [Case Studies](https://www.pachyderm.com/case-studies/)

## Documentation

[Official Documentation](https://docs.pachyderm.com/)

## Community
Keep up to date and get Pachyderm support via:
- [![Twitter](https://img.shields.io/twitter/follow/pachyderminc?style=social)](https://twitter.com/pachyderminc) Follow us on Twitter.
- [![Slack Status](https://badge.slack.pachyderm.io/badge.svg)](https://slack.pachyderm.io) Join our community [Slack Channel](https://www.pachyderm.com/slack) to get help from the Pachyderm team and other users.

## Contributing
To get started, sign the [Contributor License Agreement](https://cla-assistant.io/pachyderm/pachyderm).

You should also check out our [contributing guide](https://docs.pachyderm.com/latest/contributing/setup/).

Send us PRs, we would love to see what you do! You can also check our GH issues for things labeled "help-wanted" as a good place to start. We're sometimes bad about keeping that label up-to-date, so if you don't see any, just let us know.

## Usage Metrics

Pachyderm automatically reports anonymized usage metrics. These metrics help us
understand how people are using Pachyderm and make it better.  They can be
disabled by setting the env variable `METRICS` to `false` in the pachd
container.