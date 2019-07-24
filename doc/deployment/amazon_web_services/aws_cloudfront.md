# Deploy a Pachyderm Cluster with CloudFront

Amazon CloudFront™ is a content delivery network (CDN) that
streams data to your website, service, or application securely
and with great performance. Pachyderm recommends that you
set up Pachyderm with CloudFront for all production
deployment.

To deploy a production-ready Pachyderm cluster with CloudFront,
complete the following steps:

1. [Deploy Pachyderm with CloudFront](#deploy-pachyderm-with-cloudfront)
3. [Apply the Security Credentials](#apply-the-cloudfront-key-pair)
4. [Verify the setup](#verify-the-setup)

## Deploy Pachyderm with CloudFront

Use the Pachyderm AWS deployment script to deploy Kubernetes, CloudFront,
and Pachyderm in a single run by following these steps:

1. Download the `aws.sh` script from the Pachyderm GitHub
repository to your computer:

   ```bash
   $ curl -o aws.sh https://raw.githubusercontent.com/pachyderm/pachyderm/master/etc/deploy/aws.sh
   ```

1. Make the script executable:

   ```bash
   $ chmod +x aws.sh
   ```

1. Run the script:

   ```bash
   $ sudo -E ./aws.sh --region=us-east-1 --zone=us-east-1b --use-cloudfront &> deploy.log
   ```

   The command above redirects the output to the `deploy.log` file.
   You will need this file later.

   **Note:** If you see a few restarts on the `pachd` nodes, that means that
   Kubernetes tried to bring up those pods before `etcd` was ready. Therefore,
   Kubernetes restarted those pods. You can safely ignore this message.

## Apply the CloudFront Key Pair

Only a root AWS account can generate a CloudFront key pair. Therefore,
you might need to request your IT department to provide it.

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

The Key Pair ID is `APKAXXXXXXXXXXXXXXXX`. The other file is
the private key, which looks similar to the following text:

```
$ cat pk-APKAXXXXXXXXXXXX.pem
-----BEGIN RSA PRIVATE KEY-----
...
```

To apply this key pair to your CloudFront distribution, complete
the following steps:

1. Download the `secure-cloudfront.sh` script from the Pachyderm
repository:

   ```bash
   $ curl -o secure-cloudfront.sh https://raw.githubusercontent.com/pachyderm/pachyderm/master/etc/deploy/cloudfront/secure-cloudfront.sh
   ```

1. Make the script executable:

   ```bash
   $ chmod +x secure-cloudfront.sh
   ```

1. From the `deploy.log` file, obtain the S3 bucket name for your
deployment and the CloudFront distribution ID.

1. Apply the key pair to your CloudFront distribution:

   ```bash
   $ ./secure-cloudfront.sh --region us-west-2 --zone us-west-2c --bucket YYYY-pachyderm-store --cloudfront-distribution-id E1BEBVLIDYTLEV  --cloudfront-keypair-id APKAXXXXXXXXXXXX --cloudfront-private-key-file ~/Downloads/pk-APKAXXXXXXXXXXXX.pem
   ```

1. Restart the `pachd` pod for the
changes to take effect:

   ```
   $ kubectl scale --replicas=0 deployment/pachd && kubectl scale --replicas=1 deployment/pachd && kubectl get pod
   ```

1. Verify the setup by checking the `pachd` logs and confirming that
Kubernetes uses the CloudFront credentials:

   ```bash
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

