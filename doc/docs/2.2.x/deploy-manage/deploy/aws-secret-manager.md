# Set Up AWS Secret Manager

For production environments, we highly recommend securing and centralizing the storage and management of your secrets (database access, root token, enterprise key, etc...) in **AWS Secrets Manager**, then allow your EKS cluster to retrieve those secrets using fine-grained IAM policies.

This section will walk you through the steps to enable your EKS cluster to retrieve secrets from AWS Secrets Manager. 


## 1. Prerequisites

Note that the following steps start right after installing your EKS cluster. 
For informations on how to set your cluster up in production, refer to the [deploy Kubernetes](../aws-deploy-pachyderm/#2-deploy-kubernetes-by-using-eksctl){target=_blank} section of our deployment instructions on AWS.


## 2. Install The AWS Secrets and Configuration Provider (ASCP)

To retrieve your secrets through your workloads running on your cluster, you will first need to install:

- A Secrets Store CSI driver
- AWS Secrets Manager and Config Provider

!!! Warning
      The ASCP works with Amazon Elastic Kubernetes Service (Amazon EKS) 1.17+.

### Install the Secrets Store CSI Driver
Deploy the [**Secrets Store CSI driver**](https://secrets-store-csi-driver.sigs.k8s.io/getting-started/installation.html){target=_blank} by following the installation steps.

!!! Important
      Make sure to enable the [`Sync as Kubernetes Secret` feature](https://secrets-store-csi-driver.sigs.k8s.io/topics/sync-as-kubernetes-secret.html){target=_blank} explicitly by setting the helm parameter `syncSecret.enabled` to true.

!!! Note "TL;DR"
      ``` shell
      helm repo add secrets-store-csi-driver https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts
      helm install csi-secrets-store secrets-store-csi-driver/secrets-store-csi-driver --namespace kube-system --set syncSecret.enabled=true
      ```

### Install the AWS Provider
**AWS provider** for the Secrets Store CSI Driver allows you to make secrets stored in Secrets Manager appear as files mounted in Kubernetes pods.

!!! Note "TL;DR"
      ``` shell
      kubectl apply -f https://raw.githubusercontent.com/aws/secrets-store-csi-driver-provider-aws/main/deployment/aws-provider-installer.yaml
      ```

## 3. Store Pachyderm's Secrets in Secrets Manager
In your Secret Manager Console, click on **Store a new secret**, select the **Other type of Secret** (for generic secrets), provide the following Key/Value pairs, then choose a secret name. 


| Secret Key | Description| Value|
|------------|------------|-----|
|root_token|Root clusterAdmin of your cluster| Any|
|postgresql_password|Password to your database| Any|
|OAUTH_CLIENT_SECRET|Oauth client secret for Console <br> Required if you set an Enterprise key| Any|
|enterprise_license_key|Your enterprise license|Your enterprise License key|
|pachd_oauth_client_secret| Oauth client secret for pachd| Any|
|enterprise_secret|Needed if you connect to an enterprise server| Any|


Create your secret, then retrieve its arn. 
It will be needed in the next phase.

## 4- Grant Your EKS Cluster Access To Your Secrets Manager

Your cluster has an OpenID Connect issuer URL associated with it. 
To use IAM roles for service accounts, an IAM OIDC provider must exist for your cluster.

### Create an IAM OIDC Provider

Before granting your EKS pods the proper permissions to access your secrets, you need to **create an IAM OIDC provider** for your cluster or retrieve the arn of your provider if you already have one created.  

Follow the steps in [AWS user guide](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html){target=_blank}

!!! Example "TL;DR"
      ```shell
      eksctl utils associate-iam-oidc-provider --cluster="<cluster-name>"
      ```

### Create An IAM Policy That Grants Read Access To Your Secret 

- Create a new [Policy from your IAM Console](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_create-console.html){target=_blank}
- Select the JSON tab.
- Copy/Paste the following text in the JSON tab


```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "secretsmanager:GetSecretValue",
                "secretsmanager:DescribeSecret",
            ],
            "Resource": [
                <!-- Copy the arn of your secret HERE - see example below>
                "arn:aws:secretsmanager:<region>:<account>:secret:<your secret name>"
            ]
        }
    ]
}
```

This [policy limits the access to the secrets](https://docs.aws.amazon.com/secretsmanager/latest/userguide/auth-and-access_examples.html#auth-and-access_examples_read){target=_blank} that your EKS cluster needs to access.

### Attach Your Policy To An IAM Role and The Role To Your Service Account

[Create an IAM role and attach the IAM policy](https://docs.aws.amazon.com/eks/latest/userguide/create-service-account-iam-policy-and-role.html){target=_blank} that you specified to it. 
The role is associated with a Kubernetes service account created in the namespace that you specify (your cluster's) and annotated with `eks.amazonaws.com/role-arn:arn:aws:iam::111122223333:role/my-role-name`.


!!! Example "TL;DR"
    ``` shell
    eksctl create iamserviceaccount \
    --name "<my-service-account>" \
    --cluster "<my-cluster>" \
    --attach-policy-arn \ "<Copy the arn of your policy HERE>" \
    --approve \ 
    --override-existing-serviceaccounts
    ```


## 5. Mount Your Secrets In Your EKS Cluster
To show secrets in EKS as though they are files on the filesystem, you need to create a **SecretProviderClass** YAML file that contains information about your secrets as well as information on how to display them in the EKS pod. Use the file provided below and run `kubectl apply -f yoursecretclass.yaml`.

The SecretProviderClass must be in the same namespace as the EKS cluster.

```yaml
metadata:
  # Insert your secret name
  name: pach-secrets
spec:
  provider: aws
  parameters:
    objects: |
        - objectName: "pach-secrets"
          objectType: "secretsmanager"
          jmesPath:
            - path: root_token
              objectAlias: root-token
            - path: postgresql_password
              objectAlias: postgresql-password
            - path: OAUTH_CLIENT_SECRET
              objectAlias: OAUTH_CLIENT_SECRET
            - path: enterprise_license_key
              objectAlias: enterprise-license-key
            - path: pachd_oauth_client_secret	
              objectAlias: pachd-oauth-client-secret
            - path: enterprise_secret
              objectAlias: enterprise-secret
  secretObjects:
  - data:
    - key: root-token
      objectName: root-token
    secretName: root-token
    type: Opaque
  - data:
    - key: postgresql-password
      objectName: postgresql-password
    secretName: postgresql-password
    type: Opaque
  - data:
    - key: OAUTH_CLIENT_SECRET
      objectName: OAUTH_CLIENT_SECRET
    secretName: console-oauth-client-secret
    type: Opaque
  - data:
    - key: enterprise-license-key
      objectName: enterprise-license-key
    secretName: enterprise-license-key
    type: Opaque
  - data:
    - key: pachd-oauth-client-secret	
      objectName: pachd-oauth-client-secret	
    secretName: pachd-oauth-client-secret
    type: Opaque
  - data:
    - key: enterprise-secret
      objectName: enterprise-secret
    secretName: enterprise-secret
    type: Opaque
```

## 6. Create A Syncer Pod

Once your secret class is configured, a pod needs to request the class to trigger the CSI driver and retrieve the secrets in Kubernetes. 
Update the file below with your `serviceAccountName` and  `secretProviderClass` before you run a `kubectl apply -f syncerpod.yaml`

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: secret-syncer
spec:
  containers:
  - name: secret-syncer
    image: k8s.gcr.io/pause
    volumeMounts:
    - name: secrets-store-inline
      mountPath: "/mnt/secrets-store"
      readOnly: true
  terminationGracePeriodSeconds: 3
  serviceAccountName: "<Insert your service account name>"
  volumes:
      - name: secrets-store-inline
        csi:
          driver: secrets-store.csi.k8s.io
          readOnly: true
          volumeAttributes:
            # Insert the name of your Secret Provider
            secretProviderClass: "pach-secrets" 
```
Run a quick `kubectl get all` to check on your new pod.

## 8. Update Your Secrets In Pachyderm Values.YAML 

Finally, using the `secretName`(s) of your SecretProviderClass above, update Pachyderm's values.YAML with the list of secrets you will be needing.

Choose the ones that apply to your use case. 

```yaml
deployTarget: LOCAL
global:
  postgresql:
    postgresqlExistingSecretName: postgresql-password
    postgresqlExistingSecretKey: postgresql-password

console:
  enabled: true
  config:
    oauthClientSecretSecretName: console-oauth-client-secret

pachd:
  rootTokenSecretName: root-token
  enterpriseSecretSecretName: enterprise-secret
  oauthClientSecretSecretName: pachd-oauth-client-secret	
  enterpriseLicenseKeySecretName: enterprise-license-key
  activateEnterprise: true
```

Your Secrets Manager is now configured to provide credential values to your cluster, you can go back to your [installation of Pachyderm](../aws-deploy-pachyderm/#3-create-an-s3-bucket)  instructions.


