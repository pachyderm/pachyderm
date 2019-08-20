# Create a Complete Configuration

The following is a complete deploy command example of a
custom deployment. The command generates the manifest
and saves it as a YAML configuration file.
Also, the command includes the `local-roles` flag
to scope the deployment to the `pachyderm` service
account access permissions.

Run the following command to deploy your example cluster:

```
pachctl deploy custom --persistent-disk aws --object-store s3 \
    foobar 10 \
    pachyderm-bucket  'OBSIJRBE0PP2NO4QOA27' 'tfteSlswRu7BJ86wekitnifILbZam1KYY3TG' 'minio:9000' \
    --dynamic-etcd-nodes 10
    --local-roles --output yaml  --dry-run > custom_deploy.yaml
```

For more information about the contents of the `custom_deploy.yaml` file,
see [Pachyderm Deployment Manifest](deploy_custom_pachyderm_deployment_manifest.html).

## Deploy Your Cluster

You can either deploy manifests that you have created above
or edit them to customize them further, before deploying.

If you decide to edit your manifest, you must consult with an
experienced Kubernetes administrator.
If you are attempting a highly customized deployment,
use one of the Pachyderm support resources listed below.

To deploy your configuration, run one of the following commands:

* If you are deploying with a static persistent volume, run:

  ```bash
  $ kubectl apply -f ./pachyderm-with-static-volume.json
  ```
* If you are deploying with `StatefulSets`, run:

  ```bash
  $ kubectl apply -f ./pachyderm-with-statefulset.json
  ```

* If you are deploying with `StatefulSets` by using `StorageClasses`:

  ```bash
  $ kubectl apply -f ./pachyderm-with-statefulset-using-storageclasses.json
  ```
