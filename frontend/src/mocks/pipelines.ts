import merge from 'lodash/merge';
import range from 'lodash/range';
import {rest} from 'msw';

import {Empty} from '@dash-frontend/api/googleTypes';
import {
  InspectPipelineRequest,
  ListPipelineRequest,
  PipelineInfo,
  PipelineInfoPipelineType,
  JobState as PipelineInfoJobState,
  PipelineState,
  JobState,
} from '@dash-frontend/api/pps';
import {RequestError} from '@dash-frontend/api/utils/error';
import {getISOStringFromUnix} from '@dash-frontend/lib/dateTime';

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

export const buildPipeline = (pipeline: PipelineInfo = {}): PipelineInfo => {
  const defaultPipeline: PipelineInfo = {
    pipeline: {name: 'default'},
    version: '0',
    state: PipelineState.PIPELINE_STANDBY,
    stopped: false,
    type: PipelineInfoPipelineType.PIPELINE_TYPE_TRANSFORM,
    lastJobState: PipelineInfoJobState.JOB_CREATED,
    reason: undefined,
    __typename: 'PipelineInfo',
  };

  return merge(defaultPipeline, pipeline);
};

const MONTAGE_PIPELINE: PipelineInfo = buildPipeline({
  pipeline: {
    name: 'montage',
    project: {name: 'default'},
  },
  details: {
    description:
      'A pipeline that combines images from the `images` and `edges` repositories into a montage.',
    createdAt: '2023-07-24T17:58:26Z',
    datumTries: '3',
    outputBranch: 'master',
    recentError: '',
    reason: 'Pipeline failed because we have no memory!',
    input: {
      cross: [
        {
          pfs: {
            glob: '/',
            project: 'default',
            repo: 'images',
          },
        },
        {
          pfs: {
            glob: '/',
            project: 'default',
            repo: 'edges',
          },
        },
      ],
    },
  },
  version: '1',
  state: PipelineState.PIPELINE_FAILURE,
  stopped: false,
  lastJobState: JobState.JOB_KILLED,
  type: PipelineInfoPipelineType.PIPELINE_TYPE_TRANSFORM,
  effectiveSpecJson: JSON.stringify(effectiveSpecJsonMontage, null, 2),
  userSpecJson: JSON.stringify(userSpecJsonMontage, null, 2),
  reason: 'Pipeline failed because we have no memory!',
});

const EDGES_PIPELINE: PipelineInfo = buildPipeline({
  pipeline: {
    name: 'edges',
    project: {name: 'default'},
  },
  details: {
    description:
      'A pipeline that performs image edge detection by using the OpenCV library.',
    createdAt: '2023-07-24T17:58:25Z',
    datumTries: '3',
    outputBranch: 'master',
    recentError: '',
    reason: 'Pipeline failed because we have no memory!',
    input: {
      pfs: {
        repo: 'images',
        project: 'default',
        glob: '/*',
      },
    },
  },
  version: '1',
  state: PipelineState.PIPELINE_RUNNING,
  stopped: false,
  lastJobState: JobState.JOB_CREATED,
  type: PipelineInfoPipelineType.PIPELINE_TYPE_TRANSFORM,
  effectiveSpecJson: JSON.stringify(effectiveSpecJsonEdges, null, 2),
  userSpecJson: JSON.stringify(userSpecJsonEdges, null, 2),
});

export const ALL_PIPELINES: PipelineInfo[] = [MONTAGE_PIPELINE, EDGES_PIPELINE];

export const mockPipelines = () =>
  rest.post<ListPipelineRequest, Empty, PipelineInfo[]>(
    '/api/pps_v2.API/ListPipeline',
    async (_req, res, ctx) => {
      return res(ctx.json(ALL_PIPELINES));
    },
  );

