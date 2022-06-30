# Using a local Pachyderm Cluster

In most cases, a Pachyderm customer will receive a Console deploy as an additional pod in their Pachyderm cluster. For Console development, we need to be able to use a local frontend and backend against a real Pachyderm Cluster.

## Things you'll need
1. [Helm](https://helm.sh/docs/intro/install/)
1. [Minikube](https://minikube.sigs.k8s.io/docs/start/)
1. [Kubectl](https://kubernetes.io/docs/tasks/tools/)
1. [Pachctl](https://docs.pachyderm.com/latest/getting-started/local-installation/#install-pachctl)


## Deploy locally without the Console pod
1. Grab the latest pachyderm helm chart: `helm repo add pachyderm https://pachyderm.github.io/helmchart`
1. If you haven't already, start up minikube with `minikube start`
1. Install pachyderm locally: `helm install pachyderm --set deployTarget=LOCAL --version 2.2.3 pachyderm/pachyderm`
1. Delete your existing pachctl config file: `rm ~/.pachyderm/config.json`
1. If using Enterprise, [configure auth](#with-auth0), otherwise, your Console deploy will stay in Community Edition.
1. Run `pachctl port-forward`.

## Configuring Auth
#### With Auth0

1. Find the Auth0 console client ID and client secret in 1Password
1. Find the user login for e2e-testing@pachyderm.com in 1Password
1. Run `make setup-auth` and enter the values from 1Password when asked (they're written to .env.development.local so you shouldn't need to do this again)
1. Generate an [enterprise key](https://enterprise-token-gen.pachyderm.io/dev) and activate your license with `pachctl license activate`. For Mac OS users, `echo '<your-enterprise-token-here>' | pachctl license activate`.

#### With Github
1. [Create a Github OAuth app](https://docs.github.com/en/developers/apps/creating-an-oauth-app). For local clusters, set your callback url to `http://localhost:30658/callback`. Make sure to save the secret key, you'll need it for the next step.
1. Generate an [enterprise key](https://enterprise-token-gen.pachyderm.io/dev) and activate your license with `pachctl license activate`. For Mac OS users, `echo '<your-enterprise-token-here>' | pachctl license activate`.
1. Run `AUTH_CLIENT=github make setup-auth`. This will walk you through the setup for your local cluster.
1. (Optional) Use `pachctl auth login` to login via Github. If you opt not to do this, you will continue as the root user when creating resources.


## Deactivating Enterprise Edition
To deactivate your enterprise license and deactivate auth to return to a Community Edition console, run:
```
pachctl auth deactivate
pachctl enterprise deactivate
pachctl license delete-all
```

## Cleaning up
Any time you want to stop and restart Pachyderm, run `minikube delete`. Minikube is not meant to be a production environment
and does not handle being restarted well without a full wipe.

