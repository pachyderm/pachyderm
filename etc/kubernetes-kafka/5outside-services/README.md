## Expose Kafka outside cluster

Currently supported for brokers, not zookeeper.


### Outside access with one nodeport per broker
outside-0.yml - outside2.yml creates sample services for NodePort access, i.e. inside any firewalls that protect your cluster. Can be switched to LoadBalancer, but then the init script for brokers must be updated to retrieve external names. See discussions around https://github.com/Yolean/kubernetes-kafka/pull/78 for examples.

When using this approach a service must be created for each kafka broker.

### Outside access with hostport
An alternative is to use the hostport for the outside access. When using this only one kafka broker can run on each host, which is a good idea anyway.

In order to switch to hostport the kafka advertise address needs to be switched to the ExternalIP or ExternalDNS name of the node running the broker.
in [kafka/10broker-config.yml](../kafka/10broker-config.yml) switch to
* `OUTSIDE_HOST=$(kubectl get node "$NODE_NAME" -o jsonpath='{.status.addresses[?(@.type=="ExternalIP")].address}')`
* `OUTSIDE_PORT=${OutsidePort}`

and in [kafka/50kafka.yml ](../kafka/50kafka.yml) add the hostport:
```        
        - name: outside
          containerPort: 9094
          hostPort: 9094
```
