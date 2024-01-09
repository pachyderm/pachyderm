import merge from 'lodash/merge';
import {rest} from 'msw';

import {Empty} from '@dash-frontend/api/googleTypes';
import {
  InspectJobRequest,
  InspectJobSetRequest,
  JobInfo,
  JobState,
  ListJobRequest,
} from '@dash-frontend/api/pps';
import {getISOStringFromUnix} from '@dash-frontend/lib/dateTime';

export const buildJob = (job: Partial<JobInfo> = {}): JobInfo => {
  const defaultJob: JobInfo = {
    job: {
      id: 'default',
      pipeline: {
        name: 'default',
      },
    },
    state: JobState.JOB_SUCCESS,
    pipelineVersion: '1',
    restart: '0',
    dataProcessed: '0',
    dataSkipped: '0',
    dataFailed: '0',
    dataTotal: '0',
    dataRecovered: '0',
    stats: {
      downloadTime: '0s',
      processTime: '0s',
      uploadTime: '0s',
      downloadBytes: '0',
      uploadBytes: '0',
    },
    created: undefined,
    started: undefined,
    finished: undefined,
    reason: undefined,
    outputCommit: undefined,
    __typename: 'JobInfo',
  };

  return merge(defaultJob, job);
};

const MONTAGE_JOB_INFO_1D = buildJob({
  job: {
    id: '1dc67e479f03498badcc6180be4ee6ce',
    pipeline: {
      name: 'montage',
      project: {name: 'default'},
    },
  },
  outputCommit: {id: '1dc67e479f03498badcc6180be4ee6ce'},
  state: JobState.JOB_KILLED,
  created: getISOStringFromUnix(1690899647),
  started: getISOStringFromUnix(1690899649),
  finished: getISOStringFromUnix(1690899649),
  reason: 'job stopped',
});

const MONTAGE_JOB_INFO_A4 = buildJob({
  job: {
    id: 'a4423427351e42aabc40c1031928628e',
    pipeline: {
      name: 'montage',
      project: {name: 'default'},
    },
  },
  outputCommit: {id: 'a4423427351e42aabc40c1031928628e'},
  state: JobState.JOB_UNRUNNABLE,
  created: getISOStringFromUnix(1690899628),
  started: getISOStringFromUnix(1690899628),
  finished: getISOStringFromUnix(1690899630),
  reason: 'unrunnable because the following upstream pipelines failed: edges',
});

export const MONTAGE_JOB_INFO_5C = buildJob({
  job: {
    id: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
    pipeline: {
      name: 'montage',
      project: {name: 'default'},
    },
  },
  outputCommit: {
    id: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
    branch: {name: 'master'},
  },
  state: JobState.JOB_SUCCESS,
  created: getISOStringFromUnix(1690899612),
  started: getISOStringFromUnix(1690899615),
  finished: getISOStringFromUnix(1690899618),
  details: {
    input: {
      join: [],
      group: [],
      cross: [
        {
          pfs: {
            project: 'default',
            name: 'images',
            repo: 'images',
            repoType: 'user',
            branch: 'master',
            commit: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
            glob: '/',
            joinOn: '',
            outerJoin: false,
            groupBy: '',
            lazy: false,
            emptyFiles: false,
            s3: false,
          },
          join: [],
          group: [],
          cross: [],
          union: [],
        },
        {
          pfs: {
            project: 'default',
            name: 'edges',
            repo: 'edges',
            repoType: 'user',
            branch: 'master',
            commit: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
            glob: '/',
            joinOn: '',
            outerJoin: false,
            groupBy: '',
            lazy: false,
            emptyFiles: false,
            s3: false,
          },
          join: [],
          group: [],
          cross: [],
          union: [],
        },
      ],
      union: [],
    },
  },
  stats: {
    downloadTime: '1s',
    processTime: '1s',
    uploadTime: '1s',
    downloadBytes: '380910', // '380.91 kB'
    uploadBytes: '1420000', //'1.42 MB'
  },
  dataProcessed: '1',
  dataTotal: '1',
});

const MONTAGE_JOB_INFO_BC = buildJob({
  job: {
    id: 'bc322db1b24c4d16873e1a4db198b5c9',
    pipeline: {
      name: 'montage',
      project: {name: 'default'},
    },
  },
  outputCommit: {id: 'bc322db1b24c4d16873e1a4db198b5c9'},
  state: JobState.JOB_SUCCESS,
  created: getISOStringFromUnix(1690899573),
  started: getISOStringFromUnix(1690899582),
  finished: getISOStringFromUnix(1690899585),
  stats: {
    downloadBytes: '58650', // '58.65 kB'
    uploadBytes: '22790', // '22.79 kB'
  },
  dataProcessed: '1',
  dataTotal: '1',
  pipelineVersion: '2',
});

export const MOCK_MONTAGE_JOBS_INFO: JobInfo[] = [
  MONTAGE_JOB_INFO_1D,
  MONTAGE_JOB_INFO_A4,
  MONTAGE_JOB_INFO_5C,
  MONTAGE_JOB_INFO_BC,
];

export const EDGES_JOB_INFO_1D = buildJob({
  job: {
    id: '1dc67e479f03498badcc6180be4ee6ce',
    pipeline: {
      name: 'edges',
      project: {name: 'default'},
    },
  },
  outputCommit: {id: '1dc67e479f03498badcc6180be4ee6ce'},
  state: JobState.JOB_CREATED,
  created: getISOStringFromUnix(1690899647),
});

