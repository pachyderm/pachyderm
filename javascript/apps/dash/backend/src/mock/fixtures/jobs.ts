import {JobInfo, JobState} from '@pachyderm/proto/pb/pps/pps_pb';

import {jobInfoFromObject} from '@dash-backend/grpc/builders/pps';

const tutorial = [
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    createdAt: {seconds: 1616533099, nanos: 0},
    job: {id: '23b9af7d5d4343219bc8e02ff44cd55a'},
    pipeline: {name: 'montage'},
  }),
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    createdAt: {seconds: 1614126189, nanos: 0},
    job: {id: '23b9af7d5d4343219bc8e02ff4acd33a'},
    pipeline: {name: 'edges'},
  }),
];

const customerTeam = [
  jobInfoFromObject({
    state: JobState.JOB_FAILURE,
    createdAt: {seconds: 1614136189, nanos: 0},
    job: {id: '3'},
    pipeline: {name: 'likelihoods'},
  }),
  jobInfoFromObject({
    state: JobState.JOB_EGRESSING,
    createdAt: {seconds: 1614136189, nanos: 0},
    job: {id: '4'},
    pipeline: {name: 'models'},
  }),
  jobInfoFromObject({
    state: JobState.JOB_KILLED,
    createdAt: {seconds: 1614136189, nanos: 0},
    job: {id: '5'},
    pipeline: {name: 'joint_call'},
  }),
  jobInfoFromObject({
    state: JobState.JOB_RUNNING,
    createdAt: {seconds: 1614136189, nanos: 0},
    job: {id: '6'},
    pipeline: {name: 'split'},
  }),
  jobInfoFromObject({
    state: JobState.JOB_STARTING,
    createdAt: {seconds: 1614136189, nanos: 0},
    job: {id: '7'},
    pipeline: {name: 'test'},
  }),
];

const jobs: {[projectId: string]: JobInfo[]} = {
  '1': tutorial,
  '2': customerTeam,
  '3': tutorial,
  '4': customerTeam,
  '5': tutorial,
  default: [...tutorial, ...customerTeam],
};

export default jobs;
