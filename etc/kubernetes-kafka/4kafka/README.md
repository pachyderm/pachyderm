## Kafka manifests

This setup is designed for an image that contains an unaltered [Kafka distribution](https://kafka.apache.org/downloads). It uses a `ConfigMap` + init container instead of a custom image entrypoint script.

A caveat is that the `ConfigMap` isn't part of the `StatefulSet`s [rollout](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#rollout). To remedy this, once your're happy with your init script ([outside service names](./outside-services/) etc.) build your own kafka image that uses your script as entry point and remove the init container.

Note that brokers depend on [Zookeeper](../zookeeper/).
