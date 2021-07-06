import {JobInfo, JobState} from '@pachyderm/proto/pb/pps/pps_pb';

import {jobInfoFromObject} from '@dash-backend/grpc/builders/pps';

import jobs from './jobs';

const tutorial = {
  '23b9af7d5d4343219bc8e02ff44cd55a': [jobs['1'][0], jobs['1'][1]],
  '33b9af7d5d4343219bc8e02ff44cd55a': [jobs['1'][2]],
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

const jobSets: {[projectId: string]: {[id: string]: JobInfo[]}} = {
  '1': tutorial,
  '2': customerTeam,
  '3': tutorial,
  '4': customerTeam,
  '5': {},
  '6': {},
  default: tutorial,
};

export default jobSets;
