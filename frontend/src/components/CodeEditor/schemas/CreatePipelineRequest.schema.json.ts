import {type JSONSchema7} from 'json-schema';

export const rawCreatePipelineRequestSchema: JSONSchema7 = {
  $schema: 'http://json-schema.org/draft-04/schema#',
  $ref: '#/definitions/CreatePipelineRequest',
  definitions: {
    CreatePipelineRequest: {
      properties: {
        pipeline: {
          $ref: '#/definitions/pps_v2.Pipeline',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
        },
        tfJob: {
          $ref: '#/definitions/pps_v2.TFJob',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
          description:
            "tf_job encodes a Kubeflow TFJob spec. Pachyderm uses this to create TFJobs when running in a kubernetes cluster on which kubeflow has been installed. Exactly one of 'tf_job' and 'transform' should be set",
        },
        transform: {
          $ref: '#/definitions/pps_v2.Transform',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
        },
        parallelismSpec: {
          $ref: '#/definitions/pps_v2.ParallelismSpec',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
        },
        egress: {
          $ref: '#/definitions/pps_v2.Egress',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
        },
        outputBranch: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
        },
        s3Out: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'boolean',
            },
          ],
          description:
            "s3_out, if set, requires a pipeline's user to write to its output repo via Pachyderm's s3 gateway (if set, workers will serve Pachyderm's s3 gateway API at http://\u003cpipeline\u003e-s3.\u003cnamespace\u003e/\u003cjob id\u003e.out/my/file). In this mode /pfs_v2/out won't be walked or uploaded, and the s3 gateway service in the workers will allow writes to the job's output commit",
        },
        resourceRequests: {
          $ref: '#/definitions/pps_v2.ResourceSpec',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
        },
        resourceLimits: {
          $ref: '#/definitions/pps_v2.ResourceSpec',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
        },
        sidecarResourceLimits: {
          $ref: '#/definitions/pps_v2.ResourceSpec',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
        },
        input: {
          $ref: '#/definitions/pps_v2.Input',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
        },
        description: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
        },
        service: {
          $ref: '#/definitions/pps_v2.Service',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
        },
        spout: {
          $ref: '#/definitions/pps_v2.Spout',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
        },
        datumSetSpec: {
          $ref: '#/definitions/pps_v2.DatumSetSpec',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
        },
        datumTimeout: {
          pattern: '^([0-9]+\\.?[0-9]*|\\.[0-9]+)s$',
          type: 'string',
          format: 'regex',
        },
        jobTimeout: {
          pattern: '^([0-9]+\\.?[0-9]*|\\.[0-9]+)s$',
          type: 'string',
          format: 'regex',
        },
        salt: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
        },
        datumTries: {
          oneOf: [
            {
              type: 'integer',
            },
            {
              type: 'null',
            },
          ],
        },
        schedulingSpec: {
          $ref: '#/definitions/pps_v2.SchedulingSpec',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
        },
        podSpec: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
          description: 'deprecated, use pod_patch below',
        },
        podPatch: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
          description:
            "a json patch will be applied to the pipeline's pod_spec before it's created;",
        },
        metadata: {
          $ref: '#/definitions/pps_v2.Metadata',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
        },
        reprocessSpec: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
        },
        autoscaling: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'boolean',
            },
          ],
        },
        tolerations: {
          items: {
            $ref: '#/definitions/pps_v2.Toleration',
          },
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'array',
            },
          ],
        },
        sidecarResourceRequests: {
          $ref: '#/definitions/pps_v2.ResourceSpec',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
        },
        determined: {
          $ref: '#/definitions/pps_v2.Determined',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
        },
      },
      additionalProperties: false,
      oneOf: [
        {
          type: 'null',
        },
        {
          type: 'object',
        },
      ],
      title: 'Create Pipeline Request',
    },
    'pfs_v2.Branch': {
      properties: {
        repo: {
          $ref: '#/definitions/pfs_v2.Repo',
          additionalProperties: false,
        },
        name: {
          type: 'string',
        },
      },
      additionalProperties: false,
      type: 'object',
      title: 'Branch',
    },
    'pfs_v2.Commit': {
      properties: {
        repo: {
          $ref: '#/definitions/pfs_v2.Repo',
          additionalProperties: false,
        },
        id: {
          type: 'string',
        },
        branch: {
          $ref: '#/definitions/pfs_v2.Branch',
          additionalProperties: false,
          description: 'only used by the client',
        },
      },
      additionalProperties: false,
      type: 'object',
      title: 'Commit',
      description:
        'Commit is a reference to a commit (e.g. the collection of branches and the collection of currently-open commits in etcd are collections of Commit protos)',
    },
    'pfs_v2.ObjectStorageEgress': {
      properties: {
        url: {
          type: 'string',
        },
      },
      additionalProperties: false,
      type: 'object',
      title: 'Object Storage Egress',
    },
    'pfs_v2.Project': {
      properties: {
        name: {
          type: 'string',
        },
      },
      additionalProperties: false,
      type: 'object',
      title: 'Project',
    },
    'pfs_v2.Repo': {
      properties: {
        name: {
          type: 'string',
        },
        type: {
          type: 'string',
        },
        project: {
          $ref: '#/definitions/pfs_v2.Project',
          additionalProperties: false,
        },
      },
      additionalProperties: false,
      type: 'object',
      title: '//  PFS Data structures (stored in etcd)',
      description: '//  PFS Data structures (stored in etcd)',
    },
    'pfs_v2.SQLDatabaseEgress': {
      properties: {
        url: {
          type: 'string',
        },
        fileFormat: {
          $ref: '#/definitions/pfs_v2.SQLDatabaseEgress.FileFormat',
          additionalProperties: false,
        },
        secret: {
          $ref: '#/definitions/pfs_v2.SQLDatabaseEgress.Secret',
          additionalProperties: false,
        },
      },
      additionalProperties: false,
      type: 'object',
      title: 'SQL Database Egress',
    },
    'pfs_v2.SQLDatabaseEgress.FileFormat': {
      properties: {
        type: {
          enum: ['UNKNOWN', 'CSV', 'JSON', 'PARQUET'],
          type: 'string',
          title: 'Type',
        },
        columns: {
          items: {
            type: 'string',
          },
          type: 'array',
        },
      },
      additionalProperties: false,
      type: 'object',
      title: 'File Format',
    },
    'pfs_v2.SQLDatabaseEgress.Secret': {
      properties: {
        name: {
          type: 'string',
        },
        key: {
          type: 'string',
        },
      },
      additionalProperties: false,
      type: 'object',
      title: 'Secret',
    },
    'pfs_v2.Trigger': {
      properties: {
        branch: {
          type: 'string',
          description: 'Which branch this trigger refers to',
        },
        all: {
          type: 'boolean',
          description:
            'All indicates that all conditions must be satisfied before the trigger happens, otherwise any conditions being satisfied will trigger it.',
        },
        rateLimitSpec: {
          type: 'string',
          description:
            'Triggers if the rate limit spec (cron expression) has been satisfied since the last trigger.',
        },
        size: {
          type: 'string',
          description:
            "Triggers if there's been `size` new data added since the last trigger.",
        },
        commits: {
          type: 'integer',
          description:
            "Triggers if there's been `commits` new commits added since the last trigger.",
        },
        cronSpec: {
          type: 'string',
          description:
            'Creates a background process which fires the trigger on the schedule provided by the cron spec. This condition is mutually exclusive with respect to the others, so setting this will result with the trigger only firing based on the cron schedule.',
        },
      },
      additionalProperties: false,
      type: 'object',
      title: 'Trigger',
      description:
        'Trigger defines the conditions under which a head is moved, and to which branch it is moved.',
    },
    'pps_v2.CronInput': {
      properties: {
        name: {
          type: 'string',
        },
        project: {
          type: 'string',
        },
        repo: {
          type: 'string',
        },
        commit: {
          type: 'string',
        },
        spec: {
          type: 'string',
        },
        overwrite: {
          type: 'boolean',
          description:
            'Overwrite, if true, will expose a single datum that gets overwritten each tick. If false, it will create a new datum for each tick.',
        },
        start: {
          type: 'string',
          format: 'date-time',
        },
      },
      additionalProperties: false,
      type: 'object',
      title: 'Cron Input',
    },
    'pps_v2.DatumSetSpec': {
      properties: {
        number: {
          oneOf: [
            {
              type: 'integer',
            },
            {
              type: 'null',
            },
          ],
          description:
            "number, if nonzero, specifies that each datum set should contain `number` datums. Datum sets may contain fewer if the total number of datums don't divide evenly.",
        },
        sizeBytes: {
          oneOf: [
            {
              type: 'integer',
            },
            {
              type: 'null',
            },
          ],
          description:
            'size_bytes, if nonzero, specifies a target size for each datum set. Datum sets may be larger or smaller than size_bytes, but will usually be pretty close to size_bytes in size.',
        },
        perWorker: {
          oneOf: [
            {
              type: 'integer',
            },
            {
              type: 'null',
            },
          ],
          description:
            "per_worker, if nonzero, specifies how many datum sets should be created for each worker. It can't be set with number or size_bytes.",
        },
      },
      additionalProperties: false,
      oneOf: [
        {
          type: 'null',
        },
        {
          type: 'object',
        },
      ],
      title: 'Datum Set Spec',
      description:
        'DatumSetSpec specifies how a pipeline should split its datums into datum sets.',
    },
    'pps_v2.Determined': {
      properties: {
        workspaces: {
          items: {
            type: 'string',
          },
          type: 'array',
        },
      },
      additionalProperties: false,
      type: 'object',
      title: 'Determined',
    },
    'pps_v2.Egress': {
      properties: {
        URL: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
        },
        objectStorage: {
          $ref: '#/definitions/pfs_v2.ObjectStorageEgress',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
        },
        sqlDatabase: {
          $ref: '#/definitions/pfs_v2.SQLDatabaseEgress',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
        },
      },
      additionalProperties: false,
      oneOf: [
        {
          type: 'null',
        },
        {
          type: 'object',
        },
        {
          required: ['object_storage'],
        },
        {
          required: ['sql_database'],
        },
      ],
      title: 'Egress',
    },
    'pps_v2.GPUSpec': {
      properties: {
        type: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
          description:
            'The type of GPU (nvidia.com/gpu or amd.com/gpu for example).',
        },
        number: {
          oneOf: [
            {
              type: 'integer',
            },
            {
              type: 'null',
            },
          ],
          description: 'The number of GPUs to request.',
        },
      },
      additionalProperties: false,
      oneOf: [
        {
          type: 'null',
        },
        {
          type: 'object',
        },
      ],
      title: 'GPU Spec',
    },
    'pps_v2.Input': {
      properties: {
        pfs: {
          $ref: '#/definitions/pps_v2.PFSInput',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
        },
        join: {
          items: {
            $ref: '#/definitions/pps_v2.Input',
          },
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'array',
            },
          ],
        },
        group: {
          items: {
            $ref: '#/definitions/pps_v2.Input',
          },
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'array',
            },
          ],
        },
        cross: {
          items: {
            $ref: '#/definitions/pps_v2.Input',
          },
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'array',
            },
          ],
        },
        union: {
          items: {
            $ref: '#/definitions/pps_v2.Input',
          },
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'array',
            },
          ],
        },
        cron: {
          $ref: '#/definitions/pps_v2.CronInput',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
        },
      },
      additionalProperties: false,
      oneOf: [
        {
          type: 'null',
        },
        {
          type: 'object',
        },
      ],
      title: 'Input',
    },
    'pps_v2.Metadata': {
      properties: {
        annotations: {
          additionalProperties: {
            type: 'string',
          },
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'object',
            },
          ],
        },
        labels: {
          additionalProperties: {
            type: 'string',
          },
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'object',
            },
          ],
        },
      },
      additionalProperties: false,
      oneOf: [
        {
          type: 'null',
        },
        {
          type: 'object',
        },
      ],
      title: 'Metadata',
    },
    'pps_v2.PFSInput': {
      properties: {
        project: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
        },
        name: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
        },
        repo: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
        },
        repoType: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
        },
        branch: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
        },
        commit: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
        },
        glob: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
        },
        joinOn: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
        },
        outerJoin: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'boolean',
            },
          ],
        },
        groupBy: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
        },
        lazy: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'boolean',
            },
          ],
        },
        emptyFiles: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'boolean',
            },
          ],
          description:
            'EmptyFiles, if true, will cause files from this PFS input to be presented as empty files. This is useful in shuffle pipelines where you want to read the names of files and reorganize them using symlinks.',
        },
        s3: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'boolean',
            },
          ],
          description:
            'S3, if true, will cause the worker to NOT download or link files from this input into the /pfs_v2 directory. Instead, an instance of our S3 gateway service will run on each of the sidecars, and data can be retrieved from this input by querying http://\u003cpipeline\u003e-s3.\u003cnamespace\u003e/\u003cjob id\u003e.\u003cinput\u003e/my/file',
        },
        trigger: {
          $ref: '#/definitions/pfs_v2.Trigger',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
          description:
            "Trigger defines when this input is processed by the pipeline, if it's nil the input is processed anytime something is committed to the input branch.",
        },
      },
      additionalProperties: false,
      oneOf: [
        {
          type: 'null',
        },
        {
          type: 'object',
        },
      ],
      title: 'PFS Input',
    },
    'pps_v2.ParallelismSpec': {
      properties: {
        constant: {
          oneOf: [
            {
              type: 'integer',
            },
            {
              type: 'null',
            },
          ],
          description:
            "Starts the pipeline/job with a 'constant' workers, unless 'constant' is zero. If 'constant' is zero (which is the zero value of ParallelismSpec), then Pachyderm will choose the number of workers that is started, (currently it chooses the number of workers in the cluster)",
        },
      },
      additionalProperties: false,
      oneOf: [
        {
          type: 'null',
        },
        {
          type: 'object',
        },
      ],
      title: 'Parallelism Spec',
    },
    'pps_v2.Pipeline': {
      properties: {
        project: {
          $ref: '#/definitions/pfs_v2.Project',
          additionalProperties: false,
        },
        name: {
          type: 'string',
        },
      },
      additionalProperties: false,
      type: 'object',
      title: 'Pipeline',
    },
    'pps_v2.ResourceSpec': {
      properties: {
        cpu: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'number',
            },
          ],
          description:
            'The number of CPUs each worker needs (partial values are allowed, and encouraged)',
        },
        memory: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
          description:
            'The amount of memory each worker needs (in bytes, with allowed SI suffixes (M, K, G, Mi, Ki, Gi, etc).',
        },
        gpu: {
          $ref: '#/definitions/pps_v2.GPUSpec',
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {},
          ],
          description: 'The spec for GPU resources.',
        },
        disk: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
          description:
            'The amount of ephemeral storage each worker needs (in bytes, with allowed SI suffixes (M, K, G, Mi, Ki, Gi, etc).',
        },
      },
      additionalProperties: false,
      oneOf: [
        {
          type: 'null',
        },
        {
          type: 'object',
        },
      ],
      title: 'Resource Spec',
      description:
        'ResourceSpec describes the amount of resources that pipeline pods should request from kubernetes, for scheduling.',
    },
    'pps_v2.SchedulingSpec': {
      properties: {
        nodeSelector: {
          additionalProperties: {
            type: 'string',
          },
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'object',
            },
          ],
        },
        priorityClassName: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
        },
      },
      additionalProperties: false,
      oneOf: [
        {
          type: 'null',
        },
        {
          type: 'object',
        },
      ],
      title: 'Scheduling Spec',
    },
    'pps_v2.SecretMount': {
      properties: {
        name: {
          type: 'string',
          description: 'Name must be the name of the secret in kubernetes.',
        },
        key: {
          type: 'string',
          description:
            'Key of the secret to load into env_var, this field only has meaning if EnvVar != "".',
        },
        mountPath: {
          type: 'string',
        },
        envVar: {
          type: 'string',
        },
      },
      additionalProperties: false,
      type: 'object',
      title: 'Secret Mount',
    },
    'pps_v2.Service': {
      properties: {
        internalPort: {
          type: 'integer',
        },
        externalPort: {
          type: 'integer',
        },
        ip: {
          type: 'string',
        },
        type: {
          type: 'string',
        },
      },
      additionalProperties: false,
      type: 'object',
      title: 'Service',
    },
    'pps_v2.Spout': {
      properties: {
        service: {
          $ref: '#/definitions/pps_v2.Service',
          additionalProperties: false,
        },
      },
      additionalProperties: false,
      type: 'object',
      title: 'Spout',
    },
    'pps_v2.TFJob': {
      properties: {
        tfJob: {
          type: 'string',
          description:
            'tf_job  is a serialized Kubeflow TFJob spec. Pachyderm sends this directly to a kubernetes cluster on which kubeflow has been installed, instead of creating a pipeline ReplicationController as it normally would.',
        },
      },
      additionalProperties: false,
      type: 'object',
      title: 'TF Job',
    },
    'pps_v2.Toleration': {
      properties: {
        key: {
          type: 'string',
          description:
            'key is the taint key that the toleration applies to.  Empty means match all taint keys.',
        },
        operator: {
          enum: ['EMPTY', 'EXISTS', 'EQUAL'],
          type: 'string',
          title: 'Toleration Operator',
          description:
            "TolerationOperator relates a Toleration's key to its value.",
        },
        value: {
          type: 'string',
          description: 'value is the taint value the toleration matches to.',
        },
        effect: {
          enum: [
            'ALL_EFFECTS',
            'NO_SCHEDULE',
            'PREFER_NO_SCHEDULE',
            'NO_EXECUTE',
          ],
          type: 'string',
          title: 'Taint Effect',
          description:
            'TaintEffect is an effect that can be matched by a toleration.',
        },
        tolerationSeconds: {
          additionalProperties: false,
          type: 'integer',
          description:
            'toleration_seconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint.  If not set, tolerate the taint forever.',
        },
      },
      additionalProperties: false,
      type: 'object',
      title: 'Toleration',
      description: 'Toleration is a Kubernetes toleration.',
    },
    'pps_v2.Transform': {
      properties: {
        image: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
        },
        cmd: {
          items: {
            oneOf: [
              {
                type: 'null',
              },
              {
                type: 'string',
              },
            ],
          },
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'array',
            },
          ],
        },
        errCmd: {
          items: {
            oneOf: [
              {
                type: 'null',
              },
              {
                type: 'string',
              },
            ],
          },
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'array',
            },
          ],
        },
        env: {
          additionalProperties: {
            type: 'string',
          },
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'object',
            },
          ],
        },
        secrets: {
          items: {
            $ref: '#/definitions/pps_v2.SecretMount',
          },
          additionalProperties: false,
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'array',
            },
          ],
        },
        imagePullSecrets: {
          items: {
            oneOf: [
              {
                type: 'null',
              },
              {
                type: 'string',
              },
            ],
          },
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'array',
            },
          ],
        },
        stdin: {
          items: {
            oneOf: [
              {
                type: 'null',
              },
              {
                type: 'string',
              },
            ],
          },
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'array',
            },
          ],
        },
        errStdin: {
          items: {
            oneOf: [
              {
                type: 'null',
              },
              {
                type: 'string',
              },
            ],
          },
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'array',
            },
          ],
        },
        acceptReturnCode: {
          items: {
            oneOf: [
              {
                type: 'integer',
              },
              {
                type: 'null',
              },
            ],
          },
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'array',
            },
          ],
        },
        debug: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'boolean',
            },
          ],
        },
        user: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
        },
        workingDir: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
        },
        dockerfile: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'string',
            },
          ],
        },
        memoryVolume: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'boolean',
            },
          ],
        },
        datumBatching: {
          oneOf: [
            {
              type: 'null',
            },
            {
              type: 'boolean',
            },
          ],
        },
      },
      additionalProperties: false,
      oneOf: [
        {
          type: 'null',
        },
        {
          type: 'object',
        },
      ],
      title: 'Transform',
    },
  },
};
