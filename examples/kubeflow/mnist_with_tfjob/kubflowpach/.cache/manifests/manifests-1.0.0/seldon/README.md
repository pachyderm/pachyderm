# Seldon Kustomize 

## Install Seldon Operator

 * The yaml assumes you will install in kubeflow namespace
 * You need to have installed istio first

```
kustomize build seldon-core-operator/base | kubectl apply -n kubeflow -f -
```

## Updating

This kustomize spec was created from the seldon-core-operator helm chart with:

```
make clean seldon-core-operator/base
```
