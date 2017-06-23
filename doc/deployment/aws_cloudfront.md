# AWS Cloudfront

To deploy a production ready AWS cluster with CloudFront

1) Follow the instructions for a normal AWS deployment

You'll need to use the 'one shot' script. To do that you'll need to clone this repo, cd to it, and run the script:

```
$ git clone git@github.com:pachyderm/pachyderm
$ cd pachyderm
$ sudo -E ./etc/deploy/aws.sh --region=us-east-1 --zone=us-east-1b --use-cloudfront &> deploy.log
```

Here we've redirected the output to a file. Make sure you keep this file around for reference.

Note: You may see a few extra restarts on your pachd pod. Sometimes it takes a bit before your cloudfront distribution comes online

2) Ask your IT department for a cloudfront keypair

[You can pass along this link for instructions](http://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-trusted-signers.html#private-content-creating-cloudfront-key-pairs)

Only the root AWS account can generate this keypair.

You should receive:

- the private/public key (you only really need the private key)
- the keypair ID (usually its in the filename)

e.g.

```
rsa-APKAXXXXXXXXXXXXXXXX.pem
pk-APKAXXXXXXXXXXXXXXXX.pem
```

Here we see the Key Pair ID is `APKAXXXXXXXXXXXXXXXX`. The latter file is the private key. You can check by doing:

```
$ cat pk-APKAXXXXXXXXXXXX.pem
-----BEGIN RSA PRIVATE KEY-----
...
```

3) Run the script to apply these security credentials to your cloudfront distribution:

```
$./etc/deploy/cloudfront/secure-cloudfront.sh --region us-west-2 --zone us-west-2c --bucket YYYY-pachyderm-store --cloudfront-distribution-id E1BEBVLIDYTLEV  --cloudfront-keypair-id APKAXXXXXXXXXXXX --cloudfront-private-key-file ~/Downloads/pk-APKAXXXXXXXXXXXX.pem 
```

You'll need to look in the file you saved from above `deploy.log` for the values for the `--bucket` and `--cloudfront-distribution-id` flags

Then restart the pachd pod like so:

```
kubectl scale --replicas=0 deployment/pachd && kubectl scale --replicas=1 deployment/pachd && kubectl get pod
```

After that, you're cloudfront setup is all ready!

4) Verify the setup

To verify the setup, we can look at the pachd logs to make sure the cloudfront credentials are being used:

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

