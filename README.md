# Pachyderm File System

## Create a new cluster

First generate a new etcd discovery token:

```shell
$ curl -w "\n" https://discovery.etcd.io/new
https://discovery.etcd.io/your-new-token
```

Create the config file to launch the cluster.

`pfs.yaml`:

```yaml
#cloud-config

coreos:
  etcd:
    # generate a new token for each unique cluster from https://discovery.etcd.io/new
    discovery: https://discovery.etcd.io/your-new-token
    # multi-region and multi-cloud deployments need to use $public_ipv4
    addr: $private_ipv4:4001
    peer-addr: $private_ipv4:7001
  units:
    - name: etcd.service
      command: start
    - name: fleet.service
      command: start
```


