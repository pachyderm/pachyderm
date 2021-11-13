# Create and Manage Secrets in Pachyderm

Pachyderm uses Kubernetes' *Secrets* to store and manage sensitive data, such as passwords, OAuth tokens, or ssh keys. You can use any of [Kubernetes' types of Secrets](https://kubernetes.io/docs/concepts/configuration/secret/#secret-types) that match your use case. 
Namely, `generic` (or Opaque), `tls`, or `docker-registry`.

!!! Warning
      As of today, Pachyderm *only supports the JSON format for Kubernetes' Secrets files*.

To use a Secret in Pachyderm, you need to:

1. Create it.
1. Reference it in your pipeline's specification file.

## Create a Secret
The creation of a Secret in Pachyderm *requires a JSON configuration file*.

A good way to create this file is:

1. To generate it by calling a dry-run of the `kubectl create secret ... --dry-run=client  --output=json > myfirstsecret.json` command.
1. Then call `pachctl create secret -f myfirstsecret.json`.

!!! Info "Reminder"
      Kubernetes Secrets are, by default, stored as *unencrypted base64-encoded* strings (i.e., the values for all keys in the data field have to be base64-encoded strings). When using the `kubectl create secret` command, the encoding is done for you. If you choose to manually create your JSON file, make sure to use your own base 64 encoder.

### Generate your secret configuration file
Let's first generate your secret configuration file using the `kubectl` command. For example:

- for a generic authentication secret:
   ```shell
   $ kubectl create secret generic mysecretname --from-literal=username=<myusername> --from-literal=password=<mypassword> --dry-run=client  --output=json > myfirstsecret.json
   ```
- for a tls secret:
   ```shell
   $ kubectl create secret tls mysecretname --cert=<Path to your certificate> --key=<Path to your SSH key> --dry-run=client  --output=json > myfirstsecret.json 
   ```
- for a docker registry secret:
   ```shell
   kubectl create secret docker-registry mysecretname --dry-run=client --docker-server=<DOCKER_REGISTRY_SERVER> --docker-username=<DOCKER_USER> --docker-password=<DOCKER_PASSWORD> --output=json > myfirstsecret.json
   ```

!!! Example "Example of a generic secret"
      ```json
      {
         "apiVersion": "v1",
         "kind": "Secret",
         "metadata": {
            "name": "clearml"
         },
         "type": "Opaque",
         "stringData": {
            "access": "<CLEARML_API_ACCESS_KEY>",
            "secret": "<CLEARML_API_SECRET_KEY>"
         }
      }
      ```
Find more detailed information on the [creation of Secrets](https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-kubectl/) in Kubernetes documentation.

### Create your Secret in Pachyderm
Next, run the following to actually create the secret in the Pachyderm Kubernetes cluster:
```shell
$ pachctl create secret -f myfirstsecret.json 
```

You can run `pachctl list secret` to verify that your secret has been properly created.
You should see an output that looks like the following:

```
NAME     TYPE                           CREATED        
mysecret kubernetes.io/dockerconfigjson 11 seconds ago 
```
!!! Note
    Use `pachctl delete secret` to delete a secret given its name,  `pachctl inspect secret` to list a secret given its name.
You can now edit your pipeline specification file as follow.


## Reference a Secret in Pachyderm's specification file
Now that your secret is created on Pachyderm cluster, you will need to notify your pipeline by updating your pipeline [specification file](https://docs.pachyderm.com/latest/reference/pipeline_spec/#manifest-format).
In Pachyderm, a Secret can be used in three different ways:

1. **As a container environment variable**:

      In this case, in Pachyderm's pipeline specification file, you need to reference Kubernetes' Secret by its:

      - **`name`**
      - and specify an environment variable named **`env_var`** that the value of your  **`key`** should be bound to. 

      This makes for easy access to your Secret's data in your pipeline's code. 
      For example, this is useful for passing the password to a third-party system to your pipeline's code.

      ```json
      "transform": {
         "image": string,
         "cmd": [ string ],
         ...
         "secrets": [ {
            "name": string,
            "env_var": string,
            "key": string
         }]
      }
      ```
    !!! Example "Example of a pipeline specification file assigning a Secret's values to environment variables"  
         Look at the pipeline specification in this example and see how we used the  `"env_var"` to pass CLEARML API credentials to the pipeline code.
         ```json
         {
            "pipeline": {
               "name": "mnist"
            },
            "description": "MNIST example logging to ClearML",
            "input": {
               "pfs": {
                  "repo": "data",
                  "branch": "master",
                  "glob": "/*"
               }
            },
            "transform": {
               "cmd": [
                  "/bin/sh"
               ],
               "stdin": [
                  "python pytorch_mnist.py --lr 0.2 --save-location /pfs/out"
               ],
               "image": "pachyderm/clearml_mnist:dev0.11",
               "secrets": [
                  {
                  "name": "clearml",
                  "env_var": "CLEARML_API_ACCESS_KEY",
                  "key": "access"
                  },
                  {
                  "name": "clearml",
                  "env_var": "CLEARML_API_SECRET_KEY",
                  "key": "secret"
                  }
               ]
            }
         }
         ```


1. **As a file in a volume mounted on a container**:

      In this case, in Pachyderm's pipeline specification file, you need to reference Kubernetes' Secret by its:

      -  **`name`**
      - and specify the mount point (**`mount_path`**) to the secret (ex: `"/var/my-app-secret"`).

      Pachyderm mounts all of the keys in the secret with file names corresponding to the keys.
      This is useful for secure configuration files.

      ```json
      "transform": {
         "image": string,
         "cmd": [ string ],
         ...
         "secrets": [ {
            "name": string,
            "mount_path": string
         }]
      }
      ```

1. **When pulling images**:

      [Image pull Secrets](https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod) are a different kind of secret used to store access credentials to your private image registry. 
      
      You reference Image Pull Secrets (or Docker Registry Secrets) by setting the **`image_pull_secrets`** field of your pipeline specification file to the secret's name you created (ex: `"mysecretname"`).

      ```JSON
      "transform": {
         "image": string,
         "cmd": [ string ],
         ...
         "image_pull_secrets": [ string ]
      }
      ```

