# Deploy Enterprise Edition

To deploy and use Pachyderm's Enterprise Edition, follow
[the deployment instructions](../../deploy-manage/deploy/) for your platform
and then [activate the Enterprise Edition](#activate-pachyderm-enterprise-edition).


!!! note
    Pachyderm automatically deploys the Enterprise Console. If you want
    to deploy without the dashboard, run
    `pachctl deploy [command] --no-dashboard`.

## Activate Pachyderm Enterprise Edition

To activate Pachyderm's enterprise features, you need to have your Pachyderm Enterprise activation code
available. You should have received this from the Pachyderm sales team when
registering for the Enterprise Edition. If you are a new user evaluating Pachyderm,
you can request a FREE evaluation code on the landing page of the dashboard.
If you are having trouble locating your activation code, contact [support@pachyderm.io](mailto:support@pachyderm.io).

### Activate by Using the `pachctl` CLI

To activate the Pachyderm Enterprise Edition, complete the following steps:

1. Activate the Enterprise Edition by running:

   ```shell
   pachctl license activate <activation-code>
   ```

   If this command does not return any error, then the activation was
   successful.

1. Verify the status of the enterprise activation:

   ```shell
   pachctl enterprise get-state
   ```

   **System response:**

   ```shell
   ACTIVE
   ```

