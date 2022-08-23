# Deploy Enterprise Edition

## Retrieve Your Enterprise Token

To activate Pachyderm's enterprise features, 
you need to have your **Pachyderm Enterprise activation code** available. 
You should have received it from the Pachyderm sales team when
registering for the Enterprise Edition.

!!! Information
      - If you are a new user evaluating Pachyderm,
      you can request a [FREE trial token](https://www.pachyderm.com/trial/){target=_blank}.
      - If you are having trouble locating your activation code, contact [support@pachyderm.io](mailto:support@pachyderm.io).

## Activate The Enterprise Edition

Enabling Pachyderm's Enterprise Edition can be done in one of two flavors:

- Provide the licensing configuration as a part of the Helm deployment

    Follow [the deployment instructions](../../deploy-manage/deploy/helm-install/) for your platform, and provide your enterprise key in the `pachd.enterpriseLicenseKey: "<ENTERPRISE-LICENSE-KEY>"` field.

- Or, [activate the Enterprise Edition](#activate-pachyderm-enterprise-edition-on-an-existing-cluster) on an existing cluster as described below.

!!! Attention 
      When enterprise is enabled [through Helm, auth is automatically activated](../auth/).
       auth is automatically activated.
      Set the helm value `pachd.activateAuth` to false to prevent the bootstrap of auth on the cluster. 

### Activate Pachyderm Enterprise Edition On An Existing Cluster

To unlock Pachyderm Enterprise Features, complete the following steps:

1. Activate the Enterprise Edition by running:

      ```shell
      echo <your-activation-token> | pachctl license activate
      ```

1. Verify the status of the enterprise activation:

      ```shell
      pachctl enterprise get-state
      ```

      **System response:**
      ```
      ACTIVE
      ```

You unlocked Pachyderm's enterprise features.
