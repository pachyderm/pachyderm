# RBAC

Pachyderm has support for Kubernetes Role-Based Access Controls (RBAC).
This support is a default part of all Pachyderm deployments, there's nothing
special for you to do as a user. You can see the ClusterRole which is created
for Pachyderm's service account by doing:

```shell
kubectl get clusterrole/pachyderm -o json
```

## RBAC and DNS
Kubernetes currently (as of 1.8.0) has a bug that prevents kube-dns from
working with RBAC. Not having DNS will make Pachyderm effectively unusable. You
can tell if you're being affected by the bug like so:

```shell
$ kubectl get all --namespace=kube-system
NAME              DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deploy/kube-dns   1         1         1            0           3m

NAME                     DESIRED   CURRENT   READY     AGE
rs/kube-dns-86f6f55dd5   1         1         0         3m

NAME                            READY     STATUS    RESTARTS   AGE
po/kube-addon-manager-oryx      1/1       Running   0          3m
po/kube-dns-86f6f55dd5-xksnb    2/3       Running   4          3m
po/kubernetes-dashboard-bzjjh   1/1       Running   0          3m
po/storage-provisioner          1/1       Running   0          3m

NAME                      DESIRED   CURRENT   READY     AGE
rc/kubernetes-dashboard   1         1         1         3m

NAME                       TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)         AGE
svc/kube-dns               ClusterIP   10.96.0.10     <none>        53/UDP,53/TCP   3m
svc/kubernetes-dashboard   NodePort    10.97.194.16   <none>        80:30000/TCP    3m
```

Notice how `po/kubernetes-dashboard-bzjjh` only has 2/3 pods ready and has 4 restarts.
To fix this do:

```shell
kubectl -n kube-system create sa kube-dns
kubectl -n kube-system patch deploy/kube-dns -p '{"spec": {"template": {"spec": {"serviceAccountName": "kube-dns"}}}}'
```

this will tell Kubernetes that `kube-dns` should use the appropriate
ServiceAccount. Kubernetes creates the ServiceAccount, it just doesn't actually
use it.
