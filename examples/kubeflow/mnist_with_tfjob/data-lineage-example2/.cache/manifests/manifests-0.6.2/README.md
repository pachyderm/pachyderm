# Manifests
This repo is a [bespoke configuration](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/glossary.md#bespoke-configuration) of kustomize targets used by kubeflow. These targets are traversed by kubeflow's CLI `kfctl`. Each target is compatible with the kustomize CLI and can be processed indendently by kubectl or the kustomize command.

## Organization
Subdirectories within the repo hold kustomize targets (base or overlay subdirectory). Overlays contain additional functionality and multiple overlays may be mixed into the base (described below). Both base and overlay targets are processed by kfctl during generate and apply phases and is detailed in [Kfctl Processing](#kfctl-processing).

See [Best Practices](./docs/KustomizeBestPractices.md) for details on how kustomize targets are created.


## Kfctl Processing
Kfctl traverses directories under manifests to find and build kustomize targets based on the configuration file `app.yaml`. The contents of app.yaml is the result of running kustomize on the base and specific overlays in the kubeflow [config](https://github.com/kubeflow/kubeflow/tree/master/bootstrap/config) directory. The overlays reflect what options are chosen when calling `kfctl init...`.  The kustomize package manager in kfctl will then read app.yaml and apply the packages, components and componentParams to kustomize in the following way:

- **packages**
  - are always top-level directories under the manifests repo
- **components**
  - are also directories but may be a subdirectory in a package.
  - a component may also be a package if there is a base or overlay in the top level directory.
  - otherwise a component is a sub-directory under the package directory.
  - in all cases a component's name in app.yaml must match the directory name.
  - components are output as `<component>.yaml` under the kustomize subdirectory during `kfctl generate...`.
  - in order to output a component, a kustomization.yaml is created above the base or overlay directory and inherits common parameters, namespace and labels of the base or overlay. Additionally it adds the namespace and an application label.
```
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
  - <component>/{base|overlay/<overlay>}
commonLabels:
  app.kubernetes.io/name: <appName>
namespace:
  <namespace>
```
- **component parameters**
  - are applied to a component's params.env file. There must be an entry whose key matches the component parameter. The params.env file is used to generate a ConfigMap. Entries in params.env are resolved as kustomize vars or referenced in a deployment or statefulset env section in which case no var definition is needed.

Multiple overlays -

Kfctl may combine more than one overlay during `kfctl generate ...`. An example is shown below where the profiles target in [manifests](https://github.com/kubeflow/manifests/tree/master/profiles) can include either debug changes in the Deployment or Device information in the Namespace (the devices overlay is not fully integrated with the Profile-controller at this point in time and is intended as an example) or **both**.

```
profiles
├── base
│   └── kustomization.yaml
└── overlays
    ├── debug
    │   └── kustomization.yaml
    └── devices
        └── kustomization.yaml
```

#### What are Multiple Overlays?

Normally kustomize provides the ability to overlay a 'base' set of resources with changes that are merged into the base from resources that are located under an overlays subdirectory. For example
if the kustomize [target](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/glossary.md#target) is named foo there will be a foo/base and possibly one or more overlays such as foo/overlays/bar. A kustomization.yaml file is found in both foo/base and foo/overlays/bar. Running `kustomize build` in foo/base will generate resources as defined in kustomization.yaml. Running `kustomize build` in foo/overlays/bar will generate resources - some of which will overlay the resources in foo/base.

Kustomize doesn't provide for an easy way to combine more than one overlay for example foo/overlays/bar, foo/overlays/baz. However this is an open feature request in kustomize [issues](https://github.com/kubernetes-sigs/kustomize/issues/759). The ability to combine more than one overlay is key to handling components like tf-job-operator which has several types of overlays that can 'mix-in' whether a TFJob is submitted to a namespace or cluster-wide and whether the TFJob uses gang-scheduling.

#### Merging multiple overlays

Since each overlay includes '../../base' as its base set of resources - combining several overlays where each includes '../../base' will cause `kustomize build` to abort, complaining that it recursed on base. The approach is to create a kustomization.yaml at the target level that includes base and the contents of each overlay's kustomization file. This requires some path corrections and some awareness of the behavior of configMapGenerator, secretMapGenerator and how they are copied from each overlay. This kustomization.yaml can be constructed manually, but is integrated within kfctl via the app.yaml file. Using tf-job-operator as an example, if its componentParams has the following
```
  componentParams:
    tf-job-operator:
    - name: overlay
       value: cluster
    - name: overlay
    - value: gangscheduled
```

Then the result will be to combine these overlays eg 'mixin' an overlays in the kustomization.yaml file.

#### Merging multiple overlays to generate app.yaml

In the past when `kfctl init ...` was called it would download the kubeflow repo under `<deployment>/.cache` and read one of the config files under `.cache/kubeflow/<version>/bootstrap/config`. These config files define packages, components and component parameters (among other things). Each config file is a compatible k8 resource of kind *KfDef*. The config files are:

- kfctl_default.yaml
- kfctl_basic_auth.yaml
- kfctl_iap.yaml

Both kfctl_basic_auth.yaml and kfctl_iap.yaml contained the contents of kfctl_default.yaml plus additional changes specific to using kfctl_basic_auth.yaml when --use_basic_auth is passed in or kfctl_iap.yaml when --platform gcp is passed in . This has been refactored to use kustomize where the config/base holds kfctl_default and additional overlays add to the base. The directory now looks like:

```
.
└── config
    ├── base
    │   ├── kfctl_default.yaml
    │   └── kustomization.yaml
    └── overlays
        ├── basic_auth
        │   ├── kfctl_default-patch.yaml
        │   ├── kfctl_default.yaml
        │   └── kustomization.yaml
        ├── gcp
        │   ├── kfctl_default-patch.yaml
        │   ├── kfctl_default.yaml
        │   └── kustomization.yaml
        ├── ksonnet
        │   ├── kfctl_default-patch.yaml
        │   ├── kfctl_default.yaml
        │   └── kustomization.yaml
        └── kustomize
            ├── kfctl_default-patch.yaml
            ├── kfctl_default.yaml
            └── kustomization.yaml
```

Where ksonnet and kustomize hold differing ways of handling the pipeline manifest.

Based on the cli args to `kfctl init...`, the correct overlays will be merged to produce an app.yaml.
The original files have been left as is until UI integration can be completed in a separate PR

### Using kustomize

Generating yaml output for any target can be done using kustomize in the following way:

#### Install kustomize

`go get -u github.com/kubernetes-sigs/kustomize`

### Run kustomize

#### Example

```bash
git clone https://github.com/kubeflow/manifests
cd manifests/<target>/base
kustomize build | tee <output file>
```

Kustomize inputs to kfctl based on app.yaml which is derived from files under config/ such as [kfctl_default.yaml](https://github.com/kubeflow/kubeflow/blob/master/bootstrap/config/kfctl_default.yaml)):

```
apiVersion: kfdef.apps.kubeflow.org/v1alpha1
kind: KfDef
metadata:
  creationTimestamp: null
  name: kubeflow
  namespace: kubeflow
spec:
  appdir: /Users/kdkasrav/kubeflow
  componentParams:
    ambassador:
    - name: ambassadorServiceType
      value: NodePort
  components:
  - metacontroller
  - ambassador
  - argo
  - centraldashboard
  - jupyter-web-app
  - katib
  - notebook-controller
  - pipeline
  - profiles
  - pytorch-operator
  - tensorboard
  - tf-job-operator
  - application
  manifestsRepo: /Users/kdkasrav/kubeflow/.cache/manifests/pull/13/head
  packageManager: kustomize@pull/13
  packages:
  - application
  - argo
  - common
  - examples
  - gcp
  - jupyter
  - katib
  - metacontroller
  - modeldb
  - mpi-job
  - pipeline
  - profiles
  - pytorch-job
  - seldon
  - tensorboard
  - tf-serving
  - tf-training
  repo: /Users/kdkasrav/kubeflow/.cache/kubeflow/pull/2971/head/kubeflow
  useBasicAuth: false
  useIstio: false
  version: pull/2971
```

Outputs from kfctl (no platform specified):
```
kustomize/
├── ambassador.yaml
├── api-service.yaml
├── argo.yaml
├── centraldashboard.yaml
├── jupyter-web-app.yaml
├── katib.yaml
├── metacontroller.yaml
├── minio.yaml
├── mysql.yaml
├── notebook-controller.yaml
├── persistent-agent.yaml
├── pipelines-runner.yaml
├── pipelines-ui.yaml
├── pipelines-viewer.yaml
├── pytorch-operator.yaml
├── scheduledworkflow.yaml
├── tensorboard.yaml
└── tf-job-operator.yaml
```

