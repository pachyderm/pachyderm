import {JobInfo, JobState} from '@pachyderm/proto/pb/pps/pps_pb';

import {jobInfoFromObject} from '@dash-backend/grpc/builders/pps';

const tutorial = {
  '23b9af7d5d4343219bc8e02ff44cd55a': [
    jobInfoFromObject({
      state: JobState.JOB_SUCCESS,
      createdAt: {seconds: 1616533099, nanos: 0},
      job: {
        id: '23b9af7d5d4343219bc8e02ff44cd55a',
        pipeline: {name: 'montage'},
      },
    }),
    jobInfoFromObject({
      state: JobState.JOB_SUCCESS,
      createdAt: {seconds: 1614126189, nanos: 0},
      job: {id: '23b9af7d5d4343219bc8e02ff44cd55a', pipeline: {name: 'edges'}},
    }),
  ],
};

const customerTeam = {
  '23b9af7d5d4343219bc8e02ff4acd33a': [
    jobInfoFromObject({
      state: JobState.JOB_FAILURE,
      createdAt: {seconds: 1614136189, nanos: 0},
      job: {
        id: '23b9af7d5d4343219bc8e02ff4acd33a',
        pipeline: {name: 'likelihoods'},
      },
    }),
    jobInfoFromObject({
      state: JobState.JOB_EGRESSING,
      createdAt: {seconds: 1614136189, nanos: 0},
      job: {id: '23b9af7d5d4343219bc8e02ff4acd33a', pipeline: {name: 'models'}},
    }),
    jobInfoFromObject({
      state: JobState.JOB_KILLED,
      createdAt: {seconds: 1614136189, nanos: 0},
      job: {
        id: '23b9af7d5d4343219bc8e02ff4acd33a',
        pipeline: {name: 'joint_call'},
      },
    }),
    jobInfoFromObject({
      state: JobState.JOB_RUNNING,
      createdAt: {seconds: 1614136189, nanos: 0},
      job: {id: '23b9af7d5d4343219bc8e02ff4acd33a', pipeline: {name: 'split'}},
    }),
    jobInfoFromObject({
      state: JobState.JOB_STARTING,
      createdAt: {seconds: 1614136189, nanos: 0},
      job: {id: '23b9af7d5d4343219bc8e02ff4acd33a', pipeline: {name: 'test'}},
    }),
  ],
};

const jobsets: {[projectId: string]: {[id: string]: JobInfo[]}} = {
  '1': tutorial,
  '2': customerTeam,
  '3': tutorial,
  '4': customerTeam,
  '5': {},
  '6': {},
  default: tutorial,
};

export default jobsets;
