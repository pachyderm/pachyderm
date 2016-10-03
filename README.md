<img src='doc/pachyderm_factory_gh.png' height='225' title='Pachyderm'> 

[![GitHub release](https://img.shields.io/github/release/pachyderm/pachyderm.svg?style=flat-square)](https://github.com/pachyderm/pachyderm/releases)
[![GitHub license](https://img.shields.io/github/license/pachyderm/pachyderm.svg?style=flat-square)](https://github.com/pachyderm/pachyderm/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/pachyderm/pachyderm?status.svg)](https://godoc.org/github.com/pachyderm/pachyderm/src/client)
[![Slack Status](http://slack.pachyderm.io/badge.svg)](http://slack.pachyderm.io)

# Pachyderm: A Containerized Data Lake
Pachyderm is [Git for Data Science](http://pachyderm.io/pfs.html). We offer complete version control for data and give data scientists the same first-class development tools as software developers. Pachyderm is ideal for building machine learning pipelines and ETL workflows because we version and track every model/output directly to the raw input datasets that created it (aka: [Provenance](http://pachyderm.readthedocs.io/en/latest/advanced/provenance.html)). 

Pachyderm is built on Docker and Kubernetes. Since everything in Pachyderm is a container, data scientists can use any languages or libraries they want (e.g. R, Python, OpenCV, etc) without any additional infrastructure overhead. 

## Getting Started
[Install Pachyderm locally](http://pachyderm.readthedocs.io/en/latest/getting_started/local_installation.html) or [deploy on AWS/GCE](http://pachyderm.readthedocs.io/en/latest/development/deploying_on_the_cloud.html) in about 5 minutes. 

You can also refer to our complete [developer docs](http://pachyderm.readthedocs.io/en/latest) to see tutorials, check out example projects, and learn about advanced features of Pachyderm.

If you'd like to see some examples and learn about core use cases for Pachyderm:
- [Examples](http://pachyderm.readthedocs.io/en/latest/examples/readme.html)
- [Use Cases](http://www.pachyderm.io/use_cases.html)
- [Case Studies](http://www.pachyderm.io/usecases/generalfusion.html): Learn how [General Fusion](http://www.generalfusion.com/) uses Pachyderm to power commercial fusion research.

## What is Pachyderm?

Pachyderm is a software platform that supports the storage and processing of large data sets.
Pachyderm is inspired by the Hadoop ecosystem but _shares no code_ with it.
Instead, we leverage the container ecosystem to provide the broad functionality
of Hadoop with the ease of use of Docker.

Pachyderm was designed to enable everything from "weekend data science" projects to large-scale data collaboration, just like Git does for code. 

## What's new about Pachyderm? (How is it different from Hadoop?)

There are two bold new ideas in Pachyderm:

- Containers as the core processing primitive
- Version Control for data

These ideas lead directly to a system that's much more powerful, flexible and easy to use. 

To process data, you simply create a containerized program which reads and writes to the **local filesystem**. You can use _any_ tools you want because it's all just going in a container! Pachyderm will take your container and inject data into it by way of a FUSE volume. We'll then automatically replicate your container, showing each copy a different chunk of data. With this technique, Pachyderm can scale any code you write to process up to petabytes of data (Example: [distributed grep](http://pachyderm.readthedocs.io/en/latest/getting_started/beginner_tutorial.html)).

Pachyderm also version controls all data using a commit-based distributed
filesystem (PFS), similar to what git does with code. Version control for data
has far reaching consequences in a distributed filesystem. You get the full
history of your data, can track changes and _diffs_, collaborate with teammates, and if
anything goes wrong you can revert _the entire cluster_ with one click!

Version control is also very synergistic with our containerized processing
engine. Pachyderm understands how your data changes and thus, as new data
is ingested, can run your workload on the _diff_ of the data rather than the
whole thing. This means that there's no difference between a batched job and
a streaming job, the same code will work for both!

## Community
Keep up to date and get Pachyderm support via:
- [Twitter](http://twitter.com/pachydermio)
- [![Slack Status](http://slack.pachyderm.io/badge.svg)](http://slack.pachyderm.io) Join our community [Slack Channel](http://slack.pachyderm.io)  to get help from the Pachyderm team and other users.

### Contributing

To get started, sign the [Contributor License Agreement](https://pachyderm.wufoo.com/forms/pachyderm-contributor-license-agreement).

Send us PRs, we would love to see what you do! You can also check our GH issues for things labeled "noob-friendly" as a good place to start. We're sometimes bad about keeping that label up-to-date, so if you don't see any, just let us know. 

### Join Us

WE'RE HIRING! Love Docker, Go and distributed systems? Learn more about [our team](http://www.pachyderm.io/jobs.html) and email us at jobs@pachyderm.io.

## Usage Metrics

Pachyderm automatically reports anonymized usage metrics. These metrics help us
understand how people are using Pachyderm and make it better.  They can be
disabled by setting the env variable `METRICS` to `false` in the pachd
container.
