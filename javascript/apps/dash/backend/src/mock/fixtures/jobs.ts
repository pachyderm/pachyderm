import {JobInfo, JobState} from '@pachyderm/proto/pb/pps/pps_pb';

import {jobInfoFromObject} from '@dash-backend/grpc/builders/pps';

const tutorial = [
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    createdAt: {seconds: 1616533099, nanos: 0},
    startedAt: {seconds: 1616533100, nanos: 0},
    finishedAt: {seconds: 1616533103, nanos: 0},
    job: {id: '23b9af7d5d4343219bc8e02ff44cd55a', pipeline: {name: 'montage'}},
    input: {
      crossList: [
        {pfs: {repo: 'edges', name: 'edges', branch: 'master'}},
        {pfs: {repo: 'images', name: 'images', branch: 'master'}},
      ],
    },
  }),
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    createdAt: {seconds: 1614126189, nanos: 0},
    startedAt: {seconds: 1614126190, nanos: 0},
    finishedAt: {seconds: 1614126193, nanos: 0},
    job: {id: '23b9af7d5d4343219bc8e02ff44cd55a', pipeline: {name: 'edges'}},
    input: {
      pfs: {repo: 'images', name: 'images', branch: 'master'},
    },
  }),
  jobInfoFromObject({
    state: JobState.JOB_FAILURE,
    createdAt: {seconds: 1614126189, nanos: 0},
    startedAt: {seconds: 1614126191, nanos: 0},
    finishedAt: {seconds: 1614126194, nanos: 0},
    job: {id: '33b9af7d5d4343219bc8e02ff44cd55a', pipeline: {name: 'montage'}},
    input: {
      crossList: [
        {pfs: {repo: 'edges', name: 'edges', branch: 'master'}},
        {pfs: {repo: 'images', name: 'images', branch: 'master'}},
      ],
    },
  }),
];

const customerTeam = [
  jobInfoFromObject({
    state: JobState.JOB_FAILURE,
    createdAt: {seconds: 1614136189, nanos: 0},
    startedAt: {seconds: 1614136190, nanos: 0},
    finishedAt: {seconds: 1614136193, nanos: 0},
    job: {
      id: '23b9af7d5d4343219bc8e02ff4acd33a',
      pipeline: {name: 'likelihoods'},
    },
  }),
  jobInfoFromObject({
    state: JobState.JOB_EGRESSING,
    createdAt: {seconds: 1614136189, nanos: 0},
    startedAt: {seconds: 1614136191, nanos: 0},
    job: {id: '23b9af7d5d4343219bc8e02ff4acd33a', pipeline: {name: 'models'}},
  }),
  jobInfoFromObject({
    state: JobState.JOB_KILLED,
    createdAt: {seconds: 1614136189, nanos: 0},
    startedAt: {seconds: 1614136192, nanos: 0},
    job: {
      id: '23b9af7d5d4343219bc8e02ff4acd33a',
      pipeline: {name: 'joint_call'},
    },
  }),
  jobInfoFromObject({
    state: JobState.JOB_RUNNING,
    createdAt: {seconds: 1614136189, nanos: 0},
    startedAt: {seconds: 1614136193, nanos: 0},
    job: {id: '23b9af7d5d4343219bc8e02ff4acd33a', pipeline: {name: 'split'}},
  }),
  jobInfoFromObject({
    state: JobState.JOB_STARTING,
    createdAt: {seconds: 1614136189, nanos: 0},
    startedAt: {seconds: 1614136194, nanos: 0},
    job: {id: '23b9af7d5d4343219bc8e02ff4acd33a', pipeline: {name: 'test'}},
  }),
];

const jobs: {[projectId: string]: JobInfo[]} = {
  '1': tutorial,
  '2': customerTeam,
  '3': tutorial,
  '4': customerTeam,
  '5': [],
  '6': [],
  '7': [],
  default: [...tutorial, ...customerTeam],
};

export default jobs;