export const mockGetMontagePipeline = () =>
  rest.post<InspectPipelineRequest, Empty, PipelineInfo>(
    '/api/pps_v2.API/InspectPipeline',
    async (req, res, ctx) => {
      const args = await req.json();
      if (
        args.pipeline.name === 'montage' &&
        args.pipeline.project.name === 'default'
      ) {
        return res(ctx.json(MONTAGE_PIPELINE));
      } else {
        return res(ctx.json({}));
      }
    },
  );

export const mockGetEdgesPipeline = () =>
  rest.post<ListPipelineRequest, Empty, PipelineInfo>(
    '/api/pps_v2.API/InspectPipeline',
    async (_req, res, ctx) => {
      return res(ctx.json(EDGES_PIPELINE));
    },
  );

export const mockHealthyPipelines = () =>
  rest.post<ListPipelineRequest, Empty, PipelineInfo[]>(
    '/api/pps_v2.API/ListPipeline',
    async (_req, res, ctx) => {
      return res(ctx.json([buildPipeline(), buildPipeline(), buildPipeline()]));
    },
  );

export const mockEmptyInspectPipeline = () =>
  rest.post<ListPipelineRequest, Empty, RequestError>(
    '/api/pps_v2.API/InspectPipeline',
    async (_req, res, ctx) => {
      return res(
        ctx.status(404),
        ctx.json({
          code: 5,
          message: 'pipeline was not inspected',
          details: [],
        }),
      );
    },
  );

const SERVICE_PIPELINE_INFO = buildPipeline({
  pipeline: {
    name: 'montage',
    project: {name: 'default'},
  },
  details: {
    description:
      'A pipeline that performs image edge detection by using the OpenCV library.',
    createdAt: getISOStringFromUnix(1690221505),
    datumTries: '3',
    outputBranch: 'master',
    service: {
      type: 'LoadBalancer',
      ip: 'localhost',
    },
  },
  version: '1',
  state: PipelineState.PIPELINE_RUNNING,
  lastJobState: JobState.JOB_SUCCESS,
  type: PipelineInfoPipelineType.PIPELINE_TYPE_SERVICE,
});

export const mockGetServicePipeline = () =>
  rest.post<InspectPipelineRequest, Empty, PipelineInfo>(
    '/api/pps_v2.API/InspectPipeline',
    (req, res, ctx) => {
      return res(ctx.json(SERVICE_PIPELINE_INFO));
    },
  );

const SPOUT_PIPELINE_INFO = buildPipeline({
  pipeline: {
    name: 'montage',
    project: {name: 'default'},
  },
  details: {
    createdAt: getISOStringFromUnix(1690221505),
    datumTries: '3',
    outputBranch: 'master',
    description:
      'A pipeline that performs image edge detection by using the OpenCV library.',
    recentError: '',
  },
  version: '1',
  state: PipelineState.PIPELINE_RUNNING,
  stopped: false,
  type: PipelineInfoPipelineType.PIPELINE_TYPE_SPOUT,
  reason: '',
});

export const mockGetSpoutPipeline = () =>
  rest.post<InspectPipelineRequest, Empty, PipelineInfo>(
    '/api/pps_v2.API/InspectPipeline',
    (req, res, ctx) => {
      return res(ctx.json(SPOUT_PIPELINE_INFO));
    },
  );

export const mockGetManyPipelinesWithManyWorkers = (
  numPipelines = 1,
  numWorkers = 1,
) =>
  rest.post<ListPipelineRequest, Empty, PipelineInfo[]>(
    '/api/pps_v2.API/ListPipeline',
    (_req, res, ctx) => {
      const pipelines = range(0, numPipelines).map((i) => {
        return buildPipeline({
          pipeline: {name: `pipeline-${i}`},
          parallelism: String(numWorkers),
        });
      });

      return res(ctx.json(pipelines));
    },
  );

export const mockPipelinesEmpty = () =>
  rest.post<ListPipelineRequest, Empty, PipelineInfo[]>(
    '/api/pps_v2.API/ListPipeline',
    (_req, res, ctx) => res(ctx.json([])),
  );
