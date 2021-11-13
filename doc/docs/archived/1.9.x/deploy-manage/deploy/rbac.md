# RBAC

Pachyderm has support for Kubernetes Role-Based Access Controls (RBAC) and is a default part of all Pachyderm deployments. For most users, you shouldn't have any issues as Pachyderm takes care of setting all the RBAC permissions automatically. However, if you are deploying Pachyderm on a cluster that your company owns, security policies might not allow certain RBAC permissions by default. Therefore, it's suggested that you contact your Kubernetes admin and provide the following to ensure you don't encounter any permissions issues:

Pachyderm Permission Requirements
{% raw %}

```shell
Rules: []rbacv1.PolicyRule{{
		APIGroups: []string{""},
		Verbs:     []string{"get", "list", "watch"},
		Resources: []string{"nodes", "pods", "pods/log", "endpoints"},
		}, {
		APIGroups: []string{""},
		Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
		Resources: []string{"replicationcontrollers", "services"},
		}, {
		APIGroups:     []string{""},
		Verbs:         []string{"get", "list", "watch", "create", "update", "delete"},
		Resources:     []string{"secrets"},
		ResourceNames: []string{client.StorageSecretName},
		}},
```
{% endraw %}
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

## RBAC Permissions on GKE
If you're deploying Pachyderm on GKE and run into the following error:

```
Error from server (Forbidden): error when creating "STDIN": clusterroles.rbac.authorization.k8s.io "pachyderm" is forbidden: attempt to grant extra privileges:
```

Run the following and redeploy Pachyderm:

```
kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user=$(gcloud config get-value account)

```

