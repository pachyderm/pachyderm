package tests_test

import (
	"sigs.k8s.io/kustomize/v3/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/v3/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/loader"
	"sigs.k8s.io/kustomize/v3/pkg/plugins"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
	"sigs.k8s.io/kustomize/v3/pkg/target"
	"sigs.k8s.io/kustomize/v3/pkg/validators"
	"testing"
)

func writeSeldonCoreOperatorOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/seldon/seldon-core-operator/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: seldon-core-operator
spec:
  componentKinds:
  - group: apps/v1
    kind: StatefulSet
  - group: v1
    kind: Service
  - group: apps/v1
    kind: Deployment
  - group: v1
    kind: Secret
  - group: v1
    kind: ConfigMap
  description: Seldon allows users to create ML Inference Graphs to deploy their models
    and serve predictions
  icons: null
  keywords:
  - seldon
  - inference
  links:
  - description: Docs
    url: https://docs.seldon.io/projects/seldon-core/en/v1.0.1/
  maintainers:
  - email: dev@seldon.io
    name: Seldon
  owners:
  - email: dev@seldon.io
    name: Seldon
  selector:
    matchLabels:
      app.kubernetes.io/component: seldon
      app.kubernetes.io/instance: seldon-1.15
      app.kubernetes.io/managed-by: kfctl
      app.kubernetes.io/name: seldon
      app.kubernetes.io/part-of: kubeflow
      app.kubernetes.io/version: '1.15'
  type: seldon-core-operator
  version: v1
