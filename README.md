<p align="center">
	<img src='doc/docs/master/assets/images/Pachyderm-Character_600.png' height='225' title='Pachyderm'>
</p>

[![GitHub release](https://img.shields.io/github/release/pachyderm/pachyderm.svg?style=flat-square)](https://github.com/pachyderm/pachyderm/releases)
[![GitHub license](https://img.shields.io/badge/license-Pachyderm-blue)](https://github.com/pachyderm/pachyderm/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/pachyderm/pachyderm?status.svg)](https://godoc.org/github.com/pachyderm/pachyderm/src/client)
[![Go Report Card](https://goreportcard.com/badge/github.com/pachyderm/pachyderm)](https://goreportcard.com/report/github.com/pachyderm/pachyderm)
[![Slack Status](http://slack.pachyderm.io/badge.svg)](http://slack.pachyderm.io)
[![CLA assistant](https://cla-assistant.io/readme/badge/pachyderm/pachyderm)](https://cla-assistant.io/pachyderm/pachyderm)

# Pachyderm: Data Versioning, Data Pipelines, and Data Lineage

Pachyderm is a tool for production data pipelines. If you need to chain
together data scraping, ingestion, cleaning, munging, wrangling, processing,
modeling, and analysis in a sane way, then Pachyderm is for you. If you have an
existing set of scripts which do this in an ad-hoc fashion and you're looking
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
[Install Pachyderm locally](https://docs.pachyderm.com/latest/getting_started/local_installation/) or [deploy on AWS/GCE/Azure](https://docs.pachyderm.com/latest/deploy-manage/deploy/amazon_web_services/) in about 5 minutes. 

You can also refer to our complete [documentation](https://docs.pachyderm.com) to see tutorials, check out example projects, and learn about advanced features of Pachyderm.

If you'd like to see some examples and learn about core use cases for Pachyderm:
- [Examples](https://docs.pachyderm.com/latest/examples/examples/)
- [Use Cases](https://www.pachyderm.com/use-cases/)
- [Case Studies](https://www.pachyderm.com/case-studies/)

## Documentation

[Official Documentation](https://docs.pachyderm.com/)

## Community
Keep up to date and get Pachyderm support via:
- [![Twitter](https://img.shields.io/twitter/follow/pachyderminc?style=social)](https://twitter.com/pachyderminc) Follow us on Twitter.
- [![Slack Status](http://slack.pachyderm.io/badge.svg)](http://slack.pachyderm.io) Join our community [Slack Channel](http://slack.pachyderm.io) to get help from the Pachyderm team and other users.

## Contributing

To get started, sign the [Contributor License Agreement](https://cla-assistant.io/pachyderm/pachyderm).

You should also check out our [contributing guide](https://docs.pachyderm.com/latest/contributing/setup/).

Send us PRs, we would love to see what you do! You can also check our GH issues for things labeled "help-wanted" as a good place to start. We're sometimes bad about keeping that label up-to-date, so if you don't see any, just let us know.

## Join Us

WE'RE HIRING! Love Docker, Go and distributed systems? Learn more about [our open positions](https://jobs.lever.co/pachyderm/)

## Usage Metrics

Pachyderm automatically reports anonymized usage metrics. These metrics help us
understand how people are using Pachyderm and make it better.  They can be
disabled by setting the env variable `METRICS` to `false` in the pachd
container.

## License Information
Pachyderm has moved some components of Pachyderm Platform to a [source-available limited license](LICENSE). 

We remain committed to the culture of open source, developing our product transparently and collaboratively with our community, and giving our community and customers source code access and the ability to study and change the software to suit their needs.

Under the Pachyderm Community Edition, you can access the source code and modify or redistribute it; there is only one thing you cannot do, and that is use it to make a competing offering. 

Check out our [License FAQ Page](https://pachyderm.com/about/pachyderm-community-license-faq/) for more information.
