import {
  mockJobsQuery,
  mockJobSetQuery,
  mockJobSetsQuery,
  JobsQuery,
  Job,
  JobSet,
  JobState,
  NodeState,
  JobSetsQuery,
  mockJobQuery,
} from '@graphqlTypes';
import merge from 'lodash/merge';

export const buildJob = (job: Partial<Job>): Job => {
  const defaultJob = {
    id: 'default',
    state: JobState.JOB_SUCCESS,
    nodeState: NodeState.SUCCESS,
    pipelineName: 'default',
    pipelineVersion: 1,
    restarts: 0,
    dataProcessed: 0,
    dataSkipped: 0,
    dataFailed: 0,
    dataTotal: 0,
    dataRecovered: 0,
    downloadBytesDisplay: '0 B',
    uploadBytesDisplay: '0 B',
    jsonDetails: '{}',
    createdAt: null,
    startedAt: null,
    finishedAt: null,
    reason: '',
    outputCommit: null,
    inputString: null,
    inputBranch: null,
    outputBranch: 'master',
    transformString: null,
    transform: null,
    __typename: 'Job',
  };

  return merge(defaultJob, job);
};

export const buildJobSet = (jobset: Partial<JobSet>): JobSet => {
  const defaultJobSet = {
    id: 'default',
    createdAt: null,
    startedAt: null,
    finishedAt: null,
    inProgress: false,
    jobs: [],
    state: JobState.JOB_SUCCESS,
    __typename: 'JobSet',
  };
  return merge(defaultJobSet, jobset);
};

export const MOCK_EMPTY_JOBS: JobsQuery = {
  jobs: {
    items: [],
    cursor: null,
    hasNextPage: false,
    __typename: 'PageableJob',
  },
};

const MONTAGE_JOB_1D: Job = buildJob({
  id: '1dc67e479f03498badcc6180be4ee6ce',
  outputCommit: '1dc67e479f03498badcc6180be4ee6ce',
  state: JobState.JOB_KILLED,
  nodeState: NodeState.ERROR,
  createdAt: 1690899647,
  startedAt: 1690899649,
  finishedAt: 1690899649,
  pipelineName: 'montage',
  reason: 'job stopped',
});

const MONTAGE_JOB_A4: Job = buildJob({
  id: 'a4423427351e42aabc40c1031928628e',
  outputCommit: 'a4423427351e42aabc40c1031928628e',
  state: JobState.JOB_UNRUNNABLE,
  nodeState: NodeState.ERROR,
  createdAt: 1690899628,
  startedAt: 1690899628,
  finishedAt: 1690899630,
  reason: 'unrunnable because the following upstream pipelines failed: edges',
  pipelineName: 'montage',
});

const MONTAGE_JOB_5C: Job = buildJob({
  id: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
  outputCommit: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
  state: JobState.JOB_SUCCESS,
  nodeState: NodeState.SUCCESS,
  createdAt: 1690899612,
  startedAt: 1690899615,
  finishedAt: 1690899618,
  pipelineName: 'montage',
  downloadBytesDisplay: '380.91 kB',
  uploadBytesDisplay: '1.42 MB',
  dataProcessed: 1,
  dataTotal: 1,
  jsonDetails:
    '{\n  "pipelineVersion": 1,\n  "dataTotal": 1,\n  "stats": {\n    "downloadTime": {\n      "nanos": 35472750\n    },\n    "processTime": {\n      "seconds": 1,\n      "nanos": 998646751\n    },\n    "uploadTime": {\n      "nanos": 16216500\n    },\n    "downloadBytes": 380904,\n    "uploadBytes": 1411691\n  }\n}',
  inputString:
    '{\n  "join": [],\n  "group": [],\n  "cross": [\n    {\n      "pfs": {\n        "project": "default",\n        "name": "images",\n        "repo": "images",\n        "repoType": "user",\n        "branch": "master",\n        "commit": "5c1aa9bc87dd411ba5a1be0c80a3ebc2",\n        "glob": "/",\n        "joinOn": "",\n        "outerJoin": false,\n        "groupBy": "",\n        "lazy": false,\n        "emptyFiles": false,\n        "s3": false\n      },\n      "join": [],\n      "group": [],\n      "cross": [],\n      "union": []\n    },\n    {\n      "pfs": {\n        "project": "default",\n        "name": "edges",\n        "repo": "edges",\n        "repoType": "user",\n        "branch": "master",\n        "commit": "5c1aa9bc87dd411ba5a1be0c80a3ebc2",\n        "glob": "/",\n        "joinOn": "",\n        "outerJoin": false,\n        "groupBy": "",\n        "lazy": false,\n        "emptyFiles": false,\n        "s3": false\n      },\n      "join": [],\n      "group": [],\n      "cross": [],\n      "union": []\n    }\n  ],\n  "union": []\n}',
});

