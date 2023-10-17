import {
  mockPipelinesQuery,
  Pipeline,
  PipelineState,
  JobState,
  NodeState,
  PipelineType,
  mockPipelineQuery,
} from '@graphqlTypes';
import merge from 'lodash/merge';
import range from 'lodash/range';

const userSpecJsonEdges = {
  pipeline: {
    project: {
      name: 'default',
    },
    name: 'edges',
  },
  transform: {
    image: 'pachyderm/opencv:1.0',
    cmd: ['python3', '/edges.py'],
  },
  input: {
    pfs: {
      repo: 'images',
      glob: '/*',
    },
  },
  description:
    'A pipeline that performs image edge detection by using the OpenCV library.',
};

const effectiveSpecJsonEdges = {
  ...userSpecJsonEdges,
  resourceRequests: {
    cpu: 1,
    disk: '1Gi',
    memory: '256Mi',
  },
  sidecarResourceRequests: {
    cpu: 1,
    disk: '1Gi',
    memory: '256Mi',
  },
};

const userSpecJsonMontage = {
  pipeline: {
    project: {
      name: 'default',
    },
    name: 'montage',
  },
  transform: {
    image: 'dpokidov/imagemagick:7.1.0-23',
    cmd: ['sh'],
    stdin: [
      'montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find -L /pfs/images /pfs/edges -type f | sort) /pfs/out/montage.png',
    ],
  },
  input: {
    cross: [
      {
        pfs: {
          repo: 'images',
          glob: '/',
        },
      },
      {
        pfs: {
          repo: 'edges',
          glob: '/',
        },
      },
    ],
  },
  description:
    'A pipeline that combines images from the `images` and `edges` repositories into a montage.',
};

const effectiveSpecJsonMontage = {
  ...userSpecJsonMontage,
  resourceRequests: {
    cpu: 1,
    disk: '1Gi',
    memory: '256Mi',
  },
  sidecarResourceRequests: {
    cpu: 1,
    disk: '1Gi',
    memory: '256Mi',
  },
};

export const buildPipeline = (pipeline: Partial<Pipeline>): Pipeline => {
  const defaultPipeline = {
    id: 'default',
    name: 'default',
    description: '',
    version: 0,
    createdAt: 0,
    state: PipelineState.PIPELINE_STANDBY,
    nodeState: NodeState.SUCCESS,
    stopped: false,
    type: PipelineType.STANDARD,
    datumTries: 0,
    outputBranch: 'master',
    egress: false,
    effectiveSpecJson: '{}',
    userSpecJson: '{}',
    recentError: '',
    lastJobState: 'JOB_CREATED',
    lastJobNodeState: 'RUNNING',
    datumTimeoutS: null,
    jobTimeoutS: null,
    s3OutputRepo: null,
    parallelismSpec: 0,
    reason: null,
    __typename: 'Pipeline',
  };
  return merge(defaultPipeline, pipeline);
};

const MONTAGE_PIPELINE: Pipeline = buildPipeline({
  id: 'default_montage',
  name: 'montage',
  description:
    'A pipeline that combines images from the `images` and `edges` repositories into a montage.',
  version: 1,
  createdAt: 1690221506,
  state: PipelineState.PIPELINE_FAILURE,
  nodeState: NodeState.ERROR,
  stopped: false,
  recentError: '',
  lastJobState: JobState.JOB_KILLED,
  lastJobNodeState: NodeState.ERROR,
  type: PipelineType.STANDARD,
  datumTimeoutS: null,
  datumTries: 3,
  jobTimeoutS: null,
  outputBranch: 'master',
  s3OutputRepo: null,
  egress: false,
  effectiveSpecJson: JSON.stringify(effectiveSpecJsonMontage, null, 2),
  userSpecJson: JSON.stringify(userSpecJsonMontage, null, 2),
  reason: 'Pipeline failed because we have no memory!',
  __typename: 'Pipeline',
});

const EDGES_PIPELINE: Pipeline = buildPipeline({
  id: 'default_edges',
  name: 'edges',
  description:
    'A pipeline that performs image edge detection by using the OpenCV library.',
  version: 1,
  createdAt: 1690221505,
  state: PipelineState.PIPELINE_RUNNING,
  nodeState: NodeState.IDLE,
  lastJobState: JobState.JOB_CREATED,
  lastJobNodeState: NodeState.RUNNING,
  type: PipelineType.STANDARD,
  datumTries: 3,
  jobTimeoutS: null,
  outputBranch: 'master',
  s3OutputRepo: null,
  effectiveSpecJson: JSON.stringify(effectiveSpecJsonEdges, null, 2),
  userSpecJson: JSON.stringify(userSpecJsonEdges, null, 2),
  __typename: 'Pipeline',
});

export const ALL_PIPELINES: Pipeline[] = [MONTAGE_PIPELINE, EDGES_PIPELINE];

export const mockPipelines = () =>
  mockPipelinesQuery((_req, res, ctx) => {
    return res(ctx.data({pipelines: ALL_PIPELINES}));
  });

export const mockGetMontagePipeline = () =>
  mockPipelineQuery((req, res, ctx) => {
    if (
      req.variables.args.id === 'montage' &&
      req.variables.args.projectId === 'default'
    ) {
      return res(ctx.data({pipeline: MONTAGE_PIPELINE}));
    } else {
      return res();
    }
  });

const SERVICE_PIPELINE: Pipeline = buildPipeline({
  id: 'default_edges',
  name: 'montage',
  description:
    'A pipeline that performs image edge detection by using the OpenCV library.',
  version: 1,
  createdAt: 1690221505,
  state: PipelineState.PIPELINE_RUNNING,
  nodeState: NodeState.RUNNING,
  lastJobState: JobState.JOB_SUCCESS,
  lastJobNodeState: NodeState.IDLE,
  type: PipelineType.SERVICE,
  datumTries: 3,
  jobTimeoutS: null,
  outputBranch: 'master',
  s3OutputRepo: null,
  __typename: 'Pipeline',
});

export const mockGetServicePipeline = () =>
  mockPipelineQuery((_req, res, ctx) => {
    return res(ctx.data({pipeline: SERVICE_PIPELINE}));
  });

const SPOUT_PIPELINE: Pipeline = buildPipeline({
  id: 'default_edges',
  name: 'montage',
  description:
    'A pipeline that performs image edge detection by using the OpenCV library.',
  version: 1,
  createdAt: 1690221505,
  state: PipelineState.PIPELINE_RUNNING,
  nodeState: NodeState.RUNNING,
  stopped: false,
  recentError: '',
  lastJobState: null,
  lastJobNodeState: null,
  type: PipelineType.SPOUT,
  datumTimeoutS: null,
  datumTries: 3,
  jobTimeoutS: null,
  outputBranch: 'master',
  s3OutputRepo: null,
  egress: false,
  reason: '',
  __typename: 'Pipeline',
});

export const mockGetSpoutPipeline = () =>
  mockPipelineQuery((req, res, ctx) => {
    if (
      req.variables.args.id === 'montage' &&
      req.variables.args.projectId === 'default'
    ) {
      return res(ctx.data({pipeline: SPOUT_PIPELINE}));
    } else {
      return res();
    }
  });

export const mockGetManyPipelinesWithManyWorkers = (
  numPipelines = 1,
  numWorkers = 1,
) =>
  mockPipelinesQuery((_req, res, ctx) => {
    const pipelines = range(0, numPipelines).map((i) => {
      return buildPipeline({
        name: `pipeline-${i}`,
        parallelismSpec: numWorkers,
      });
    });

    return res(ctx.data({pipelines}));
  });
