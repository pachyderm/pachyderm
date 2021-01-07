# Create a Custom Pachyderm Deployment

Pachyderm provides the `pachctl deploy custom` command for
creating customized deployments for cloud providers or on-premises use.

This section describes how to use `pachctl deploy custom`
to create a manifest for a custom, on-premises deployment. Although
deployment automation is out of the scope of this section, Pachyderm
strongly encourages you to treat your infrastructure as code
[Deploy On-Premises](../on_premises/#infrastructure-as-code).

The topics in this section walk you through the process of using the
available flags to create the following components of your Pachyderm
infrastructure:

- A Pachyderm deployment using StatefulSets.
- An on-premises Kubernetes cluster with StatefulSets configured. It
  has the standard etcd StorageClass, along with access controls that
  limit the deployment to namespace-local roles only.
- An on-premises MinIO object store with the following parameters:

  -   SSL is enabled.
  -   Authentication requests are signed with the S3v4 signatures.
  -   The endpoint is `minio:9000`.
  -   The access key is `OBSIJRBE0PP2NO4QOA27`.
  -   The secret key is
        `tfteSlswRu7BJ86wekitnifILbZam1KYY3TG`.
  -   The S3 bucket name is `pachyderm-bucket`.

After configuring these parameters, you save the output of the
invocation to a configuration file that you can later use to deploy and
configure your environment. For the purposes of our example, all scripts
in that hypothetical infrastructure work with YAML manifests.

Complete the steps described in the following topics to deploy your
custom environment:

* [Before You Begin](deploy_custom_before_you_begin.md)
* [Pachyderm Deployment Manifest](deploy_custom_pachyderm_deployment_manifest.md)
* [Configuring Persistent Disk Parameters](deploy_custom_configuring_persistent_disk_parameters.md)
* [Configuring Object Store](deploy_custom_configuring_object_store.md)
* [Create a Complete Configuration](deploy_custom_complete_example_invocation.md)
* [Additional Flags](deploy_custom_additional_flags.md)
