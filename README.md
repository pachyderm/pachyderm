# Trino

This branch was created by msteffen for the purpose of demoing an integration between Trino and Pachyderm at the Feb 2023 Miami offsite.

See the directory [_trino_stuff](tree/msteffen/20230207/master/trino/_trino_stuff) for the files I authored to do this demo

## Why Trino

- "Data Science" currently describes two currently-mostly-disjoint groups: people training deep learning models using the latest AI techniques, and people querying line-of-business datasets with SQL
  - The number of people doing the latter is much larger than the number of people doing the former. The reason Databricks is rich is because there are a lot of people doing the latter with SparkSQL.
- Pachyderm is currently very useful to the former group. It could be very useful to the latter, and, more importantly, could bridge the two (e.g. by allowing people to plug a sentiment analysis model into their existing data and trivially add a "user sentiment" column—called a "feature" in the lingo—to one of their tables).
- This branch contains a small demo that begins to explore this possibility.
- Trino is a distributed SQL engine. You give it an OLAP (i.e. read-only) SQL query, it turns that SQL query into a bunch of low-level tasks to e.g. read parquet files from S3, join it with the result of a postgres query, etc. and stream the result back to you. It's sort of like Google's Dremel, but while Dremel could only do OLAP queries on large datasets in a specific form on Google's infra, Trina can do OLAP queries on a wide variety of data structures in a wide variety of storage (S3, Postgres, etc).
- This examples deploys a small trino cluster, connects it to Pachyderm via Pachyderm's HTTP API (S3 gateway was hard for some reason that I no longer remember), and queries some data living in Pachyderm. It is slow and inefficient and mostly silly, but it does work.
- Longer-term, it would make sense to write a Trino connector for Pachyderm. This would allow users to query arbitrary versioned data in Pachyderm, hopefully efficiently.

## How

- Per above, see the `_trino_stuff` directory
- The `trino-server-406/etc` directory has most of the Trino config files that make the demo work
  - `etc/catalog/pach.properties` is a big one. It tells Trino that there's a dataset in Pachyderm, and to use Trino's `example-http` connector to talk to it, and passes the address of the dataset metadata file (as a Pachyderm HTTP path) to the `example-http` connector.
- `config.properties` just configures the single-node Trino cluster that the demo deploys. It tells the node to act as a coordinator (compiles SQL queries into low-level tasks, assigns the low-level tasks, and coordinates their execution), a discovery service endpoint, and worker (by not setting `node-scheduler.include-coordinator`)
- `jvm.config` just configures the JVM that runs the Trino node
- `log.properties` sets the trino node's logging behavior. Not important
- `node.properties` I don't remember what this does

- See `trino-example-data/generate-and-load.sh` for info on placing all data correctly for the demo

---

<p align="center">
	<img src='./Pachyderm_Icon-01.svg' height='225' title='Pachyderm'>
</p>

[![GitHub release](https://img.shields.io/github/release/pachyderm/pachyderm.svg?style=flat-square)](https://github.com/pachyderm/pachyderm/releases)
[![GitHub license](https://img.shields.io/badge/license-Pachyderm-blue)](https://github.com/pachyderm/pachyderm/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/pachyderm/pachyderm?status.svg)](https://pkg.go.dev/github.com/pachyderm/pachyderm/v2/src/client)
[![Go Report Card](https://goreportcard.com/badge/github.com/pachyderm/pachyderm)](https://goreportcard.com/report/github.com/pachyderm/pachyderm)
[![Slack Status](https://badge.slack.pachyderm.io/badge.svg)](https://slack.pachyderm.io)
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
To start deploying your end-to-end version-controlled data pipelines, run Pachyderm [locally](https://docs.pachyderm.com/latest/getting-started/local-installation/) or you can also [deploy on AWS/GCE/Azure](https://docs.pachyderm.com/latest/deploy-manage/deploy/amazon_web_services/) in about 5 minutes. 

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
Pachyderm has moved some components of Pachyderm Platform to a [source-available limited license](LICENSE). 

We remain committed to the culture of open source, developing our product transparently and collaboratively with our community, and giving our community and customers source code access and the ability to study and change the software to suit their needs.

Under the Pachyderm Community License, you can access the source code and modify or redistribute it; there is only one thing you cannot do, and that is use it to make a competing offering. 

Check out our [License FAQ Page](https://www.pachyderm.com/community-license-faq/) for more information.
