
# Pachyderm: Data Versioning, Data Pipelines, and Data Lineage

Pachyderm is a tool for version-controlled, automated, end-to-end data pipelines for data science. If you need to chain together data scraping, ingestion, cleaning, munging, wrangling, processing,
modeling, and analysis in a sane way, while ensuring the traceability and provenance of your data, Pachyderm is for you. If you have an existing set of scripts which do this in an ad-hoc fashion and you're looking
for a way to "productionize" them, Pachyderm can make this easy for you.

## Features

- Containerized: Pachyderm is built on Docker and Kubernetes. Whatever
  languages or libraries your pipeline needs, they can run on Pachyderm which
  can easily be deployed on any cloud provider or on prem.
- Version Control: Pachyderm version controls your data as it's processed. You
  can always ask the system how data has changed, see a diff, and, if something
  doesn't look right, revert.
- Provenance (aka data lineage): Pachyderm tracks where data comes from. Pachyderm keeps track of all the code and  data that created a result.
- Parallelization: Pachyderm can efficiently schedule massively parallel
  workloads.
- Incremental Processing: Pachyderm understands how your data has changed and
  is smart enough to only process the new data.

## Getting Started

Running on Minikube: (or similar)

```
$ helm repo add pachyderm https://pachyderm.github.io/helmchart
$ helm install --set deployTarget=LOCAL pachd pachyderm/pachyderm
```

If you'd like to see some examples and learn about core use cases for Pachyderm:
- [Examples](https://docs.pachyderm.com/latest/examples/examples/)
- [Use Cases](https://www.pachyderm.com/use-cases/)
- [Case Studies](https://www.pachyderm.com/case-studies/)

## Documentation

[Official Documentation](https://docs.pachyderm.com/)

## Community
Keep up to date and get Pachyderm support via:
- [![Twitter](https://img.shields.io/twitter/follow/pachyderminc?style=social)](https://twitter.com/pachyderminc) Follow us on Twitter.
- [![Slack Status](https://badge.slack.pachyderm.io/badge.svg)](https://slack.pachyderm.io) Join our community [Slack Channel](https://slack.pachyderm.io) to get help from the Pachyderm team and other users.

## Contributing
To get started, sign the [Contributor License Agreement](https://cla-assistant.io/pachyderm/pachyderm).

You should also check out our [contributing guide](https://docs.pachyderm.com/latest/contributing/setup/).

Send us PRs, we would love to see what you do! You can also check our GH issues for things labeled "help-wanted" as a good place to start. We're sometimes bad about keeping that label up-to-date, so if you don't see any, just let us know.

## Join Us

WE'RE HIRING! Love Docker, Go and distributed systems? Learn more about [our open positions](https://boards.greenhouse.io/pachyderm)

## Usage Metrics

Pachyderm automatically reports anonymized usage metrics. These metrics help us
understand how people are using Pachyderm and make it better.  They can be
disabled by setting the env variable `METRICS` to `false` in the pachd
container.

## License Information
The Helm Chart uses the [Apache License](../LICENSE/Apache-2.0.txt). 

<!-- SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
SPDX-License-Identifier: Apache-2.0 -->