`)
	th.writeK("/manifests/seldon/seldon-core-operator/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
bases:
- ../../base
commonLabels:
  app.kubernetes.io/component: seldon
  app.kubernetes.io/instance: seldon-1.15
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/name: seldon-core-operator
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: '1.15'
kind: Kustomization
resources:
- application.yaml
`)
	th.writeF("/manifests/seldon/seldon-core-operator/base/resources.yaml", `
---
# Source: seldon-core-operator/templates/configmap_seldon-config.yaml
apiVersion: v1
data:
  credentials: '{"gcs":{"gcsCredentialFileName":"gcloud-application-credentials.json"},"s3":{"s3AccessKeyIDName":"awsAccessKeyID","s3SecretAccessKeyName":"awsSecretAccessKey"}}'
  predictor_servers: '{"MLFLOW_SERVER":{"grpc":{"defaultImageVersion":"0.2","image":"seldonio/mlflowserver_grpc"},"rest":{"defaultImageVersion":"0.2","image":"seldonio/mlflowserver_rest"}},"SKLEARN_SERVER":{"grpc":{"defaultImageVersion":"0.2","image":"seldonio/sklearnserver_grpc"},"rest":{"defaultImageVersion":"0.2","image":"seldonio/sklearnserver_rest"}},"TENSORFLOW_SERVER":{"grpc":{"defaultImageVersion":"0.7","image":"seldonio/tfserving-proxy_grpc"},"rest":{"defaultImageVersion":"0.7","image":"seldonio/tfserving-proxy_rest"},"tensorflow":true,"tfImage":"tensorflow/serving:latest"},"XGBOOST_SERVER":{"grpc":{"defaultImageVersion":"0.2","image":"seldonio/xgboostserver_grpc"},"rest":{"defaultImageVersion":"0.2","image":"seldonio/xgboostserver_rest"}}}'
  storageInitializer: '{"cpuLimit":"1","cpuRequest":"100m","image":"gcr.io/kfserving/storage-initializer:0.2.1","memoryLimit":"1Gi","memoryRequest":"100Mi"}'
kind: ConfigMap
metadata:
  labels:
    app: seldon
    app.kubernetes.io/instance: 'RELEASE-NAME'
    app.kubernetes.io/name: 'seldon-core-operator'
    app.kubernetes.io/version: '1.0.1'
    control-plane: seldon-controller-manager
  name: seldon-config
  namespace: 'kubeflow'
---
# Source: seldon-core-operator/templates/serviceaccount_seldon-manager.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: seldon
    app.kubernetes.io/instance: 'RELEASE-NAME'
    app.kubernetes.io/name: 'seldon-core-operator'
    app.kubernetes.io/version: '1.0.1'
  name: 'seldon-manager'
  namespace: 'kubeflow'
---
# Source: seldon-core-operator/templates/customresourcedefinition_seldondeployments.machinelearning.seldon.io.yaml
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  annotations:
    cert-manager.io/inject-ca-from: 'kubeflow/seldon-serving-cert'
  creationTimestamp: null
  labels:
    app: seldon
    app.kubernetes.io/instance: 'RELEASE-NAME'
    app.kubernetes.io/name: 'seldon-core-operator'
    app.kubernetes.io/version: '1.0.1'
  name: seldondeployments.machinelearning.seldon.io
spec:
  group: machinelearning.seldon.io
  names:
    kind: SeldonDeployment
    plural: seldondeployments
    shortNames:
    - sdep
  scope: Namespaced
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: SeldonDeployment is the Schema for the seldondeployments API
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          description: SeldonDeploymentSpec defines the desired state of SeldonDeployment
          properties:
            annotations:
              additionalProperties:
                type: string
              type: object
            name:
              type: string
            oauth_key:
              type: string
            oauth_secret:
              type: string
            predictors:
              items:
                properties:
                  annotations:
                    additionalProperties:
                      type: string
                    type: object
                  componentSpecs:
                    items:
                      properties:
                        hpaSpec:
                          properties:
                            maxReplicas:
                              format: int32
                              type: integer
                            metrics:
                              items:
                                description: MetricSpec specifies how to scale based on a single metric (only `+"`"+`type`+"`"+` and one other matching field should be set at once).
                                properties:
                                  external:
                                    description: external refers to a global metric that is not associated with any Kubernetes object. It allows autoscaling based on information coming from components running outside of cluster (for example length of queue in cloud messaging service, or QPS from loadbalancer running outside of cluster).
                                    properties:
                                      metricName:
                                        description: metricName is the name of the metric in question.
                                        type: string
                                      metricSelector:
                                        description: metricSelector is used to identify a specific time series within a given metric.
                                        properties:
                                          matchExpressions:
                                            description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                            items:
                                              description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                              properties:
                                                key:
                                                  description: key is the label key that the selector applies to.
                                                  type: string
                                                operator:
                                                  description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                  type: string
                                                values:
                                                  description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                  items:
                                                    type: string
                                                  type: array
                                              required:
                                              - key
                                              - operator
                                              type: object
                                            type: array
                                          matchLabels:
                                            additionalProperties:
                                              type: string
                                            description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                            type: object
                                        type: object
                                      targetAverageValue:
                                        description: targetAverageValue is the target per-pod value of global metric (as a quantity). Mutually exclusive with TargetValue.
                                        type: string
                                      targetValue:
                                        description: targetValue is the target value of the metric (as a quantity). Mutually exclusive with TargetAverageValue.
                                        type: string
                                    required:
                                    - metricName
                                    type: object
                                  object:
                                    description: object refers to a metric describing a single kubernetes object (for example, hits-per-second on an Ingress object).
                                    properties:
                                      averageValue:
                                        description: averageValue is the target value of the average of the metric across all relevant pods (as a quantity)
                                        type: string
                                      metricName:
                                        description: metricName is the name of the metric in question.
                                        type: string
                                      selector:
                                        description: selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping When unset, just the metricName will be used to gather metrics.
                                        properties:
                                          matchExpressions:
                                            description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                            items:
                                              description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                              properties:
                                                key:
                                                  description: key is the label key that the selector applies to.
                                                  type: string
                                                operator:
                                                  description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                  type: string
                                                values:
                                                  description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                  items:
                                                    type: string
                                                  type: array
                                              required:
                                              - key
                                              - operator
                                              type: object
                                            type: array
                                          matchLabels:
                                            additionalProperties:
                                              type: string
                                            description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                            type: object
                                        type: object
                                      target:
                                        description: target is the described Kubernetes object.
                                        properties:
                                          apiVersion:
                                            description: API version of the referent
                                            type: string
                                          kind:
                                            description: 'Kind of the referent; More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds"'
                                            type: string
                                          name:
                                            description: 'Name of the referent; More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                                            type: string
                                        required:
                                        - kind
                                        - name
                                        type: object
                                      targetValue:
                                        description: targetValue is the target value of the metric (as a quantity).
                                        type: string
                                    required:
                                    - metricName
                                    - target
                                    - targetValue
                                    type: object
                                  pods:
                                    description: pods refers to a metric describing each pod in the current scale target (for example, transactions-processed-per-second).  The values will be averaged together before being compared to the target value.
                                    properties:
                                      metricName:
                                        description: metricName is the name of the metric in question
                                        type: string
                                      selector:
                                        description: selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping When unset, just the metricName will be used to gather metrics.
                                        properties:
                                          matchExpressions:
                                            description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                            items:
                                              description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                              properties:
                                                key:
                                                  description: key is the label key that the selector applies to.
                                                  type: string
                                                operator:
                                                  description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                  type: string
                                                values:
                                                  description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                  items:
                                                    type: string
                                                  type: array
                                              required:
                                              - key
                                              - operator
                                              type: object
                                            type: array
                                          matchLabels:
                                            additionalProperties:
                                              type: string
                                            description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                            type: object
                                        type: object
                                      targetAverageValue:
                                        description: targetAverageValue is the target value of the average of the metric across all relevant pods (as a quantity)
                                        type: string
                                    required:
                                    - metricName
                                    - targetAverageValue
                                    type: object
                                  resource:
                                    description: resource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing each pod in the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source.
                                    properties:
                                      name:
                                        description: name is the name of the resource in question.
                                        type: string
                                      targetAverageUtilization:
                                        description: targetAverageUtilization is the target value of the average of the resource metric across all relevant pods, represented as a percentage of the requested value of the resource for the pods.
                                        format: int32
                                        type: integer
                                      targetAverageValue:
                                        description: targetAverageValue is the target value of the average of the resource metric across all relevant pods, as a raw value (instead of as a percentage of the request), similar to the "pods" metric source type.
                                        type: string
                                    required:
                                    - name
                                    type: object
                                  type:
                                    description: type is the type of metric source.  It should be one of "Object", "Pods" or "Resource", each mapping to a matching field in the object.
                                    type: string
                                required:
                                - type
                                type: object
                              type: array
                            minReplicas:
                              format: int32
                              type: integer
                          required:
                          - maxReplicas
                          type: object
                        metadata:
                          type: object
                        spec:
                          description: PodSpec is a description of a pod.
                          properties:
                            activeDeadlineSeconds:
                              description: Optional duration in seconds the pod may be active on the node relative to StartTime before the system will actively try to mark it failed and kill associated containers. Value must be a positive integer.
                              format: int64
                              type: integer
                            affinity:
                              description: If specified, the pod's scheduling constraints
                              properties:
                                nodeAffinity:
                                  description: Describes node affinity scheduling rules for the pod.
                                  properties:
                                    preferredDuringSchedulingIgnoredDuringExecution:
                                      description: The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node matches the corresponding matchExpressions; the node(s) with the highest sum are the most preferred.
                                      items:
                                        description: An empty preferred scheduling term matches all objects with implicit weight 0 (i.e. it's a no-op). A null preferred scheduling term matches no objects (i.e. is also a no-op).
                                        properties:
                                          preference:
                                            description: A node selector term, associated with the corresponding weight.
                                            properties:
                                              matchExpressions:
                                                description: A list of node selector requirements by node's labels.
                                                items:
                                                  description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                  properties:
                                                    key:
                                                      description: The label key that the selector applies to.
                                                      type: string
                                                    operator:
                                                      description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                                      type: string
                                                    values:
                                                      description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                                      items:
                                                        type: string
                                                      type: array
                                                  required:
                                                  - key
                                                  - operator
                                                  type: object
                                                type: array
                                              matchFields:
                                                description: A list of node selector requirements by node's fields.
                                                items:
                                                  description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                  properties:
                                                    key:
                                                      description: The label key that the selector applies to.
                                                      type: string
                                                    operator:
                                                      description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                                      type: string
                                                    values:
                                                      description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                                      items:
                                                        type: string
                                                      type: array
                                                  required:
                                                  - key
                                                  - operator
                                                  type: object
                                                type: array
                                            type: object
                                          weight:
                                            description: Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.
                                            format: int32
                                            type: integer
                                        required:
                                        - preference
                                        - weight
                                        type: object
                                      type: array
                                    requiredDuringSchedulingIgnoredDuringExecution:
                                      description: If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to an update), the system may or may not try to eventually evict the pod from its node.
                                      properties:
                                        nodeSelectorTerms:
                                          description: Required. A list of node selector terms. The terms are ORed.
                                          items:
                                            description: A null or empty node selector term matches no objects. The requirements of them are ANDed. The TopologySelectorTerm type implements a subset of the NodeSelectorTerm.
                                            properties:
                                              matchExpressions:
                                                description: A list of node selector requirements by node's labels.
                                                items:
                                                  description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                  properties:
                                                    key:
                                                      description: The label key that the selector applies to.
                                                      type: string
                                                    operator:
                                                      description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                                      type: string
                                                    values:
                                                      description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                                      items:
                                                        type: string
                                                      type: array
                                                  required:
                                                  - key
                                                  - operator
                                                  type: object
                                                type: array
                                              matchFields:
                                                description: A list of node selector requirements by node's fields.
                                                items:
                                                  description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                  properties:
                                                    key:
                                                      description: The label key that the selector applies to.
                                                      type: string
                                                    operator:
                                                      description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                                      type: string
                                                    values:
                                                      description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                                      items:
                                                        type: string
                                                      type: array
                                                  required:
                                                  - key
                                                  - operator
                                                  type: object
                                                type: array
                                            type: object
                                          type: array
                                      required:
                                      - nodeSelectorTerms
                                      type: object
                                  type: object
                                podAffinity:
                                  description: Describes pod affinity scheduling rules (e.g. co-locate this pod in the same node, zone, etc. as some other pod(s)).
                                  properties:
                                    preferredDuringSchedulingIgnoredDuringExecution:
                                      description: The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.
                                      items:
                                        description: The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)
                                        properties:
                                          podAffinityTerm:
                                            description: Required. A pod affinity term, associated with the corresponding weight.
                                            properties:
                                              labelSelector:
                                                description: A label query over a set of resources, in this case pods.
                                                properties:
                                                  matchExpressions:
                                                    description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                                    items:
                                                      description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                      properties:
                                                        key:
                                                          description: key is the label key that the selector applies to.
                                                          type: string
                                                        operator:
                                                          description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                          type: string
                                                        values:
                                                          description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                          items:
                                                            type: string
                                                          type: array
                                                      required:
                                                      - key
                                                      - operator
                                                      type: object
                                                    type: array
                                                  matchLabels:
                                                    additionalProperties:
                                                      type: string
                                                    description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                                    type: object
                                                type: object
                                              namespaces:
                                                description: namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means "this pod's namespace"
                                                items:
                                                  type: string
                                                type: array
                                              topologyKey:
                                                description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                                type: string
                                            required:
                                            - topologyKey
                                            type: object
                                          weight:
                                            description: weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
                                            format: int32
                                            type: integer
                                        required:
                                        - podAffinityTerm
                                        - weight
                                        type: object
                                      type: array
                                    requiredDuringSchedulingIgnoredDuringExecution:
                                      description: If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.
                                      items:
                                        description: Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running
                                        properties:
                                          labelSelector:
                                            description: A label query over a set of resources, in this case pods.
                                            properties:
                                              matchExpressions:
                                                description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                                items:
                                                  description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                  properties:
                                                    key:
                                                      description: key is the label key that the selector applies to.
                                                      type: string
                                                    operator:
                                                      description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                      type: string
                                                    values:
                                                      description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                      items:
                                                        type: string
                                                      type: array
                                                  required:
                                                  - key
                                                  - operator
                                                  type: object
                                                type: array
                                              matchLabels:
                                                additionalProperties:
                                                  type: string
                                                description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                                type: object
                                            type: object
                                          namespaces:
                                            description: namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means "this pod's namespace"
                                            items:
                                              type: string
                                            type: array
                                          topologyKey:
                                            description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                            type: string
                                        required:
                                        - topologyKey
                                        type: object
                                      type: array
                                  type: object
                                podAntiAffinity:
                                  description: Describes pod anti-affinity scheduling rules (e.g. avoid putting this pod in the same node, zone, etc. as some other pod(s)).
                                  properties:
                                    preferredDuringSchedulingIgnoredDuringExecution:
                                      description: The scheduler will prefer to schedule pods to nodes that satisfy the anti-affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling anti-affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.
                                      items:
                                        description: The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)
                                        properties:
                                          podAffinityTerm:
                                            description: Required. A pod affinity term, associated with the corresponding weight.
                                            properties:
                                              labelSelector:
                                                description: A label query over a set of resources, in this case pods.
                                                properties:
                                                  matchExpressions:
                                                    description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                                    items:
                                                      description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                      properties:
                                                        key:
                                                          description: key is the label key that the selector applies to.
                                                          type: string
                                                        operator:
                                                          description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                          type: string
                                                        values:
                                                          description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                          items:
                                                            type: string
                                                          type: array
                                                      required:
                                                      - key
                                                      - operator
                                                      type: object
                                                    type: array
                                                  matchLabels:
                                                    additionalProperties:
                                                      type: string
                                                    description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                                    type: object
                                                type: object
                                              namespaces:
                                                description: namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means "this pod's namespace"
                                                items:
                                                  type: string
                                                type: array
                                              topologyKey:
                                                description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                                type: string
                                            required:
                                            - topologyKey
                                            type: object
                                          weight:
                                            description: weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
                                            format: int32
                                            type: integer
                                        required:
                                        - podAffinityTerm
                                        - weight
                                        type: object
                                      type: array
                                    requiredDuringSchedulingIgnoredDuringExecution:
                                      description: If the anti-affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the anti-affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.
                                      items:
                                        description: Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running
                                        properties:
                                          labelSelector:
                                            description: A label query over a set of resources, in this case pods.
                                            properties:
                                              matchExpressions:
                                                description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                                items:
                                                  description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                  properties:
                                                    key:
                                                      description: key is the label key that the selector applies to.
                                                      type: string
                                                    operator:
                                                      description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                      type: string
                                                    values:
                                                      description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                      items:
                                                        type: string
                                                      type: array
                                                  required:
                                                  - key
                                                  - operator
                                                  type: object
                                                type: array
                                              matchLabels:
                                                additionalProperties:
                                                  type: string
                                                description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                                type: object
                                            type: object
                                          namespaces:
                                            description: namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means "this pod's namespace"
                                            items:
                                              type: string
                                            type: array
                                          topologyKey:
                                            description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                            type: string
                                        required:
                                        - topologyKey
                                        type: object
                                      type: array
                                  type: object
                              type: object
                            automountServiceAccountToken:
                              description: AutomountServiceAccountToken indicates whether a service account token should be automatically mounted.
                              type: boolean
                            containers:
                              description: List of containers belonging to the pod. Containers cannot currently be added or removed. There must be at least one container in a Pod. Cannot be updated.
                              items:
                                description: A single application container that you want to run within a pod.
                                properties:
                                  args:
                                    description: 'Arguments to the entrypoint. The docker image''s CMD is used if this is not provided. Variable references $(VAR_NAME) are expanded using the container''s environment. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell'
                                    items:
                                      type: string
                                    type: array
                                  command:
                                    description: 'Entrypoint array. Not executed within a shell. The docker image''s ENTRYPOINT is used if this is not provided. Variable references $(VAR_NAME) are expanded using the container''s environment. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell'
                                    items:
                                      type: string
                                    type: array
                                  env:
                                    description: List of environment variables to set in the container. Cannot be updated.
                                    items:
                                      description: EnvVar represents an environment variable present in a Container.
                                      properties:
                                        name:
                                          description: Name of the environment variable. Must be a C_IDENTIFIER.
                                          type: string
                                        value:
                                          description: 'Variable references $(VAR_NAME) are expanded using the previous defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to "".'
                                          type: string
                                        valueFrom:
                                          description: Source for the environment variable's value. Cannot be used if value is not empty.
                                          properties:
                                            configMapKeyRef:
                                              description: Selects a key of a ConfigMap.
                                              properties:
                                                key:
                                                  description: The key to select.
                                                  type: string
                                                name:
                                                  description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?'
                                                  type: string
                                                optional:
                                                  description: Specify whether the ConfigMap or it's key must be defined
                                                  type: boolean
                                              required:
                                              - key
                                              type: object
                                            fieldRef:
                                              description: 'Selects a field of the pod: supports metadata.name, metadata.namespace, metadata.labels, metadata.annotations, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP.'
                                              properties:
                                                apiVersion:
                                                  description: Version of the schema the FieldPath is written in terms of, defaults to "v1".
                                                  type: string
                                                fieldPath:
                                                  description: Path of the field to select in the specified API version.
                                                  type: string
                                              required:
                                              - fieldPath
                                              type: object
                                            resourceFieldRef:
                                              description: 'Selects a resource of the container: only resources limits and requests (limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.'
                                              properties:
                                                containerName:
                                                  description: 'Container name: required for volumes, optional for env vars'
                                                  type: string
                                                divisor:
                                                  description: Specifies the output format of the exposed resources, defaults to "1"
                                                  type: string
                                                resource:
                                                  description: 'Required: resource to select'
                                                  type: string
                                              required:
                                              - resource
                                              type: object
                                            secretKeyRef:
                                              description: Selects a key of a secret in the pod's namespace
                                              properties:
                                                key:
                                                  description: The key of the secret to select from.  Must be a valid secret key.
                                                  type: string
                                                name:
                                                  description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?'
                                                  type: string
                                                optional:
                                                  description: Specify whether the Secret or it's key must be defined
                                                  type: boolean
                                              required:
                                              - key
                                              type: object
                                          type: object
                                      required:
                                      - name
                                      type: object
                                    type: array
                                  envFrom:
                                    description: List of sources to populate environment variables in the container. The keys defined within a source must be a C_IDENTIFIER. All invalid keys will be reported as an event when the container is starting. When a key exists in multiple sources, the value associated with the last source will take precedence. Values defined by an Env with a duplicate key will take precedence. Cannot be updated.
                                    items:
                                      description: EnvFromSource represents the source of a set of ConfigMaps
                                      properties:
                                        configMapRef:
                                          description: The ConfigMap to select from
                                          properties:
                                            name:
                                              description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?'
                                              type: string
                                            optional:
                                              description: Specify whether the ConfigMap must be defined
                                              type: boolean
                                          type: object
                                        prefix:
                                          description: An optional identifier to prepend to each key in the ConfigMap. Must be a C_IDENTIFIER.
                                          type: string
                                        secretRef:
                                          description: The Secret to select from
                                          properties:
                                            name:
                                              description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?'
                                              type: string
                                            optional:
                                              description: Specify whether the Secret must be defined
                                              type: boolean
                                          type: object
                                      type: object
                                    type: array
                                  image:
                                    description: 'Docker image name. More info: https://kubernetes.io/docs/concepts/containers/images This field is optional to allow higher level config management to default or override container images in workload controllers like Deployments and StatefulSets.'
                                    type: string
                                  imagePullPolicy:
                                    description: 'Image pull policy. One of Always, Never, IfNotPresent. Defaults to Always if :latest tag is specified, or IfNotPresent otherwise. Cannot be updated. More info: https://kubernetes.io/docs/concepts/containers/images#updating-images'
                                    type: string
                                  lifecycle:
                                    description: Actions that the management system should take in response to container lifecycle events. Cannot be updated.
                                    properties:
                                      postStart:
                                        description: 'PostStart is called immediately after a container is created. If the handler fails, the container is terminated and restarted according to its restart policy. Other management of the container blocks until the hook completes. More info: https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks'
                                        properties:
                                          exec:
                                            description: One and only one of the following should be specified. Exec specifies the action to take.
                                            properties:
                                              command:
                                                description: Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
                                                items:
                                                  type: string
                                                type: array
                                            type: object
                                          httpGet:
                                            description: HTTPGet specifies the http request to perform.
                                            properties:
                                              host:
                                                description: Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.
                                                type: string
                                              httpHeaders:
                                                description: Custom headers to set in the request. HTTP allows repeated headers.
                                                items:
                                                  description: HTTPHeader describes a custom header to be used in HTTP probes
                                                  properties:
                                                    name:
                                                      description: The header field name
                                                      type: string
                                                    value:
                                                      description: The header field value
                                                      type: string
                                                  required:
                                                  - name
                                                  - value
                                                  type: object
                                                type: array
                                              path:
                                                description: Path to access on the HTTP server.
                                                type: string
                                              port:
                                                description: Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                                x-kubernetes-int-or-string: true
                                              scheme:
                                                description: Scheme to use for connecting to the host. Defaults to HTTP.
                                                type: string
                                            required:
                                            - port
                                            type: object
                                          tcpSocket:
                                            description: 'TCPSocket specifies an action involving a TCP port. TCP hooks not yet supported TODO: implement a realistic TCP lifecycle hook'
                                            properties:
                                              host:
                                                description: 'Optional: Host name to connect to, defaults to the pod IP.'
                                                type: string
                                              port:
                                                description: Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                                x-kubernetes-int-or-string: true
                                            required:
                                            - port
                                            type: object
                                        type: object
                                      preStop:
                                        description: 'PreStop is called immediately before a container is terminated due to an API request or management event such as liveness probe failure, preemption, resource contention, etc. The handler is not called if the container crashes or exits. The reason for termination is passed to the handler. The Pod''s termination grace period countdown begins before the PreStop hooked is executed. Regardless of the outcome of the handler, the container will eventually terminate within the Pod''s termination grace period. Other management of the container blocks until the hook completes or until the termination grace period is reached. More info: https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks'
                                        properties:
                                          exec:
                                            description: One and only one of the following should be specified. Exec specifies the action to take.
                                            properties:
                                              command:
                                                description: Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
                                                items:
                                                  type: string
                                                type: array
                                            type: object
                                          httpGet:
                                            description: HTTPGet specifies the http request to perform.
                                            properties:
                                              host:
                                                description: Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.
                                                type: string
                                              httpHeaders:
                                                description: Custom headers to set in the request. HTTP allows repeated headers.
                                                items:
                                                  description: HTTPHeader describes a custom header to be used in HTTP probes
                                                  properties:
                                                    name:
                                                      description: The header field name
                                                      type: string
                                                    value:
                                                      description: The header field value
                                                      type: string
                                                  required:
                                                  - name
                                                  - value
                                                  type: object
                                                type: array
                                              path:
                                                description: Path to access on the HTTP server.
                                                type: string
                                              port:
                                                description: Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                                x-kubernetes-int-or-string: true
                                              scheme:
                                                description: Scheme to use for connecting to the host. Defaults to HTTP.
                                                type: string
                                            required:
                                            - port
                                            type: object
                                          tcpSocket:
                                            description: 'TCPSocket specifies an action involving a TCP port. TCP hooks not yet supported TODO: implement a realistic TCP lifecycle hook'
                                            properties:
                                              host:
                                                description: 'Optional: Host name to connect to, defaults to the pod IP.'
                                                type: string
                                              port:
                                                description: Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                                x-kubernetes-int-or-string: true
                                            required:
                                            - port
                                            type: object
                                        type: object
                                    type: object
                                  livenessProbe:
                                    description: 'Periodic probe of container liveness. Container will be restarted if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                    properties:
                                      exec:
                                        description: One and only one of the following should be specified. Exec specifies the action to take.
                                        properties:
                                          command:
                                            description: Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
                                            items:
                                              type: string
                                            type: array
                                        type: object
                                      failureThreshold:
                                        description: Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3. Minimum value is 1.
                                        format: int32
                                        type: integer
                                      httpGet:
                                        description: HTTPGet specifies the http request to perform.
                                        properties:
                                          host:
                                            description: Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.
                                            type: string
                                          httpHeaders:
                                            description: Custom headers to set in the request. HTTP allows repeated headers.
                                            items:
                                              description: HTTPHeader describes a custom header to be used in HTTP probes
                                              properties:
                                                name:
                                                  description: The header field name
                                                  type: string
                                                value:
                                                  description: The header field value
                                                  type: string
                                              required:
                                              - name
                                              - value
                                              type: object
                                            type: array
                                          path:
                                            description: Path to access on the HTTP server.
                                            type: string
                                          port:
                                            description: Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                            x-kubernetes-int-or-string: true
                                          scheme:
                                            description: Scheme to use for connecting to the host. Defaults to HTTP.
                                            type: string
                                        required:
                                        - port
                                        type: object
                                      initialDelaySeconds:
                                        description: 'Number of seconds after the container has started before liveness probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                        format: int32
                                        type: integer
                                      periodSeconds:
                                        description: How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1.
                                        format: int32
                                        type: integer
                                      successThreshold:
                                        description: Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1. Must be 1 for liveness. Minimum value is 1.
                                        format: int32
                                        type: integer
                                      tcpSocket:
                                        description: 'TCPSocket specifies an action involving a TCP port. TCP hooks not yet supported TODO: implement a realistic TCP lifecycle hook'
                                        properties:
                                          host:
                                            description: 'Optional: Host name to connect to, defaults to the pod IP.'
                                            type: string
                                          port:
                                            description: Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                            x-kubernetes-int-or-string: true
                                        required:
                                        - port
                                        type: object
                                      timeoutSeconds:
                                        description: 'Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                        format: int32
                                        type: integer
                                    type: object
                                  name:
                                    description: Name of the container specified as a DNS_LABEL. Each container in a pod must have a unique name (DNS_LABEL). Cannot be updated.
                                    type: string
                                  ports:
                                    description: List of ports to expose from the container. Exposing a port here gives the system additional information about the network connections a container uses, but is primarily informational. Not specifying a port here DOES NOT prevent that port from being exposed. Any port which is listening on the default "0.0.0.0" address inside a container will be accessible from the network. Cannot be updated.
                                    items:
                                      description: ContainerPort represents a network port in a single container.
                                      properties:
                                        containerPort:
                                          description: Number of port to expose on the pod's IP address. This must be a valid port number, 0 < x < 65536.
                                          format: int32
                                          type: integer
                                        hostIP:
                                          description: What host IP to bind the external port to.
                                          type: string
                                        hostPort:
                                          description: Number of port to expose on the host. If specified, this must be a valid port number, 0 < x < 65536. If HostNetwork is specified, this must match ContainerPort. Most containers do not need this.
                                          format: int32
                                          type: integer
                                        name:
                                          description: If specified, this must be an IANA_SVC_NAME and unique within the pod. Each named port in a pod must have a unique name. Name for the port that can be referred to by services.
                                          type: string
                                        protocol:
                                          description: Protocol for port. Must be UDP, TCP, or SCTP. Defaults to "TCP".
                                          type: string
                                      required:
                                      - containerPort
                                      type: object
                                    type: array
                                  readinessProbe:
                                    description: 'Periodic probe of container service readiness. Container will be removed from service endpoints if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                    properties:
                                      exec:
                                        description: One and only one of the following should be specified. Exec specifies the action to take.
                                        properties:
                                          command:
                                            description: Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
                                            items:
                                              type: string
                                            type: array
                                        type: object
                                      failureThreshold:
                                        description: Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3. Minimum value is 1.
                                        format: int32
                                        type: integer
                                      httpGet:
                                        description: HTTPGet specifies the http request to perform.
                                        properties:
                                          host:
                                            description: Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.
                                            type: string
                                          httpHeaders:
                                            description: Custom headers to set in the request. HTTP allows repeated headers.
                                            items:
                                              description: HTTPHeader describes a custom header to be used in HTTP probes
                                              properties:
                                                name:
                                                  description: The header field name
                                                  type: string
                                                value:
                                                  description: The header field value
                                                  type: string
                                              required:
                                              - name
                                              - value
                                              type: object
                                            type: array
                                          path:
                                            description: Path to access on the HTTP server.
                                            type: string
                                          port:
                                            description: Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                            x-kubernetes-int-or-string: true
                                          scheme:
                                            description: Scheme to use for connecting to the host. Defaults to HTTP.
                                            type: string
                                        required:
                                        - port
                                        type: object
                                      initialDelaySeconds:
                                        description: 'Number of seconds after the container has started before liveness probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                        format: int32
                                        type: integer
                                      periodSeconds:
                                        description: How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1.
                                        format: int32
                                        type: integer
                                      successThreshold:
                                        description: Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1. Must be 1 for liveness. Minimum value is 1.
                                        format: int32
                                        type: integer
                                      tcpSocket:
                                        description: 'TCPSocket specifies an action involving a TCP port. TCP hooks not yet supported TODO: implement a realistic TCP lifecycle hook'
                                        properties:
                                          host:
                                            description: 'Optional: Host name to connect to, defaults to the pod IP.'
                                            type: string
                                          port:
                                            description: Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                            x-kubernetes-int-or-string: true
                                        required:
                                        - port
                                        type: object
                                      timeoutSeconds:
                                        description: 'Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                        format: int32
                                        type: integer
                                    type: object
                                  resources:
                                    description: 'Compute Resources required by this container. Cannot be updated. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                                    properties:
                                      limits:
                                        additionalProperties:
                                          type: string
                                        description: 'Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                                        type: object
                                      requests:
                                        additionalProperties:
                                          type: string
                                        description: 'Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                                        type: object
                                    type: object
                                  securityContext:
                                    description: 'Security options the pod should run with. More info: https://kubernetes.io/docs/concepts/policy/security-context/ More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/'
                                    properties:
                                      allowPrivilegeEscalation:
                                        description: 'AllowPrivilegeEscalation controls whether a process can gain more privileges than its parent process. This bool directly controls if the no_new_privs flag will be set on the container process. AllowPrivilegeEscalation is true always when the container is: 1) run as Privileged 2) has CAP_SYS_ADMIN'
                                        type: boolean
                                      capabilities:
                                        description: The capabilities to add/drop when running containers. Defaults to the default set of capabilities granted by the container runtime.
                                        properties:
                                          add:
                                            description: Added capabilities
                                            items:
                                              description: Capability represent POSIX capabilities type
                                              type: string
                                            type: array
                                          drop:
                                            description: Removed capabilities
                                            items:
                                              description: Capability represent POSIX capabilities type
                                              type: string
                                            type: array
                                        type: object
                                      privileged:
                                        description: Run container in privileged mode. Processes in privileged containers are essentially equivalent to root on the host. Defaults to false.
                                        type: boolean
                                      procMount:
                                        description: procMount denotes the type of proc mount to use for the containers. The default is DefaultProcMount which uses the container runtime defaults for readonly paths and masked paths. This requires the ProcMountType feature flag to be enabled.
                                        type: string
                                      readOnlyRootFilesystem:
                                        description: Whether this container has a read-only root filesystem. Default is false.
                                        type: boolean
                                      runAsGroup:
                                        description: The GID to run the entrypoint of the container process. Uses runtime default if unset. May also be set in PodSecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.
                                        format: int64
                                        type: integer
                                      runAsNonRoot:
                                        description: Indicates that the container must run as a non-root user. If true, the Kubelet will validate the image at runtime to ensure that it does not run as UID 0 (root) and fail to start the container if it does. If unset or false, no such validation will be performed. May also be set in PodSecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.
                                        type: boolean
                                      runAsUser:
                                        description: The UID to run the entrypoint of the container process. Defaults to user specified in image metadata if unspecified. May also be set in PodSecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.
                                        format: int64
                                        type: integer
                                      seLinuxOptions:
                                        description: The SELinux context to be applied to the container. If unspecified, the container runtime will allocate a random SELinux context for each container.  May also be set in PodSecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.
                                        properties:
                                          level:
                                            description: Level is SELinux level label that applies to the container.
                                            type: string
                                          role:
                                            description: Role is a SELinux role label that applies to the container.
                                            type: string
                                          type:
                                            description: Type is a SELinux type label that applies to the container.
                                            type: string
                                          user:
                                            description: User is a SELinux user label that applies to the container.
                                            type: string
                                        type: object
                                    type: object
                                  stdin:
                                    description: Whether this container should allocate a buffer for stdin in the container runtime. If this is not set, reads from stdin in the container will always result in EOF. Default is false.
                                    type: boolean
                                  stdinOnce:
                                    description: Whether the container runtime should close the stdin channel after it has been opened by a single attach. When stdin is true the stdin stream will remain open across multiple attach sessions. If stdinOnce is set to true, stdin is opened on container start, is empty until the first client attaches to stdin, and then remains open and accepts data until the client disconnects, at which time stdin is closed and remains closed until the container is restarted. If this flag is false, a container processes that reads from stdin will never receive an EOF. Default is false
                                    type: boolean
                                  terminationMessagePath:
                                    description: 'Optional: Path at which the file to which the container''s termination message will be written is mounted into the container''s filesystem. Message written is intended to be brief final status, such as an assertion failure message. Will be truncated by the node if greater than 4096 bytes. The total message length across all containers will be limited to 12kb. Defaults to /dev/termination-log. Cannot be updated.'
                                    type: string
                                  terminationMessagePolicy:
                                    description: Indicate how the termination message should be populated. File will use the contents of terminationMessagePath to populate the container status message on both success and failure. FallbackToLogsOnError will use the last chunk of container log output if the termination message file is empty and the container exited with an error. The log output is limited to 2048 bytes or 80 lines, whichever is smaller. Defaults to File. Cannot be updated.
                                    type: string
                                  tty:
                                    description: Whether this container should allocate a TTY for itself, also requires 'stdin' to be true. Default is false.
                                    type: boolean
                                  volumeDevices:
                                    description: volumeDevices is the list of block devices to be used by the container. This is a beta feature.
                                    items:
                                      description: volumeDevice describes a mapping of a raw block device within a container.
                                      properties:
                                        devicePath:
                                          description: devicePath is the path inside of the container that the device will be mapped to.
                                          type: string
                                        name:
                                          description: name must match the name of a persistentVolumeClaim in the pod
                                          type: string
                                      required:
                                      - devicePath
                                      - name
                                      type: object
                                    type: array
                                  volumeMounts:
                                    description: Pod volumes to mount into the container's filesystem. Cannot be updated.
                                    items:
                                      description: VolumeMount describes a mounting of a Volume within a container.
                                      properties:
                                        mountPath:
                                          description: Path within the container at which the volume should be mounted.  Must not contain ':'.
                                          type: string
                                        mountPropagation:
                                          description: mountPropagation determines how mounts are propagated from the host to container and the other way around. When not set, MountPropagationNone is used. This field is beta in 1.10.
                                          type: string
                                        name:
                                          description: This must match the Name of a Volume.
                                          type: string
                                        readOnly:
                                          description: Mounted read-only if true, read-write otherwise (false or unspecified). Defaults to false.
                                          type: boolean
                                        subPath:
                                          description: Path within the volume from which the container's volume should be mounted. Defaults to "" (volume's root).
                                          type: string
                                        subPathExpr:
                                          description: Expanded path within the volume from which the container's volume should be mounted. Behaves similarly to SubPath but environment variable references $(VAR_NAME) are expanded using the container's environment. Defaults to "" (volume's root). SubPathExpr and SubPath are mutually exclusive. This field is alpha in 1.14.
                                          type: string
                                      required:
                                      - mountPath
                                      - name
                                      type: object
                                    type: array
                                  workingDir:
                                    description: Container's working directory. If not specified, the container runtime's default will be used, which might be configured in the container image. Cannot be updated.
                                    type: string
                                required:
                                - name
                                type: object
                              type: array
                            dnsConfig:
                              description: Specifies the DNS parameters of a pod. Parameters specified here will be merged to the generated DNS configuration based on DNSPolicy.
                              properties:
                                nameservers:
                                  description: A list of DNS name server IP addresses. This will be appended to the base nameservers generated from DNSPolicy. Duplicated nameservers will be removed.
                                  items:
                                    type: string
                                  type: array
                                options:
                                  description: A list of DNS resolver options. This will be merged with the base options generated from DNSPolicy. Duplicated entries will be removed. Resolution options given in Options will override those that appear in the base DNSPolicy.
                                  items:
                                    description: PodDNSConfigOption defines DNS resolver options of a pod.
                                    properties:
                                      name:
                                        description: Required.
                                        type: string
                                      value:
                                        type: string
                                    type: object
                                  type: array
                                searches:
                                  description: A list of DNS search domains for host-name lookup. This will be appended to the base search paths generated from DNSPolicy. Duplicated search paths will be removed.
                                  items:
                                    type: string
                                  type: array
                              type: object
                            dnsPolicy:
                              description: Set DNS policy for the pod. Defaults to "ClusterFirst". Valid values are 'ClusterFirstWithHostNet', 'ClusterFirst', 'Default' or 'None'. DNS parameters given in DNSConfig will be merged with the policy selected with DNSPolicy. To have DNS options set along with hostNetwork, you have to specify DNS policy explicitly to 'ClusterFirstWithHostNet'.
                              type: string
                            enableServiceLinks:
                              description: 'EnableServiceLinks indicates whether information about services should be injected into pod''s environment variables, matching the syntax of Docker links. Optional: Defaults to true.'
                              type: boolean
                            hostAliases:
                              description: HostAliases is an optional list of hosts and IPs that will be injected into the pod's hosts file if specified. This is only valid for non-hostNetwork pods.
                              items:
                                description: HostAlias holds the mapping between IP and hostnames that will be injected as an entry in the pod's hosts file.
                                properties:
                                  hostnames:
                                    description: Hostnames for the above IP address.
                                    items:
                                      type: string
                                    type: array
                                  ip:
                                    description: IP address of the host file entry.
                                    type: string
                                type: object
                              type: array
                            hostIPC:
                              description: 'Use the host''s ipc namespace. Optional: Default to false.'
                              type: boolean
                            hostNetwork:
                              description: Host networking requested for this pod. Use the host's network namespace. If this option is set, the ports that will be used must be specified. Default to false.
                              type: boolean
                            hostPID:
                              description: 'Use the host''s pid namespace. Optional: Default to false.'
                              type: boolean
                            hostname:
                              description: Specifies the hostname of the Pod If not specified, the pod's hostname will be set to a system-defined value.
                              type: string
                            imagePullSecrets:
                              description: 'ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec. If specified, these secrets will be passed to individual puller implementations for them to use. For example, in the case of docker, only DockerConfig type secrets are honored. More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod'
                              items:
                                description: LocalObjectReference contains enough information to let you locate the referenced object inside the same namespace.
                                properties:
                                  name:
                                    description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?'
                                    type: string
                                type: object
                              type: array
                            initContainers:
                              description: 'List of initialization containers belonging to the pod. Init containers are executed in order prior to containers being started. If any init container fails, the pod is considered to have failed and is handled according to its restartPolicy. The name for an init container or normal container must be unique among all containers. Init containers may not have Lifecycle actions, Readiness probes, or Liveness probes. The resourceRequirements of an init container are taken into account during scheduling by finding the highest request/limit for each resource type, and then using the max of of that value or the sum of the normal containers. Limits are applied to init containers in a similar fashion. Init containers cannot currently be added or removed. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/'
                              items:
                                description: A single application container that you want to run within a pod.
                                properties:
                                  args:
                                    description: 'Arguments to the entrypoint. The docker image''s CMD is used if this is not provided. Variable references $(VAR_NAME) are expanded using the container''s environment. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell'
                                    items:
                                      type: string
                                    type: array
                                  command:
                                    description: 'Entrypoint array. Not executed within a shell. The docker image''s ENTRYPOINT is used if this is not provided. Variable references $(VAR_NAME) are expanded using the container''s environment. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell'
                                    items:
                                      type: string
                                    type: array
                                  env:
                                    description: List of environment variables to set in the container. Cannot be updated.
                                    items:
                                      description: EnvVar represents an environment variable present in a Container.
                                      properties:
                                        name:
                                          description: Name of the environment variable. Must be a C_IDENTIFIER.
                                          type: string
                                        value:
                                          description: 'Variable references $(VAR_NAME) are expanded using the previous defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to "".'
                                          type: string
                                        valueFrom:
                                          description: Source for the environment variable's value. Cannot be used if value is not empty.
                                          properties:
                                            configMapKeyRef:
                                              description: Selects a key of a ConfigMap.
                                              properties:
                                                key:
                                                  description: The key to select.
                                                  type: string
                                                name:
                                                  description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?'
                                                  type: string
                                                optional:
                                                  description: Specify whether the ConfigMap or it's key must be defined
                                                  type: boolean
                                              required:
                                              - key
                                              type: object
                                            fieldRef:
                                              description: 'Selects a field of the pod: supports metadata.name, metadata.namespace, metadata.labels, metadata.annotations, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP.'
                                              properties:
                                                apiVersion:
                                                  description: Version of the schema the FieldPath is written in terms of, defaults to "v1".
                                                  type: string
                                                fieldPath:
                                                  description: Path of the field to select in the specified API version.
                                                  type: string
                                              required:
                                              - fieldPath
                                              type: object
                                            resourceFieldRef:
                                              description: 'Selects a resource of the container: only resources limits and requests (limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.'
                                              properties:
                                                containerName:
                                                  description: 'Container name: required for volumes, optional for env vars'
                                                  type: string
                                                divisor:
                                                  description: Specifies the output format of the exposed resources, defaults to "1"
                                                  type: string
                                                resource:
                                                  description: 'Required: resource to select'
                                                  type: string
                                              required:
                                              - resource
                                              type: object
                                            secretKeyRef:
                                              description: Selects a key of a secret in the pod's namespace
                                              properties:
                                                key:
                                                  description: The key of the secret to select from.  Must be a valid secret key.
                                                  type: string
                                                name:
                                                  description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?'
                                                  type: string
                                                optional:
                                                  description: Specify whether the Secret or it's key must be defined
                                                  type: boolean
                                              required:
                                              - key
                                              type: object
                                          type: object
                                      required:
                                      - name
                                      type: object
                                    type: array
                                  envFrom:
                                    description: List of sources to populate environment variables in the container. The keys defined within a source must be a C_IDENTIFIER. All invalid keys will be reported as an event when the container is starting. When a key exists in multiple sources, the value associated with the last source will take precedence. Values defined by an Env with a duplicate key will take precedence. Cannot be updated.
                                    items:
                                      description: EnvFromSource represents the source of a set of ConfigMaps
                                      properties:
                                        configMapRef:
                                          description: The ConfigMap to select from
                                          properties:
                                            name:
                                              description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?'
                                              type: string
                                            optional:
                                              description: Specify whether the ConfigMap must be defined
                                              type: boolean
                                          type: object
                                        prefix:
                                          description: An optional identifier to prepend to each key in the ConfigMap. Must be a C_IDENTIFIER.
                                          type: string
                                        secretRef:
                                          description: The Secret to select from
                                          properties:
                                            name:
                                              description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?'
                                              type: string
                                            optional:
                                              description: Specify whether the Secret must be defined
                                              type: boolean
                                          type: object
                                      type: object
                                    type: array
                                  image:
                                    description: 'Docker image name. More info: https://kubernetes.io/docs/concepts/containers/images This field is optional to allow higher level config management to default or override container images in workload controllers like Deployments and StatefulSets.'
                                    type: string
                                  imagePullPolicy:
                                    description: 'Image pull policy. One of Always, Never, IfNotPresent. Defaults to Always if :latest tag is specified, or IfNotPresent otherwise. Cannot be updated. More info: https://kubernetes.io/docs/concepts/containers/images#updating-images'
                                    type: string
                                  lifecycle:
                                    description: Actions that the management system should take in response to container lifecycle events. Cannot be updated.
                                    properties:
                                      postStart:
                                        description: 'PostStart is called immediately after a container is created. If the handler fails, the container is terminated and restarted according to its restart policy. Other management of the container blocks until the hook completes. More info: https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks'
                                        properties:
                                          exec:
                                            description: One and only one of the following should be specified. Exec specifies the action to take.
                                            properties:
                                              command:
                                                description: Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
                                                items:
                                                  type: string
                                                type: array
                                            type: object
                                          httpGet:
                                            description: HTTPGet specifies the http request to perform.
                                            properties:
                                              host:
                                                description: Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.
                                                type: string
                                              httpHeaders:
                                                description: Custom headers to set in the request. HTTP allows repeated headers.
                                                items:
                                                  description: HTTPHeader describes a custom header to be used in HTTP probes
                                                  properties:
                                                    name:
                                                      description: The header field name
                                                      type: string
                                                    value:
                                                      description: The header field value
                                                      type: string
                                                  required:
                                                  - name
                                                  - value
                                                  type: object
                                                type: array
                                              path:
                                                description: Path to access on the HTTP server.
                                                type: string
                                              port:
                                                description: Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                                x-kubernetes-int-or-string: true
                                              scheme:
                                                description: Scheme to use for connecting to the host. Defaults to HTTP.
                                                type: string
                                            required:
                                            - port
                                            type: object
                                          tcpSocket:
                                            description: 'TCPSocket specifies an action involving a TCP port. TCP hooks not yet supported TODO: implement a realistic TCP lifecycle hook'
                                            properties:
                                              host:
                                                description: 'Optional: Host name to connect to, defaults to the pod IP.'
                                                type: string
                                              port:
                                                description: Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                                x-kubernetes-int-or-string: true
                                            required:
                                            - port
                                            type: object
                                        type: object
                                      preStop:
                                        description: 'PreStop is called immediately before a container is terminated due to an API request or management event such as liveness probe failure, preemption, resource contention, etc. The handler is not called if the container crashes or exits. The reason for termination is passed to the handler. The Pod''s termination grace period countdown begins before the PreStop hooked is executed. Regardless of the outcome of the handler, the container will eventually terminate within the Pod''s termination grace period. Other management of the container blocks until the hook completes or until the termination grace period is reached. More info: https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks'
                                        properties:
                                          exec:
                                            description: One and only one of the following should be specified. Exec specifies the action to take.
                                            properties:
                                              command:
                                                description: Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
                                                items:
                                                  type: string
                                                type: array
                                            type: object
                                          httpGet:
                                            description: HTTPGet specifies the http request to perform.
                                            properties:
                                              host:
                                                description: Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.
                                                type: string
                                              httpHeaders:
                                                description: Custom headers to set in the request. HTTP allows repeated headers.
                                                items:
                                                  description: HTTPHeader describes a custom header to be used in HTTP probes
                                                  properties:
                                                    name:
                                                      description: The header field name
                                                      type: string
                                                    value:
                                                      description: The header field value
                                                      type: string
                                                  required:
                                                  - name
                                                  - value
                                                  type: object
                                                type: array
                                              path:
                                                description: Path to access on the HTTP server.
                                                type: string
                                              port:
                                                description: Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                                x-kubernetes-int-or-string: true
                                              scheme:
                                                description: Scheme to use for connecting to the host. Defaults to HTTP.
                                                type: string
                                            required:
                                            - port
                                            type: object
                                          tcpSocket:
                                            description: 'TCPSocket specifies an action involving a TCP port. TCP hooks not yet supported TODO: implement a realistic TCP lifecycle hook'
                                            properties:
                                              host:
                                                description: 'Optional: Host name to connect to, defaults to the pod IP.'
                                                type: string
                                              port:
                                                description: Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                                x-kubernetes-int-or-string: true
                                            required:
                                            - port
                                            type: object
                                        type: object
                                    type: object
                                  livenessProbe:
                                    description: 'Periodic probe of container liveness. Container will be restarted if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                    properties:
                                      exec:
                                        description: One and only one of the following should be specified. Exec specifies the action to take.
                                        properties:
                                          command:
                                            description: Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
                                            items:
                                              type: string
                                            type: array
                                        type: object
                                      failureThreshold:
                                        description: Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3. Minimum value is 1.
                                        format: int32
                                        type: integer
                                      httpGet:
                                        description: HTTPGet specifies the http request to perform.
                                        properties:
                                          host:
                                            description: Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.
                                            type: string
                                          httpHeaders:
                                            description: Custom headers to set in the request. HTTP allows repeated headers.
                                            items:
                                              description: HTTPHeader describes a custom header to be used in HTTP probes
                                              properties:
                                                name:
                                                  description: The header field name
                                                  type: string
                                                value:
                                                  description: The header field value
                                                  type: string
                                              required:
                                              - name
                                              - value
                                              type: object
                                            type: array
                                          path:
                                            description: Path to access on the HTTP server.
                                            type: string
                                          port:
                                            description: Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                            x-kubernetes-int-or-string: true
                                          scheme:
                                            description: Scheme to use for connecting to the host. Defaults to HTTP.
                                            type: string
                                        required:
                                        - port
                                        type: object
                                      initialDelaySeconds:
                                        description: 'Number of seconds after the container has started before liveness probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                        format: int32
                                        type: integer
                                      periodSeconds:
                                        description: How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1.
                                        format: int32
                                        type: integer
                                      successThreshold:
                                        description: Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1. Must be 1 for liveness. Minimum value is 1.
                                        format: int32
                                        type: integer
                                      tcpSocket:
                                        description: 'TCPSocket specifies an action involving a TCP port. TCP hooks not yet supported TODO: implement a realistic TCP lifecycle hook'
                                        properties:
                                          host:
                                            description: 'Optional: Host name to connect to, defaults to the pod IP.'
                                            type: string
                                          port:
                                            description: Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                            x-kubernetes-int-or-string: true
                                        required:
                                        - port
                                        type: object
                                      timeoutSeconds:
                                        description: 'Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                        format: int32
                                        type: integer
                                    type: object
                                  name:
                                    description: Name of the container specified as a DNS_LABEL. Each container in a pod must have a unique name (DNS_LABEL). Cannot be updated.
                                    type: string
                                  ports:
                                    description: List of ports to expose from the container. Exposing a port here gives the system additional information about the network connections a container uses, but is primarily informational. Not specifying a port here DOES NOT prevent that port from being exposed. Any port which is listening on the default "0.0.0.0" address inside a container will be accessible from the network. Cannot be updated.
                                    items:
                                      description: ContainerPort represents a network port in a single container.
                                      properties:
                                        containerPort:
                                          description: Number of port to expose on the pod's IP address. This must be a valid port number, 0 < x < 65536.
                                          format: int32
                                          type: integer
                                        hostIP:
                                          description: What host IP to bind the external port to.
                                          type: string
                                        hostPort:
                                          description: Number of port to expose on the host. If specified, this must be a valid port number, 0 < x < 65536. If HostNetwork is specified, this must match ContainerPort. Most containers do not need this.
                                          format: int32
                                          type: integer
                                        name:
                                          description: If specified, this must be an IANA_SVC_NAME and unique within the pod. Each named port in a pod must have a unique name. Name for the port that can be referred to by services.
                                          type: string
                                        protocol:
                                          description: Protocol for port. Must be UDP, TCP, or SCTP. Defaults to "TCP".
                                          type: string
                                      required:
                                      - containerPort
                                      type: object
                                    type: array
                                  readinessProbe:
                                    description: 'Periodic probe of container service readiness. Container will be removed from service endpoints if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                    properties:
                                      exec:
                                        description: One and only one of the following should be specified. Exec specifies the action to take.
                                        properties:
                                          command:
                                            description: Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
                                            items:
                                              type: string
                                            type: array
                                        type: object
                                      failureThreshold:
                                        description: Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3. Minimum value is 1.
                                        format: int32
                                        type: integer
                                      httpGet:
                                        description: HTTPGet specifies the http request to perform.
                                        properties:
                                          host:
                                            description: Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.
                                            type: string
                                          httpHeaders:
                                            description: Custom headers to set in the request. HTTP allows repeated headers.
                                            items:
                                              description: HTTPHeader describes a custom header to be used in HTTP probes
                                              properties:
                                                name:
                                                  description: The header field name
                                                  type: string
                                                value:
                                                  description: The header field value
                                                  type: string
                                              required:
                                              - name
                                              - value
                                              type: object
                                            type: array
                                          path:
                                            description: Path to access on the HTTP server.
                                            type: string
                                          port:
                                            description: Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                            x-kubernetes-int-or-string: true
                                          scheme:
                                            description: Scheme to use for connecting to the host. Defaults to HTTP.
                                            type: string
                                        required:
                                        - port
                                        type: object
                                      initialDelaySeconds:
                                        description: 'Number of seconds after the container has started before liveness probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                        format: int32
                                        type: integer
                                      periodSeconds:
                                        description: How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1.
                                        format: int32
                                        type: integer
                                      successThreshold:
                                        description: Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1. Must be 1 for liveness. Minimum value is 1.
                                        format: int32
                                        type: integer
                                      tcpSocket:
                                        description: 'TCPSocket specifies an action involving a TCP port. TCP hooks not yet supported TODO: implement a realistic TCP lifecycle hook'
                                        properties:
                                          host:
                                            description: 'Optional: Host name to connect to, defaults to the pod IP.'
                                            type: string
                                          port:
                                            description: Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                            x-kubernetes-int-or-string: true
                                        required:
                                        - port
                                        type: object
                                      timeoutSeconds:
                                        description: 'Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                        format: int32
                                        type: integer
                                    type: object
                                  resources:
                                    description: 'Compute Resources required by this container. Cannot be updated. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                                    properties:
                                      limits:
                                        additionalProperties:
                                          type: string
                                        description: 'Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                                        type: object
                                      requests:
                                        additionalProperties:
                                          type: string
                                        description: 'Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                                        type: object
                                    type: object
                                  securityContext:
                                    description: 'Security options the pod should run with. More info: https://kubernetes.io/docs/concepts/policy/security-context/ More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/'
                                    properties:
                                      allowPrivilegeEscalation:
                                        description: 'AllowPrivilegeEscalation controls whether a process can gain more privileges than its parent process. This bool directly controls if the no_new_privs flag will be set on the container process. AllowPrivilegeEscalation is true always when the container is: 1) run as Privileged 2) has CAP_SYS_ADMIN'
                                        type: boolean
                                      capabilities:
                                        description: The capabilities to add/drop when running containers. Defaults to the default set of capabilities granted by the container runtime.
                                        properties:
                                          add:
                                            description: Added capabilities
                                            items:
                                              description: Capability represent POSIX capabilities type
                                              type: string
                                            type: array
                                          drop:
                                            description: Removed capabilities
                                            items:
                                              description: Capability represent POSIX capabilities type
                                              type: string
                                            type: array
                                        type: object
                                      privileged:
                                        description: Run container in privileged mode. Processes in privileged containers are essentially equivalent to root on the host. Defaults to false.
                                        type: boolean
                                      procMount:
                                        description: procMount denotes the type of proc mount to use for the containers. The default is DefaultProcMount which uses the container runtime defaults for readonly paths and masked paths. This requires the ProcMountType feature flag to be enabled.
                                        type: string
                                      readOnlyRootFilesystem:
                                        description: Whether this container has a read-only root filesystem. Default is false.
                                        type: boolean
                                      runAsGroup:
                                        description: The GID to run the entrypoint of the container process. Uses runtime default if unset. May also be set in PodSecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.
                                        format: int64
                                        type: integer
                                      runAsNonRoot:
                                        description: Indicates that the container must run as a non-root user. If true, the Kubelet will validate the image at runtime to ensure that it does not run as UID 0 (root) and fail to start the container if it does. If unset or false, no such validation will be performed. May also be set in PodSecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.
                                        type: boolean
                                      runAsUser:
                                        description: The UID to run the entrypoint of the container process. Defaults to user specified in image metadata if unspecified. May also be set in PodSecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.
                                        format: int64
                                        type: integer
                                      seLinuxOptions:
                                        description: The SELinux context to be applied to the container. If unspecified, the container runtime will allocate a random SELinux context for each container.  May also be set in PodSecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.
                                        properties:
                                          level:
                                            description: Level is SELinux level label that applies to the container.
                                            type: string
                                          role:
                                            description: Role is a SELinux role label that applies to the container.
                                            type: string
                                          type:
                                            description: Type is a SELinux type label that applies to the container.
                                            type: string
                                          user:
                                            description: User is a SELinux user label that applies to the container.
                                            type: string
                                        type: object
                                    type: object
                                  stdin:
                                    description: Whether this container should allocate a buffer for stdin in the container runtime. If this is not set, reads from stdin in the container will always result in EOF. Default is false.
                                    type: boolean
                                  stdinOnce:
                                    description: Whether the container runtime should close the stdin channel after it has been opened by a single attach. When stdin is true the stdin stream will remain open across multiple attach sessions. If stdinOnce is set to true, stdin is opened on container start, is empty until the first client attaches to stdin, and then remains open and accepts data until the client disconnects, at which time stdin is closed and remains closed until the container is restarted. If this flag is false, a container processes that reads from stdin will never receive an EOF. Default is false
                                    type: boolean
                                  terminationMessagePath:
                                    description: 'Optional: Path at which the file to which the container''s termination message will be written is mounted into the container''s filesystem. Message written is intended to be brief final status, such as an assertion failure message. Will be truncated by the node if greater than 4096 bytes. The total message length across all containers will be limited to 12kb. Defaults to /dev/termination-log. Cannot be updated.'
                                    type: string
                                  terminationMessagePolicy:
                                    description: Indicate how the termination message should be populated. File will use the contents of terminationMessagePath to populate the container status message on both success and failure. FallbackToLogsOnError will use the last chunk of container log output if the termination message file is empty and the container exited with an error. The log output is limited to 2048 bytes or 80 lines, whichever is smaller. Defaults to File. Cannot be updated.
                                    type: string
                                  tty:
                                    description: Whether this container should allocate a TTY for itself, also requires 'stdin' to be true. Default is false.
                                    type: boolean
                                  volumeDevices:
                                    description: volumeDevices is the list of block devices to be used by the container. This is a beta feature.
                                    items:
                                      description: volumeDevice describes a mapping of a raw block device within a container.
                                      properties:
                                        devicePath:
                                          description: devicePath is the path inside of the container that the device will be mapped to.
                                          type: string
                                        name:
                                          description: name must match the name of a persistentVolumeClaim in the pod
                                          type: string
                                      required:
                                      - devicePath
                                      - name
                                      type: object
                                    type: array
                                  volumeMounts:
                                    description: Pod volumes to mount into the container's filesystem. Cannot be updated.
                                    items:
                                      description: VolumeMount describes a mounting of a Volume within a container.
                                      properties:
                                        mountPath:
                                          description: Path within the container at which the volume should be mounted.  Must not contain ':'.
                                          type: string
                                        mountPropagation:
                                          description: mountPropagation determines how mounts are propagated from the host to container and the other way around. When not set, MountPropagationNone is used. This field is beta in 1.10.
                                          type: string
                                        name:
                                          description: This must match the Name of a Volume.
                                          type: string
                                        readOnly:
                                          description: Mounted read-only if true, read-write otherwise (false or unspecified). Defaults to false.
                                          type: boolean
                                        subPath:
                                          description: Path within the volume from which the container's volume should be mounted. Defaults to "" (volume's root).
                                          type: string
                                        subPathExpr:
                                          description: Expanded path within the volume from which the container's volume should be mounted. Behaves similarly to SubPath but environment variable references $(VAR_NAME) are expanded using the container's environment. Defaults to "" (volume's root). SubPathExpr and SubPath are mutually exclusive. This field is alpha in 1.14.
                                          type: string
                                      required:
                                      - mountPath
                                      - name
                                      type: object
                                    type: array
                                  workingDir:
                                    description: Container's working directory. If not specified, the container runtime's default will be used, which might be configured in the container image. Cannot be updated.
                                    type: string
                                required:
                                - name
                                type: object
                              type: array
                            nodeName:
                              description: NodeName is a request to schedule this pod onto a specific node. If it is non-empty, the scheduler simply schedules this pod onto that node, assuming that it fits resource requirements.
                              type: string
                            nodeSelector:
                              additionalProperties:
                                type: string
                              description: 'NodeSelector is a selector which must be true for the pod to fit on a node. Selector which must match a node''s labels for the pod to be scheduled on that node. More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/'
                              type: object
                            priority:
                              description: The priority value. Various system components use this field to find the priority of the pod. When Priority Admission Controller is enabled, it prevents users from setting this field. The admission controller populates this field from PriorityClassName. The higher the value, the higher the priority.
                              format: int32
                              type: integer
                            priorityClassName:
                              description: If specified, indicates the pod's priority. "system-node-critical" and "system-cluster-critical" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default.
                              type: string
                            readinessGates:
                              description: 'If specified, all readiness gates will be evaluated for pod readiness. A pod is ready when all its containers are ready AND all conditions specified in the readiness gates have status equal to "True" More info: https://git.k8s.io/enhancements/keps/sig-network/0007-pod-ready%2B%2B.md'
                              items:
                                description: PodReadinessGate contains the reference to a pod condition
                                properties:
                                  conditionType:
                                    description: ConditionType refers to a condition in the pod's condition list with matching type.
                                    type: string
                                required:
                                - conditionType
                                type: object
                              type: array
                            restartPolicy:
                              description: 'Restart policy for all containers within the pod. One of Always, OnFailure, Never. Default to Always. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy'
                              type: string
                            runtimeClassName:
                              description: 'RuntimeClassName refers to a RuntimeClass object in the node.k8s.io group, which should be used to run this pod.  If no RuntimeClass resource matches the named class, the pod will not be run. If unset or empty, the "legacy" RuntimeClass will be used, which is an implicit class with an empty definition that uses the default runtime handler. More info: https://git.k8s.io/enhancements/keps/sig-node/runtime-class.md This is an alpha feature and may change in the future.'
                              type: string
                            schedulerName:
                              description: If specified, the pod will be dispatched by specified scheduler. If not specified, the pod will be dispatched by default scheduler.
                              type: string
                            securityContext:
                              description: 'SecurityContext holds pod-level security attributes and common container settings. Optional: Defaults to empty.  See type description for default values of each field.'
                              properties:
                                fsGroup:
                                  description: "A special supplemental group that applies to all containers in a pod. Some volume types allow the Kubelet to change the ownership of that volume to be owned by the pod: \n 1. The owning GID will be the FSGroup 2. The setgid bit is set (new files created in the volume will be owned by FSGroup) 3. The permission bits are OR'd with rw-rw---- \n If unset, the Kubelet will not modify the ownership and permissions of any volume."
                                  format: int64
                                  type: integer
                                runAsGroup:
                                  description: The GID to run the entrypoint of the container process. Uses runtime default if unset. May also be set in SecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence for that container.
                                  format: int64
                                  type: integer
                                runAsNonRoot:
                                  description: Indicates that the container must run as a non-root user. If true, the Kubelet will validate the image at runtime to ensure that it does not run as UID 0 (root) and fail to start the container if it does. If unset or false, no such validation will be performed. May also be set in SecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.
                                  type: boolean
                                runAsUser:
                                  description: The UID to run the entrypoint of the container process. Defaults to user specified in image metadata if unspecified. May also be set in SecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence for that container.
                                  format: int64
                                  type: integer
                                seLinuxOptions:
                                  description: The SELinux context to be applied to all containers. If unspecified, the container runtime will allocate a random SELinux context for each container.  May also be set in SecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence for that container.
                                  properties:
                                    level:
                                      description: Level is SELinux level label that applies to the container.
                                      type: string
                                    role:
                                      description: Role is a SELinux role label that applies to the container.
                                      type: string
                                    type:
                                      description: Type is a SELinux type label that applies to the container.
                                      type: string
                                    user:
                                      description: User is a SELinux user label that applies to the container.
                                      type: string
                                  type: object
                                supplementalGroups:
                                  description: A list of groups applied to the first process run in each container, in addition to the container's primary GID.  If unspecified, no groups will be added to any container.
                                  items:
                                    format: int64
                                    type: integer
                                  type: array
                                sysctls:
                                  description: Sysctls hold a list of namespaced sysctls used for the pod. Pods with unsupported sysctls (by the container runtime) might fail to launch.
                                  items:
                                    description: Sysctl defines a kernel parameter to be set
                                    properties:
                                      name:
                                        description: Name of a property to set
                                        type: string
                                      value:
                                        description: Value of a property to set
                                        type: string
                                    required:
                                    - name
                                    - value
                                    type: object
                                  type: array
                              type: object
                            serviceAccount:
                              description: 'DeprecatedServiceAccount is a depreciated alias for ServiceAccountName. Deprecated: Use serviceAccountName instead.'
                              type: string
                            serviceAccountName:
                              description: 'ServiceAccountName is the name of the ServiceAccount to use to run this pod. More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/'
                              type: string
                            shareProcessNamespace:
                              description: 'Share a single process namespace between all of the containers in a pod. When this is set containers will be able to view and signal processes from other containers in the same pod, and the first process in each container will not be assigned PID 1. HostPID and ShareProcessNamespace cannot both be set. Optional: Default to false. This field is beta-level and may be disabled with the PodShareProcessNamespace feature.'
                              type: boolean
                            subdomain:
                              description: If specified, the fully qualified Pod hostname will be "<hostname>.<subdomain>.<pod namespace>.svc.<cluster domain>". If not specified, the pod will not have a domainname at all.
                              type: string
                            terminationGracePeriodSeconds:
                              description: Optional duration in seconds the pod needs to terminate gracefully. May be decreased in delete request. Value must be non-negative integer. The value zero indicates delete immediately. If this value is nil, the default grace period will be used instead. The grace period is the duration in seconds after the processes running in the pod are sent a termination signal and the time when the processes are forcibly halted with a kill signal. Set this value longer than the expected cleanup time for your process. Defaults to 30 seconds.
                              format: int64
                              type: integer
                            tolerations:
                              description: If specified, the pod's tolerations.
                              items:
                                description: The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>.
                                properties:
                                  effect:
                                    description: Effect indicates the taint effect to match. Empty means match all taint effects. When specified, allowed values are NoSchedule, PreferNoSchedule and NoExecute.
                                    type: string
                                  key:
                                    description: Key is the taint key that the toleration applies to. Empty means match all taint keys. If the key is empty, operator must be Exists; this combination means to match all values and all keys.
                                    type: string
                                  operator:
                                    description: Operator represents a key's relationship to the value. Valid operators are Exists and Equal. Defaults to Equal. Exists is equivalent to wildcard for value, so that a pod can tolerate all taints of a particular category.
                                    type: string
                                  tolerationSeconds:
                                    description: TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default, it is not set, which means tolerate the taint forever (do not evict). Zero and negative values will be treated as 0 (evict immediately) by the system.
                                    format: int64
                                    type: integer
                                  value:
                                    description: Value is the taint value the toleration matches to. If the operator is Exists, the value should be empty, otherwise just a regular string.
                                    type: string
                                type: object
                              type: array
                            volumes:
                              description: 'List of volumes that can be mounted by containers belonging to the pod. More info: https://kubernetes.io/docs/concepts/storage/volumes'
                              items:
                                type: object
                              type: array
                          required:
                          - containers
                          type: object
                      type: object
                    type: array
                  engineResources:
                    description: ResourceRequirements describes the compute resource requirements.
                    properties:
                      limits:
                        additionalProperties:
                          type: string
                        description: 'Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                        type: object
                      requests:
                        additionalProperties:
                          type: string
                        description: 'Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                        type: object
                    type: object
                  explainer:
                    properties:
                      config:
                        additionalProperties:
                          type: string
                        type: object
                      containerSpec:
                        description: A single application container that you want to run within a pod.
                        properties:
                          args:
                            description: 'Arguments to the entrypoint. The docker image''s CMD is used if this is not provided. Variable references $(VAR_NAME) are expanded using the container''s environment. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell'
                            items:
                              type: string
                            type: array
                          command:
                            description: 'Entrypoint array. Not executed within a shell. The docker image''s ENTRYPOINT is used if this is not provided. Variable references $(VAR_NAME) are expanded using the container''s environment. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell'
                            items:
                              type: string
                            type: array
                          env:
                            description: List of environment variables to set in the container. Cannot be updated.
                            items:
                              description: EnvVar represents an environment variable present in a Container.
                              properties:
                                name:
                                  description: Name of the environment variable. Must be a C_IDENTIFIER.
                                  type: string
                                value:
                                  description: 'Variable references $(VAR_NAME) are expanded using the previous defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to "".'
                                  type: string
                                valueFrom:
                                  description: Source for the environment variable's value. Cannot be used if value is not empty.
                                  properties:
                                    configMapKeyRef:
                                      description: Selects a key of a ConfigMap.
                                      properties:
                                        key:
                                          description: The key to select.
                                          type: string
                                        name:
                                          description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?'
                                          type: string
                                        optional:
                                          description: Specify whether the ConfigMap or it's key must be defined
                                          type: boolean
                                      required:
                                      - key
                                      type: object
                                    fieldRef:
                                      description: 'Selects a field of the pod: supports metadata.name, metadata.namespace, metadata.labels, metadata.annotations, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP.'
                                      properties:
                                        apiVersion:
                                          description: Version of the schema the FieldPath is written in terms of, defaults to "v1".
                                          type: string
                                        fieldPath:
                                          description: Path of the field to select in the specified API version.
                                          type: string
                                      required:
                                      - fieldPath
                                      type: object
                                    resourceFieldRef:
                                      description: 'Selects a resource of the container: only resources limits and requests (limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.'
                                      properties:
                                        containerName:
                                          description: 'Container name: required for volumes, optional for env vars'
                                          type: string
                                        divisor:
                                          description: Specifies the output format of the exposed resources, defaults to "1"
                                          type: string
                                        resource:
                                          description: 'Required: resource to select'
                                          type: string
                                      required:
                                      - resource
                                      type: object
                                    secretKeyRef:
                                      description: Selects a key of a secret in the pod's namespace
                                      properties:
                                        key:
                                          description: The key of the secret to select from.  Must be a valid secret key.
                                          type: string
                                        name:
                                          description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?'
                                          type: string
                                        optional:
                                          description: Specify whether the Secret or it's key must be defined
                                          type: boolean
                                      required:
                                      - key
                                      type: object
                                  type: object
                              required:
                              - name
                              type: object
                            type: array
                          envFrom:
                            description: List of sources to populate environment variables in the container. The keys defined within a source must be a C_IDENTIFIER. All invalid keys will be reported as an event when the container is starting. When a key exists in multiple sources, the value associated with the last source will take precedence. Values defined by an Env with a duplicate key will take precedence. Cannot be updated.
                            items:
                              description: EnvFromSource represents the source of a set of ConfigMaps
                              properties:
                                configMapRef:
                                  description: The ConfigMap to select from
                                  properties:
                                    name:
                                      description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?'
                                      type: string
                                    optional:
                                      description: Specify whether the ConfigMap must be defined
                                      type: boolean
                                  type: object
                                prefix:
                                  description: An optional identifier to prepend to each key in the ConfigMap. Must be a C_IDENTIFIER.
                                  type: string
                                secretRef:
                                  description: The Secret to select from
                                  properties:
                                    name:
                                      description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?'
                                      type: string
                                    optional:
                                      description: Specify whether the Secret must be defined
                                      type: boolean
                                  type: object
                              type: object
                            type: array
                          image:
                            description: 'Docker image name. More info: https://kubernetes.io/docs/concepts/containers/images This field is optional to allow higher level config management to default or override container images in workload controllers like Deployments and StatefulSets.'
                            type: string
                          imagePullPolicy:
                            description: 'Image pull policy. One of Always, Never, IfNotPresent. Defaults to Always if :latest tag is specified, or IfNotPresent otherwise. Cannot be updated. More info: https://kubernetes.io/docs/concepts/containers/images#updating-images'
                            type: string
                          lifecycle:
                            description: Actions that the management system should take in response to container lifecycle events. Cannot be updated.
                            properties:
                              postStart:
                                description: 'PostStart is called immediately after a container is created. If the handler fails, the container is terminated and restarted according to its restart policy. Other management of the container blocks until the hook completes. More info: https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks'
                                properties:
                                  exec:
                                    description: One and only one of the following should be specified. Exec specifies the action to take.
                                    properties:
                                      command:
                                        description: Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
                                        items:
                                          type: string
                                        type: array
                                    type: object
                                  httpGet:
                                    description: HTTPGet specifies the http request to perform.
                                    properties:
                                      host:
                                        description: Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.
                                        type: string
                                      httpHeaders:
                                        description: Custom headers to set in the request. HTTP allows repeated headers.
                                        items:
                                          description: HTTPHeader describes a custom header to be used in HTTP probes
                                          properties:
                                            name:
                                              description: The header field name
                                              type: string
                                            value:
                                              description: The header field value
                                              type: string
                                          required:
                                          - name
                                          - value
                                          type: object
                                        type: array
                                      path:
                                        description: Path to access on the HTTP server.
                                        type: string
                                      port:
                                        description: Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                        x-kubernetes-int-or-string: true
                                      scheme:
                                        description: Scheme to use for connecting to the host. Defaults to HTTP.
                                        type: string
                                    required:
                                    - port
                                    type: object
                                  tcpSocket:
                                    description: 'TCPSocket specifies an action involving a TCP port. TCP hooks not yet supported TODO: implement a realistic TCP lifecycle hook'
                                    properties:
                                      host:
                                        description: 'Optional: Host name to connect to, defaults to the pod IP.'
                                        type: string
                                      port:
                                        description: Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                        x-kubernetes-int-or-string: true
                                    required:
                                    - port
                                    type: object
                                type: object
                              preStop:
                                description: 'PreStop is called immediately before a container is terminated due to an API request or management event such as liveness probe failure, preemption, resource contention, etc. The handler is not called if the container crashes or exits. The reason for termination is passed to the handler. The Pod''s termination grace period countdown begins before the PreStop hooked is executed. Regardless of the outcome of the handler, the container will eventually terminate within the Pod''s termination grace period. Other management of the container blocks until the hook completes or until the termination grace period is reached. More info: https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks'
                                properties:
                                  exec:
                                    description: One and only one of the following should be specified. Exec specifies the action to take.
                                    properties:
                                      command:
                                        description: Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
                                        items:
                                          type: string
                                        type: array
                                    type: object
                                  httpGet:
                                    description: HTTPGet specifies the http request to perform.
                                    properties:
                                      host:
                                        description: Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.
                                        type: string
                                      httpHeaders:
                                        description: Custom headers to set in the request. HTTP allows repeated headers.
                                        items:
                                          description: HTTPHeader describes a custom header to be used in HTTP probes
                                          properties:
                                            name:
                                              description: The header field name
                                              type: string
                                            value:
                                              description: The header field value
                                              type: string
                                          required:
                                          - name
                                          - value
                                          type: object
                                        type: array
                                      path:
                                        description: Path to access on the HTTP server.
                                        type: string
                                      port:
                                        description: Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                        x-kubernetes-int-or-string: true
                                      scheme:
                                        description: Scheme to use for connecting to the host. Defaults to HTTP.
                                        type: string
                                    required:
                                    - port
                                    type: object
                                  tcpSocket:
                                    description: 'TCPSocket specifies an action involving a TCP port. TCP hooks not yet supported TODO: implement a realistic TCP lifecycle hook'
                                    properties:
                                      host:
                                        description: 'Optional: Host name to connect to, defaults to the pod IP.'
                                        type: string
                                      port:
                                        description: Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                        x-kubernetes-int-or-string: true
                                    required:
                                    - port
                                    type: object
                                type: object
                            type: object
                          livenessProbe:
                            description: 'Periodic probe of container liveness. Container will be restarted if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                            properties:
                              exec:
                                description: One and only one of the following should be specified. Exec specifies the action to take.
                                properties:
                                  command:
                                    description: Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
                                    items:
                                      type: string
                                    type: array
                                type: object
                              failureThreshold:
                                description: Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3. Minimum value is 1.
                                format: int32
                                type: integer
                              httpGet:
                                description: HTTPGet specifies the http request to perform.
                                properties:
                                  host:
                                    description: Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.
                                    type: string
                                  httpHeaders:
                                    description: Custom headers to set in the request. HTTP allows repeated headers.
                                    items:
                                      description: HTTPHeader describes a custom header to be used in HTTP probes
                                      properties:
                                        name:
                                          description: The header field name
                                          type: string
                                        value:
                                          description: The header field value
                                          type: string
                                      required:
                                      - name
                                      - value
                                      type: object
                                    type: array
                                  path:
                                    description: Path to access on the HTTP server.
                                    type: string
                                  port:
                                    description: Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                    x-kubernetes-int-or-string: true
                                  scheme:
                                    description: Scheme to use for connecting to the host. Defaults to HTTP.
                                    type: string
                                required:
                                - port
                                type: object
                              initialDelaySeconds:
                                description: 'Number of seconds after the container has started before liveness probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                format: int32
                                type: integer
                              periodSeconds:
                                description: How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1.
                                format: int32
                                type: integer
                              successThreshold:
                                description: Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1. Must be 1 for liveness. Minimum value is 1.
                                format: int32
                                type: integer
                              tcpSocket:
                                description: 'TCPSocket specifies an action involving a TCP port. TCP hooks not yet supported TODO: implement a realistic TCP lifecycle hook'
                                properties:
                                  host:
                                    description: 'Optional: Host name to connect to, defaults to the pod IP.'
                                    type: string
                                  port:
                                    description: Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                    x-kubernetes-int-or-string: true
                                required:
                                - port
                                type: object
                              timeoutSeconds:
                                description: 'Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                format: int32
                                type: integer
                            type: object
                          name:
                            description: Name of the container specified as a DNS_LABEL. Each container in a pod must have a unique name (DNS_LABEL). Cannot be updated.
                            type: string
                          ports:
                            description: List of ports to expose from the container. Exposing a port here gives the system additional information about the network connections a container uses, but is primarily informational. Not specifying a port here DOES NOT prevent that port from being exposed. Any port which is listening on the default "0.0.0.0" address inside a container will be accessible from the network. Cannot be updated.
                            items:
                              description: ContainerPort represents a network port in a single container.
                              properties:
                                containerPort:
                                  description: Number of port to expose on the pod's IP address. This must be a valid port number, 0 < x < 65536.
                                  format: int32
                                  type: integer
                                hostIP:
                                  description: What host IP to bind the external port to.
                                  type: string
                                hostPort:
                                  description: Number of port to expose on the host. If specified, this must be a valid port number, 0 < x < 65536. If HostNetwork is specified, this must match ContainerPort. Most containers do not need this.
                                  format: int32
                                  type: integer
                                name:
                                  description: If specified, this must be an IANA_SVC_NAME and unique within the pod. Each named port in a pod must have a unique name. Name for the port that can be referred to by services.
                                  type: string
                                protocol:
                                  description: Protocol for port. Must be UDP, TCP, or SCTP. Defaults to "TCP".
                                  type: string
                              required:
                              - containerPort
                              type: object
                            type: array
                          readinessProbe:
                            description: 'Periodic probe of container service readiness. Container will be removed from service endpoints if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                            properties:
                              exec:
                                description: One and only one of the following should be specified. Exec specifies the action to take.
                                properties:
                                  command:
                                    description: Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
                                    items:
                                      type: string
                                    type: array
                                type: object
                              failureThreshold:
                                description: Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3. Minimum value is 1.
                                format: int32
                                type: integer
                              httpGet:
                                description: HTTPGet specifies the http request to perform.
                                properties:
                                  host:
                                    description: Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.
                                    type: string
                                  httpHeaders:
                                    description: Custom headers to set in the request. HTTP allows repeated headers.
                                    items:
                                      description: HTTPHeader describes a custom header to be used in HTTP probes
                                      properties:
                                        name:
                                          description: The header field name
                                          type: string
                                        value:
                                          description: The header field value
                                          type: string
                                      required:
                                      - name
                                      - value
                                      type: object
                                    type: array
                                  path:
                                    description: Path to access on the HTTP server.
                                    type: string
                                  port:
                                    description: Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                    x-kubernetes-int-or-string: true
                                  scheme:
                                    description: Scheme to use for connecting to the host. Defaults to HTTP.
                                    type: string
                                required:
                                - port
                                type: object
                              initialDelaySeconds:
                                description: 'Number of seconds after the container has started before liveness probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                format: int32
                                type: integer
                              periodSeconds:
                                description: How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1.
                                format: int32
                                type: integer
                              successThreshold:
                                description: Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1. Must be 1 for liveness. Minimum value is 1.
                                format: int32
                                type: integer
                              tcpSocket:
                                description: 'TCPSocket specifies an action involving a TCP port. TCP hooks not yet supported TODO: implement a realistic TCP lifecycle hook'
                                properties:
                                  host:
                                    description: 'Optional: Host name to connect to, defaults to the pod IP.'
                                    type: string
                                  port:
                                    description: Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                    x-kubernetes-int-or-string: true
                                required:
                                - port
                                type: object
                              timeoutSeconds:
                                description: 'Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                format: int32
                                type: integer
                            type: object
                          resources:
                            description: 'Compute Resources required by this container. Cannot be updated. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                            properties:
                              limits:
                                additionalProperties:
                                  type: string
                                description: 'Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                                type: object
                              requests:
                                additionalProperties:
                                  type: string
                                description: 'Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                                type: object
                            type: object
                          securityContext:
                            description: 'Security options the pod should run with. More info: https://kubernetes.io/docs/concepts/policy/security-context/ More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/'
                            properties:
                              allowPrivilegeEscalation:
                                description: 'AllowPrivilegeEscalation controls whether a process can gain more privileges than its parent process. This bool directly controls if the no_new_privs flag will be set on the container process. AllowPrivilegeEscalation is true always when the container is: 1) run as Privileged 2) has CAP_SYS_ADMIN'
                                type: boolean
                              capabilities:
                                description: The capabilities to add/drop when running containers. Defaults to the default set of capabilities granted by the container runtime.
                                properties:
                                  add:
                                    description: Added capabilities
                                    items:
                                      description: Capability represent POSIX capabilities type
                                      type: string
                                    type: array
                                  drop:
                                    description: Removed capabilities
                                    items:
                                      description: Capability represent POSIX capabilities type
                                      type: string
                                    type: array
                                type: object
                              privileged:
                                description: Run container in privileged mode. Processes in privileged containers are essentially equivalent to root on the host. Defaults to false.
                                type: boolean
                              procMount:
                                description: procMount denotes the type of proc mount to use for the containers. The default is DefaultProcMount which uses the container runtime defaults for readonly paths and masked paths. This requires the ProcMountType feature flag to be enabled.
                                type: string
                              readOnlyRootFilesystem:
                                description: Whether this container has a read-only root filesystem. Default is false.
                                type: boolean
                              runAsGroup:
                                description: The GID to run the entrypoint of the container process. Uses runtime default if unset. May also be set in PodSecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.
                                format: int64
                                type: integer
                              runAsNonRoot:
                                description: Indicates that the container must run as a non-root user. If true, the Kubelet will validate the image at runtime to ensure that it does not run as UID 0 (root) and fail to start the container if it does. If unset or false, no such validation will be performed. May also be set in PodSecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.
                                type: boolean
                              runAsUser:
                                description: The UID to run the entrypoint of the container process. Defaults to user specified in image metadata if unspecified. May also be set in PodSecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.
                                format: int64
                                type: integer
                              seLinuxOptions:
                                description: The SELinux context to be applied to the container. If unspecified, the container runtime will allocate a random SELinux context for each container.  May also be set in PodSecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.
                                properties:
                                  level:
                                    description: Level is SELinux level label that applies to the container.
                                    type: string
                                  role:
                                    description: Role is a SELinux role label that applies to the container.
                                    type: string
                                  type:
                                    description: Type is a SELinux type label that applies to the container.
                                    type: string
                                  user:
                                    description: User is a SELinux user label that applies to the container.
                                    type: string
                                type: object
                            type: object
                          stdin:
                            description: Whether this container should allocate a buffer for stdin in the container runtime. If this is not set, reads from stdin in the container will always result in EOF. Default is false.
                            type: boolean
                          stdinOnce:
                            description: Whether the container runtime should close the stdin channel after it has been opened by a single attach. When stdin is true the stdin stream will remain open across multiple attach sessions. If stdinOnce is set to true, stdin is opened on container start, is empty until the first client attaches to stdin, and then remains open and accepts data until the client disconnects, at which time stdin is closed and remains closed until the container is restarted. If this flag is false, a container processes that reads from stdin will never receive an EOF. Default is false
                            type: boolean
                          terminationMessagePath:
                            description: 'Optional: Path at which the file to which the container''s termination message will be written is mounted into the container''s filesystem. Message written is intended to be brief final status, such as an assertion failure message. Will be truncated by the node if greater than 4096 bytes. The total message length across all containers will be limited to 12kb. Defaults to /dev/termination-log. Cannot be updated.'
                            type: string
                          terminationMessagePolicy:
                            description: Indicate how the termination message should be populated. File will use the contents of terminationMessagePath to populate the container status message on both success and failure. FallbackToLogsOnError will use the last chunk of container log output if the termination message file is empty and the container exited with an error. The log output is limited to 2048 bytes or 80 lines, whichever is smaller. Defaults to File. Cannot be updated.
                            type: string
                          tty:
                            description: Whether this container should allocate a TTY for itself, also requires 'stdin' to be true. Default is false.
                            type: boolean
                          volumeDevices:
                            description: volumeDevices is the list of block devices to be used by the container. This is a beta feature.
                            items:
                              description: volumeDevice describes a mapping of a raw block device within a container.
                              properties:
                                devicePath:
                                  description: devicePath is the path inside of the container that the device will be mapped to.
                                  type: string
                                name:
                                  description: name must match the name of a persistentVolumeClaim in the pod
                                  type: string
                              required:
                              - devicePath
                              - name
                              type: object
                            type: array
                          volumeMounts:
                            description: Pod volumes to mount into the container's filesystem. Cannot be updated.
                            items:
                              description: VolumeMount describes a mounting of a Volume within a container.
                              properties:
                                mountPath:
                                  description: Path within the container at which the volume should be mounted.  Must not contain ':'.
                                  type: string
                                mountPropagation:
                                  description: mountPropagation determines how mounts are propagated from the host to container and the other way around. When not set, MountPropagationNone is used. This field is beta in 1.10.
                                  type: string
                                name:
                                  description: This must match the Name of a Volume.
                                  type: string
                                readOnly:
                                  description: Mounted read-only if true, read-write otherwise (false or unspecified). Defaults to false.
                                  type: boolean
                                subPath:
                                  description: Path within the volume from which the container's volume should be mounted. Defaults to "" (volume's root).
                                  type: string
                                subPathExpr:
                                  description: Expanded path within the volume from which the container's volume should be mounted. Behaves similarly to SubPath but environment variable references $(VAR_NAME) are expanded using the container's environment. Defaults to "" (volume's root). SubPathExpr and SubPath are mutually exclusive. This field is alpha in 1.14.
                                  type: string
                              required:
                              - mountPath
                              - name
                              type: object
                            type: array
                          workingDir:
                            description: Container's working directory. If not specified, the container runtime's default will be used, which might be configured in the container image. Cannot be updated.
                            type: string
                        required:
                        - name
                        type: object
                      endpoint:
                        properties:
                          service_host:
                            type: string
                          service_port:
                            format: int32
                            type: integer
                          type:
                            type: string
                        type: object
                      envSecretRefName:
                        type: string
                      modelUri:
                        type: string
                      serviceAccountName:
                        type: string
                      type:
                        type: string
                    type: object
                  graph:
                    properties:
                      children:
                        items:
                          properties:
                            children:
                              items:
                                properties:
                                  children:
                                    items:
                                      properties:
                                        children:
                                          items:
                                            properties:
                                              endpoint:
                                                properties:
                                                  service_host:
                                                    type: string
                                                  service_port:
                                                    format: int32
                                                    type: integer
                                                  type:
                                                    type: string
                                                type: object
                                              envSecretRefName:
                                                type: string
                                              implementation:
                                                type: string
                                              methods:
                                                items:
                                                  type: string
                                                type: array
                                              modelUri:
                                                type: string
                                              name:
                                                type: string
                                              parameters:
                                                items:
                                                  properties:
                                                    name:
                                                      type: string
                                                    type:
                                                      type: string
                                                    value:
                                                      type: string
                                                  type: object
                                                type: array
                                              serviceAccountName:
                                                type: string
                                              type:
                                                type: string
                                            type: object
                                          type: array
                                        endpoint:
                                          properties:
                                            service_host:
                                              type: string
                                            service_port:
                                              format: int32
                                              type: integer
                                            type:
                                              type: string
                                          type: object
                                        envSecretRefName:
                                          type: string
                                        implementation:
                                          type: string
                                        methods:
                                          items:
                                            type: string
                                          type: array
                                        modelUri:
                                          type: string
                                        name:
                                          type: string
                                        parameters:
                                          items:
                                            properties:
                                              name:
                                                type: string
                                              type:
                                                type: string
                                              value:
                                                type: string
                                            type: object
                                          type: array
                                        serviceAccountName:
                                          type: string
                                        type:
                                          type: string
                                      type: object
                                    type: array
                                  endpoint:
                                    properties:
                                      service_host:
                                        type: string
                                      service_port:
                                        format: int32
                                        type: integer
                                      type:
                                        type: string
                                    type: object
                                  envSecretRefName:
                                    type: string
                                  implementation:
                                    type: string
                                  methods:
                                    items:
                                      type: string
                                    type: array
                                  modelUri:
                                    type: string
                                  name:
                                    type: string
                                  parameters:
                                    items:
                                      properties:
                                        name:
                                          type: string
                                        type:
                                          type: string
                                        value:
                                          type: string
                                      type: object
                                    type: array
                                  serviceAccountName:
                                    type: string
                                  type:
                                    type: string
                                type: object
                              type: array
                            endpoint:
                              properties:
                                service_host:
                                  type: string
                                service_port:
                                  format: int32
                                  type: integer
                                type:
                                  type: string
                              type: object
                            envSecretRefName:
                              type: string
                            implementation:
                              type: string
                            methods:
                              items:
                                type: string
                              type: array
                            modelUri:
                              type: string
                            name:
                              type: string
                            parameters:
                              items:
                                properties:
                                  name:
                                    type: string
                                  type:
                                    type: string
                                  value:
                                    type: string
                                type: object
                              type: array
                            serviceAccountName:
                              type: string
                            type:
                              type: string
                          type: object
                        type: array
                      endpoint:
                        properties:
                          service_host:
                            type: string
                          service_port:
                            format: int32
                            type: integer
                          type:
                            type: string
                        type: object
                      envSecretRefName:
                        type: string
                      implementation:
                        type: string
                      methods:
                        items:
                          type: string
                        type: array
                      modelUri:
                        type: string
                      name:
                        type: string
                      parameters:
                        items:
                          properties:
                            name:
                              type: string
                            type:
                              type: string
                            value:
                              type: string
                          required:
                          - name
                          - type
                          - value
                          type: object
                        type: array
                      serviceAccountName:
                        type: string
                      type:
                        type: string
                    required:
                    - name
                    type: object
                  labels:
                    additionalProperties:
                      type: string
                    type: object
                  name:
                    type: string
                  replicas:
                    format: int32
                    type: integer
                  shadow:
                    type: boolean
                  svcOrchSpec:
                    properties:
                      env:
                        items:
                          description: EnvVar represents an environment variable present in a Container.
                          properties:
                            name:
                              description: Name of the environment variable. Must be a C_IDENTIFIER.
                              type: string
                            value:
                              description: 'Variable references $(VAR_NAME) are expanded using the previous defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to "".'
                              type: string
                            valueFrom:
                              description: Source for the environment variable's value. Cannot be used if value is not empty.
                              properties:
                                configMapKeyRef:
                                  description: Selects a key of a ConfigMap.
                                  properties:
                                    key:
                                      description: The key to select.
                                      type: string
                                    name:
                                      description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?'
                                      type: string
                                    optional:
                                      description: Specify whether the ConfigMap or it's key must be defined
                                      type: boolean
                                  required:
                                  - key
                                  type: object
                                fieldRef:
                                  description: 'Selects a field of the pod: supports metadata.name, metadata.namespace, metadata.labels, metadata.annotations, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP.'
                                  properties:
                                    apiVersion:
                                      description: Version of the schema the FieldPath is written in terms of, defaults to "v1".
                                      type: string
                                    fieldPath:
                                      description: Path of the field to select in the specified API version.
                                      type: string
                                  required:
                                  - fieldPath
                                  type: object
                                resourceFieldRef:
                                  description: 'Selects a resource of the container: only resources limits and requests (limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.'
                                  properties:
                                    containerName:
                                      description: 'Container name: required for volumes, optional for env vars'
                                      type: string
                                    divisor:
                                      description: Specifies the output format of the exposed resources, defaults to "1"
                                      type: string
                                    resource:
                                      description: 'Required: resource to select'
                                      type: string
                                  required:
                                  - resource
                                  type: object
                                secretKeyRef:
                                  description: Selects a key of a secret in the pod's namespace
                                  properties:
                                    key:
                                      description: The key of the secret to select from.  Must be a valid secret key.
                                      type: string
                                    name:
                                      description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?'
                                      type: string
                                    optional:
                                      description: Specify whether the Secret or it's key must be defined
                                      type: boolean
                                  required:
                                  - key
                                  type: object
                              type: object
                          required:
                          - name
                          type: object
                        type: array
                      resources:
                        description: ResourceRequirements describes the compute resource requirements.
                        properties:
                          limits:
                            additionalProperties:
                              type: string
                            description: 'Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                            type: object
                          requests:
                            additionalProperties:
                              type: string
                            description: 'Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                            type: object
                        type: object
                    type: object
                  traffic:
                    format: int32
                    type: integer
                required:
                - graph
                - name
                type: object
              type: array
          required:
          - predictors
          type: object
        status:
          description: SeldonDeploymentStatus defines the observed state of SeldonDeployment
          properties:
            deploymentStatus:
              additionalProperties:
                properties:
                  availableReplicas:
                    format: int32
                    type: integer
                  description:
                    type: string
                  explainerFor:
                    type: string
                  name:
                    type: string
                  replicas:
                    format: int32
                    type: integer
                  status:
                    type: string
                type: object
              type: object
            description:
              type: string
            serviceStatus:
              additionalProperties:
                properties:
                  explainerFor:
                    type: string
                  grpcEndpoint:
                    type: string
                  httpEndpoint:
                    type: string
                  svcName:
                    type: string
                type: object
              type: object
            state:
              type: string
          type: object
      type: object
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
  - name: v1alpha2
    served: true
    storage: false
  - name: v1alpha3
    served: true
    storage: false
status:
  acceptedNames:
    kind: ''
    plural: ''
  conditions: []
  storedVersions: []
---
# Source: seldon-core-operator/templates/clusterrole_seldon-manager-role.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  creationTimestamp: null
  labels:
    app: seldon
    app.kubernetes.io/instance: 'RELEASE-NAME'
    app.kubernetes.io/name: 'seldon-core-operator'
    app.kubernetes.io/version: '1.0.1'
  name: seldon-manager-role-kubeflow
rules:
- apiGroups:
  - ''
  resources:
  - namespaces
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ''
  resources:
  - services
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - apps
  resources:
  - deployments
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - apps
  resources:
  - deployments/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - autoscaling
  resources:
  - horizontalpodautoscalers
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - autoscaling
  resources:
  - horizontalpodautoscalers/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - machinelearning.seldon.io
  resources:
  - seldondeployments
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - machinelearning.seldon.io
  resources:
  - seldondeployments/finalizers
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - machinelearning.seldon.io
  resources:
  - seldondeployments/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - networking.istio.io
  resources:
  - destinationrules
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - networking.istio.io
  resources:
  - destinationrules/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - networking.istio.io
  resources:
  - virtualservices
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - networking.istio.io
  resources:
  - virtualservices/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - v1
  resources:
  - namespaces
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - v1
  resources:
  - services
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - v1
  resources:
  - services/status
  verbs:
  - get
  - patch
  - update
---
# Source: seldon-core-operator/templates/clusterrole_seldon-manager-sas-role.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  creationTimestamp: null
  labels:
    app: seldon
    app.kubernetes.io/instance: 'RELEASE-NAME'
    app.kubernetes.io/name: 'seldon-core-operator'
    app.kubernetes.io/version: '1.0.1'
  name: seldon-manager-sas-role-kubeflow
rules:
- apiGroups:
  - ''
  resources:
  - secrets
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ''
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ''
  resources:
  - serviceaccounts
  verbs:
  - get
  - list
  - watch
---
# Source: seldon-core-operator/templates/clusterrolebinding_seldon-manager-rolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: seldon
    app.kubernetes.io/instance: 'RELEASE-NAME'
    app.kubernetes.io/name: 'seldon-core-operator'
    app.kubernetes.io/version: '1.0.1'
  name: seldon-manager-rolebinding-kubeflow
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: seldon-manager-role-kubeflow
subjects:
- kind: ServiceAccount
  name: 'seldon-manager'
  namespace: 'kubeflow'
---
# Source: seldon-core-operator/templates/clusterrolebinding_seldon-manager-sas-rolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: seldon
    app.kubernetes.io/instance: 'RELEASE-NAME'
    app.kubernetes.io/name: 'seldon-core-operator'
    app.kubernetes.io/version: '1.0.1'
  name: seldon-manager-sas-rolebinding-kubeflow
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: seldon-manager-sas-role-kubeflow
subjects:
- kind: ServiceAccount
  name: seldon-manager
  namespace: 'kubeflow'
---
# Source: seldon-core-operator/templates/role_seldon-leader-election-role.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  labels:
    app: seldon
    app.kubernetes.io/instance: 'RELEASE-NAME'
    app.kubernetes.io/name: 'seldon-core-operator'
    app.kubernetes.io/version: '1.0.1'
  name: seldon-leader-election-role
  namespace: 'kubeflow'
rules:
- apiGroups:
  - ''
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - ''
  resources:
  - configmaps/status
  verbs:
  - get
  - update
  - patch
- apiGroups:
  - ''
  resources:
  - events
  verbs:
  - create
---
# Source: seldon-core-operator/templates/role_seldon-manager-cm-role.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  creationTimestamp: null
  labels:
    app: seldon
    app.kubernetes.io/instance: 'RELEASE-NAME'
    app.kubernetes.io/name: 'seldon-core-operator'
    app.kubernetes.io/version: '1.0.1'
  name: seldon-manager-cm-role
  namespace: 'kubeflow'
rules:
- apiGroups:
  - ''
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
---
# Source: seldon-core-operator/templates/rolebinding_seldon-leader-election-rolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app: seldon
    app.kubernetes.io/instance: 'RELEASE-NAME'
    app.kubernetes.io/name: 'seldon-core-operator'
    app.kubernetes.io/version: '1.0.1'
  name: seldon-leader-election-rolebinding
  namespace: 'kubeflow'
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: seldon-leader-election-role
subjects:
- kind: ServiceAccount
  name: seldon-manager
  namespace: 'kubeflow'
---
# Source: seldon-core-operator/templates/rolebinding_seldon-manager-cm-rolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app: seldon
    app.kubernetes.io/instance: 'RELEASE-NAME'
    app.kubernetes.io/name: 'seldon-core-operator'
    app.kubernetes.io/version: '1.0.1'
  name: seldon-manager-cm-rolebinding
  namespace: 'kubeflow'
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: seldon-manager-cm-role
subjects:
- kind: ServiceAccount
  name: seldon-manager
  namespace: 'kubeflow'
---
# Source: seldon-core-operator/templates/service_seldon-webhook-service.yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app: seldon
    app.kubernetes.io/instance: 'RELEASE-NAME'
    app.kubernetes.io/name: 'seldon-core-operator'
    app.kubernetes.io/version: '1.0.1'
  name: seldon-webhook-service
  namespace: 'kubeflow'
spec:
  ports:
  - port: 443
    targetPort: 443
  selector:
    app: seldon
    app.kubernetes.io/instance: seldon1
    app.kubernetes.io/name: seldon
    app.kubernetes.io/version: v0.5
    control-plane: seldon-controller-manager
---
# Source: seldon-core-operator/templates/deployment_seldon-controller-manager.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: seldon
    app.kubernetes.io/instance: 'RELEASE-NAME'
    app.kubernetes.io/name: 'seldon-core-operator'
    app.kubernetes.io/version: '1.0.1'
    control-plane: seldon-controller-manager
  name: seldon-controller-manager
  namespace: 'kubeflow'
spec:
  replicas: 1
  selector:
    matchLabels:
      app: seldon
      app.kubernetes.io/instance: seldon1
      app.kubernetes.io/name: seldon
      app.kubernetes.io/version: v0.5
      control-plane: seldon-controller-manager
  template:
    metadata:
      annotations:
        prometheus.io/scrape: 'true'
        sidecar.istio.io/inject: 'false'
      labels:
        app: seldon
        app.kubernetes.io/instance: seldon1
        app.kubernetes.io/name: seldon
        app.kubernetes.io/version: v0.5
        control-plane: seldon-controller-manager
    spec:
      containers:
      - args:
        - --enable-leader-election
        - --webhook-port=443
        - ''
        command:
        - /manager
        env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: CONTROLLER_ID
          value: ''
        - name: AMBASSADOR_ENABLED
          value: 'true'
        - name: AMBASSADOR_SINGLE_NAMESPACE
          value: 'false'
        - name: ENGINE_CONTAINER_IMAGE_AND_VERSION
          value: 'docker.io/seldonio/engine:1.0.1'
        - name: ENGINE_CONTAINER_IMAGE_PULL_POLICY
          value: 'IfNotPresent'
        - name: ENGINE_CONTAINER_SERVICE_ACCOUNT_NAME
          value: 'default'
        - name: ENGINE_CONTAINER_USER
          value: '8888'
        - name: ENGINE_LOG_MESSAGES_EXTERNALLY
          value: 'false'
        - name: PREDICTIVE_UNIT_SERVICE_PORT
          value: '9000'
        - name: ENGINE_SERVER_GRPC_PORT
          value: '5001'
        - name: ENGINE_SERVER_PORT
          value: '8000'
        - name: ENGINE_PROMETHEUS_PATH
          value: 'prometheus'
        - name: ISTIO_ENABLED
          value: 'true'
        - name: ISTIO_GATEWAY
          value: 'kubeflow-gateway'
        - name: ISTIO_TLS_MODE
          value: ''
        image: 'docker.io/seldonio/seldon-core-operator:1.0.1'
        imagePullPolicy: 'IfNotPresent'
        name: manager
        ports:
        - containerPort: 443
          name: webhook-server
          protocol: TCP
        - containerPort: 8080
          name: metrics
          protocol: TCP
        resources:
          limits:
            cpu: 100m
            memory: 30Mi
          requests:
            cpu: 100m
            memory: 20Mi
        volumeMounts:
        - mountPath: /tmp/k8s-webhook-server/serving-certs
          name: cert
          readOnly: true
      serviceAccountName: seldon-manager
      terminationGracePeriodSeconds: 10
      volumes:
      - name: cert
        secret:
          defaultMode: 420
          secretName: seldon-webhook-server-cert
---
# Source: seldon-core-operator/templates/certificate_seldon-serving-cert.yaml
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  labels:
    app: seldon
    app.kubernetes.io/instance: 'RELEASE-NAME'
    app.kubernetes.io/name: 'seldon-core-operator'
    app.kubernetes.io/version: '1.0.1'
  name: seldon-serving-cert
  namespace: 'kubeflow'
spec:
  commonName: 'seldon-webhook-service.kubeflow.svc'
  dnsNames:
  - 'seldon-webhook-service.kubeflow.svc.cluster.local'
  - 'seldon-webhook-service.kubeflow.svc'
  issuerRef:
    kind: Issuer
    name: seldon-selfsigned-issuer
  secretName: seldon-webhook-server-cert
---
# Source: seldon-core-operator/templates/issuer_seldon-selfsigned-issuer.yaml
apiVersion: cert-manager.io/v1alpha2
kind: Issuer
metadata:
  labels:
    app: seldon
    app.kubernetes.io/instance: 'RELEASE-NAME'
    app.kubernetes.io/name: 'seldon-core-operator'
    app.kubernetes.io/version: '1.0.1'
  name: seldon-selfsigned-issuer
  namespace: 'kubeflow'
spec:
  selfSigned: {}
---
# Source: seldon-core-operator/templates/webhook.yaml
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  annotations:
    cert-manager.io/inject-ca-from: 'kubeflow/seldon-serving-cert'
  creationTimestamp: null
  labels:
    app: seldon
    app.kubernetes.io/instance: 'RELEASE-NAME'
    app.kubernetes.io/name: 'seldon-core-operator'
    app.kubernetes.io/version: '1.0.1'
  name: seldon-mutating-webhook-configuration-kubeflow
webhooks:
- clientConfig:
    caBundle: 'LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURCRENDQWV5Z0F3SUJBZ0lRREpTUVN2d0RJbzdaTCtybngwNmJSakFOQmdrcWhraUc5dzBCQVFzRkFEQWMKTVJvd0dBWURWUVFERXhGamRYTjBiMjB0YldWMGNtbGpjeTFqWVRBZUZ3MHlNREF5TURNd05qSTBOVFZhRncweQpNVEF5TURJd05qSTBOVFZhTUJ3eEdqQVlCZ05WQkFNVEVXTjFjM1J2YlMxdFpYUnlhV056TFdOaE1JSUJJakFOCkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXRldldVOVFvN2lGUnlzMS9TTlh2VmFkSGFVSFIKTmVaSW9FQnVvN0RXRFI1cnFlbk9Ubmdpc1BrYzE5VDVTelpLRWtZVk52TnIrZ1RsUEo2OXE4N241VzRlc3UxbgpKa045SVAzWHA4enBtQ0FXTnQvaHFrOFJNZ0hFVml0Vkh6czdEVS9UMW5DRnAwRmltQTVUUkVNK1hyUU1EVFBmCkNVYnUycFh5UU44WHkrQmdHTWJDcXpIUmZ1Z1YrbHJDMHRxZzd2djZDWGRmdExyM3dvdlRKQXVOcjhjdkRUaEkKc3dYY3ZGdXFrdkhvM3hEWlgrQksyeTMwd1NmUnFoYWMwWjZGNlVESWdqbFRka0w0N0N1c2RISDNWeUZKczFxUApTLzRuN29iV2FsZG5DSDFUWlZtZmZIc3EvTjZJZ0F6TDE3RGlLUS9yWG9GV3RxZEM3Vmhtb3JnT093SURBUUFCCm8wSXdRREFPQmdOVkhROEJBZjhFQkFNQ0FxUXdIUVlEVlIwbEJCWXdGQVlJS3dZQkJRVUhBd0VHQ0NzR0FRVUYKQndNQ01BOEdBMVVkRXdFQi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFDZHh2aElVR2ZWMQpKRHc4eTdpUUt5NmRZMFc1L3hrR2JFeDc1YXNpSzA3SmJPYnFQeGJlUVJ3RGRYcm02U2JSVUphbjFkMmFYWjlTCmVHejhRUVIvUzVyQmFnN0lMUlREdmx6bmoxMnJ3RzVCZTRkYlJES25KNWdGdUU4M0hMYTZneU1rZ0NYUjNGenYKWG84eTA0L09yVGRGRlhLVWNnRHlBK2dQQlFkNFFSNUtBdVN4Z1dCR2RjbmUwdkNyS3FsRERhVGtMd3ZWVDE1ZgpvaWpzQ01GQ1drY0pzVEtyVzcxZ1cxZ0pmcHEzS0VNN1YxNERIbHVkSlJ0bXM2SlBITHFNdS8zWEpCWENzNTdsCkdVYTRCZnUzTVVmQmdESTFjYThnYzluT3Y2NTZrL2lBTnZuRHQyZGdBeDJscDVLbEp3cWRMMTVZN0llUVJieVAKb3J6WndESGIwTEE9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K'
    service:
      name: seldon-webhook-service
      namespace: 'kubeflow'
      path: /mutate-machinelearning-seldon-io-v1-seldondeployment
  failurePolicy: Fail
  name: mseldondeployment.kb.io
  namespaceSelector:
    matchExpressions:
    - key: seldon.io/controller-id
      operator: DoesNotExist
    matchLabels:
      serving.kubeflow.org/inferenceservice: enabled
  rules:
  - apiGroups:
    - machinelearning.seldon.io
    apiVersions:
    - v1
    operations:
    - CREATE
    - UPDATE
    resources:
    - seldondeployments
- clientConfig:
    caBundle: 'LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURCRENDQWV5Z0F3SUJBZ0lRREpTUVN2d0RJbzdaTCtybngwNmJSakFOQmdrcWhraUc5dzBCQVFzRkFEQWMKTVJvd0dBWURWUVFERXhGamRYTjBiMjB0YldWMGNtbGpjeTFqWVRBZUZ3MHlNREF5TURNd05qSTBOVFZhRncweQpNVEF5TURJd05qSTBOVFZhTUJ3eEdqQVlCZ05WQkFNVEVXTjFjM1J2YlMxdFpYUnlhV056TFdOaE1JSUJJakFOCkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXRldldVOVFvN2lGUnlzMS9TTlh2VmFkSGFVSFIKTmVaSW9FQnVvN0RXRFI1cnFlbk9Ubmdpc1BrYzE5VDVTelpLRWtZVk52TnIrZ1RsUEo2OXE4N241VzRlc3UxbgpKa045SVAzWHA4enBtQ0FXTnQvaHFrOFJNZ0hFVml0Vkh6czdEVS9UMW5DRnAwRmltQTVUUkVNK1hyUU1EVFBmCkNVYnUycFh5UU44WHkrQmdHTWJDcXpIUmZ1Z1YrbHJDMHRxZzd2djZDWGRmdExyM3dvdlRKQXVOcjhjdkRUaEkKc3dYY3ZGdXFrdkhvM3hEWlgrQksyeTMwd1NmUnFoYWMwWjZGNlVESWdqbFRka0w0N0N1c2RISDNWeUZKczFxUApTLzRuN29iV2FsZG5DSDFUWlZtZmZIc3EvTjZJZ0F6TDE3RGlLUS9yWG9GV3RxZEM3Vmhtb3JnT093SURBUUFCCm8wSXdRREFPQmdOVkhROEJBZjhFQkFNQ0FxUXdIUVlEVlIwbEJCWXdGQVlJS3dZQkJRVUhBd0VHQ0NzR0FRVUYKQndNQ01BOEdBMVVkRXdFQi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFDZHh2aElVR2ZWMQpKRHc4eTdpUUt5NmRZMFc1L3hrR2JFeDc1YXNpSzA3SmJPYnFQeGJlUVJ3RGRYcm02U2JSVUphbjFkMmFYWjlTCmVHejhRUVIvUzVyQmFnN0lMUlREdmx6bmoxMnJ3RzVCZTRkYlJES25KNWdGdUU4M0hMYTZneU1rZ0NYUjNGenYKWG84eTA0L09yVGRGRlhLVWNnRHlBK2dQQlFkNFFSNUtBdVN4Z1dCR2RjbmUwdkNyS3FsRERhVGtMd3ZWVDE1ZgpvaWpzQ01GQ1drY0pzVEtyVzcxZ1cxZ0pmcHEzS0VNN1YxNERIbHVkSlJ0bXM2SlBITHFNdS8zWEpCWENzNTdsCkdVYTRCZnUzTVVmQmdESTFjYThnYzluT3Y2NTZrL2lBTnZuRHQyZGdBeDJscDVLbEp3cWRMMTVZN0llUVJieVAKb3J6WndESGIwTEE9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K'
    service:
      name: seldon-webhook-service
      namespace: 'kubeflow'
      path: /mutate-machinelearning-seldon-io-v1alpha2-seldondeployment
  failurePolicy: Fail
  name: mseldondeployment.kb.io
  namespaceSelector:
    matchExpressions:
    - key: seldon.io/controller-id
      operator: DoesNotExist
    matchLabels:
      serving.kubeflow.org/inferenceservice: enabled
  rules:
  - apiGroups:
    - machinelearning.seldon.io
    apiVersions:
    - v1alpha2
    operations:
    - CREATE
    - UPDATE
    resources:
    - seldondeployments
- clientConfig:
    caBundle: 'LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURCRENDQWV5Z0F3SUJBZ0lRREpTUVN2d0RJbzdaTCtybngwNmJSakFOQmdrcWhraUc5dzBCQVFzRkFEQWMKTVJvd0dBWURWUVFERXhGamRYTjBiMjB0YldWMGNtbGpjeTFqWVRBZUZ3MHlNREF5TURNd05qSTBOVFZhRncweQpNVEF5TURJd05qSTBOVFZhTUJ3eEdqQVlCZ05WQkFNVEVXTjFjM1J2YlMxdFpYUnlhV056TFdOaE1JSUJJakFOCkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXRldldVOVFvN2lGUnlzMS9TTlh2VmFkSGFVSFIKTmVaSW9FQnVvN0RXRFI1cnFlbk9Ubmdpc1BrYzE5VDVTelpLRWtZVk52TnIrZ1RsUEo2OXE4N241VzRlc3UxbgpKa045SVAzWHA4enBtQ0FXTnQvaHFrOFJNZ0hFVml0Vkh6czdEVS9UMW5DRnAwRmltQTVUUkVNK1hyUU1EVFBmCkNVYnUycFh5UU44WHkrQmdHTWJDcXpIUmZ1Z1YrbHJDMHRxZzd2djZDWGRmdExyM3dvdlRKQXVOcjhjdkRUaEkKc3dYY3ZGdXFrdkhvM3hEWlgrQksyeTMwd1NmUnFoYWMwWjZGNlVESWdqbFRka0w0N0N1c2RISDNWeUZKczFxUApTLzRuN29iV2FsZG5DSDFUWlZtZmZIc3EvTjZJZ0F6TDE3RGlLUS9yWG9GV3RxZEM3Vmhtb3JnT093SURBUUFCCm8wSXdRREFPQmdOVkhROEJBZjhFQkFNQ0FxUXdIUVlEVlIwbEJCWXdGQVlJS3dZQkJRVUhBd0VHQ0NzR0FRVUYKQndNQ01BOEdBMVVkRXdFQi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFDZHh2aElVR2ZWMQpKRHc4eTdpUUt5NmRZMFc1L3hrR2JFeDc1YXNpSzA3SmJPYnFQeGJlUVJ3RGRYcm02U2JSVUphbjFkMmFYWjlTCmVHejhRUVIvUzVyQmFnN0lMUlREdmx6bmoxMnJ3RzVCZTRkYlJES25KNWdGdUU4M0hMYTZneU1rZ0NYUjNGenYKWG84eTA0L09yVGRGRlhLVWNnRHlBK2dQQlFkNFFSNUtBdVN4Z1dCR2RjbmUwdkNyS3FsRERhVGtMd3ZWVDE1ZgpvaWpzQ01GQ1drY0pzVEtyVzcxZ1cxZ0pmcHEzS0VNN1YxNERIbHVkSlJ0bXM2SlBITHFNdS8zWEpCWENzNTdsCkdVYTRCZnUzTVVmQmdESTFjYThnYzluT3Y2NTZrL2lBTnZuRHQyZGdBeDJscDVLbEp3cWRMMTVZN0llUVJieVAKb3J6WndESGIwTEE9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K'
    service:
      name: seldon-webhook-service
      namespace: 'kubeflow'
      path: /mutate-machinelearning-seldon-io-v1alpha3-seldondeployment
  failurePolicy: Fail
  name: mseldondeployment.kb.io
  namespaceSelector:
    matchExpressions:
    - key: seldon.io/controller-id
      operator: DoesNotExist
    matchLabels:
      serving.kubeflow.org/inferenceservice: enabled
  rules:
  - apiGroups:
    - machinelearning.seldon.io
    apiVersions:
    - v1alpha3
    operations:
    - CREATE
    - UPDATE
    resources:
    - seldondeployments
---
# Source: seldon-core-operator/templates/webhook.yaml
---

apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  annotations:
    cert-manager.io/inject-ca-from: 'kubeflow/seldon-serving-cert'
  creationTimestamp: null
  labels:
    app: seldon
    app.kubernetes.io/instance: 'RELEASE-NAME'
    app.kubernetes.io/name: 'seldon-core-operator'
    app.kubernetes.io/version: '1.0.1'
  name: seldon-validating-webhook-configuration-kubeflow
webhooks:
- clientConfig:
    caBundle: 'LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURCRENDQWV5Z0F3SUJBZ0lRREpTUVN2d0RJbzdaTCtybngwNmJSakFOQmdrcWhraUc5dzBCQVFzRkFEQWMKTVJvd0dBWURWUVFERXhGamRYTjBiMjB0YldWMGNtbGpjeTFqWVRBZUZ3MHlNREF5TURNd05qSTBOVFZhRncweQpNVEF5TURJd05qSTBOVFZhTUJ3eEdqQVlCZ05WQkFNVEVXTjFjM1J2YlMxdFpYUnlhV056TFdOaE1JSUJJakFOCkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXRldldVOVFvN2lGUnlzMS9TTlh2VmFkSGFVSFIKTmVaSW9FQnVvN0RXRFI1cnFlbk9Ubmdpc1BrYzE5VDVTelpLRWtZVk52TnIrZ1RsUEo2OXE4N241VzRlc3UxbgpKa045SVAzWHA4enBtQ0FXTnQvaHFrOFJNZ0hFVml0Vkh6czdEVS9UMW5DRnAwRmltQTVUUkVNK1hyUU1EVFBmCkNVYnUycFh5UU44WHkrQmdHTWJDcXpIUmZ1Z1YrbHJDMHRxZzd2djZDWGRmdExyM3dvdlRKQXVOcjhjdkRUaEkKc3dYY3ZGdXFrdkhvM3hEWlgrQksyeTMwd1NmUnFoYWMwWjZGNlVESWdqbFRka0w0N0N1c2RISDNWeUZKczFxUApTLzRuN29iV2FsZG5DSDFUWlZtZmZIc3EvTjZJZ0F6TDE3RGlLUS9yWG9GV3RxZEM3Vmhtb3JnT093SURBUUFCCm8wSXdRREFPQmdOVkhROEJBZjhFQkFNQ0FxUXdIUVlEVlIwbEJCWXdGQVlJS3dZQkJRVUhBd0VHQ0NzR0FRVUYKQndNQ01BOEdBMVVkRXdFQi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFDZHh2aElVR2ZWMQpKRHc4eTdpUUt5NmRZMFc1L3hrR2JFeDc1YXNpSzA3SmJPYnFQeGJlUVJ3RGRYcm02U2JSVUphbjFkMmFYWjlTCmVHejhRUVIvUzVyQmFnN0lMUlREdmx6bmoxMnJ3RzVCZTRkYlJES25KNWdGdUU4M0hMYTZneU1rZ0NYUjNGenYKWG84eTA0L09yVGRGRlhLVWNnRHlBK2dQQlFkNFFSNUtBdVN4Z1dCR2RjbmUwdkNyS3FsRERhVGtMd3ZWVDE1ZgpvaWpzQ01GQ1drY0pzVEtyVzcxZ1cxZ0pmcHEzS0VNN1YxNERIbHVkSlJ0bXM2SlBITHFNdS8zWEpCWENzNTdsCkdVYTRCZnUzTVVmQmdESTFjYThnYzluT3Y2NTZrL2lBTnZuRHQyZGdBeDJscDVLbEp3cWRMMTVZN0llUVJieVAKb3J6WndESGIwTEE9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K'
    service:
      name: seldon-webhook-service
      namespace: 'kubeflow'
      path: /validate-machinelearning-seldon-io-v1-seldondeployment
  failurePolicy: Fail
  name: vseldondeployment.kb.io
  namespaceSelector:
    matchExpressions:
    - key: seldon.io/controller-id
      operator: DoesNotExist
    matchLabels:
      serving.kubeflow.org/inferenceservice: enabled
  rules:
  - apiGroups:
    - machinelearning.seldon.io
    apiVersions:
    - v1
    operations:
    - CREATE
    - UPDATE
    resources:
    - seldondeployments
- clientConfig:
    caBundle: 'LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURCRENDQWV5Z0F3SUJBZ0lRREpTUVN2d0RJbzdaTCtybngwNmJSakFOQmdrcWhraUc5dzBCQVFzRkFEQWMKTVJvd0dBWURWUVFERXhGamRYTjBiMjB0YldWMGNtbGpjeTFqWVRBZUZ3MHlNREF5TURNd05qSTBOVFZhRncweQpNVEF5TURJd05qSTBOVFZhTUJ3eEdqQVlCZ05WQkFNVEVXTjFjM1J2YlMxdFpYUnlhV056TFdOaE1JSUJJakFOCkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXRldldVOVFvN2lGUnlzMS9TTlh2VmFkSGFVSFIKTmVaSW9FQnVvN0RXRFI1cnFlbk9Ubmdpc1BrYzE5VDVTelpLRWtZVk52TnIrZ1RsUEo2OXE4N241VzRlc3UxbgpKa045SVAzWHA4enBtQ0FXTnQvaHFrOFJNZ0hFVml0Vkh6czdEVS9UMW5DRnAwRmltQTVUUkVNK1hyUU1EVFBmCkNVYnUycFh5UU44WHkrQmdHTWJDcXpIUmZ1Z1YrbHJDMHRxZzd2djZDWGRmdExyM3dvdlRKQXVOcjhjdkRUaEkKc3dYY3ZGdXFrdkhvM3hEWlgrQksyeTMwd1NmUnFoYWMwWjZGNlVESWdqbFRka0w0N0N1c2RISDNWeUZKczFxUApTLzRuN29iV2FsZG5DSDFUWlZtZmZIc3EvTjZJZ0F6TDE3RGlLUS9yWG9GV3RxZEM3Vmhtb3JnT093SURBUUFCCm8wSXdRREFPQmdOVkhROEJBZjhFQkFNQ0FxUXdIUVlEVlIwbEJCWXdGQVlJS3dZQkJRVUhBd0VHQ0NzR0FRVUYKQndNQ01BOEdBMVVkRXdFQi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFDZHh2aElVR2ZWMQpKRHc4eTdpUUt5NmRZMFc1L3hrR2JFeDc1YXNpSzA3SmJPYnFQeGJlUVJ3RGRYcm02U2JSVUphbjFkMmFYWjlTCmVHejhRUVIvUzVyQmFnN0lMUlREdmx6bmoxMnJ3RzVCZTRkYlJES25KNWdGdUU4M0hMYTZneU1rZ0NYUjNGenYKWG84eTA0L09yVGRGRlhLVWNnRHlBK2dQQlFkNFFSNUtBdVN4Z1dCR2RjbmUwdkNyS3FsRERhVGtMd3ZWVDE1ZgpvaWpzQ01GQ1drY0pzVEtyVzcxZ1cxZ0pmcHEzS0VNN1YxNERIbHVkSlJ0bXM2SlBITHFNdS8zWEpCWENzNTdsCkdVYTRCZnUzTVVmQmdESTFjYThnYzluT3Y2NTZrL2lBTnZuRHQyZGdBeDJscDVLbEp3cWRMMTVZN0llUVJieVAKb3J6WndESGIwTEE9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K'
    service:
      name: seldon-webhook-service
      namespace: 'kubeflow'
      path: /validate-machinelearning-seldon-io-v1alpha2-seldondeployment
  failurePolicy: Fail
  name: vseldondeployment.kb.io
  namespaceSelector:
    matchExpressions:
    - key: seldon.io/controller-id
      operator: DoesNotExist
    matchLabels:
      serving.kubeflow.org/inferenceservice: enabled
  rules:
  - apiGroups:
    - machinelearning.seldon.io
    apiVersions:
    - v1alpha2
    operations:
    - CREATE
    - UPDATE
    resources:
    - seldondeployments
- clientConfig:
    caBundle: 'LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURCRENDQWV5Z0F3SUJBZ0lRREpTUVN2d0RJbzdaTCtybngwNmJSakFOQmdrcWhraUc5dzBCQVFzRkFEQWMKTVJvd0dBWURWUVFERXhGamRYTjBiMjB0YldWMGNtbGpjeTFqWVRBZUZ3MHlNREF5TURNd05qSTBOVFZhRncweQpNVEF5TURJd05qSTBOVFZhTUJ3eEdqQVlCZ05WQkFNVEVXTjFjM1J2YlMxdFpYUnlhV056TFdOaE1JSUJJakFOCkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXRldldVOVFvN2lGUnlzMS9TTlh2VmFkSGFVSFIKTmVaSW9FQnVvN0RXRFI1cnFlbk9Ubmdpc1BrYzE5VDVTelpLRWtZVk52TnIrZ1RsUEo2OXE4N241VzRlc3UxbgpKa045SVAzWHA4enBtQ0FXTnQvaHFrOFJNZ0hFVml0Vkh6czdEVS9UMW5DRnAwRmltQTVUUkVNK1hyUU1EVFBmCkNVYnUycFh5UU44WHkrQmdHTWJDcXpIUmZ1Z1YrbHJDMHRxZzd2djZDWGRmdExyM3dvdlRKQXVOcjhjdkRUaEkKc3dYY3ZGdXFrdkhvM3hEWlgrQksyeTMwd1NmUnFoYWMwWjZGNlVESWdqbFRka0w0N0N1c2RISDNWeUZKczFxUApTLzRuN29iV2FsZG5DSDFUWlZtZmZIc3EvTjZJZ0F6TDE3RGlLUS9yWG9GV3RxZEM3Vmhtb3JnT093SURBUUFCCm8wSXdRREFPQmdOVkhROEJBZjhFQkFNQ0FxUXdIUVlEVlIwbEJCWXdGQVlJS3dZQkJRVUhBd0VHQ0NzR0FRVUYKQndNQ01BOEdBMVVkRXdFQi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFDZHh2aElVR2ZWMQpKRHc4eTdpUUt5NmRZMFc1L3hrR2JFeDc1YXNpSzA3SmJPYnFQeGJlUVJ3RGRYcm02U2JSVUphbjFkMmFYWjlTCmVHejhRUVIvUzVyQmFnN0lMUlREdmx6bmoxMnJ3RzVCZTRkYlJES25KNWdGdUU4M0hMYTZneU1rZ0NYUjNGenYKWG84eTA0L09yVGRGRlhLVWNnRHlBK2dQQlFkNFFSNUtBdVN4Z1dCR2RjbmUwdkNyS3FsRERhVGtMd3ZWVDE1ZgpvaWpzQ01GQ1drY0pzVEtyVzcxZ1cxZ0pmcHEzS0VNN1YxNERIbHVkSlJ0bXM2SlBITHFNdS8zWEpCWENzNTdsCkdVYTRCZnUzTVVmQmdESTFjYThnYzluT3Y2NTZrL2lBTnZuRHQyZGdBeDJscDVLbEp3cWRMMTVZN0llUVJieVAKb3J6WndESGIwTEE9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K'
    service:
      name: seldon-webhook-service
      namespace: 'kubeflow'
      path: /validate-machinelearning-seldon-io-v1alpha3-seldondeployment
  failurePolicy: Fail
  name: vseldondeployment.kb.io
  namespaceSelector:
    matchExpressions:
    - key: seldon.io/controller-id
      operator: DoesNotExist
    matchLabels:
      serving.kubeflow.org/inferenceservice: enabled
  rules:
  - apiGroups:
    - machinelearning.seldon.io
    apiVersions:
    - v1alpha3
    operations:
    - CREATE
    - UPDATE
    resources:
    - seldondeployments
`)
	th.writeK("/manifests/seldon/seldon-core-operator/base", `
# List of resource files that kustomize reads, modifies
# and emits as a YAML string
resources:
- resources.yaml

`)
}

func TestSeldonCoreOperatorOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/seldon/seldon-core-operator/overlays/application")
	writeSeldonCoreOperatorOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../seldon/seldon-core-operator/overlays/application"
	fsys := fs.MakeRealFS()
	lrc := loader.RestrictionRootOnly
	_loader, loaderErr := loader.NewLoader(lrc, validators.MakeFakeValidator(), targetPath, fsys)
	if loaderErr != nil {
		t.Fatalf("could not load kustomize loader: %v", loaderErr)
	}
	rf := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())
	pc := plugins.DefaultPluginConfig()
	kt, err := target.NewKustTarget(_loader, rf, transformer.NewFactoryImpl(), plugins.NewLoader(pc, rf))
	if err != nil {
		th.t.Fatalf("Unexpected construction error %v", err)
	}
	actual, err := kt.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(actual, string(expected))
}
