# Upgrade Pachyderm

Upgrades between minor releases or patch releases, such as `2.1.0` to version `2.2.0`,
should be seamless.
Therefore, the upgrade procedure is simple and requires little to no downtime.
As a good practice, we recommend that you check the [release notes](https://github.com/pachyderm/pachyderm/blob/master/CHANGELOG.md){target=_blank} before an upgrade to get an understanding of the changes introduced between your current version and your target. 

!!! Warning 
       Do not use these steps to upgrade between major versions as it might result in data corruption.

Complete the following steps to upgrade Pachyderm from one minor release to another.
## 1- Backup Your Cluster

As a general good practice, start with the backup of your cluster as described in the [Backup and Restore](../backup-restore/)
section of this documentation.

## 2- Update Your Helm Values

This phase depends on whether you need to modify your existing configuration (for example, enter an enterprise key, plug an identity provider, reference an enterprise server, etc...).

In the case of a simple upgrade of version on a cluster, and provided that you do not need to change any additional configuration, no change in the values.yaml should be required. The new version of Pachyderm will be directly set in the `helm upgrade` command.

## 3- Upgrade `pachctl` Version

!!!Warning 
       Migrating from a **pachd** version older than 2.3.0? Do not skip upgrading **pachctl**.

 
 - To update to the latest version of Pachyderm, run the steps below depending on your operating system:
  
      * For macOS, run:  
  
      ```shell  
      brew tap pachyderm/tap && brew install pachyderm/tap/pachctl@{{ config.pach_major_minor_version }}  
      ```  
  
      * For a Debian-based Linux 64-bit or Windows 10 or later running on  
      WSL:  
  
      ```shell  
      curl -o /tmp/pachctl.deb -L https://github.com/pachyderm/pachyderm/releases/download/v{{ config.pach_latest_version }}/pachctl_{{ config.pach_latest_version }}_amd64.deb && sudo dpkg -i /tmp/pachctl.deb  
      ```  
  
      * For all other Linux flavors:  
  
      ```shell  
      curl -o /tmp/pachctl.tar.gz -L https://github.com/pachyderm/pachyderm/releases/download/v{{ config.pach_latest_version }}/pachctl_{{ config.pach_latest_version }}_linux_amd64.tar.gz && tar -xvf /tmp/pachctl.tar.gz -C /tmp && sudo cp /tmp/pachctl_{{ config.pach_latest_version }}_linux_amd64/pachctl /usr/local/bin  
      ```  

!!! Note
      For a specific target release, specify the targeted major/minor version of `pachctl` for brew and major/minor/patch release for curl in the commands above.


 - Verify that the installation was successful by running `pachctl version --client-only`:  
  
      ```shell  
      pachctl version --client-only  
      ```  
  
      **System Response:**  
  
      ```shell  
      COMPONENT           VERSION  
      pachctl             <This should display the version you installed>  
      ```  

## 4- Helm Upgrade

- Redeploy Pachyderm by running the [helm upgrade](https://helm.sh/docs/helm/helm_upgrade/){target=_blank} command with your updated values.yaml:

      ```shell
      helm repo add pach https://helm.pachyderm.com
      helm repo update
      helm upgrade pachd -f my_pachyderm_values.yaml pach/pachyderm --version <your_chart_version>
      ```

!!! Note 
      Each chart version is associated with a given version of Pachyderm. You will find the list of all available chart versions and their associated version of Pachyderm on [Artifacthub](https://artifacthub.io/packages/helm/pachyderm/pachyderm){target=_blank}.

- The upgrade can take some time. You can run `kubectl get pods` periodically in a separate tab to check the status of the deployment. When Pachyderm is deployed, the command shows all pods as `READY`:

      ```shell
      kubectl get pods
      ```
      Once the pods are up, you should see a pod for `pachd` running 
      (alongside etcd, pg-bouncer, postgres, console etc... depending on your installation). 

      **System response:**

      ```shell
      NAME                     READY     STATUS    RESTARTS   AGE
      pachd-3677268306-9sqm0   1/1       Running   0          4m
      ...
      ```

- Verify that the new version has been deployed:

      ```shell
      pachctl version
      ```

      **System response:**

      ```shell
      COMPONENT           VERSION
      pachctl             {{ config.pach_latest_version }}
      pachd               {{ config.pach_latest_version }}
      ```

      The `pachd` and `pachctl` versions must both match the new version.

