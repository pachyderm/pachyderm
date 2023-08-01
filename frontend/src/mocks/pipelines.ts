import {
  mockPipelinesQuery,
  Pipeline,
  PipelineState,
  JobState,
  NodeState,
  PipelineType,
} from '@graphqlTypes';

export const ALL_PIPELINES: Pipeline[] = [
  {
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
    reason: '',
    __typename: 'Pipeline',
  },
  {
    id: 'default_edges',
    name: 'edges',
    description:
      'A pipeline that performs image edge detection by using the OpenCV library.',
    version: 1,
    createdAt: 1690221505,
    state: PipelineState.PIPELINE_RUNNING,
    nodeState: NodeState.IDLE,
    stopped: false,
    recentError: '',
    lastJobState: JobState.JOB_CREATED,
    lastJobNodeState: NodeState.RUNNING,
    type: PipelineType.STANDARD,
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
  },
];

export const mockPipelines = () =>
  mockPipelinesQuery((_req, res, ctx) => {
    return res(ctx.data({pipelines: ALL_PIPELINES}));
  });
