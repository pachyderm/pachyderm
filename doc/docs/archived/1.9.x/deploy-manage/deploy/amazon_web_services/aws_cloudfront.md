# Deploy a Pachyderm Cluster with CloudFront

After you have an EKS cluster or a Kubernetes cluster
deployed with `kops` ready,
you can integrate it with Amazon
CloudFront™.

Amazon CloudFront is a content delivery network (CDN) that
streams data to your website, service, or application securely
and with great performance. Pachyderm recommends that you
set up Pachyderm with CloudFront for all production
deployments.

To deploy Pachyderm cluster with CloudFront,
complete the following steps:

1. [Create a CloudFront Distribution](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/GettingStarted.html#GettingStartedCreateDistribution)
1. [Deploy Pachyderm with an IAM role](aws-deploy-pachyderm.md)
1. [Apply the CloudFront Key Pair](#apply-the-cloudfront-key-pair)

## Apply the CloudFront Key Pair

If you need to create signed URLs and
signed cookies for the data that goes to Pachyderm, you need to
configure your AWS account to use a valid CloudFront key pair.
Only a root AWS account can generate these secure credentials. Therefore,
you might need to request your IT department to create them for you.

For more information, see the [Amazon documentation](http://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-trusted-signers.html#private-content-creating-cloudfront-key-pairs).

The CloudFront key pair includes the following attributes:

- The private and public key. For this deployment, you only need the private
key.
- The key pair ID. Typically, the key pair ID is recorded in the filename.

**Example:**

```
rsa-APKAXXXXXXXXXXXXXXXX.pem
pk-APKAXXXXXXXXXXXXXXXX.pem
```

The key-pair ID is `APKAXXXXXXXXXXXXXXXX`. The other file is
the private key, which looks similar to the following text:

!!! example
    ```shell
    $ cat pk-APKAXXXXXXXXXXXX.pem
    -----BEGIN RSA PRIVATE KEY-----
    ...
    ```

To apply this key pair to your CloudFront distribution, complete
the following steps:

1. Download the `secure-cloudfront.sh` script from the Pachyderm
repository:

   ```shell
   $ curl -o secure-cloudfront.sh https://raw.githubusercontent.com/pachyderm/pachyderm/master/etc/deploy/cloudfront/secure-cloudfront.sh
   ```

1. Make the script executable:

   ```shell
   $ chmod +x secure-cloudfront.sh
   ```

1. From the `deploy.log` file, obtain the S3 bucket name for your
deployment and the CloudFront distribution ID.

1. Apply the key pair to your CloudFront distribution:

   ```shell
   $ ./secure-cloudfront.sh --region us-west-2 --zone us-west-2c --bucket YYYY-pachyderm-store --cloudfront-distribution-id E1BEBVLIDYTLEV  --cloudfront-keypair-id APKAXXXXXXXXXXXX --cloudfront-private-key-file ~/Downloads/pk-APKAXXXXXXXXXXXX.pem
   ```

1. Restart the `pachd` pod for the
changes to take effect:

   ```shell
   $ kubectl scale --replicas=0 deployment/pachd && kubectl scale --replicas=1 deployment/pachd && kubectl get pod
   ```

1. Verify the setup by checking the `pachd` logs and confirming that
Kubernetes uses the CloudFront credentials:

   ```shell
   $ kubectl get pod
   NAME                        READY     STATUS             RESTARTS   AGE
   etcd-0                   1/1       Running            0          19h
   etcd-1                   1/1       Running            0          19h
   etcd-2                   1/1       Running            0          19h
   pachd-2796595787-9x0qf   1/1       Running            0          16h

   $ kubectl logs pachd-2796595787-9x0qf | grep cloudfront
   2017-06-09T22:56:27Z INFO  AWS deployed with cloudfront distribution at d3j9kenawdv8p0
   2017-06-09T22:56:27Z INFO  Using cloudfront security credentials - keypair ID (APKAXXXXXXXXX) - to sign cloudfront URLs
   ```