const EDGES_JOB_INFO_A4 = buildJob({
  job: {
    id: 'a4423427351e42aabc40c1031928628e',
    pipeline: {
      name: 'edges',
      project: {name: 'default'},
    },
  },
  outputCommit: {id: 'a4423427351e42aabc40c1031928628e'},
  state: JobState.JOB_FAILURE,
  created: getISOStringFromUnix(1690899628),
  started: getISOStringFromUnix(1690899628),
  finished: getISOStringFromUnix(1690899630),
  details: {},
  stats: {
    downloadBytes: '1370000', // '1.37 MB'
  },
  reason:
    'datum f6630a15d15771cc01bb949bbad96f82c5e16f792d6c2d7e5c2294abe97c5ada failed',
  dataFailed: '1',
  dataTotal: '4',
});

const EDGES_JOB_INFO_5C = buildJob({
  job: {
    id: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
    pipeline: {
      name: 'edges',
      project: {name: 'default'},
    },
  },
  outputCommit: {id: '5c1aa9bc87dd411ba5a1be0c80a3ebc2'},
  state: JobState.JOB_SUCCESS,
  created: getISOStringFromUnix(1690899573),
  started: getISOStringFromUnix(1690899581),
  finished: getISOStringFromUnix(1690899582),
  details: {},
  stats: {
    downloadBytes: '185430', // '185.43 kB'
    uploadBytes: '114060', // '114.06 kB'
  },
  dataProcessed: '1',
  dataSkipped: '2',
  dataTotal: '3',
});

const EDGES_JOB_INFO_CF = buildJob({
  job: {
    id: 'cf302e9203874015be2d453d75864721',
    pipeline: {
      name: 'edges',
      project: {name: 'default'},
    },
  },
  outputCommit: {id: 'cf302e9203874015be2d453d75864721'},
  state: JobState.JOB_SUCCESS,
  created: getISOStringFromUnix(1690899572),
  started: getISOStringFromUnix(1690899578),
  finished: getISOStringFromUnix(1690899580),
  pipelineVersion: '2',
  details: {},
  stats: {
    downloadBytes: '58650', // '58.65 kB'
    uploadBytes: '22790', // '22.79 kB'
  },
  dataProcessed: '1',
  dataTotal: '1',
});

export const MOCK_EDGES_JOBS_INFO: JobInfo[] = [
  EDGES_JOB_INFO_1D,
  EDGES_JOB_INFO_A4,
  EDGES_JOB_INFO_5C,
  EDGES_JOB_INFO_CF,
];

export const ALL_JOBS_INFOS: JobInfo[] = [
  ...MOCK_MONTAGE_JOBS_INFO,
  ...MOCK_EDGES_JOBS_INFO,
];

export const mockEmptyJob = () =>
  rest.post<ListJobRequest, Empty, JobInfo[]>(
    '/api/pps_v2.API/ListJob',
    (_req, res, ctx) => {
      return res(ctx.json([]));
    },
  );

export const mockGetMontageJob1D = () =>
  rest.post<InspectJobRequest, Empty, JobInfo>(
    '/api/pps_v2.API/InspectJob',
    async (req, res, ctx) => {
      const body = await req.json();
      if (body?.job?.id === '1dc67e479f03498badcc6180be4ee6ce') {
        return res(ctx.json(MONTAGE_JOB_INFO_1D));
      }
    },
  );

export const mockGetMontageJobs = () =>
  rest.post<ListJobRequest, Empty, JobInfo[]>(
    '/api/pps_v2.API/ListJob',
    async (req, res, ctx) => {
      const body = await req.json();
      if (body?.projects?.[0]?.name === 'default') {
        return res(ctx.json(MOCK_EDGES_JOBS_INFO));
      }
    },
  );

export const mockGetAllJobs = () =>
  rest.post<ListJobRequest, Empty, JobInfo[]>(
    '/api/pps_v2.API/ListJob',
    async (req, res, ctx) => {
      const body = await req.json();
      if (body?.projects?.[0]?.name === 'default') {
        return res(ctx.json(ALL_JOBS_INFOS));
      }
    },
  );

export const mockGetMontageJob5C = () =>
  rest.post<ListJobRequest, Empty, JobInfo[]>(
    '/api/pps_v2.API/ListJob',
    async (req, res, ctx) => {
      const body = await req.json();
      if (body?.projects?.[0]?.name === 'default') {
        return res(ctx.json([MONTAGE_JOB_INFO_5C]));
      }
    },
  );

export const mockInspectJobMontage5C = () =>
  rest.post<InspectJobRequest, Empty, JobInfo>(
    '/api/pps_v2.API/InspectJob',
    async (req, res, ctx) => {
      const body = await req.json();
      if (body?.job?.id === '5c1aa9bc87dd411ba5a1be0c80a3ebc2') {
        return res(ctx.json(MONTAGE_JOB_INFO_5C));
      }
    },
  );

export const mockEmptyJobSet = () =>
  rest.post<InspectJobSetRequest, Empty, JobInfo[]>(
    '/api/pps_v2.API/InspectJobSet',
    (_req, res, ctx) => {
      return res(ctx.json([]));
    },
  );

export const mockGetJobSet1D = () =>
  rest.post<InspectJobSetRequest, Empty, JobInfo[]>(
    '/api/pps_v2.API/InspectJobSet',
    async (req, res, ctx) => {
      const body = await req.json();
      if (body?.jobSet?.id === '1dc67e479f03498badcc6180be4ee6ce') {
        return res(ctx.json([MONTAGE_JOB_INFO_1D]));
      }
    },
  );
