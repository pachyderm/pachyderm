# Mount a Volume in a Pipeline

You may have a local or a network-attached storage that you want your
pipeline to write files to.
You can mount that folder as a volume in Kubernetes
and make it available in your pipeline worker by using the
`pod_patch` pipeline parameter.
The `pod_patch` parameter takes a string that specifies the changes
that you want to add to your existing manifest. To create
a patch, you need to generate a diff of the original ReplicationController
and the one with your changes. You can use one of the online JSON patch
utilities, such as [JSON Patch Generator](https://extendsclass.com/json-patch.html)
to create a diff. A diff for mounting a volume might look like this:

```json
[
 {
  "op": "add",
  "path": "/volumes/-",
  "value": {
   "name": "task-pv-storage",
   "persistentVolumeClaim": {
    "claimName": "task-pv-claim"
   }
  }
 },
 {
  "op": "add",
  "path": "/containers/0/volumeMounts/-",
  "value": {
   "mountPath": "/data",
   "name": "task-pv-volume"
  }
 }
]
```

This output needs to be converted into a one-liner and added to the
pipeline spec.

We will use the [OpenCV example](../getting_started/beginner_tutorial/).
to demonstrate this functionality.

To mount a volume, complete the following steps:

1. Create a PersistentVolume and a PersistentVolumeClaim as
described in [Configure a Pod to Use a PersistentVolume for Storage](https://kubernetes.io/docs/tasks/configure-pod-container/configure-persistent-volume-storage/). Modify `mountPath` and `path` as needed.

   For testing purposes, you might want to add an `index.html`
   file as described in [Create an index.html file](https://kubernetes.io/docs/tasks/configure-pod-container/configure-persistent-volume-storage/#create-an-index-html-file-on-your-node).

1. Get the ReplicationController (RC) manifest from your pipeline:

   ```shell
   kubectl get rc <rc-pipeline> -o json > <filename>.yaml
   ```

   **Example:**

   ```shell
   kubectl get rc pipeline-edges-v7 -o json > test-rc.yaml
   ```

1. Open the generated RC manifest for editing.
1. Under `spec`, find the `volumeMounts` section.
1. Add your volume in the list of mounts. 

   **Example:**

   ```json
   {
        "mountPath": "/data",
        "name": "task-pv-storage"
   }
   ```

   `mountPath` is where your volume will be mounted inside of the
   container.

1. Find the `volumes` section.
1. Add the information about the volume.

   **Example:**

   ```json
   {
        "name": "task-pv-storage",
        "persistentVolumeClaim": {
            "claimName": "task-pv-claim"
        }
   }
   ```

   In this section, you need to specify the PersistentVolumeClaim you have
   created in Step 1.

1. Save these changes to a new file.
1. Copy the contents of the original RC to the clipboard.
1. Go to a JSON patch generator, such as [JSON Patch Generator](https://extendsclass.com/json-patch.html),
and paste the contents of the original RC manifest to the **Source JSON**
field.
1. Copy the contents of the modified RC manifest to clipboard
as described above.
1. Paste the contents of the modified RC manifest to the **Target JSON**
field.
1. Copy the generated JSON Patch.
1. Go to your terminal and open the pipeline manifest for editing.

   For example, if you are modifying the `edges` pipeline, open the
   `edges.json` file.

1. Add the patch as a one-liner under the `pod_patch` parameter.

   **Example:**

   ```json
   "pod_patch": "[{\"op\": \"add\",\"path\": \"/volumes/-\",\"value\": {\"name\": \"task-pv-storage\",\"persistentVolumeClaim\": {\"claimName\": \"task-pv-claim\"}}}, {\"op\": \"add\",\"path\": \"/containers/0/volumeMounts/-\",\"value\": {\"mountPath\": \"/data\",\"name\": \"task-pv-storage\"}}]"
   ```

   You need to add a backslash (\) before every quote (") sign
   that is enclosed in square brackets ([]). Also, you might need
   to modify the path to `volumeMounts` and `volumes` by removing
   the `/spec/template/spec/` prefix and replacing the assigned
   volume number with a dash (-). For example, if a
   path in the JSON patch is `/spec/template/spec/volumes/5`, you
   might need to replace it with `/volumes/-`. See the example
   above for details.

1. After modifying the pipeline spec, update the pipeline:

   ```shell
   pachctl update pipeline -f <pipeline-spec.yaml>
   ```

   A new pod and new replication controller should be created with
   your modified changes.

1. Verify that your file was mounted by connecting to your pod and
listing the directory that you have specified as a mountpoint. In this
example, it is `/data`.

   **Example:**

   ```shell
   kubectl exec -it <pipeline-pod> -- /bin/bash
   ```

   ```shell
   ls /data
   ```

   If you have added the `index.html` file for testing as described
   in Step 1, you should see that file in the mounted directory.

   You might want to adjust your pipeline code to read from or write to
   the mounted directory. For example, in the aforementioned
   [OpenCV example](https://docs.pachyderm.com/latest/getting_started/beginner_tutorial/#create-a-pipeline),
   the code reads from the `/pfs/images` directory and writes to the
   `/pfs/out` directory. If you want to read or write to the `/data`
   directory, you need to change those to `/data`.

   !!! important
       Pachyderm has no notion of the files stored in the mounted directory
       before it is mounted to Pachyderm. Moreover, if you have mounted a
       network share to which you write files from other than Pachyderm
       sources, Pachyderm does not guarantee the provenance of those changes.

