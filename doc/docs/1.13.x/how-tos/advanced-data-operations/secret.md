# Create and Manage Secrets in Pachyderm

Pachyderm uses Kubernetes' *Secrets* to store and manage sensitive data, such as passwords, OAuth tokens or ssh keys. 

!!! Warning
   As of today, Pachyderm *only supports the JSON format for Kubernetes' Secrets*.

Pachyderm let's you use any of [Kubernetes' various types of Secrets](https://kubernetes.io/docs/concepts/configuration/secret/#secret-types) that match your use case. 
Namely, `generic` (or Opaque), `tls`, or `docker-registry`.


## Create a Secret
Pachyderm Secret creation *uses a JSON configuration file*.
A good way to create this file is to generate it by calling a dry-run of the `kubectl create secret ...` command then call `pachctl create secret -f <yourgeneratedsecretfile.json>`.

!!! Info "Reminder"
   Kubernetes Secrets are, by default, stored as *unencrypted base64-encoded* strings (i.e., the values for all keys in the data field have to be base64-encoded strings). When using the `kubectl create secret` command, the encoding is done for you. If you choose to manually create your JSON file, make sure to use your own base 64 encoder.

For example, 
- for a simple authentication secret:
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

Next, run the following to actually create the secret in the Pachyderm Kubernetes cluster:
```shell
$ pachctl create secret -f myfirstsecret.json 
```
You can now edit your pipeline specification file as follow.

## Reference a Secret in Pachyderm's specification file
In Pachyderm, a Secret can be used in three different ways:

1. As a container environment variable.
   In this case, you need to reference Kubernetes's Secret by its *name* and specify *an environment variable name `env_var` that the value of your  `key` should be bound to* in Pachyderm's pipeline specification file. This makes for an easy access to your Secret's data in your pipeline's code. 
   For example, this is useful for passing the password to a third party system to your pipeline's code.
   
   Look at this pipeline specification in our [RabbitMQ Spout example](https://github.com/pachyderm/pachyderm/blob/master/examples/spouts/go-rabbitmq-spout/pipelines/spout.pipeline.json) and see how we used the  `"env_var": "RABBITMQ_PASSWORD"` to access a RabbitMQ instance from inside our pipeline.


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

1. As a file in a volume mounted on a container.
   In this case, you need to reference Kubernetes's Secret by its *`name`* and specify the *path (`mount_path`) to the secret* in Pachyderm's pipeline specification file. Pachyderm mounts all of the keys in the secret with file names corresponding to the keys.
   For example, this would be useful for secure configuration files.

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

1. When pulling images.
   [Image pull Secrets](https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod) are a different kind of secret used to store access credentials to your private image registry. You reference Image Pull Secrets by listing their *name* in the `image_pull_secrets` field of your pipeline specification file.

  ```json
    "transform": {
      "image": string,
      "cmd": [ string ],
      ...
      "image_pull_secrets": [ string ]
   }
   ```
For more general information on Secrets, read [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/secret/).





    transform.secrets is an array of secrets. You can use the secrets to embed sensitive data, such as credentials. The secrets reference Kubernetes secrets by name and specify a path to map the secrets or an environment variable (env_var) that the value should be bound to. Secrets must set name which should be the name of a secret in Kubernetes. Secrets must also specify either mount_path or env_var and key. See more information about Kubernetes secrets here.

transform.image_pull_secrets is an array of image pull secrets, image pull secrets are similar to secrets except that they are mounted before the containers are created so they can be used to provide credentials for image pulling. For example, if you are using a private Docker registry for your images, you can specify it by running the following command:


kubectl create secret docker-registry myregistrykey --docker-server=DOCKER_REGISTRY_SERVER --docker-username=DOCKER_USER --docker-password=DOCKER_PASSWORD --docker-email=DOCKER_EMAIL
And then, notify your pipeline about it by using "image_pull_secrets": [ "myregistrykey" ]. Read more about image pull secrets here
