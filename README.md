# Pachyderm
[![GitHub release](https://img.shields.io/github/release/pachyderm/pachyderm.svg?style=flat-square)](https://github.com/pachyderm/pachyderm/releases)
[![GitHub license](https://img.shields.io/github/license/pachyderm/pachyderm.svg?style=flat-square)](https://github.com/pachyderm/pachyderm/blob/master/LICENSE)

* [News](#news)
* [What is Pachyderm?](#what-is-pachyderm)
* [What's new about Pachyderm? (How is it different from Hadoop?)](#-new-about-pachyderm-how-is-it-different-from-hadoop)
* [Our Vision](#our-vision)
* [Contributing](#contributing)

### News

WE'RE HIRING! Love Docker, Go and distributed systems? Learn more about [our team](http://www.pachyderm.io/jobs.html) and email us at jobs@pachyderm.io.

### Getting Started

Already got a kube cluster:

```sh
$ kubectl create -f http://pachyderm.io/manifest.json
```

Get up and running with the [fruit stand example](https://github.com/pachyderm/pachyderm/blob/master/examples/fruit_stand/GUIDE.md).

### What is Pachyderm?

Pachyderm is a Data Lake -- a place to dump and process gigantic data sets.
Pachyderm is inspired by the Hadoop ecosystem but _shares no code_ with it.
Instead, we leverage the container ecosystem to provide the broad functionality
of Hadoop with the ease of use of Docker.

Pachyderm offers the following core functionality:

- Virtually limitless storage for any data.
- Virtually limitless processing power using any tools.
- Tracking of data history, provenance and ownership. (Version Control for data).
- Automatic processing on new data as it’s ingested. (Streaming).
- Chaining processes together. (Pipelining)

### What's new about Pachyderm? (How is it different from Hadoop?)

There are two bold new ideas in Pachyderm:

- Containers as the core processing primitive
- Version Control for data

These ideas lead directly to a system that's much more powerful, flexible and easy to use. 

To process data, you simply create a containerized program which reads and writes to the local filesystem. You can use _any_ tools you want because it's all just going in a container! Pachyderm will take your container and inject data into it by way of a FUSE volume. We'll then automatically replicate your container, showing each copy a different chunk of data. With this technique, Pachyderm can scale any code you write to process up to petabytes of data (Example: [distributed grep](https://github.com/pachyderm/pachyderm/examples/fruit_stand/GUIDE.md)).

Pachyderm also version controls all data using a commit-based distributed
filesystem (PFS), similar to what git does with code. Version control for data
has far reaching consequences in a distributed filesystem. You get the full
history of your data, it's much easier to collaborate with teammates, and if
anything goes wrong you can revert _the entire cluster_ with one click!

Version control is also very synergistic with our containerized processing
engine. Pachyderm understands how your data changes and thus, as new data
is ingested, can run your workload on the _diff_ of the data rather than the
whole thing. This means that there's no difference between a batched job and
a streaming job, the same code will work for both!

### Our Vision

Containers are a revolutionary new technology with a compelling application to
big data. Our goal is to fully realize that use case. Hadoop has spawned a
sprawling ecosystem of tools but with each new tool the complexity of your
cluster grows until maintaining it becomes a full-time job. Containers are the
_perfect_ antidote to this problem. What if adding a new tool to your data
infrastructure was as easy as installing an app? Thanks to the magic of
containers in Pachyderm, it really is!

The most exciting thing about this vision though is what comes next. Pachyderm
can do big data with _anything_ that runs on Linux. And anything you build can be
easily shared with the rest of the community, afterall it's just a
container so it's completely reusable and will run the same every time. We have some ideas of our own about what the best starting building blocks will be, but it's just the
tip of the iceburg -- we expect our users will have many more interesting ideas.
We can't wait to see what they are!

### Contributing

[Deploying Pachyderm](https://github.com/pachyderm/pachyderm/blob/master/examples/grep/GUIDE.md#setup).

To get started, sign the [Contributor License Agreement](https://pachyderm.wufoo.com/forms/pachyderm-contributor-license-agreement).

Send us PRs, we would love to see what you do!

