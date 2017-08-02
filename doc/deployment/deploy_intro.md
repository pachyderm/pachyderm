# Overview

Pachyderm runs on [Kubernetes](http://kubernetes.io/) and is backed by an object store of your choice.  As such, Pachyderm can run on any platform that supports Kubernetes and an object store. These docs cover the following commonly used deployments:

* [Google Cloud Platform](http://pachyderm.readthedocs.io/en/stable/deployment/google_cloud_platform.html)
* [Amazon Web Services](http://pachyderm.readthedocs.io/en/stable/deployment/amazon_web_services.html)
* [Azure](http://pachyderm.readthedocs.io/en/stable/deployment/azure.html)
* [OpenShift](http://pachyderm.readthedocs.io/en/stable/deployment/openshift.html)
* [On Premises](http://pachyderm.readthedocs.io/en/stable/deployment/on_premises.html)
* [Custom Object Stores](http://pachyderm.readthedocs.io/en/stable/deployment/custom_object_stores.html)
* [Migrations](http://pachyderm.readthedocs.io/en/stable/deployment/migrations.html)

## Usage Metrics

Pachyderm automatically reports anonymized usage metrics. These metrics help us
understand how people are using Pachyderm and make it better.  They can be
disabled by setting the env variable `METRICS` to `false` in the pachd
container.
