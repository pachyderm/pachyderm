# Connecting to your Pachyderm Cluster


## Directly

This is the recommended approach, especially if you're using the Pachyderm Dashboard or doing large uploads/downloads.

### Pachd

To expose the pachd service, you need to change the k8s service:

```
$ kubectl edit svc/pachd
```

Then mark the `type` as `LoadBalancer`

If you've gone to the trouble of [deploying within an existing VPC](./existing_vpc), you probably want to limit access to the cluster to IPs originating from this VPC.

In this case, you want an internal load balancer. To expose the pachd service (for `pachctl` access), you'll need to also add the annotation:

```
service.beta.kubernetes.io/aws-load-balancer-internal: '0.0.0.0/0'
```

Once the load balancer is provisioned, you'll see the address that was provisioned via `kubectl get svc/pachd -o yaml`

E.g. `internal-afsdfasdlkfjh34lkjh-3485763487.us-west-1.elb.amazonaws.com:650`

Now set the pachd address of the associated context, so `pachctl` knows to directly connect:

```
$ pachctl config update context `pachctl config get active-context` --pachd-address=internal-afsdfasdlkfjh34lkjh-3485763487.us-west-1.elb.amazonaws.com:650
```

Finally, you can test that it's working by doing:

```

$ pachctl version
COMPONENT           VERSION                                          
pachctl             1.7.3  
pachd               1.7.3
```

### Dash

Similar to exposing the pachd service above, you'll want to make the same modifications to the dash service:

```
$ kubectl edit svc/dash
```

The one additional configuration you'll need to add is the following annotation:

```
service.beta.kubernetes.io/aws-load-balancer-connection-idle-timeout: "3600"
```

(This is to allow the long lived websockets that the dash uses to stay alive through the load balancer)

## Port Forwarding

Whenever you run a `pachctl` command and the pachd address config is not set, `pachctl` implicitly starts port forwarding to try to connect to your cluster. Port forwarding is the easiest way to poke around and verify your cluster is working, however, we don't recommend using it for production workloads, as it is rate limited to about 1MB/s.

You can also explicitly start port forwarding via `pachctl port-forward`. This has the added bonus of port forwarding for Pachyderm Dashboard-related functionality. From there, you'll be able to access the Dashboard at `localhost:30080`.
