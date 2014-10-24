# Pachyderm File System

## Quick Start

### Creating a CoreOS cluster

Pfs is designed to run on CoreOS. To start you'll need a working CoreOS
cluster.

Google Compute Engine (recommended): [https://coreos.com/docs/running-coreos/cloud-providers/google-compute-engine/]
Amazon EC2: [https://coreos.com/docs/running-coreos/cloud-providers/ec2/]

### Deploy pfs
`$ curl https://raw.githubusercontent.com/pachyderm-io/pfs/master/deploy/static/3Node.tar.gz | tar -xvf`

`$ fleetctl start 3Node/*`
