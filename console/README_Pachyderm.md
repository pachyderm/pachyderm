# Using a local Pachyderm Cluster

In most cases, a Pachyderm customer will receive a Console deploy as an additional pod in their Pachyderm cluster. For Console development, we need to be able to use a local frontend and backend against a real Pachyderm Cluster.

## Getting started

1. Install these things
   1. [Helm](https://helm.sh/docs/intro/install/)
   1. One of the following local k8s cluster runners:
      - [Kind](https://kind.sigs.k8s.io/)
      - [Minikube](https://minikube.sigs.k8s.io/docs/start/)
      - [Docker Desktop](https://www.docker.com/products/docker-desktop/)
   1. [Kubectl](https://kubernetes.io/docs/tasks/tools/)
   1. [Pachctl](https://docs.pachyderm.com/latest/getting-started/local-installation/#install-pachctl)
1. Grab the latest pachyderm helm chart:

   ```bash
   helm repo add pachyderm https://pachyderm.github.io/helmchart
   ```

## Enterprise Key

1. Generate an [enterprise key](https://enterprise-token-gen.pachyderm.io/dev).

1. Add PACHYDERM_ENTERPRISE_KEY to your environment variables. Optionally add it this to your terminal's rc file so you only need to run it once. Run:

   ```bash
   export PACHYDERM_ENTERPRISE_KEY='yourKey'
   ```

## Starting your Cluster

### Kind

Kind does not expose the proxy ports by default. Therefore you need to start your cluster with the following config:

```bash
$ cat <<EOF | kind create cluster --name=kind --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
    - role: control-plane
      kubeadmConfigPatches:
          - |
              kind: InitConfiguration
              nodeRegistration:
                  kubeletExtraArgs:
                      node-labels: "ingress-ready=true"
      extraPortMappings:
          - containerPort: 30080
            hostPort: 80
            protocol: TCP
EOF
```

You can then use the following to set your kubectl context to kind

```bash
kubectl config set current-context kind-kind
```

### Minikube

[https://www.notion.so/Console-Local-Dev-9e3b1dffb36a44d0af04ab17fcdc0c75](https://www.notion.so/Console-Local-Dev-9e3b1dffb36a44d0af04ab17fcdc0c75)

### Docker Desktop

You don't need to do anything special here. Just enable k8s.

## Deploying pach

1. Run Pachyderm using the values from `enterpriseProxyHelmValues.yaml`. Additionally, we will be passing the enterprise key through helm. This will activate _both_ enterprise and RBAC authorization. If you want to deploy into community edition, just remove the set licenseKey line.

   - If you are using **Docker Desktop**:

     ```bash
     # cd to root of this project
     $ helm install \
             --wait --timeout 10m pachd pachyderm/pachyderm \
             --version=2.6.0 \
             -f enterpriseProxyHelmValues.yaml \
             --set pachd.metrics.enabled=false \
             --set pachd.enterpriseLicenseKey=$PACHYDERM_ENTERPRISE_KEY
     ```

   - If you are using **Kind**:

     ```bash
     # cd to root of this project
     $ helm install \
             --wait --timeout 10m pachd pachyderm/pachyderm \
             --version=2.6.0 \
             -f enterpriseProxyHelmValues.yaml \
             --set pachd.metrics.enabled=false \
             -f kind.yaml \
             --set pachd.enterpriseLicenseKey=$PACHYDERM_ENTERPRISE_KEY
     ```

## Deploy Console Dev

1. Run the frontend and backend server with:

   ```bash
   make launch-dev | make bunyan
   ```

1. Open [localhost:4000](http://localhost:4000)

## Login to Mock IDP

1. Once Console is running, you should be redirected to the Login page from the Mock IDP. Login with user `admin` and password `password`

1. If you are experiencing any errors with logging in, ensure your `.env.development.local` file is not injecting incorrect values.

You are good to dev now!

---

# Extra info

## Deactivating Enterprise Edition

- To deactivate your enterprise license and deactivate auth to return to a Community Edition console, run:

  ```bash
  pachctl auth deactivate
  pachctl enterprise deactivate
  pachctl license delete-all
  ```

- Alternatively you can use use `pachctl delete all` if you don't mind losing all of your repos and pipelines.

## Cleaning up

Any time you want to stop and restart Pachyderm, run `minikube delete`. Minikube is not meant to be a production environment
and does not handle being restarted well without a full wipe.
