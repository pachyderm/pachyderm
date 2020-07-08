# Kafka for Kubernetes

This community seeks to provide:

-   Production-worthy Kafka setup for persistent (domain- and ops-) data at
    small scale.
-   Operational knowledge, biased towards resilience over throughput, as
    Kubernetes manifest.
-   A platform for event-driven (streaming!) microservices design using
    Kubernetes.

To quote
[@arthurk](https://github.com/Yolean/kubernetes-kafka/issues/82#issuecomment-337532548):

> thanks for creating and maintaining this Kubernetes files, they're up-to-date
> (unlike the kubernetes contrib files, don't require helm and work great!

## Getting started

We suggest you `apply -f` manifests in the following order:

-   Your choice of storage classes from [./configure](./configure/)
-   [namespace](./00-namespace.yml)
-   [./rbac-namespace-default](./rbac-namespace-default/)
-   [./zookeeper](./zookeeper/)
-   [./kafka](./kafka/)

That'll give you client "bootstrap" `bootstrap.kafka.svc.cluster.local:9092`.

## Fork

Our only dependency is `kubectl`. Not because we dislike Helm or Operators, but
because we think plain manifests make it easier to collaborate. If you begin to
rely on this kafka setup we recommend you fork, for example to edit
[broker config](https://github.com/Yolean/kubernetes-kafka/blob/master/kafka/10broker-config.yml#L47).

## Version history

| tag    | k8s ≥ | highlights                                                                                                                                                                 |
| ------ | ----- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| v5.1.0 | 1.11+ | Kafka 2.1.1                                                                                                                                                                |
| v5.0.3 | 1.11+ | Zookeeper fix [#227](https://github.com/Yolean/kubernetes-kafka/pull/227) + [maxClientCnxns=1](https://github.com/Yolean/kubernetes-kafka/pull/230#issuecomment-445953857) |
| v5.0   | 1.11+ | Destabilize because in Docker we want Java 11 [#197](https://github.com/Yolean/kubernetes-kafka/pull/197) [#191](https://github.com/Yolean/kubernetes-kafka/pull/191)      |
| v4.3.1 | 1.9+  | Critical Zookeeper persistence fix [#228](https://github.com/Yolean/kubernetes-kafka/pull/228)                                                                             |
| v4.3   | 1.9+  | Adds a proper shutdown hook [#207](https://github.com/Yolean/kubernetes-kafka/pull/207)                                                                                    |
| v4.2   | 1.9+  | Kafka 1.0.2 and tools upgrade                                                                                                                                              |
|        |       | ... see [releases](https://github.com/Yolean/kubernetes-kafka/releases) for full history ...                                                                               |
| v1.0   | 1     | Stateful? In Kubernetes? In 2016? Yes.                                                                                                                                     |

## Monitoring

Have a look at:

-   [./prometheus](./prometheus/)
-   [./linkedin-burrow](./linkedin-burrow/)
-   [or plain JMX](https://github.com/Yolean/kubernetes-kafka/pull/96)
-   what's happening in the
    [monitoring](https://github.com/Yolean/kubernetes-kafka/labels/monitoring)
    label.
-   Note that this repo is intentionally light on
    [automation](https://github.com/Yolean/kubernetes-kafka/labels/automation).
    We think every SRE team must build the operational knowledge first.

## Outside (out-of-cluster) access

Available for:

-   [Brokers](./outside-services/)

## Fewer than three nodes?

For [minikube](https://github.com/kubernetes/minikube/),
[youkube](https://github.com/Yolean/youkube) etc:

-   [Scale 1](https://github.com/Yolean/kubernetes-kafka/pull/44)
-   [Scale 2](https://github.com/Yolean/kubernetes-kafka/pull/118)

## Stream...

-   [Kubernetes events to Kafka](./events-kube/)
-   [Container logs to Kafka](https://github.com/Yolean/kubernetes-kafka/pull/131)
-   [Heapster metrics to Kafka](https://github.com/Yolean/kubernetes-kafka/pull/120)