const MONTAGE_JOB_BC: Job = buildJob({
  id: 'bc322db1b24c4d16873e1a4db198b5c9',
  outputCommit: 'bc322db1b24c4d16873e1a4db198b5c9',
  state: JobState.JOB_SUCCESS,
  nodeState: NodeState.SUCCESS,
  createdAt: 1690899573,
  startedAt: 1690899582,
  finishedAt: 1690899585,
  pipelineName: 'montage',
  pipelineVersion: 2,
  downloadBytesDisplay: '58.65 kB',
  uploadBytesDisplay: '22.79 kB',
  dataProcessed: 1,
  dataTotal: 1,
  jsonDetails:
    '{\n  "pipelineVersion": 1,\n  "dataTotal": 1,\n  "stats": {\n    "downloadTime": {\n      "nanos": 65156542\n    },\n    "processTime": {\n      "seconds": 1,\n      "nanos": 867227126\n    },\n    "uploadTime": {\n      "nanos": 30792875\n    },\n    "downloadBytes": 380904,\n    "uploadBytes": 1411760\n  }\n}',
});

export const MOCK_MONTAGE_JOBS: Job[] = [
  MONTAGE_JOB_1D,
  MONTAGE_JOB_A4,
  MONTAGE_JOB_5C,
  MONTAGE_JOB_BC,
];

const EDGES_JOB_1D: Job = buildJob({
  id: '1dc67e479f03498badcc6180be4ee6ce',
  outputCommit: '1dc67e479f03498badcc6180be4ee6ce',
  state: JobState.JOB_CREATED,
  nodeState: NodeState.RUNNING,
  createdAt: 1690899647,
  pipelineName: 'edges',
  jsonDetails: '{\n  "pipelineVersion": 1,\n  "stats": {}\n}',
});

const EDGES_JOB_A4: Job = buildJob({
  id: 'a4423427351e42aabc40c1031928628e',
  outputCommit: 'a4423427351e42aabc40c1031928628e',
  state: JobState.JOB_FAILURE,
  nodeState: NodeState.ERROR,
  createdAt: 1690899628,
  startedAt: 1690899628,
  finishedAt: 1690899630,
  reason:
    'datum f6630a15d15771cc01bb949bbad96f82c5e16f792d6c2d7e5c2294abe97c5ada failed',
  pipelineName: 'edges',
  downloadBytesDisplay: '1.37 MB',
  dataFailed: 1,
  dataTotal: 4,
  jsonDetails:
    '{\n  "pipelineVersion": 1,\n  "dataTotal": 4,\n  "dataFailed": 1,\n  "stats": {\n    "downloadTime": {\n      "nanos": 30827750\n    },\n    "processTime": {\n      "nanos": 587331917\n    },\n    "uploadTime": {},\n    "downloadBytes": 1368152\n  }\n}',
});

const EDGES_JOB_5C: Job = buildJob({
  id: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
  outputCommit: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
  state: JobState.JOB_SUCCESS,
  nodeState: NodeState.SUCCESS,
  createdAt: 1690899573,
  startedAt: 1690899581,
  finishedAt: 1690899582,
  pipelineName: 'edges',
  downloadBytesDisplay: '185.43 kB',
  uploadBytesDisplay: '114.06 kB',
  dataProcessed: 1,
  dataSkipped: 2,
  dataTotal: 3,
  jsonDetails:
    '{\n  "pipelineVersion": 1,\n  "dataTotal": 3,\n  "stats": {\n    "downloadTime": {\n      "nanos": 11066000\n    },\n    "processTime": {\n      "nanos": 505251417\n    },\n    "uploadTime": {\n      "nanos": 2163334\n    },\n    "downloadBytes": 104836,\n    "uploadBytes": 75995\n  }\n}',
});

const EDGES_JOB_CF: Job = buildJob({
  id: 'cf302e9203874015be2d453d75864721',
  outputCommit: 'cf302e9203874015be2d453d75864721',
  state: JobState.JOB_SUCCESS,
  nodeState: NodeState.SUCCESS,
  createdAt: 1690899572,
  startedAt: 1690899578,
  finishedAt: 1690899580,
  pipelineName: 'edges',
  pipelineVersion: 2,
  downloadBytesDisplay: '58.65 kB',
  uploadBytesDisplay: '22.79 kB',
  dataProcessed: 1,
  dataTotal: 1,
  jsonDetails:
    '{\n  "pipelineVersion": 1,\n  "dataTotal": 1,\n  "stats": {\n    "downloadTime": {\n      "nanos": 48069458\n    },\n    "processTime": {\n      "seconds": 2,\n      "nanos": 363580668\n    },\n    "uploadTime": {\n      "nanos": 9280708\n    },\n    "downloadBytes": 58644,\n    "uploadBytes": 22783\n  }\n}',
});

