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
    jsonSpec: '{}',
    recentError: '',
    lastJobState: 'JOB_CREATED',
    lastJobNodeState: 'RUNNING',
    datumTimeoutS: null,
    jobTimeoutS: null,
    s3OutputRepo: null,
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
  jsonSpec:
    '{\n  "transform": {\n    "image": "dpokidov/imagemagick:7.1.0-23",\n    "cmd": [\n      "sh"\n    ],\n    "stdin": [\n      "montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find -L /pfs/images /pfs/edges -type f | sort) /pfs/out/montage.png"\n    ]\n  },\n  "input": {\n    "cross": [\n      {\n        "pfs": {\n          "project": "default",\n          "name": "images",\n          "repo": "images",\n          "repoType": "user",\n          "branch": "master",\n          "glob": "/"\n        }\n      },\n      {\n        "pfs": {\n          "project": "default",\n          "name": "edges",\n          "repo": "edges",\n          "repoType": "user",\n          "branch": "master",\n          "glob": "/"\n        }\n      }\n    ]\n  },\n  "reprocessSpec": "until_success"\n}',
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
  jsonSpec:
    '{\n  "transform": {\n    "image": "pachyderm/opencv:1.0",\n    "cmd": [\n      "python3",\n      "/edges.py"\n    ]\n  },\n  "input": {\n    "pfs": {\n      "project": "default",\n      "name": "images",\n      "repo": "images",\n      "repoType": "user",\n      "branch": "master",\n      "glob": "/*"\n    }\n  },\n  "reprocessSpec": "until_success"\n}',
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
  jsonSpec: JSON.stringify(
    {
      transform: {
        image: 'pachyderm/opencv:1.0',
        cmd: ['python3', '/edges.py'],
      },
      input: {
        pfs: {
          project: 'default',
          name: 'images',
          repo: 'images',
          repoType: 'user',
          branch: 'master',
          glob: '/*',
        },
      },
      reprocessSpec: 'until_success',
    },
    null,
    2,
  ),
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
  jsonSpec:
    '{\n  "transform": {\n    "image": "pachyderm/opencv:1.0",\n    "cmd": [\n      "python3",\n      "/edges.py"\n    ]\n  },\n  "input": {\n    "pfs": {\n      "project": "default",\n      "name": "images",\n      "repo": "images",\n      "repoType": "user",\n      "branch": "master",\n      "glob": "/*"\n    }\n  },\n  "reprocessSpec": "until_success"\n}',
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
        jsonSpec: JSON.stringify(
          {
            parallelismSpec: {
              constant: numWorkers,
            },
            transform: {
              image: 'pachyderm/opencv:1.0',
              cmd: ['python3', '/edges.py'],
            },
            input: {
              pfs: {
                project: 'default',
                name: 'images',
                repo: 'images',
                repoType: 'user',
                branch: 'master',
                glob: '/*',
              },
            },
          },
          null,
          2,
        ),
      });
    });

    return res(ctx.data({pipelines}));
  });
