# Gcloud cluster setup

In order to develop pachyderm against a gcloud-deployed cluster, follow these instructions.

## First steps

First follow the [general setup instructions](https://github.com/pachyderm/pachyderm/blob/master/doc/contributing/setup.md).

## gcloud

[Download Page](https://cloud.google.com/sdk/)

Setup Google Cloud Platform via the web

- login with your Gmail or G Suite account
  - click the silhouette in the upper right to make sure you're logged in with the right account
- get your owner/admin to setup a project for you (e.g. YOURNAME-dev)
- then they need to go into the project > settings > permissions and add you
  - hint to owner/admin: its the permissions button in one of the left hand popin menus (GKE UI can be confusing)
- you should have an email invite to accept
- click 'use google APIS' (or something along the lines of enable/manage APIs)
- click through to google compute engine API and enable it or click the 'get started' button to make it provision

Then, locally, run the following commands one at a time:

    gcloud auth login
    gcloud init

    # This should have you logged in / w gcloud
    # The following will only work after your GKE owner/admin adds you to the right project on gcloud:

    gcloud config set project YOURNAME-dev
    gcloud compute instances list

    # Now create instance using our bash helper
    create_docker_machine

    # And attach to the right docker daemon
    eval "$(docker-machine env dev)"

Setup a project on gcloud

- go to console.cloud.google.com/start
- make sure you're logged in w your gmail account
- create project 'YOURNAME-dev'

## kubectl

Now that you have gcloud, just do:

    gcloud components update kubectl
    # Now you need to start port forwarding to allow kubectl client talk to the kubernetes service on GCE

    portfowarding
    # To see this alias, look at the bash_helpers

    kubectl version
    # should report a client version, not a server version yet

    make launch-kube
    # to deploy kubernetes service

    kubectl version
    # now you should see a client and server version

    docker ps
    # you should see a few processes

## Pachyderm cluster deployment

    make launch