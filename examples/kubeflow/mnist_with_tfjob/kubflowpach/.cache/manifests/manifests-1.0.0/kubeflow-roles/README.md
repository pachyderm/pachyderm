# Default Kubeflow ClusterRoles

This manifest package contains the default ClusterRoles Kubeflow uses for defining roles for Kubeflow user Profiles.
These roles are currently assigned to users by Profiles (profile-controller and kfam) Service with the help of Manage Users page in Central Dashboard.

*Note*: `kfctl` assigns the default Kubernetes role `cluster-admin` to the user who deploys Kubeflow for the [GCP IAP configuration](https://github.com/kubeflow/manifests/blob/master/kfdef/kfctl_gcp_iap.yaml).

## How to define role privileges for your Kubeflow application?
Each application defines its own ClusterRole for each role here in kubeflow-roles. We use [ClusterRole Aggregation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles) for these application ClusterRoles to be aggregated to their corresponding Kubeflow roles. An example implementation showing the same can be found here:  

The example is taken from [istio manifests](istio/istio/base/cluster-roles.yaml).
```
---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-istio-admin
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-admin: "true"
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.authorization.kubeflow.org/aggregate-to-kubeflow-istio-admin: "true"
rules: []

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-istio-edit
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-edit: "true"
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-istio-admin: "true"
rules:
- apiGroups: ["istio.io"]
  resources: ["*"]
  verbs:
  - get
  - list
  - watch
  - create
  - delete
  - deletecollection
  - patch
  - update

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-istio-view
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-view: "true"
rules:
- apiGroups: ["istio.io"]
  resources: ["*"]
  verbs:
  - get
  - list
  - watch
```

Note the usage of labels in each ClusterRole to indicate ClusterRole Aggregation with Kubeflow ClusterRoles for this application.

## Reference Links

- [Define Kubeflow cluster role and combine roles](https://github.com/kubeflow/kubeflow/issues/3938)
