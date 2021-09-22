# Deploy Enterprise Edition

To deploy and use Pachyderm's Enterprise Edition, follow
[the deployment instructions](../deploy-manage/deploy/) for your platform,
then [activate the Enterprise Edition](#activate-pachyderm-enterprise-edition).

## Activate Pachyderm Enterprise Edition

To activate Pachyderm's enterprise features, 
you need to have your **Pachyderm Enterprise activation code** available. 
You should have received this from the Pachyderm sales team when
registering for the Enterprise Edition.

!!! Information
      - If you are a new user evaluating Pachyderm,
      you can request a [FREE evaluation code](https://www.pachyderm.com/trial).
      - If you are having trouble locating your activation code, contact [support@pachyderm.io](mailto:support@pachyderm.io).


To unlock Pachyderm Enterprise Features, complete the following steps:

1. Activate the Enterprise Edition by running:

      ```shell
      $ echo <your-activation-token> | pachctl license activate
      ```

1. Verify the status of the enterprise activation:

      ```shell
      $ pachctl enterprise get-state
      ```

      **System response:**
      ```
      ACTIVE
      ```

