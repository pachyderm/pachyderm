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

func writeSparkOperatorBase(th *KustTestHarness) {
	th.writeF("/manifests/spark/spark-operator/base/spark-sa.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: spark-sa
`)
	th.writeF("/manifests/spark/spark-operator/base/cr-clusterrole.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: operator-cr
rules:
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - services
  - configmaps
  - secrets
  verbs:
  - create
  - get
  - delete
  - update
- apiGroups:
  - extensions
  - networking.k8s.io
  resources:
  - ingresses
  verbs:
  - create
  - get
  - delete
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - update
  - patch
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - create
  - get
  - update
  - delete
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - mutatingwebhookconfigurations
  verbs:
  - create
  - get
  - update
  - delete
- apiGroups:
  - sparkoperator.k8s.io
  resources:
  - sparkapplications
  - scheduledsparkapplications
  verbs:
  - '*'
`)
	th.writeF("/manifests/spark/spark-operator/base/crb.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: sparkoperator-crb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: operator-cr
subjects:
- kind: ServiceAccount
  name: operator-sa
`)
	th.writeF("/manifests/spark/spark-operator/base/crd-cleanup-job.yaml", `
apiVersion: batch/v1
kind: Job
metadata:
  name: crd-cleanup
  namespace: default
spec:
  template:
    spec:
      containers:
      - command:
        - /bin/sh
        - -c
        - 'curl -ik -X DELETE -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)"
          -H "Accept: application/json" -H "Content-Type: application/json" https://kubernetes.default.svc/apis/apiextensions.k8s.io/v1beta1/customresourcedefinitions/sparkapplications.sparkoperator.k8s.io'
        image: gcr.io/spark-operator/spark-operator:v1beta2-1.0.0-2.4.4
        imagePullPolicy: IfNotPresent
        name: delete-sparkapp-crd
      - command:
        - /bin/sh
        - -c
        - 'curl -ik -X DELETE -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)"
          -H "Accept: application/json" -H "Content-Type: application/json" https://kubernetes.default.svc/apis/apiextensions.k8s.io/v1beta1/customresourcedefinitions/scheduledsparkapplications.sparkoperator.k8s.io'
        image: gcr.io/spark-operator/spark-operator:v1beta2-1.0.0-2.4.4
        imagePullPolicy: IfNotPresent
        name: delete-scheduledsparkapp-crd
      restartPolicy: OnFailure
      serviceAccountName: operator-sa
`)
	th.writeF("/manifests/spark/spark-operator/base/deploy.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sparkoperator
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: sparkoperator
      app.kubernetes.io/version: v1beta2-1.0.0-2.4.4
  strategy:
    type: Recreate
  template:
    metadata:
      annotations:
        prometheus.io/path: /metrics
        prometheus.io/port: "10254"
        prometheus.io/scrape: "true"
      initializers:
        pending: []
      labels:
        app.kubernetes.io/name: sparkoperator
        app.kubernetes.io/version: v1beta2-1.0.0-2.4.4
    spec:
      containers:
      - args:
        - -v=2
        - -namespace=
        - -ingress-url-format=
        - -controller-threads=10
        - -resync-interval=30
        - -logtostderr
        - -enable-metrics=true
        - -metrics-labels=app_type
        - -metrics-port=10254
        - -metrics-endpoint=/metrics
        - -metrics-prefix=
        image: gcr.io/spark-operator/spark-operator:v1beta2-1.0.0-2.4.4
        imagePullPolicy: IfNotPresent
        name: sparkoperator
        ports:
        - containerPort: 10254
      serviceAccountName: operator-sa
`)
	th.writeF("/manifests/spark/spark-operator/base/operator-sa.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: operator-sa
`)
	th.writeF("/manifests/spark/spark-operator/base/sparkapplications.sparkoperator.k8s.io-crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: sparkapplications.sparkoperator.k8s.io
spec:
  group: sparkoperator.k8s.io
  names:
    kind: SparkApplication
    listKind: SparkApplicationList
    plural: sparkapplications
    shortNames:
    - sparkapp
    singular: sparkapplication
  scope: Namespaced
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          type: string
        kind:
          type: string
        metadata:
          type: object
        spec:
          properties:
            arguments:
              items:
                type: string
              type: array
            batchScheduler:
              type: string
            batchSchedulerOptions:
              properties:
                priorityClassName:
                  type: string
                queue:
                  type: string
              type: object
            deps:
              properties:
                downloadTimeout:
                  format: int32
                  minimum: 1
                  type: integer
                files:
                  items:
                    type: string
                  type: array
                filesDownloadDir:
                  type: string
                jars:
                  items:
                    type: string
                  type: array
                jarsDownloadDir:
                  type: string
                maxSimultaneousDownloads:
                  format: int32
                  minimum: 1
                  type: integer
                pyFiles:
                  items:
                    type: string
                  type: array
              type: object
            driver:
              properties:
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
                            - preference
                            - weight
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
                            - podAffinityTerm
                            - weight
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
                            - podAffinityTerm
                            - weight
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
                annotations:
                  additionalProperties:
                    type: string
                  type: object
                configMaps:
                  items:
                    properties:
                      name:
                        type: string
                      path:
                        type: string
                    required:
                    - name
                    - path
                    type: object
                  type: array
                coreLimit:
                  type: string
                cores:
                  format: int32
                  minimum: 1
                  type: integer
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
                envSecretKeyRefs:
                  additionalProperties:
                    properties:
                      key:
                        type: string
                      name:
                        type: string
                    required:
                    - key
                    - name
                    type: object
                  type: object
                envVars:
                  additionalProperties:
                    type: string
                  type: object
                gpu:
                  properties:
                    name:
                      type: string
                    quantity:
                      format: int64
                      type: integer
                  required:
                  - name
                  - quantity
                  type: object
                hostNetwork:
                  type: boolean
                image:
                  type: string
                javaOptions:
                  type: string
                labels:
                  additionalProperties:
                    type: string
                  type: object
                memory:
                  type: string
                memoryOverhead:
                  type: string
                nodeSelector:
                  additionalProperties:
                    type: string
                  type: object
                podName:
                  pattern: '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*'
                  type: string
                schedulerName:
                  type: string
                secrets:
                  items:
                    properties:
                      name:
                        type: string
                      path:
                        type: string
                      secretType:
                        type: string
                    required:
                    - name
                    - path
                    - secretType
                    type: object
                  type: array
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
                  type: object
                serviceAccount:
                  type: string
                sidecars:
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
                                    anyOf:
                                    - type: string
                                    - type: integer
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
                                    anyOf:
                                    - type: string
                                    - type: integer
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
                                    anyOf:
                                    - type: string
                                    - type: integer
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
                                    anyOf:
                                    - type: string
                                    - type: integer
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
                                anyOf:
                                - type: string
                                - type: integer
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
                                anyOf:
                                - type: string
                                - type: integer
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
                                anyOf:
                                - type: string
                                - type: integer
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
                                anyOf:
                                - type: string
                                - type: integer
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
                            additionalProperties:
                              type: string
                            type: object
                          requests:
                            additionalProperties:
                              type: string
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
                          - devicePath
                          - name
                          type: object
                        type: array
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
                          required:
                          - mountPath
                          - name
                          type: object
                        type: array
                      workingDir:
                        type: string
                    required:
                    - name
                    type: object
                  type: array
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
                    required:
                    - mountPath
                    - name
                    type: object
                  type: array
              type: object
            executor:
              properties:
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
                            - preference
                            - weight
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
                            - podAffinityTerm
                            - weight
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
                            - podAffinityTerm
                            - weight
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
                annotations:
                  additionalProperties:
                    type: string
                  type: object
                configMaps:
                  items:
                    properties:
                      name:
                        type: string
                      path:
                        type: string
                    required:
                    - name
                    - path
                    type: object
                  type: array
                coreLimit:
                  type: string
                coreRequest:
                  type: string
                cores:
                  format: int32
                  minimum: 1
                  type: integer
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
                envSecretKeyRefs:
                  additionalProperties:
                    properties:
                      key:
                        type: string
                      name:
                        type: string
                    required:
                    - key
                    - name
                    type: object
                  type: object
                envVars:
                  additionalProperties:
                    type: string
                  type: object
                gpu:
                  properties:
                    name:
                      type: string
                    quantity:
                      format: int64
                      type: integer
                  required:
                  - name
                  - quantity
                  type: object
                hostNetwork:
                  type: boolean
                image:
                  type: string
                instances:
                  format: int32
                  minimum: 1
                  type: integer
                javaOptions:
                  type: string
                labels:
                  additionalProperties:
                    type: string
                  type: object
                memory:
                  type: string
                memoryOverhead:
                  type: string
                nodeSelector:
                  additionalProperties:
                    type: string
                  type: object
                schedulerName:
                  type: string
                secrets:
                  items:
                    properties:
                      name:
                        type: string
                      path:
                        type: string
                      secretType:
                        type: string
                    required:
                    - name
                    - path
                    - secretType
                    type: object
                  type: array
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
                  type: object
                sidecars:
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
                                    anyOf:
                                    - type: string
                                    - type: integer
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
                                    anyOf:
                                    - type: string
                                    - type: integer
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
                                    anyOf:
                                    - type: string
                                    - type: integer
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
                                    anyOf:
                                    - type: string
                                    - type: integer
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
                                anyOf:
                                - type: string
                                - type: integer
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
                                anyOf:
                                - type: string
                                - type: integer
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
                                anyOf:
                                - type: string
                                - type: integer
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
                                anyOf:
                                - type: string
                                - type: integer
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
                            additionalProperties:
                              type: string
                            type: object
                          requests:
                            additionalProperties:
                              type: string
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
                          - devicePath
                          - name
                          type: object
                        type: array
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
                          required:
                          - mountPath
                          - name
                          type: object
                        type: array
                      workingDir:
                        type: string
                    required:
                    - name
                    type: object
                  type: array
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
                    required:
                    - mountPath
                    - name
                    type: object
                  type: array
              type: object
            failureRetries:
              format: int32
              type: integer
            hadoopConf:
              additionalProperties:
                type: string
              type: object
            hadoopConfigMap:
              type: string
            image:
              type: string
            imagePullPolicy:
              type: string
            imagePullSecrets:
              items:
                type: string
              type: array
            initContainerImage:
              type: string
            mainApplicationFile:
              type: string
            mainClass:
              type: string
            memoryOverheadFactor:
              type: string
            mode:
              enum:
              - cluster
              - client
              type: string
            monitoring:
              properties:
                exposeDriverMetrics:
                  type: boolean
                exposeExecutorMetrics:
                  type: boolean
                metricsProperties:
                  type: string
                prometheus:
                  properties:
                    configFile:
                      type: string
                    configuration:
                      type: string
                    jmxExporterJar:
                      type: string
                    port:
                      format: int32
                      maximum: 49151
                      minimum: 1024
                      type: integer
                  required:
                  - jmxExporterJar
                  type: object
              required:
              - exposeDriverMetrics
              - exposeExecutorMetrics
              type: object
            nodeSelector:
              additionalProperties:
                type: string
              type: object
            pythonVersion:
              enum:
              - "2"
              - "3"
              type: string
            restartPolicy:
              properties:
                onFailureRetries:
                  format: int32
                  minimum: 0
                  type: integer
                onFailureRetryInterval:
                  format: int64
                  minimum: 1
                  type: integer
                onSubmissionFailureRetries:
                  format: int32
                  minimum: 0
                  type: integer
                onSubmissionFailureRetryInterval:
                  format: int64
                  minimum: 1
                  type: integer
                type:
                  enum:
                  - Never
                  - Always
                  - OnFailure
                  type: string
              type: object
            retryInterval:
              format: int64
              type: integer
            sparkConf:
              additionalProperties:
                type: string
              type: object
            sparkConfigMap:
              type: string
            sparkVersion:
              type: string
            timeToLiveSeconds:
              format: int64
              type: integer
            type:
              enum:
              - Java
              - Python
              - Scala
              - R
              type: string
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
                  configMap:
                    properties:
                      defaultMode:
                        format: int32
                        type: integer
                      items:
                        items:
                          properties:
                            key:
                              type: string
                            mode:
                              format: int32
                              type: integer
                            path:
                              type: string
                          required:
                          - key
                          - path
                          type: object
                        type: array
                      name:
                        type: string
                      optional:
                        type: boolean
                    type: object
                  downwardAPI:
                    properties:
                      defaultMode:
                        format: int32
                        type: integer
                      items:
                        items:
                          properties:
                            fieldRef:
                              properties:
                                apiVersion:
                                  type: string
                                fieldPath:
                                  type: string
                              required:
                              - fieldPath
                              type: object
                            mode:
                              format: int32
                              type: integer
                            path:
                              type: string
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
                          required:
                          - path
                          type: object
                        type: array
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
                    - iqn
                    - lun
                    - targetPortal
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
                    - path
                    - server
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
                            configMap:
                              properties:
                                items:
                                  items:
                                    properties:
                                      key:
                                        type: string
                                      mode:
                                        format: int32
                                        type: integer
                                      path:
                                        type: string
                                    required:
                                    - key
                                    - path
                                    type: object
                                  type: array
                                name:
                                  type: string
                                optional:
                                  type: boolean
                              type: object
                            downwardAPI:
                              properties:
                                items:
                                  items:
                                    properties:
                                      fieldRef:
                                        properties:
                                          apiVersion:
                                            type: string
                                          fieldPath:
                                            type: string
                                        required:
                                        - fieldPath
                                        type: object
                                      mode:
                                        format: int32
                                        type: integer
                                      path:
                                        type: string
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
                                    required:
                                    - path
                                    type: object
                                  type: array
                              type: object
                            secret:
                              properties:
                                items:
                                  items:
                                    properties:
                                      key:
                                        type: string
                                      mode:
                                        format: int32
                                        type: integer
                                      path:
                                        type: string
                                    required:
                                    - key
                                    - path
                                    type: object
                                  type: array
                                name:
                                  type: string
                                optional:
                                  type: boolean
                              type: object
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
                    - image
                    - monitors
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
                    - secretRef
                    - system
                    type: object
                  secret:
                    properties:
                      defaultMode:
                        format: int32
                        type: integer
                      items:
                        items:
                          properties:
                            key:
                              type: string
                            mode:
                              format: int32
                              type: integer
                            path:
                              type: string
                          required:
                          - key
                          - path
                          type: object
                        type: array
                      optional:
                        type: boolean
                      secretName:
                        type: string
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
          required:
          - driver
          - executor
          - mainApplicationFile
          - sparkVersion
          - type
          type: object
      required:
      - metadata
      - spec
      type: object
  version: v1beta2
  versions:
  - name: v1beta2
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`)
	th.writeF("/manifests/spark/spark-operator/base/scheduledsparkapplications.sparkoperator.k8s.io-crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: scheduledsparkapplications.sparkoperator.k8s.io
spec:
  group: sparkoperator.k8s.io
  names:
    kind: ScheduledSparkApplication
    listKind: ScheduledSparkApplicationList
    plural: scheduledsparkapplications
    shortNames:
    - scheduledsparkapp
    singular: scheduledsparkapplication
  scope: Namespaced
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          type: string
        kind:
          type: string
        metadata:
          type: object
        spec:
          properties:
            concurrencyPolicy:
              type: string
            failedRunHistoryLimit:
              format: int32
              type: integer
            schedule:
              type: string
            successfulRunHistoryLimit:
              format: int32
              type: integer
            suspend:
              type: boolean
            template:
              properties:
                arguments:
                  items:
                    type: string
                  type: array
                batchScheduler:
                  type: string
                batchSchedulerOptions:
                  properties:
                    priorityClassName:
                      type: string
                    queue:
                      type: string
                  type: object
                deps:
                  properties:
                    downloadTimeout:
                      format: int32
                      minimum: 1
                      type: integer
                    files:
                      items:
                        type: string
                      type: array
                    filesDownloadDir:
                      type: string
                    jars:
                      items:
                        type: string
                      type: array
                    jarsDownloadDir:
                      type: string
                    maxSimultaneousDownloads:
                      format: int32
                      minimum: 1
                      type: integer
                    pyFiles:
                      items:
                        type: string
                      type: array
                  type: object
                driver:
                  properties:
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
                                - preference
                                - weight
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
                                - podAffinityTerm
                                - weight
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
                                - podAffinityTerm
                                - weight
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
                    annotations:
                      additionalProperties:
                        type: string
                      type: object
                    configMaps:
                      items:
                        properties:
                          name:
                            type: string
                          path:
                            type: string
                        required:
                        - name
                        - path
                        type: object
                      type: array
                    coreLimit:
                      type: string
                    cores:
                      format: int32
                      minimum: 1
                      type: integer
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
                    envSecretKeyRefs:
                      additionalProperties:
                        properties:
                          key:
                            type: string
                          name:
                            type: string
                        required:
                        - key
                        - name
                        type: object
                      type: object
                    envVars:
                      additionalProperties:
                        type: string
                      type: object
                    gpu:
                      properties:
                        name:
                          type: string
                        quantity:
                          format: int64
                          type: integer
                      required:
                      - name
                      - quantity
                      type: object
                    hostNetwork:
                      type: boolean
                    image:
                      type: string
                    javaOptions:
                      type: string
                    labels:
                      additionalProperties:
                        type: string
                      type: object
                    memory:
                      type: string
                    memoryOverhead:
                      type: string
                    nodeSelector:
                      additionalProperties:
                        type: string
                      type: object
                    podName:
                      pattern: '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*'
                      type: string
                    schedulerName:
                      type: string
                    secrets:
                      items:
                        properties:
                          name:
                            type: string
                          path:
                            type: string
                          secretType:
                            type: string
                        required:
                        - name
                        - path
                        - secretType
                        type: object
                      type: array
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
                      type: object
                    serviceAccount:
                      type: string
                    sidecars:
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
                                        anyOf:
                                        - type: string
                                        - type: integer
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
                                        anyOf:
                                        - type: string
                                        - type: integer
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
                                        anyOf:
                                        - type: string
                                        - type: integer
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
                                        anyOf:
                                        - type: string
                                        - type: integer
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
                                    anyOf:
                                    - type: string
                                    - type: integer
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
                                    anyOf:
                                    - type: string
                                    - type: integer
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
                                    anyOf:
                                    - type: string
                                    - type: integer
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
                                    anyOf:
                                    - type: string
                                    - type: integer
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
                                additionalProperties:
                                  type: string
                                type: object
                              requests:
                                additionalProperties:
                                  type: string
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
                              - devicePath
                              - name
                              type: object
                            type: array
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
                              required:
                              - mountPath
                              - name
                              type: object
                            type: array
                          workingDir:
                            type: string
                        required:
                        - name
                        type: object
                      type: array
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
                        required:
                        - mountPath
                        - name
                        type: object
                      type: array
                  type: object
                executor:
                  properties:
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
                                - preference
                                - weight
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
                                - podAffinityTerm
                                - weight
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
                                - podAffinityTerm
                                - weight
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
                    annotations:
                      additionalProperties:
                        type: string
                      type: object
                    configMaps:
                      items:
                        properties:
                          name:
                            type: string
                          path:
                            type: string
                        required:
                        - name
                        - path
                        type: object
                      type: array
                    coreLimit:
                      type: string
                    coreRequest:
                      type: string
                    cores:
                      format: int32
                      minimum: 1
                      type: integer
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
                    envSecretKeyRefs:
                      additionalProperties:
                        properties:
                          key:
                            type: string
                          name:
                            type: string
                        required:
                        - key
                        - name
                        type: object
                      type: object
                    envVars:
                      additionalProperties:
                        type: string
                      type: object
                    gpu:
                      properties:
                        name:
                          type: string
                        quantity:
                          format: int64
                          type: integer
                      required:
                      - name
                      - quantity
                      type: object
                    hostNetwork:
                      type: boolean
                    image:
                      type: string
                    instances:
                      format: int32
                      minimum: 1
                      type: integer
                    javaOptions:
                      type: string
                    labels:
                      additionalProperties:
                        type: string
                      type: object
                    memory:
                      type: string
                    memoryOverhead:
                      type: string
                    nodeSelector:
                      additionalProperties:
                        type: string
                      type: object
                    schedulerName:
                      type: string
                    secrets:
                      items:
                        properties:
                          name:
                            type: string
                          path:
                            type: string
                          secretType:
                            type: string
                        required:
                        - name
                        - path
                        - secretType
                        type: object
                      type: array
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
                      type: object
                    sidecars:
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
                                        anyOf:
                                        - type: string
                                        - type: integer
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
                                        anyOf:
                                        - type: string
                                        - type: integer
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
                                        anyOf:
                                        - type: string
                                        - type: integer
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
                                        anyOf:
                                        - type: string
                                        - type: integer
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
                                    anyOf:
                                    - type: string
                                    - type: integer
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
                                    anyOf:
                                    - type: string
                                    - type: integer
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
                                    anyOf:
                                    - type: string
                                    - type: integer
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
                                    anyOf:
                                    - type: string
                                    - type: integer
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
                                additionalProperties:
                                  type: string
                                type: object
                              requests:
                                additionalProperties:
                                  type: string
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
                              - devicePath
                              - name
                              type: object
                            type: array
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
                              required:
                              - mountPath
                              - name
                              type: object
                            type: array
                          workingDir:
                            type: string
                        required:
                        - name
                        type: object
                      type: array
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
                        required:
                        - mountPath
                        - name
                        type: object
                      type: array
                  type: object
                failureRetries:
                  format: int32
                  type: integer
                hadoopConf:
                  additionalProperties:
                    type: string
                  type: object
                hadoopConfigMap:
                  type: string
                image:
                  type: string
                imagePullPolicy:
                  type: string
                imagePullSecrets:
                  items:
                    type: string
                  type: array
                initContainerImage:
                  type: string
                mainApplicationFile:
                  type: string
                mainClass:
                  type: string
                memoryOverheadFactor:
                  type: string
                mode:
                  enum:
                  - cluster
                  - client
                  type: string
                monitoring:
                  properties:
                    exposeDriverMetrics:
                      type: boolean
                    exposeExecutorMetrics:
                      type: boolean
                    metricsProperties:
                      type: string
                    prometheus:
                      properties:
                        configFile:
                          type: string
                        configuration:
                          type: string
                        jmxExporterJar:
                          type: string
                        port:
                          format: int32
                          maximum: 49151
                          minimum: 1024
                          type: integer
                      required:
                      - jmxExporterJar
                      type: object
                  required:
                  - exposeDriverMetrics
                  - exposeExecutorMetrics
                  type: object
                nodeSelector:
                  additionalProperties:
                    type: string
                  type: object
                pythonVersion:
                  enum:
                  - "2"
                  - "3"
                  type: string
                restartPolicy:
                  properties:
                    onFailureRetries:
                      format: int32
                      minimum: 0
                      type: integer
                    onFailureRetryInterval:
                      format: int64
                      minimum: 1
                      type: integer
                    onSubmissionFailureRetries:
                      format: int32
                      minimum: 0
                      type: integer
                    onSubmissionFailureRetryInterval:
                      format: int64
                      minimum: 1
                      type: integer
                    type:
                      enum:
                      - Never
                      - Always
                      - OnFailure
                      type: string
                  type: object
                retryInterval:
                  format: int64
                  type: integer
                sparkConf:
                  additionalProperties:
                    type: string
                  type: object
                sparkConfigMap:
                  type: string
                sparkVersion:
                  type: string
                timeToLiveSeconds:
                  format: int64
                  type: integer
                type:
                  enum:
                  - Java
                  - Python
                  - Scala
                  - R
                  type: string
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
                      configMap:
                        properties:
                          defaultMode:
                            format: int32
                            type: integer
                          items:
                            items:
                              properties:
                                key:
                                  type: string
                                mode:
                                  format: int32
                                  type: integer
                                path:
                                  type: string
                              required:
                              - key
                              - path
                              type: object
                            type: array
                          name:
                            type: string
                          optional:
                            type: boolean
                        type: object
                      downwardAPI:
                        properties:
                          defaultMode:
                            format: int32
                            type: integer
                          items:
                            items:
                              properties:
                                fieldRef:
                                  properties:
                                    apiVersion:
                                      type: string
                                    fieldPath:
                                      type: string
                                  required:
                                  - fieldPath
                                  type: object
                                mode:
                                  format: int32
                                  type: integer
                                path:
                                  type: string
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
                              required:
                              - path
                              type: object
                            type: array
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
                        - iqn
                        - lun
                        - targetPortal
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
                        - path
                        - server
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
                                configMap:
                                  properties:
                                    items:
                                      items:
                                        properties:
                                          key:
                                            type: string
                                          mode:
                                            format: int32
                                            type: integer
                                          path:
                                            type: string
                                        required:
                                        - key
                                        - path
                                        type: object
                                      type: array
                                    name:
                                      type: string
                                    optional:
                                      type: boolean
                                  type: object
                                downwardAPI:
                                  properties:
                                    items:
                                      items:
                                        properties:
                                          fieldRef:
                                            properties:
                                              apiVersion:
                                                type: string
                                              fieldPath:
                                                type: string
                                            required:
                                            - fieldPath
                                            type: object
                                          mode:
                                            format: int32
                                            type: integer
                                          path:
                                            type: string
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
                                        required:
                                        - path
                                        type: object
                                      type: array
                                  type: object
                                secret:
                                  properties:
                                    items:
                                      items:
                                        properties:
                                          key:
                                            type: string
                                          mode:
                                            format: int32
                                            type: integer
                                          path:
                                            type: string
                                        required:
                                        - key
                                        - path
                                        type: object
                                      type: array
                                    name:
                                      type: string
                                    optional:
                                      type: boolean
                                  type: object
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
                        - image
                        - monitors
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
                        - secretRef
                        - system
                        type: object
                      secret:
                        properties:
                          defaultMode:
                            format: int32
                            type: integer
                          items:
                            items:
                              properties:
                                key:
                                  type: string
                                mode:
                                  format: int32
                                  type: integer
                                path:
                                  type: string
                              required:
                              - key
                              - path
                              type: object
                            type: array
                          optional:
                            type: boolean
                          secretName:
                            type: string
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
              required:
              - driver
              - executor
              - mainApplicationFile
              - sparkVersion
              - type
              type: object
          required:
          - schedule
          - template
          type: object
      required:
      - metadata
      - spec
      type: object
  version: v1beta2
  versions:
  - name: v1beta2
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`)
	th.writeK("/manifests/spark/spark-operator/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kubeflow
# Labels to add to all resources and selectors.
commonLabels:
  kustomize.component: spark-operator
  app.kubernetes.io/instance: spark-operator
  app.kubernetes.io/name: sparkoperator

# Images modify the tags for images without
# creating patches.
images:
- name: gcr.io/spark-operator/spark-operator
  newTag: v1beta2-1.0.0-2.4.4

# Value of this field is prepended to the
# names of all resources
  newName: gcr.io/spark-operator/spark-operator
namePrefix: spark-operator

# List of resource files that kustomize reads, modifies
# and emits as a YAML string
resources:
- spark-sa.yaml
- cr-clusterrole.yaml
- crb.yaml
- crd-cleanup-job.yaml
- deploy.yaml
- operator-sa.yaml
- sparkapplications.sparkoperator.k8s.io-crd.yaml
- scheduledsparkapplications.sparkoperator.k8s.io-crd.yaml
`)
}

func TestSparkOperatorBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/spark/spark-operator/base")
	writeSparkOperatorBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../spark/spark-operator/base"
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
