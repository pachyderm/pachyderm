import {
  ProjectDetailsQuery,
  NodeState,
  JobState,
  ProjectStatus,
  mockProjectDetailsQuery,
  mockProjectsQuery,
  ProjectsQuery,
  mockProjectStatusQuery,
} from '@graphqlTypes';

export const ALL_PROJECTS: ProjectsQuery = {
  projects: [
    {
      id: 'ProjectA',
      description: 'A description for project a',
      __typename: 'Project',
      status: null,
    },
    {
      id: 'ProjectB',
      description: 'A description for project b',
      __typename: 'Project',
      status: null,
    },
    {
      id: 'ProjectC',
      description: 'A description for project c',
      __typename: 'Project',
      status: null,
    },
  ],
};

export const MOCK_PROJECT_DETAILS: ProjectDetailsQuery = {
  projectDetails: {
    sizeDisplay: '3 kB',
    repoCount: 3,
    pipelineCount: 2,
    jobSets: [
      {
        id: '23b9af7d5d4343219bc8e02ff44cd55a',
        state: JobState.JOB_SUCCESS,
        createdAt: 1614126189,
        startedAt: 1614126190,
        finishedAt: 1616533103,
        inProgress: false,
        jobs: [
          {
            id: '23b9af7d5d4343219bc8e02ff44cd55a',
            state: JobState.JOB_SUCCESS,
            nodeState: NodeState.SUCCESS,
            createdAt: 1614126189,
            startedAt: 1614126190,
            finishedAt: 1614126193,
            restarts: 0,
            pipelineName: 'edges',
            reason: '',
            dataProcessed: 0,
            dataSkipped: 0,
            dataFailed: 0,
            dataTotal: 100,
            dataRecovered: 0,
            downloadBytesDisplay: '0 B',
            uploadBytesDisplay: '0 B',
            outputCommit: '23b9af7d5d4343219bc8e02ff44cd55a',
            __typename: 'Job',
            inputString:
              '{\n  "pfs": {\n    "project": "Solar-Panel-Data-Sorting",\n    "name": "images",\n    "repo": "images",\n    "repoType": "",\n    "branch": "master",\n    "commit": "",\n    "glob": "",\n    "joinOn": "",\n    "outerJoin": false,\n    "groupBy": "",\n    "lazy": false,\n    "emptyFiles": false,\n    "s3": false\n  },\n  "join": [],\n  "group": [],\n  "cross": [],\n  "union": []\n}',
            inputBranch: 'master',
            transformString: null,
            transform: null,
          },
          {
            id: '23b9af7d5d4343219bc8e02ff44cd55a',
            state: JobState.JOB_SUCCESS,
            nodeState: NodeState.SUCCESS,
            createdAt: 1616533099,
            startedAt: 1616533100,
            finishedAt: 1616533103,
            restarts: 0,
            pipelineName: 'montage',
            reason: '',
            dataProcessed: 2,
            dataSkipped: 1,
            dataFailed: 1,
            dataTotal: 4,
            dataRecovered: 0,
            downloadBytesDisplay: '0 B',
            uploadBytesDisplay: '0 B',
            outputCommit: '23b9af7d5d4343219bc8e02ff44cd55a',
            __typename: 'Job',
            inputString:
              '{\n  "join": [],\n  "group": [],\n  "cross": [\n    {\n      "pfs": {\n        "project": "Solar-Panel-Data-Sorting",\n        "name": "edges",\n        "repo": "edges",\n        "repoType": "",\n        "branch": "master",\n        "commit": "",\n        "glob": "",\n        "joinOn": "",\n        "outerJoin": false,\n        "groupBy": "",\n        "lazy": false,\n        "emptyFiles": false,\n        "s3": false\n      },\n      "join": [],\n      "group": [],\n      "cross": [],\n      "union": []\n    },\n    {\n      "pfs": {\n        "project": "Solar-Panel-Data-Sorting",\n        "name": "images",\n        "repo": "images",\n        "repoType": "",\n        "branch": "master",\n        "commit": "",\n        "glob": "",\n        "joinOn": "",\n        "outerJoin": false,\n        "groupBy": "",\n        "lazy": false,\n        "emptyFiles": false,\n        "s3": false\n      },\n      "join": [],\n      "group": [],\n      "cross": [],\n      "union": []\n    }\n  ],\n  "union": []\n}',
            inputBranch: null,
            transformString: null,
            transform: null,
          },
        ],
        __typename: 'JobSet',
      },
      {
        id: '33b9af7d5d4343219bc8e02ff44cd55a',
        state: JobState.JOB_FAILURE,
        createdAt: 1614126190,
        startedAt: 1614126191,
        finishedAt: 1614126194,
        inProgress: false,
        jobs: [
          {
            id: '33b9af7d5d4343219bc8e02ff44cd55a',
            state: JobState.JOB_FAILURE,
            nodeState: NodeState.ERROR,
            createdAt: 1614126190,
            startedAt: 1614126191,
            finishedAt: 1614126194,
            restarts: 0,
            pipelineName: 'montage',
            reason:
              'datum 64b95f0fe1a787b6c26ec7ede800be6f2b97616f3224592d91cbfe1cfccd00a1 failed',
            dataProcessed: 0,
            dataSkipped: 0,
            dataFailed: 4,
            dataTotal: 5,
            dataRecovered: 0,
            downloadBytesDisplay: '2.9 kB',
            uploadBytesDisplay: '0 B',
            outputCommit: null,
            __typename: 'Job',
            inputString:
              '{\n  "join": [],\n  "group": [],\n  "cross": [\n    {\n      "pfs": {\n        "project": "Solar-Panel-Data-Sorting",\n        "name": "edges",\n        "repo": "edges",\n        "repoType": "",\n        "branch": "master",\n        "commit": "",\n        "glob": "",\n        "joinOn": "",\n        "outerJoin": false,\n        "groupBy": "",\n        "lazy": false,\n        "emptyFiles": false,\n        "s3": false\n      },\n      "join": [],\n      "group": [],\n      "cross": [],\n      "union": []\n    },\n    {\n      "pfs": {\n        "project": "Solar-Panel-Data-Sorting",\n        "name": "images",\n        "repo": "images",\n        "repoType": "",\n        "branch": "master",\n        "commit": "",\n        "glob": "",\n        "joinOn": "",\n        "outerJoin": false,\n        "groupBy": "",\n        "lazy": false,\n        "emptyFiles": false,\n        "s3": false\n      },\n      "join": [],\n      "group": [],\n      "cross": [],\n      "union": []\n    }\n  ],\n  "union": []\n}',
            inputBranch: null,
            transformString: null,
            transform: null,
          },
        ],
        __typename: 'JobSet',
      },
      {
        id: '7798fhje5d4343219bc8e02ff4acd33a',
        state: JobState.JOB_FINISHING,
        createdAt: 1614125000,
        startedAt: 1614125000,
        finishedAt: null,
        inProgress: true,
        jobs: [
          {
            id: '7798fhje5d4343219bc8e02ff4acd33a',
            state: JobState.JOB_FINISHING,
            nodeState: NodeState.RUNNING,
            createdAt: 1614125000,
            startedAt: 1614125000,
            finishedAt: null,
            restarts: 0,
            pipelineName: 'montage',
            reason: '',
            dataProcessed: 0,
            dataSkipped: 0,
            dataFailed: 0,
            dataTotal: 0,
            dataRecovered: 0,
            downloadBytesDisplay: '0 B',
            uploadBytesDisplay: '0 B',
            outputCommit: null,
            __typename: 'Job',
            inputString: null,
            inputBranch: null,
            transformString: null,
            transform: null,
          },
        ],
        __typename: 'JobSet',
      },
      {
        id: 'o90du4js5d4343219bc8e02ff4acd33a',
        state: JobState.JOB_KILLED,
        createdAt: 1614123000,
        startedAt: 1614123000,
        finishedAt: null,
        inProgress: true,
        jobs: [
          {
            id: 'o90du4js5d4343219bc8e02ff4acd33a',
            state: JobState.JOB_KILLED,
            nodeState: NodeState.ERROR,
            createdAt: 1614123000,
            startedAt: 1614123000,
            finishedAt: null,
            restarts: 0,
            pipelineName: 'montage',
            reason: '',
            dataProcessed: 0,
            dataSkipped: 0,
            dataFailed: 0,
            dataTotal: 0,
            dataRecovered: 0,
            downloadBytesDisplay: '0 B',
            uploadBytesDisplay: '0 B',
            outputCommit: null,
            __typename: 'Job',
            inputString: null,
            inputBranch: null,
            transformString: null,
            transform: null,
          },
        ],
        __typename: 'JobSet',
      },
    ],
    __typename: 'ProjectDetails',
  },
};

export const MOCK_PROJECT_DETAILS_NO_JOBS: ProjectDetailsQuery = {
  projectDetails: {
    sizeDisplay: '3 kB',
    repoCount: 3,
    pipelineCount: 2,
    jobSets: [],
    __typename: 'ProjectDetails',
  },
};

export const MOCK_EMPTY_PROJECT_DETAILS: ProjectDetailsQuery = {
  projectDetails: {
    sizeDisplay: '0 B',
    repoCount: 0,
    pipelineCount: 0,
    jobSets: [],
    __typename: 'ProjectDetails',
  },
};

export const mockProjects = () =>
  mockProjectsQuery((_req, res, ctx) => {
    return res(ctx.data(ALL_PROJECTS));
  });

export const mockHealthyProjectStatus = () =>
  mockProjectStatusQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        projectStatus: {
          id: 'ProjectA',
          status: ProjectStatus.HEALTHY,
          __typename: 'Project',
        },
      }),
    );
  });

export const mockEmptyProjectDetails = () =>
  mockProjectDetailsQuery((_req, res, ctx) => {
    return res(ctx.data(MOCK_EMPTY_PROJECT_DETAILS));
  });
