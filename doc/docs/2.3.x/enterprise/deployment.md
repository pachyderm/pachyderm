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

- (Recommended) Provide the licensing configuration as a part of the Helm deployment

    - Provide your enterprise key in the `pachd.enterpriseLicenseKey: "<ENTERPRISE-LICENSE-KEY>"` field of your helm values (See the [Helm deployment instructions](../../deploy-manage/deploy/helm-install/) that match your targeted platform). Pachyderm will store this value in the platform secret `pachyderm-license` in the key `enterprise-license-key`.
    
    - Alternatively, you can [create a secret](../../how-tos/advanced-data-operations/secrets/#create-and-manage-secrets-in-pachyderm){target=_blank} containing your License (Key: `enterprise-license-key`) and reference its name in the field `pachd.enterpriseLicenseKeySecretName`.

    !!! Attention 
        - When enabling the enterprise features through Helm, [**auth is automatically activated**](../auth).

             In this case, a `pachyderm-auth` k8s secret is automatically created (Key: `rootToken`) containing an entry for your rootToken. Use `{{"kubectl get secret pachyderm-auth -o go-template='{{.data.rootToken | base64decode }}'"}}` to retrieve it and save it where you see fit.

        - Set the helm value `pachd.activateAuth` to false to prevent the bootstrap of auth on the cluster.

        - Note that you have the ability to provide your own rootToken value in the field `pachd.rootToken` or create your own secret holding the rootToken value in the key `rootToken`, then reference your secret name in `pachd.rootTokenSecretName`.

            Generate your own value using:

            ```shell
            openssl rand -hex 16
            f438329fc98302779be65eef226d32c1
            ```

    !!! See Also
         See the [Helm Values reference](../../reference/helm-values/#pachd){target=_blank} for details. 

- Or, [activate the Enterprise Edition](#activate-pachyderm-enterprise-edition-on-an-existing-cluster) on an existing cluster using `pachctl` as described below.

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
