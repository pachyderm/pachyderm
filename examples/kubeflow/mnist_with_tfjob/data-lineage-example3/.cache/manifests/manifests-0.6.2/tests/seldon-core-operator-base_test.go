package tests_test

import (
	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/target"
	"testing"
)

func writeSeldonCoreOperatorBase(th *KustTestHarness) {
	th.writeF("/manifests/seldon/seldon-core-operator/base/clusterrole.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: seldon-operator-manager-role
rules:
- apiGroups:
  - apps
  resources:
  - deployments
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - apps
  resources:
  - deployments/status
  verbs:
  - get
  - update
  - patch
- apiGroups:
  - v1
  resources:
  - services
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - v1
  resources:
  - services/status
  verbs:
  - get
  - update
  - patch
- apiGroups:
  - networking.istio.io
  resources:
  - virtualservices
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - networking.istio.io
  resources:
  - virtualservices/status
  verbs:
  - get
  - update
  - patch
- apiGroups:
  - networking.istio.io
  resources:
  - destinationrules
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - networking.istio.io
  resources:
  - destinationrules/status
  verbs:
  - get
  - update
  - patch
- apiGroups:
  - autoscaling
  resources:
  - horizontalpodautoscalers
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - autoscaling
  resources:
  - horizontalpodautoscalers/status
  verbs:
  - get
  - update
  - patch
- apiGroups:
  - machinelearning.seldon.io
  resources:
  - seldondeployments
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - machinelearning.seldon.io
  resources:
  - seldondeployments/status
  verbs:
  - get
  - update
  - patch
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - mutatingwebhookconfigurations
  - validatingwebhookconfigurations
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
`)
	th.writeF("/manifests/seldon/seldon-core-operator/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: seldondeployments.machinelearning.seldon.io
spec:
  group: machinelearning.seldon.io
  names:
    kind: SeldonDeployment
    plural: seldondeployments
    shortNames:
    - sdep
    singular: seldondeployment
  scope: Namespaced
  validation:
    openAPIV3Schema:
      properties:
        spec:
          properties:
            annotations:
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
                                properties:
                                  external:
                                    properties:
                                      metricName:
                                        type: string
                                      metricSelector:
                                        properties:
                                          matchExpressions:
                                            items:
                                              properties:
                                                key:
                                                  type: string
                                                  x-kubernetes-patch-merge-key: key
                                                  x-kubernetes-patch-strategy: merge
                                                operator:
                                                  type: string
                                                values:
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
                                            type: object
                                        type: object
                                      targetAverageValue:
                                        type: string
                                      targetValue:
                                        type: string
                                    required:
                                    - metricName
                                    type: object
                                  object:
                                    properties:
                                      averageValue:
                                        type: string
                                      metricName:
                                        type: string
                                      selector:
                                        properties:
                                          matchExpressions:
                                            items:
                                              properties:
                                                key:
                                                  type: string
                                                  x-kubernetes-patch-merge-key: key
                                                  x-kubernetes-patch-strategy: merge
                                                operator:
                                                  type: string
                                                values:
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
                                            type: object
                                        type: object
                                      target:
                                        properties:
                                          apiVersion:
                                            type: string
                                          kind:
                                            type: string
                                          name:
                                            type: string
                                        required:
                                        - kind
                                        - name
                                        type: object
                                      targetValue:
                                        type: string
                                    required:
                                    - target
                                    - metricName
                                    - targetValue
                                    type: object
                                  pods:
                                    properties:
                                      metricName:
                                        type: string
                                      selector:
                                        properties:
                                          matchExpressions:
                                            items:
                                              properties:
                                                key:
                                                  type: string
                                                  x-kubernetes-patch-merge-key: key
                                                  x-kubernetes-patch-strategy: merge
                                                operator:
                                                  type: string
                                                values:
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
                                            type: object
                                        type: object
                                      targetAverageValue:
                                        type: string
                                    required:
                                    - metricName
                                    - targetAverageValue
                                    type: object
                                  resource:
                                    properties:
                                      name:
                                        type: string
                                      targetAverageUtilization:
                                        format: int32
                                        type: integer
                                      targetAverageValue:
                                        type: string
                                    required:
                                    - name
                                    type: object
                                  type:
                                    type: string
                                required:
                                - type
                                type: object
                              type: array
                            minReplicas:
                              format: int32
                              type: integer
                          type: object
                        metadata:
                          properties:
                            annotations:
                              additionalProperties:
                                type: string
                              type: object
                            clusterName:
                              type: string
                            creationTimestamp:
                              format: date-time
                              type: string
                            deletionGracePeriodSeconds:
                              format: int64
                              type: integer
                            deletionTimestamp:
                              format: date-time
                              type: string
                            finalizers:
                              items:
                                type: string
                              type: array
                              x-kubernetes-patch-strategy: merge
                            generateName:
                              type: string
                            generation:
                              format: int64
                              type: integer
                            initializers:
                              properties:
                                pending:
                                  items:
                                    properties:
                                      name:
                                        type: string
                                    required:
                                    - name
                                    type: object
                                  type: array
                                  x-kubernetes-patch-merge-key: name
                                  x-kubernetes-patch-strategy: merge
                                result:
                                  properties:
                                    apiVersion:
                                      type: string
                                    code:
                                      format: int32
                                      type: integer
                                    details:
                                      properties:
                                        causes:
                                          items:
                                            properties:
                                              field:
                                                type: string
                                              message:
                                                type: string
                                              reason:
                                                type: string
                                            type: object
                                          type: array
                                        group:
                                          type: string
                                        kind:
                                          type: string
                                        name:
                                          type: string
                                        retryAfterSeconds:
                                          format: int32
                                          type: integer
                                        uid:
                                          type: string
                                      type: object
                                    kind:
                                      type: string
                                    message:
                                      type: string
                                    metadata:
                                      properties:
                                        continue:
                                          type: string
                                        remainingItemCount:
                                          format: int64
                                          type: integer
                                        resourceVersion:
                                          type: string
                                        selfLink:
                                          type: string
                                      type: object
                                    reason:
                                      type: string
                                    status:
                                      type: string
                                  type: object
                                  x-kubernetes-group-version-kind:
                                  - group: ""
                                    kind: Status
                                    version: v1
                              required:
                              - pending
                              type: object
                            labels:
                              additionalProperties:
                                type: string
                              type: object
                            managedFields:
                              items:
                                properties:
                                  apiVersion:
                                    type: string
                                  fields:
                                    type: object
                                  manager:
                                    type: string
                                  operation:
                                    type: string
                                  time:
                                    format: date-time
                                    type: string
                                type: object
                              type: array
                            name:
                              type: string
                            namespace:
                              type: string
                            ownerReferences:
                              items:
                                properties:
                                  apiVersion:
                                    type: string
                                  blockOwnerDeletion:
                                    type: boolean
                                  controller:
                                    type: boolean
                                  kind:
                                    type: string
                                  name:
                                    type: string
                                  uid:
                                    type: string
                                required:
                                - apiVersion
                                - kind
                                - name
                                - uid
                                type: object
                              type: array
                              x-kubernetes-patch-merge-key: uid
                              x-kubernetes-patch-strategy: merge
                            resourceVersion:
                              type: string
                            selfLink:
                              type: string
                            uid:
                              type: string
                          type: object
                        spec:
                          properties:
                            activeDeadlineSeconds:
                              format: int64
                              type: integer
                            affinity:
                              properties:
                                nodeAffinity:
                                  properties:
                                    preferredDuringSchedulingIgnoredDuringExecution:
                                      items:
                                        properties:
                                          preference:
                                            properties:
                                              matchExpressions:
                                                items:
                                                  properties:
                                                    key:
                                                      type: string
                                                    operator:
                                                      type: string
                                                    values:
                                                      items:
                                                        type: string
                                                      type: array
                                                  required:
                                                  - key
                                                  - operator
                                                  type: object
                                                type: array
                                              matchFields:
                                                items:
                                                  properties:
                                                    key:
                                                      type: string
                                                    operator:
                                                      type: string
                                                    values:
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
                                            format: int32
                                            type: integer
                                        required:
                                        - weight
                                        - preference
                                        type: object
                                      type: array
                                    requiredDuringSchedulingIgnoredDuringExecution:
                                      properties:
                                        nodeSelectorTerms:
                                          items:
                                            properties:
                                              matchExpressions:
                                                items:
                                                  properties:
                                                    key:
                                                      type: string
                                                    operator:
                                                      type: string
                                                    values:
                                                      items:
                                                        type: string
                                                      type: array
                                                  required:
                                                  - key
                                                  - operator
                                                  type: object
                                                type: array
                                              matchFields:
                                                items:
                                                  properties:
                                                    key:
                                                      type: string
                                                    operator:
                                                      type: string
                                                    values:
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
                                  properties:
                                    preferredDuringSchedulingIgnoredDuringExecution:
                                      items:
                                        properties:
                                          podAffinityTerm:
                                            properties:
                                              labelSelector:
                                                properties:
                                                  matchExpressions:
                                                    items:
                                                      properties:
                                                        key:
                                                          type: string
                                                          x-kubernetes-patch-merge-key: key
                                                          x-kubernetes-patch-strategy: merge
                                                        operator:
                                                          type: string
                                                        values:
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
                                                    type: object
                                                type: object
                                              namespaces:
                                                items:
                                                  type: string
                                                type: array
                                              topologyKey:
                                                type: string
                                            required:
                                            - topologyKey
                                            type: object
                                          weight:
                                            format: int32
                                            type: integer
                                        required:
                                        - weight
                                        - podAffinityTerm
                                        type: object
                                      type: array
                                    requiredDuringSchedulingIgnoredDuringExecution:
                                      items:
                                        properties:
                                          labelSelector:
                                            properties:
                                              matchExpressions:
                                                items:
                                                  properties:
                                                    key:
                                                      type: string
                                                      x-kubernetes-patch-merge-key: key
                                                      x-kubernetes-patch-strategy: merge
                                                    operator:
                                                      type: string
                                                    values:
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
                                                type: object
                                            type: object
                                          namespaces:
                                            items:
                                              type: string
                                            type: array
                                          topologyKey:
                                            type: string
                                        required:
                                        - topologyKey
                                        type: object
                                      type: array
                                  type: object
                                podAntiAffinity:
                                  properties:
                                    preferredDuringSchedulingIgnoredDuringExecution:
                                      items:
                                        properties:
                                          podAffinityTerm:
                                            properties:
                                              labelSelector:
                                                properties:
                                                  matchExpressions:
                                                    items:
                                                      properties:
                                                        key:
                                                          type: string
                                                          x-kubernetes-patch-merge-key: key
                                                          x-kubernetes-patch-strategy: merge
                                                        operator:
                                                          type: string
                                                        values:
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
                                                    type: object
                                                type: object
                                              namespaces:
                                                items:
                                                  type: string
                                                type: array
                                              topologyKey:
                                                type: string
                                            required:
                                            - topologyKey
                                            type: object
                                          weight:
                                            format: int32
                                            type: integer
                                        required:
                                        - weight
                                        - podAffinityTerm
                                        type: object
                                      type: array
                                    requiredDuringSchedulingIgnoredDuringExecution:
                                      items:
                                        properties:
                                          labelSelector:
                                            properties:
                                              matchExpressions:
                                                items:
                                                  properties:
                                                    key:
                                                      type: string
                                                      x-kubernetes-patch-merge-key: key
                                                      x-kubernetes-patch-strategy: merge
                                                    operator:
                                                      type: string
                                                    values:
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
                                                type: object
                                            type: object
                                          namespaces:
                                            items:
                                              type: string
                                            type: array
                                          topologyKey:
                                            type: string
                                        required:
                                        - topologyKey
                                        type: object
                                      type: array
                                  type: object
                              type: object
                            automountServiceAccountToken:
                              type: boolean
                            containers:
                              items:
                                properties:
                                  args:
                                    items:
                                      type: string
                                    type: array
                                  command:
                                    items:
                                      type: string
                                    type: array
                                  env:
                                    items:
                                      properties:
                                        name:
                                          type: string
                                        value:
                                          type: string
                                        valueFrom:
                                          properties:
                                            configMapKeyRef:
                                              properties:
                                                key:
                                                  type: string
                                                name:
                                                  type: string
                                                optional:
                                                  type: boolean
                                              required:
                                              - key
                                              type: object
                                            fieldRef:
                                              properties:
                                                apiVersion:
                                                  type: string
                                                fieldPath:
                                                  type: string
                                              required:
                                              - fieldPath
                                              type: object
                                            resourceFieldRef:
                                              properties:
                                                containerName:
                                                  type: string
                                                divisor:
                                                  type: string
                                                resource:
                                                  type: string
                                              required:
                                              - resource
                                              type: object
                                            secretKeyRef:
                                              properties:
                                                key:
                                                  type: string
                                                name:
                                                  type: string
                                                optional:
                                                  type: boolean
                                              required:
                                              - key
                                              type: object
                                          type: object
                                      required:
                                      - name
                                      type: object
                                    type: array
                                    x-kubernetes-patch-merge-key: name
                                    x-kubernetes-patch-strategy: merge
                                  envFrom:
                                    items:
                                      properties:
                                        configMapRef:
                                          properties:
                                            name:
                                              type: string
                                            optional:
                                              type: boolean
                                          type: object
                                        prefix:
                                          type: string
                                        secretRef:
                                          properties:
                                            name:
                                              type: string
                                            optional:
                                              type: boolean
                                          type: object
                                      type: object
                                    type: array
                                  image:
                                    type: string
                                  imagePullPolicy:
                                    type: string
                                  lifecycle:
                                    properties:
                                      postStart:
                                        properties:
                                          exec:
                                            properties:
                                              command:
                                                items:
                                                  type: string
                                                type: array
                                            type: object
                                          httpGet:
                                            properties:
                                              host:
                                                type: string
                                              httpHeaders:
                                                items:
                                                  properties:
                                                    name:
                                                      type: string
                                                    value:
                                                      type: string
                                                  required:
                                                  - name
                                                  - value
                                                  type: object
                                                type: array
                                              path:
                                                type: string
                                              port:
                                                format: int-or-string
                                                type: string
                                              scheme:
                                                type: string
                                            required:
                                            - port
                                            type: object
                                          tcpSocket:
                                            properties:
                                              host:
                                                type: string
                                              port:
                                                format: int-or-string
                                                type: string
                                            required:
                                            - port
                                            type: object
                                        type: object
                                      preStop:
                                        properties:
                                          exec:
                                            properties:
                                              command:
                                                items:
                                                  type: string
                                                type: array
                                            type: object
                                          httpGet:
                                            properties:
                                              host:
                                                type: string
                                              httpHeaders:
                                                items:
                                                  properties:
                                                    name:
                                                      type: string
                                                    value:
                                                      type: string
                                                  required:
                                                  - name
                                                  - value
                                                  type: object
                                                type: array
                                              path:
                                                type: string
                                              port:
                                                format: int-or-string
                                                type: string
                                              scheme:
                                                type: string
                                            required:
                                            - port
                                            type: object
                                          tcpSocket:
                                            properties:
                                              host:
                                                type: string
                                              port:
                                                format: int-or-string
                                                type: string
                                            required:
                                            - port
                                            type: object
                                        type: object
                                    type: object
                                  livenessProbe:
                                    properties:
                                      exec:
                                        properties:
                                          command:
                                            items:
                                              type: string
                                            type: array
                                        type: object
                                      failureThreshold:
                                        format: int32
                                        type: integer
                                      httpGet:
                                        properties:
                                          host:
                                            type: string
                                          httpHeaders:
                                            items:
                                              properties:
                                                name:
                                                  type: string
                                                value:
                                                  type: string
                                              required:
                                              - name
                                              - value
                                              type: object
                                            type: array
                                          path:
                                            type: string
                                          port:
                                            format: int-or-string
                                            type: string
                                          scheme:
                                            type: string
                                        required:
                                        - port
                                        type: object
                                      initialDelaySeconds:
                                        format: int32
                                        type: integer
                                      periodSeconds:
                                        format: int32
                                        type: integer
                                      successThreshold:
                                        format: int32
                                        type: integer
                                      tcpSocket:
                                        properties:
                                          host:
                                            type: string
                                          port:
                                            format: int-or-string
                                            type: string
                                        required:
                                        - port
                                        type: object
                                      timeoutSeconds:
                                        format: int32
                                        type: integer
                                    type: object
                                  name:
                                    type: string
                                  ports:
                                    items:
                                      properties:
                                        containerPort:
                                          format: int32
                                          type: integer
                                        hostIP:
                                          type: string
                                        hostPort:
                                          format: int32
                                          type: integer
                                        name:
                                          type: string
                                        protocol:
                                          type: string
                                      required:
                                      - containerPort
                                      type: object
                                    type: array
                                    x-kubernetes-list-map-keys:
                                    - containerPort
                                    - protocol
                                    x-kubernetes-list-type: map
                                    x-kubernetes-patch-merge-key: containerPort
                                    x-kubernetes-patch-strategy: merge
                                  readinessProbe:
                                    properties:
                                      exec:
                                        properties:
                                          command:
                                            items:
                                              type: string
                                            type: array
                                        type: object
                                      failureThreshold:
                                        format: int32
                                        type: integer
                                      httpGet:
                                        properties:
                                          host:
                                            type: string
                                          httpHeaders:
                                            items:
                                              properties:
                                                name:
                                                  type: string
                                                value:
                                                  type: string
                                              required:
                                              - name
                                              - value
                                              type: object
                                            type: array
                                          path:
                                            type: string
                                          port:
                                            format: int-or-string
                                            type: string
                                          scheme:
                                            type: string
                                        required:
                                        - port
                                        type: object
                                      initialDelaySeconds:
                                        format: int32
                                        type: integer
                                      periodSeconds:
                                        format: int32
                                        type: integer
                                      successThreshold:
                                        format: int32
                                        type: integer
                                      tcpSocket:
                                        properties:
                                          host:
                                            type: string
                                          port:
                                            format: int-or-string
                                            type: string
                                        required:
                                        - port
                                        type: object
                                      timeoutSeconds:
                                        format: int32
                                        type: integer
                                    type: object
                                  resources:
                                    properties:
                                      limits:
                                        type: object
                                      requests:
                                        type: object
                                    type: object
                                  securityContext:
                                    properties:
                                      allowPrivilegeEscalation:
                                        type: boolean
                                      capabilities:
                                        properties:
                                          add:
                                            items:
                                              type: string
                                            type: array
                                          drop:
                                            items:
                                              type: string
                                            type: array
                                        type: object
                                      privileged:
                                        type: boolean
                                      procMount:
                                        type: string
                                      readOnlyRootFilesystem:
                                        type: boolean
                                      runAsGroup:
                                        format: int64
                                        type: integer
                                      runAsNonRoot:
                                        type: boolean
                                      runAsUser:
                                        format: int64
                                        type: integer
                                      seLinuxOptions:
                                        properties:
                                          level:
                                            type: string
                                          role:
                                            type: string
                                          type:
                                            type: string
                                          user:
                                            type: string
                                        type: object
                                      windowsOptions:
                                        properties:
                                          gmsaCredentialSpec:
                                            type: string
                                          gmsaCredentialSpecName:
                                            type: string
                                        type: object
                                    type: object
                                  stdin:
                                    type: boolean
                                  stdinOnce:
                                    type: boolean
                                  terminationMessagePath:
                                    type: string
                                  terminationMessagePolicy:
                                    type: string
                                  tty:
                                    type: boolean
                                  volumeDevices:
                                    items:
                                      properties:
                                        devicePath:
                                          type: string
                                        name:
                                          type: string
                                      required:
                                      - name
                                      - devicePath
                                      type: object
                                    type: array
                                    x-kubernetes-patch-merge-key: devicePath
                                    x-kubernetes-patch-strategy: merge
                                  volumeMounts:
                                    items:
                                      properties:
                                        mountPath:
                                          type: string
                                        mountPropagation:
                                          type: string
                                        name:
                                          type: string
                                        readOnly:
                                          type: boolean
                                        subPath:
                                          type: string
                                        subPathExpr:
                                          type: string
                                      required:
                                      - name
                                      - mountPath
                                      type: object
                                    type: array
                                    x-kubernetes-patch-merge-key: mountPath
                                    x-kubernetes-patch-strategy: merge
                                  workingDir:
                                    type: string
                                required:
                                - name
                                type: object
                              type: array
                              x-kubernetes-patch-merge-key: name
                              x-kubernetes-patch-strategy: merge
                            dnsConfig:
                              properties:
                                nameservers:
                                  items:
                                    type: string
                                  type: array
                                options:
                                  items:
                                    properties:
                                      name:
                                        type: string
                                      value:
                                        type: string
                                    type: object
                                  type: array
                                searches:
                                  items:
                                    type: string
                                  type: array
                              type: object
                            dnsPolicy:
                              type: string
                            enableServiceLinks:
                              type: boolean
                            hostAliases:
                              items:
                                properties:
                                  hostnames:
                                    items:
                                      type: string
                                    type: array
                                  ip:
                                    type: string
                                type: object
                              type: array
                              x-kubernetes-patch-merge-key: ip
                              x-kubernetes-patch-strategy: merge
                            hostIPC:
                              type: boolean
                            hostNetwork:
                              type: boolean
                            hostPID:
                              type: boolean
                            hostname:
                              type: string
                            imagePullSecrets:
                              items:
                                properties:
                                  name:
                                    type: string
                                type: object
                              type: array
                              x-kubernetes-patch-merge-key: name
                              x-kubernetes-patch-strategy: merge
                            initContainers:
                              items:
                                properties:
                                  args:
                                    items:
                                      type: string
                                    type: array
                                  command:
                                    items:
                                      type: string
                                    type: array
                                  env:
                                    items:
                                      properties:
                                        name:
                                          type: string
                                        value:
                                          type: string
                                        valueFrom:
                                          properties:
                                            configMapKeyRef:
                                              properties:
                                                key:
                                                  type: string
                                                name:
                                                  type: string
                                                optional:
                                                  type: boolean
                                              required:
                                              - key
                                              type: object
                                            fieldRef:
                                              properties:
                                                apiVersion:
                                                  type: string
                                                fieldPath:
                                                  type: string
                                              required:
                                              - fieldPath
                                              type: object
                                            resourceFieldRef:
                                              properties:
                                                containerName:
                                                  type: string
                                                divisor:
                                                  type: string
                                                resource:
                                                  type: string
                                              required:
                                              - resource
                                              type: object
                                            secretKeyRef:
                                              properties:
                                                key:
                                                  type: string
                                                name:
                                                  type: string
                                                optional:
                                                  type: boolean
                                              required:
                                              - key
                                              type: object
                                          type: object
                                      required:
                                      - name
                                      type: object
                                    type: array
                                    x-kubernetes-patch-merge-key: name
                                    x-kubernetes-patch-strategy: merge
                                  envFrom:
                                    items:
                                      properties:
                                        configMapRef:
                                          properties:
                                            name:
                                              type: string
                                            optional:
                                              type: boolean
                                          type: object
                                        prefix:
                                          type: string
                                        secretRef:
                                          properties:
                                            name:
                                              type: string
                                            optional:
                                              type: boolean
                                          type: object
                                      type: object
                                    type: array
                                  image:
                                    type: string
                                  imagePullPolicy:
                                    type: string
                                  lifecycle:
                                    properties:
                                      postStart:
                                        properties:
                                          exec:
                                            properties:
                                              command:
                                                items:
                                                  type: string
                                                type: array
                                            type: object
                                          httpGet:
                                            properties:
                                              host:
                                                type: string
                                              httpHeaders:
                                                items:
                                                  properties:
                                                    name:
                                                      type: string
                                                    value:
                                                      type: string
                                                  required:
                                                  - name
                                                  - value
                                                  type: object
                                                type: array
                                              path:
                                                type: string
                                              port:
                                                format: int-or-string
                                                type: string
                                              scheme:
                                                type: string
                                            required:
                                            - port
                                            type: object
                                          tcpSocket:
                                            properties:
                                              host:
                                                type: string
                                              port:
                                                format: int-or-string
                                                type: string
                                            required:
                                            - port
                                            type: object
                                        type: object
                                      preStop:
                                        properties:
                                          exec:
                                            properties:
                                              command:
                                                items:
                                                  type: string
                                                type: array
                                            type: object
                                          httpGet:
                                            properties:
                                              host:
                                                type: string
                                              httpHeaders:
                                                items:
                                                  properties:
                                                    name:
                                                      type: string
                                                    value:
                                                      type: string
                                                  required:
                                                  - name
                                                  - value
                                                  type: object
                                                type: array
                                              path:
                                                type: string
                                              port:
                                                format: int-or-string
                                                type: string
                                              scheme:
                                                type: string
                                            required:
                                            - port
                                            type: object
                                          tcpSocket:
                                            properties:
                                              host:
                                                type: string
                                              port:
                                                format: int-or-string
                                                type: string
                                            required:
                                            - port
                                            type: object
                                        type: object
                                    type: object
                                  livenessProbe:
                                    properties:
                                      exec:
                                        properties:
                                          command:
                                            items:
                                              type: string
                                            type: array
                                        type: object
                                      failureThreshold:
                                        format: int32
                                        type: integer
                                      httpGet:
                                        properties:
                                          host:
                                            type: string
                                          httpHeaders:
                                            items:
                                              properties:
                                                name:
                                                  type: string
                                                value:
                                                  type: string
                                              required:
                                              - name
                                              - value
                                              type: object
                                            type: array
                                          path:
                                            type: string
                                          port:
                                            format: int-or-string
                                            type: string
                                          scheme:
                                            type: string
                                        required:
                                        - port
                                        type: object
                                      initialDelaySeconds:
                                        format: int32
                                        type: integer
                                      periodSeconds:
                                        format: int32
                                        type: integer
                                      successThreshold:
                                        format: int32
                                        type: integer
                                      tcpSocket:
                                        properties:
                                          host:
                                            type: string
                                          port:
                                            format: int-or-string
                                            type: string
                                        required:
                                        - port
                                        type: object
                                      timeoutSeconds:
                                        format: int32
                                        type: integer
                                    type: object
                                  name:
                                    type: string
                                  ports:
                                    items:
                                      properties:
                                        containerPort:
                                          format: int32
                                          type: integer
                                        hostIP:
                                          type: string
                                        hostPort:
                                          format: int32
                                          type: integer
                                        name:
                                          type: string
                                        protocol:
                                          type: string
                                      required:
                                      - containerPort
                                      type: object
                                    type: array
                                    x-kubernetes-list-map-keys:
                                    - containerPort
                                    - protocol
                                    x-kubernetes-list-type: map
                                    x-kubernetes-patch-merge-key: containerPort
                                    x-kubernetes-patch-strategy: merge
                                  readinessProbe:
                                    properties:
                                      exec:
                                        properties:
                                          command:
                                            items:
                                              type: string
                                            type: array
                                        type: object
                                      failureThreshold:
                                        format: int32
                                        type: integer
                                      httpGet:
                                        properties:
                                          host:
                                            type: string
                                          httpHeaders:
                                            items:
                                              properties:
                                                name:
                                                  type: string
                                                value:
                                                  type: string
                                              required:
                                              - name
                                              - value
                                              type: object
                                            type: array
                                          path:
                                            type: string
                                          port:
                                            format: int-or-string
                                            type: string
                                          scheme:
                                            type: string
                                        required:
                                        - port
                                        type: object
                                      initialDelaySeconds:
                                        format: int32
                                        type: integer
                                      periodSeconds:
                                        format: int32
                                        type: integer
                                      successThreshold:
                                        format: int32
                                        type: integer
                                      tcpSocket:
                                        properties:
                                          host:
                                            type: string
                                          port:
                                            format: int-or-string
                                            type: string
                                        required:
                                        - port
                                        type: object
                                      timeoutSeconds:
                                        format: int32
                                        type: integer
                                    type: object
                                  resources:
                                    properties:
                                      limits:
                                        type: object
                                      requests:
                                        type: object
                                    type: object
                                  securityContext:
                                    properties:
                                      allowPrivilegeEscalation:
                                        type: boolean
                                      capabilities:
                                        properties:
                                          add:
                                            items:
                                              type: string
                                            type: array
                                          drop:
                                            items:
                                              type: string
                                            type: array
                                        type: object
                                      privileged:
                                        type: boolean
                                      procMount:
                                        type: string
                                      readOnlyRootFilesystem:
                                        type: boolean
                                      runAsGroup:
                                        format: int64
                                        type: integer
                                      runAsNonRoot:
                                        type: boolean
                                      runAsUser:
                                        format: int64
                                        type: integer
                                      seLinuxOptions:
                                        properties:
                                          level:
                                            type: string
                                          role:
                                            type: string
                                          type:
                                            type: string
                                          user:
                                            type: string
                                        type: object
                                      windowsOptions:
                                        properties:
                                          gmsaCredentialSpec:
                                            type: string
                                          gmsaCredentialSpecName:
                                            type: string
                                        type: object
                                    type: object
                                  stdin:
                                    type: boolean
                                  stdinOnce:
                                    type: boolean
                                  terminationMessagePath:
                                    type: string
                                  terminationMessagePolicy:
                                    type: string
                                  tty:
                                    type: boolean
                                  volumeDevices:
                                    items:
                                      properties:
                                        devicePath:
                                          type: string
                                        name:
                                          type: string
                                      required:
                                      - name
                                      - devicePath
                                      type: object
                                    type: array
                                    x-kubernetes-patch-merge-key: devicePath
                                    x-kubernetes-patch-strategy: merge
                                  volumeMounts:
                                    items:
                                      properties:
                                        mountPath:
                                          type: string
                                        mountPropagation:
                                          type: string
                                        name:
                                          type: string
                                        readOnly:
                                          type: boolean
                                        subPath:
                                          type: string
                                        subPathExpr:
                                          type: string
                                      required:
                                      - name
                                      - mountPath
                                      type: object
                                    type: array
                                    x-kubernetes-patch-merge-key: mountPath
                                    x-kubernetes-patch-strategy: merge
                                  workingDir:
                                    type: string
                                required:
                                - name
                                type: object
                              type: array
                              x-kubernetes-patch-merge-key: name
                              x-kubernetes-patch-strategy: merge
                            nodeName:
                              type: string
                            nodeSelector:
                              additionalProperties:
                                type: string
                              type: object
                            overhead:
                              type: object
                            preemptionPolicy:
                              type: string
                            priority:
                              format: int32
                              type: integer
                            priorityClassName:
                              type: string
                            readinessGates:
                              items:
                                properties:
                                  conditionType:
                                    type: string
                                required:
                                - conditionType
                                type: object
                              type: array
                            restartPolicy:
                              type: string
                            runtimeClassName:
                              type: string
                            schedulerName:
                              type: string
                            securityContext:
                              properties:
                                fsGroup:
                                  format: int64
                                  type: integer
                                runAsGroup:
                                  format: int64
                                  type: integer
                                runAsNonRoot:
                                  type: boolean
                                runAsUser:
                                  format: int64
                                  type: integer
                                seLinuxOptions:
                                  properties:
                                    level:
                                      type: string
                                    role:
                                      type: string
                                    type:
                                      type: string
                                    user:
                                      type: string
                                  type: object
                                supplementalGroups:
                                  items:
                                    format: int64
                                    type: integer
                                  type: array
                                sysctls:
                                  items:
                                    properties:
                                      name:
                                        type: string
                                      value:
                                        type: string
                                    required:
                                    - name
                                    - value
                                    type: object
                                  type: array
                                windowsOptions:
                                  properties:
                                    gmsaCredentialSpec:
                                      type: string
                                    gmsaCredentialSpecName:
                                      type: string
                                  type: object
                              type: object
                            serviceAccount:
                              type: string
                            serviceAccountName:
                              type: string
                            shareProcessNamespace:
                              type: boolean
                            subdomain:
                              type: string
                            terminationGracePeriodSeconds:
                              format: int64
                              type: integer
                            tolerations:
                              items:
                                properties:
                                  effect:
                                    type: string
                                  key:
                                    type: string
                                  operator:
                                    type: string
                                  tolerationSeconds:
                                    format: int64
                                    type: integer
                                  value:
                                    type: string
                                type: object
                              type: array
                            volumes:
                              items:
                                properties:
                                  awsElasticBlockStore:
                                    properties:
                                      fsType:
                                        type: string
                                      partition:
                                        format: int32
                                        type: integer
                                      readOnly:
                                        type: boolean
                                      volumeID:
                                        type: string
                                    required:
                                    - volumeID
                                    type: object
                                  azureDisk:
                                    properties:
                                      cachingMode:
                                        type: string
                                      diskName:
                                        type: string
                                      diskURI:
                                        type: string
                                      fsType:
                                        type: string
                                      kind:
                                        type: string
                                      readOnly:
                                        type: boolean
                                    required:
                                    - diskName
                                    - diskURI
                                    type: object
                                  azureFile:
                                    properties:
                                      readOnly:
                                        type: boolean
                                      secretName:
                                        type: string
                                      shareName:
                                        type: string
                                    required:
                                    - secretName
                                    - shareName
                                    type: object
                                  cephfs:
                                    properties:
                                      monitors:
                                        items:
                                          type: string
                                        type: array
                                      path:
                                        type: string
                                      readOnly:
                                        type: boolean
                                      secretFile:
                                        type: string
                                      secretRef:
                                        properties:
                                          name:
                                            type: string
                                        type: object
                                      user:
                                        type: string
                                    required:
                                    - monitors
                                    type: object
                                  cinder:
                                    properties:
                                      fsType:
                                        type: string
                                      readOnly:
                                        type: boolean
                                      secretRef:
                                        properties:
                                          name:
                                            type: string
                                        type: object
                                      volumeID:
                                        type: string
                                    required:
                                    - volumeID
                                    type: object
                                  csi:
                                    properties:
                                      driver:
                                        type: string
                                      fsType:
                                        type: string
                                      nodePublishSecretRef:
                                        properties:
                                          name:
                                            type: string
                                        type: object
                                      readOnly:
                                        type: boolean
                                      volumeAttributes:
                                        additionalProperties:
                                          type: string
                                        type: object
                                    required:
                                    - driver
                                    type: object
                                  emptyDir:
                                    properties:
                                      medium:
                                        type: string
                                      sizeLimit:
                                        type: string
                                    type: object
                                  fc:
                                    properties:
                                      fsType:
                                        type: string
                                      lun:
                                        format: int32
                                        type: integer
                                      readOnly:
                                        type: boolean
                                      targetWWNs:
                                        items:
                                          type: string
                                        type: array
                                      wwids:
                                        items:
                                          type: string
                                        type: array
                                    type: object
                                  flexVolume:
                                    properties:
                                      driver:
                                        type: string
                                      fsType:
                                        type: string
                                      options:
                                        additionalProperties:
                                          type: string
                                        type: object
                                      readOnly:
                                        type: boolean
                                      secretRef:
                                        properties:
                                          name:
                                            type: string
                                        type: object
                                    required:
                                    - driver
                                    type: object
                                  flocker:
                                    properties:
                                      datasetName:
                                        type: string
                                      datasetUUID:
                                        type: string
                                    type: object
                                  gcePersistentDisk:
                                    properties:
                                      fsType:
                                        type: string
                                      partition:
                                        format: int32
                                        type: integer
                                      pdName:
                                        type: string
                                      readOnly:
                                        type: boolean
                                    required:
                                    - pdName
                                    type: object
                                  gitRepo:
                                    properties:
                                      directory:
                                        type: string
                                      repository:
                                        type: string
                                      revision:
                                        type: string
                                    required:
                                    - repository
                                    type: object
                                  glusterfs:
                                    properties:
                                      endpoints:
                                        type: string
                                      path:
                                        type: string
                                      readOnly:
                                        type: boolean
                                    required:
                                    - endpoints
                                    - path
                                    type: object
                                  hostPath:
                                    properties:
                                      path:
                                        type: string
                                      type:
                                        type: string
                                    required:
                                    - path
                                    type: object
                                  iscsi:
                                    properties:
                                      chapAuthDiscovery:
                                        type: boolean
                                      chapAuthSession:
                                        type: boolean
                                      fsType:
                                        type: string
                                      initiatorName:
                                        type: string
                                      iqn:
                                        type: string
                                      iscsiInterface:
                                        type: string
                                      lun:
                                        format: int32
                                        type: integer
                                      portals:
                                        items:
                                          type: string
                                        type: array
                                      readOnly:
                                        type: boolean
                                      secretRef:
                                        properties:
                                          name:
                                            type: string
                                        type: object
                                      targetPortal:
                                        type: string
                                    required:
                                    - targetPortal
                                    - iqn
                                    - lun
                                    type: object
                                  name:
                                    type: string
                                  nfs:
                                    properties:
                                      path:
                                        type: string
                                      readOnly:
                                        type: boolean
                                      server:
                                        type: string
                                    required:
                                    - server
                                    - path
                                    type: object
                                  persistentVolumeClaim:
                                    properties:
                                      claimName:
                                        type: string
                                      readOnly:
                                        type: boolean
                                    required:
                                    - claimName
                                    type: object
                                  photonPersistentDisk:
                                    properties:
                                      fsType:
                                        type: string
                                      pdID:
                                        type: string
                                    required:
                                    - pdID
                                    type: object
                                  portworxVolume:
                                    properties:
                                      fsType:
                                        type: string
                                      readOnly:
                                        type: boolean
                                      volumeID:
                                        type: string
                                    required:
                                    - volumeID
                                    type: object
                                  projected:
                                    properties:
                                      defaultMode:
                                        format: int32
                                        type: integer
                                      sources:
                                        items:
                                          properties:
                                            serviceAccountToken:
                                              properties:
                                                audience:
                                                  type: string
                                                expirationSeconds:
                                                  format: int64
                                                  type: integer
                                                path:
                                                  type: string
                                              required:
                                              - path
                                              type: object
                                          type: object
                                        type: array
                                    required:
                                    - sources
                                    type: object
                                  quobyte:
                                    properties:
                                      group:
                                        type: string
                                      readOnly:
                                        type: boolean
                                      registry:
                                        type: string
                                      tenant:
                                        type: string
                                      user:
                                        type: string
                                      volume:
                                        type: string
                                    required:
                                    - registry
                                    - volume
                                    type: object
                                  rbd:
                                    properties:
                                      fsType:
                                        type: string
                                      image:
                                        type: string
                                      keyring:
                                        type: string
                                      monitors:
                                        items:
                                          type: string
                                        type: array
                                      pool:
                                        type: string
                                      readOnly:
                                        type: boolean
                                      secretRef:
                                        properties:
                                          name:
                                            type: string
                                        type: object
                                      user:
                                        type: string
                                    required:
                                    - monitors
                                    - image
                                    type: object
                                  scaleIO:
                                    properties:
                                      fsType:
                                        type: string
                                      gateway:
                                        type: string
                                      protectionDomain:
                                        type: string
                                      readOnly:
                                        type: boolean
                                      secretRef:
                                        properties:
                                          name:
                                            type: string
                                        type: object
                                      sslEnabled:
                                        type: boolean
                                      storageMode:
                                        type: string
                                      storagePool:
                                        type: string
                                      system:
                                        type: string
                                      volumeName:
                                        type: string
                                    required:
                                    - gateway
                                    - system
                                    - secretRef
                                    type: object
                                  storageos:
                                    properties:
                                      fsType:
                                        type: string
                                      readOnly:
                                        type: boolean
                                      secretRef:
                                        properties:
                                          name:
                                            type: string
                                        type: object
                                      volumeName:
                                        type: string
                                      volumeNamespace:
                                        type: string
                                    type: object
                                  vsphereVolume:
                                    properties:
                                      fsType:
                                        type: string
                                      storagePolicyID:
                                        type: string
                                      storagePolicyName:
                                        type: string
                                      volumePath:
                                        type: string
                                    required:
                                    - volumePath
                                    type: object
                                required:
                                - name
                                type: object
                              type: array
                              x-kubernetes-patch-merge-key: name
                              x-kubernetes-patch-strategy: merge,retainKeys
                          required:
                          - containers
                          type: object
                      type: object
                    type: array
                  graph:
                    properties:
                      children:
                        items:
                          properties:
                            children:
                              items:
                                properties:
                                  children:
                                    type: array
                                  endpoint:
                                    properties:
                                      service_host:
                                        type: string
                                      service_port:
                                        type: integer
                                      type:
                                        enum:
                                        - REST
                                        - GRPC
                                        type: string
                                  implementation:
                                    enum:
                                    - UNKNOWN_IMPLEMENTATION
                                    - SIMPLE_MODEL
                                    - SIMPLE_ROUTER
                                    - RANDOM_ABTEST
                                    - AVERAGE_COMBINER
                                    type: string
                                  methods:
                                    items:
                                      enum:
                                      - TRANSFORM_INPUT
                                      - TRANSFORM_OUTPUT
                                      - ROUTE
                                      - AGGREGATE
                                      - SEND_FEEDBACK
                                      type: string
                                    type: array
                                  name:
                                    type: string
                                  type:
                                    enum:
                                    - UNKNOWN_TYPE
                                    - ROUTER
                                    - COMBINER
                                    - MODEL
                                    - TRANSFORMER
                                    - OUTPUT_TRANSFORMER
                                    type: string
                              type: array
                            endpoint:
                              properties:
                                service_host:
                                  type: string
                                service_port:
                                  type: integer
                                type:
                                  enum:
                                  - REST
                                  - GRPC
                                  type: string
                            implementation:
                              enum:
                              - UNKNOWN_IMPLEMENTATION
                              - SIMPLE_MODEL
                              - SIMPLE_ROUTER
                              - RANDOM_ABTEST
                              - AVERAGE_COMBINER
                              type: string
                            methods:
                              items:
                                enum:
                                - TRANSFORM_INPUT
                                - TRANSFORM_OUTPUT
                                - ROUTE
                                - AGGREGATE
                                - SEND_FEEDBACK
                                type: string
                              type: array
                            name:
                              type: string
                            type:
                              enum:
                              - UNKNOWN_TYPE
                              - ROUTER
                              - COMBINER
                              - MODEL
                              - TRANSFORMER
                              - OUTPUT_TRANSFORMER
                              type: string
                        type: array
                      endpoint:
                        properties:
                          service_host:
                            type: string
                          service_port:
                            type: integer
                          type:
                            enum:
                            - REST
                            - GRPC
                            type: string
                      implementation:
                        enum:
                        - UNKNOWN_IMPLEMENTATION
                        - SIMPLE_MODEL
                        - SIMPLE_ROUTER
                        - RANDOM_ABTEST
                        - AVERAGE_COMBINER
                        type: string
                      methods:
                        items:
                          enum:
                          - TRANSFORM_INPUT
                          - TRANSFORM_OUTPUT
                          - ROUTE
                          - AGGREGATE
                          - SEND_FEEDBACK
                          type: string
                        type: array
                      name:
                        type: string
                      type:
                        enum:
                        - UNKNOWN_TYPE
                        - ROUTER
                        - COMBINER
                        - MODEL
                        - TRANSFORMER
                        - OUTPUT_TRANSFORMER
                        type: string
                  labels:
                    type: object
                  name:
                    type: string
                  replicas:
                    type: integer
                  svcOrchSpec:
                    properties:
                      env:
                        items:
                          properties:
                            name:
                              type: string
                            value:
                              type: string
                            valueFrom:
                              properties:
                                configMapKeyRef:
                                  properties:
                                    key:
                                      type: string
                                    name:
                                      type: string
                                    optional:
                                      type: boolean
                                  required:
                                  - key
                                fieldRef:
                                  properties:
                                    apiVersion:
                                      type: string
                                    fieldPath:
                                      type: string
                                  required:
                                  - fieldPath
                                resourceFieldRef:
                                  properties:
                                    containerName:
                                      type: string
                                    divisor:
                                      type: string
                                    resource:
                                      type: string
                                  required:
                                  - resource
                                secretKeyRef:
                                  properties:
                                    key:
                                      type: string
                                    name:
                                      type: string
                                    optional:
                                      type: boolean
                                  required:
                                  - key
                          required:
                          - name
                        type: array
                        x-kubernetes-patch-merge-key: name
                        x-kubernetes-patch-strategy: merge
                      resources:
                        properties:
                          limits:
                            additionalProperties: true
                            type: object
                          requests:
                            additionalProperties: true
                            type: object
                    type: object
              type: array
  version: v1alpha2
  versions:
  - name: v1alpha2
    served: true
    storage: true
`)
	th.writeF("/manifests/seldon/seldon-core-operator/base/rolebinding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: seldon-operator-manager-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: seldon-operator-manager-role
subjects:
- kind: ServiceAccount
  name: seldon-manager
  namespace: kubeflow
`)
	th.writeF("/manifests/seldon/seldon-core-operator/base/secret.yaml", `
apiVersion: v1
kind: Secret
metadata:
  name: seldon-operator-webhook-server-secret
`)
	th.writeF("/manifests/seldon/seldon-core-operator/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: seldon-manager
`)
	th.writeF("/manifests/seldon/seldon-core-operator/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    control-plane: seldon-controller-manager
    controller-tools.k8s.io: "1.0"
  name: seldon-operator-controller-manager-service
spec:
  ports:
  - port: 443
  selector:
    control-plane: seldon-controller-manager
    controller-tools.k8s.io: "1.0"
`)
	th.writeF("/manifests/seldon/seldon-core-operator/base/statefulset.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    control-plane: seldon-controller-manager
    controller-tools.k8s.io: "1.0"
  name: seldon-operator-controller-manager
spec:
  selector:
    matchLabels:
      app.kubernetes.io/instance: seldon-core-operator
      app.kubernetes.io/name: seldon-core-operator
      control-plane: seldon-controller-manager
      controller-tools.k8s.io: "1.0"
  serviceName: seldon-operator-controller-manager-service
  template:
    metadata:
      annotations:
        prometheus.io/scrape: "true"
      labels:
        app.kubernetes.io/instance: seldon-core-operator
        app.kubernetes.io/name: seldon-core-operator
        control-plane: seldon-controller-manager
        controller-tools.k8s.io: "1.0"
    spec:
      containers:
      - command:
        - /manager
        env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: SECRET_NAME
          value: seldon-operator-webhook-server-secret
        - name: AMBASSADOR_ENABLED
          value: "true"
        - name: AMBASSADOR_SINGLE_NAMESPACE
          value: "false"
        - name: ENGINE_CONTAINER_IMAGE_AND_VERSION
          value: docker.io/seldonio/engine:0.3.0
        - name: ENGINE_CONTAINER_IMAGE_PULL_POLICY
          value: IfNotPresent
        - name: ENGINE_CONTAINER_SERVICE_ACCOUNT_NAME
          value: default
        - name: ENGINE_CONTAINER_USER
          value: "8888"
        - name: PREDICTIVE_UNIT_SERVICE_PORT
          value: "9000"
        - name: ENGINE_SERVER_GRPC_PORT
          value: "5001"
        - name: ENGINE_SERVER_PORT
          value: "8000"
        - name: ENGINE_PROMETHEUS_PATH
          value: prometheus
        - name: ISTIO_ENABLED
          value: "true"
        - name: ISTIO_GATEWAY
          value: kubeflow-gateway
        image: docker.io/seldonio/seldon-core-operator:0.3.1
        imagePullPolicy: Always
        name: manager
        ports:
        - containerPort: 8080
          name: metrics
          protocol: TCP
        - containerPort: 9876
          name: webhook-server
          protocol: TCP
        resources:
          requests:
            cpu: 100m
            memory: 20Mi
        volumeMounts:
        - mountPath: /tmp/cert
          name: cert
          readOnly: true
      serviceAccountName: seldon-manager
      terminationGracePeriodSeconds: 10
      volumes:
      - name: cert
        secret:
          defaultMode: 420
          secretName: seldon-operator-webhook-server-secret
  volumeClaimTemplates: []
`)
	th.writeK("/manifests/seldon/seldon-core-operator/base", `
resources:
- clusterrole.yaml
- crd.yaml
- rolebinding.yaml
- secret.yaml
- service-account.yaml
- service.yaml
- statefulset.yaml
`)
}

func TestSeldonCoreOperatorBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/seldon/seldon-core-operator/base")
	writeSeldonCoreOperatorBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../seldon/seldon-core-operator/base"
	fsys := fs.MakeRealFS()
	_loader, loaderErr := loader.NewLoader(targetPath, fsys)
	if loaderErr != nil {
		t.Fatalf("could not load kustomize loader: %v", loaderErr)
	}
	rf := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()))
	kt, err := target.NewKustTarget(_loader, rf, transformer.NewFactoryImpl())
	if err != nil {
		th.t.Fatalf("Unexpected construction error %v", err)
	}
	n, err := kt.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := n.EncodeAsYaml()
	th.assertActualEqualsExpected(m, string(expected))
}
