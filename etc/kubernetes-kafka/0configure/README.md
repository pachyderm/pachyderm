## Sample storage classes

Volume sizes are configured in the respective statefulset, but for everything else adapt the examples here to your needs. Kafka is said to be optimized for local storage, but this repository has taken a more pragmatic approach, typically using the cloud provider's volumes.

Kubernetes has lately introduced better support for [local volumes](https://kubernetes.io/docs/concepts/storage/volumes/#local) and for example [GKE](https://cloud.google.com/kubernetes-engine/docs/concepts/local-ssd) has so too, but such options are yet to be explored here.