export const MOCK_EDGES_JOBS: Job[] = [
  EDGES_JOB_1D,
  EDGES_JOB_A4,
  EDGES_JOB_5C,
  EDGES_JOB_CF,
];

export const ALL_JOBS: Job[] = [...MOCK_MONTAGE_JOBS, ...MOCK_EDGES_JOBS];

export const mockEmptyJobs = () =>
  mockJobsQuery((_req, res, ctx) => {
    return res(ctx.data(MOCK_EMPTY_JOBS));
  });

export const mockGetEdgesJobs = () =>
  mockJobsQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        jobs: {
          items: MOCK_EDGES_JOBS,
          cursor: null,
          hasNextPage: false,
          __typename: 'PageableJob',
        },
      }),
    );
  });

export const mockGetMontageJob_5C = () =>
  mockJobQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        job: MONTAGE_JOB_5C,
      }),
    );
  });

export const mockGetMontageJob_1D = () =>
  mockJobQuery((req, res, ctx) => {
    if (req.variables.args.id === '1dc67e479f03498badcc6180be4ee6ce') {
      return res(
        ctx.data({
          job: MONTAGE_JOB_1D,
        }),
      );
    } else {
      return res(ctx.errors([]));
    }
  });

export const mockEmptyJob = () =>
  mockJobQuery((_req, res, ctx) => {
    return res(
      ctx.errors([
        {
          message: 'resource not found',
          path: ['job'],
        },
      ]),
    );
  });

export const mockGetMontageJobs = () =>
  mockJobsQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        jobs: {
          items: MOCK_EDGES_JOBS,
          cursor: null,
          hasNextPage: false,
          __typename: 'PageableJob',
        },
      }),
    );
  });

export const mockGetAllJobs = () =>
  mockJobsQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        jobs: {
          items: ALL_JOBS,
          cursor: null,
          hasNextPage: false,
          __typename: 'PageableJob',
        },
      }),
    );
  });

export const MOCK_EMPTY_JOBSETS: JobSetsQuery = {
  jobSets: {
    items: [],
    cursor: null,
    hasNextPage: false,
    __typename: 'PageableJobSet',
  },
};

export const JOBSET_1D: JobSet = buildJobSet({
  id: '1dc67e479f03498badcc6180be4ee6ce',
  state: JobState.JOB_CREATED,
  createdAt: 1690899647,
  startedAt: 1690899578,
  inProgress: true,
  jobs: [MONTAGE_JOB_1D, EDGES_JOB_1D],
});

export const JOBSET_A4: JobSet = buildJobSet({
  id: 'a4423427351e42aabc40c1031928628e',
  state: JobState.JOB_FAILURE,
  createdAt: 1690899628,
  startedAt: 1690899628,
  finishedAt: 1690899630,
  jobs: [MONTAGE_JOB_A4, EDGES_JOB_A4],
});

export const JOBSET_5C: JobSet = buildJobSet({
  id: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
  state: JobState.JOB_SUCCESS,
  createdAt: 1690899612,
  startedAt: 1690899615,
  finishedAt: 1690899618,
  jobs: [MONTAGE_JOB_5C, EDGES_JOB_5C],
});

export const JOBSET_CF: JobSet = buildJobSet({
  id: 'cf302e9203874015be2d453d75864721',
  state: JobState.JOB_SUCCESS,
  createdAt: 1690899572,
  startedAt: 1690899578,
  finishedAt: 1690899580,
  jobs: [EDGES_JOB_CF],
});

export const JOBSET_BC: JobSet = buildJobSet({
  id: 'bc322db1b24c4d16873e1a4db198b5c9',
  state: JobState.JOB_SUCCESS,
  createdAt: 1690899573,
  startedAt: 1690899778,
  finishedAt: 1690903580,
  jobs: [MONTAGE_JOB_BC],
});

export const ALL_JOBSETS: JobSet[] = [
  JOBSET_1D,
  JOBSET_A4,
  JOBSET_5C,
  JOBSET_CF,
  JOBSET_BC,
];

export const mockEmptyJobsets = () =>
  mockJobSetsQuery((_req, res, ctx) => {
    return res(ctx.data(MOCK_EMPTY_JOBSETS));
  });

export const mockGetAllJobsets = () =>
  mockJobSetsQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        jobSets: {
          items: ALL_JOBSETS,
          cursor: null,
          hasNextPage: false,
          __typename: 'PageableJobSet',
        },
      }),
    );
  });

export const mockGetJobSet_1D = () =>
  mockJobSetQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        jobSet: JOBSET_1D,
      }),
    );
  });
