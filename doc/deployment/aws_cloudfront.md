# AWS CloudFront

To deploy a production ready AWS cluster with CloudFront:

1. [Deploy a cloudfront enabled Pachyderm cluster in AWS](#deploy-a-cloudfront-enabled-cluster-in-aws)
2. [Obtain a cloudfront keypair](#obtain-a-cloudfront-keypair)
3. [Apply the security credentials](#apply-the-security-credentials)
4. [Verify the setup](#verify-the-setup)

## Deploy a cloudfront enabled cluster in AWS

  You'll need to use our "one shot" AWS deployment script to deploy this cluster as follows:

  ```
  $ curl -o aws.sh https://raw.githubusercontent.com/pachyderm/pachyderm/master/etc/deploy/aws.sh
  $ chmod +x aws.sh
  $ sudo -E ./aws.sh --region=us-east-1 --zone=us-east-1b --use-cloudfront &> deploy.log
  ```

  Here we've redirected the output to a file. Make sure you keep this file around for reference.

  **Note:** You may see a few extra restarts on your pachd pod. Sometimes it takes a bit before your cloudfront distribution comes online

## Obtain a cloudfront keypair

  You will most likely need to Ask your IT department for a cloudfront keypair, because only a root AWS account can generate this keypair. You can pass along [this link with instructions](http://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-trusted-signers.html#private-content-creating-cloudfront-key-pairs).

  When you get the keypair, you should receive:

  - the private/public key (although you only need the private key)
  - the keypair ID (which is usually in the filename)

  For example,

  ```
  rsa-APKAXXXXXXXXXXXXXXXX.pem
  pk-APKAXXXXXXXXXXXXXXXX.pem
  ```

  Here we see that the Key Pair ID is `APKAXXXXXXXXXXXXXXXX`, and the second file is the private key, which should look similar to the following:

  ```
  $ cat pk-APKAXXXXXXXXXXXX.pem
  -----BEGIN RSA PRIVATE KEY-----
  ...
  ```

## Apply the security credentials

  You can now run the following script to apply these security credentials to your cloudfront distribution:

  ```
  $ curl -o secure-cloudfront.sh https://raw.githubusercontent.com/pachyderm/pachyderm/master/etc/deploy/cloudfront/secure-cloudfront.sh
  $ chmod +x secure-cloudfront.sh
  $ ./secure-cloudfront.sh --region us-west-2 --zone us-west-2c --bucket YYYY-pachyderm-store --cloudfront-distribution-id E1BEBVLIDYTLEV  --cloudfront-keypair-id APKAXXXXXXXXXXXX --cloudfront-private-key-file ~/Downloads/pk-APKAXXXXXXXXXXXX.pem 
  ```

  where the values for the `--bucket` and `--cloudfront-distribution-id` flags can be obtained from the `deploy.log` file containing your deployment logs.

  You will then need to restart the `pachd` pod in kubernetes for the changes to take effect:

  ```
  $ kubectl scale --replicas=0 deployment/pachd && kubectl scale --replicas=1 deployment/pachd && kubectl get pod
  ```

## Verify the setup

  To verify the setup, we can look at the pachd logs to confirm usage of the cloudfront credentials:

  ```
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

