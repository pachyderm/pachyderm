# Before You Begin

Before you start creating a custom deployment, verify that you have
completed the following steps:

1. Read and complete the steps described in the [Introduction](../../on_premises/#introduction)
in the *On-Premises* section. This section explains the differences between
static persistent volumes, StatefulSets, and StatefulSets with StorageClasses.
Also, it explains the meanings of the variables, such as  `PVC_STORAGE_SIZE`
and `OS_ENDPOINT` that are used in the examples below.
1. Install [kubectl](https://kubernetes.io/docs/user-guide/prereqs/).
1. Install [pachctl](../../../../../getting_started/local_installation/#install-pachctl).
1. Proceed to [Pachyderm Deployment Manifest](./deploy_custom_pachyderm_deployment_manifest.md).
