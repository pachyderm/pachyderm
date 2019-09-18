## Zookeeper for Kafka for Kubernetes

This zookeeper setup uses the Kafka image's (i.e. the Kafka distribution's) zookeeper start script. There's a configmap to detect the essential "my id" index at startup.

For background on ephemeral and persistent zookeeper config see https://github.com/Yolean/kubernetes-kafka/issues/123.
