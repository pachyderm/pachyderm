# Deploy Pachyderm on Amazon AWS

Pachyderm can run in a Kubernetes cluster deployed in Amazon®
Web Services (AWS), whether it is an Elastic Container
Service (EKS) or a Kubernetes cluster deployed directly on
EC2 by using a deployment tool.
AWS removes the need to maintain the underlying virtual cloud.
This advantage makes AWS a logical choice for organizations that
decide to offload the cloud infrastructure operational burden
to a third-party vendor. Pachyderm seamlessly integrates with
Amazon EKS and runs in the same fashion as on your computer.

You can install Pachyderm on Amazon AWS by using one of the following
options:

Deploy Pachyderm on Amazon EKS
:   If you already have an Amazon EKS cluster, you can quickly deploy
    Pachyderm on top of it. If you are just starting with Amazon EKS,
    this section guides you through the EKS deployment process.

Deploy Pachyderm on Amazon EC2 by using `kops`
:   Instead of EKS, you can deploy Kubernetes on AWS EC2 directly by
    using a Kubernetes deployment tool such as `kops`
    and then deploy Pachyderm on that Kubernetes cluster.
    If you deploy a cluster with `kops`, you
    remain responsible for the Kubernetes operations and maintenance.

Deploy Pachyderm with CloudFront
:   Use this option in production environments that require
    high throughput and secure data delivery.
