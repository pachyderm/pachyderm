import {JobInfo, JobState} from '@pachyderm/proto/pb/pps/pps_pb';

import {jobInfoFromObject} from '@dash-backend/grpc/builders/pps';

const tutorial = [
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    createdAt: {seconds: 1616533099, nanos: 0},
    id: '23b9af7d5d4343219bc8e02ff44cd55a',
  }),
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    createdAt: {seconds: 1614126189, nanos: 0},
    id: '23b9af7d5d4343219bc8e02ff4acd33a',
  }),
];

const customerTeam = [
  jobInfoFromObject({
    state: JobState.JOB_FAILURE,
    createdAt: {seconds: 1614126189, nanos: 0},
    id: '3',
  }),
  jobInfoFromObject({
    state: JobState.JOB_EGRESSING,
    createdAt: {seconds: 1614126189, nanos: 0},
    id: '4',
  }),
  jobInfoFromObject({
    state: JobState.JOB_KILLED,
    createdAt: {seconds: 1614126189, nanos: 0},
    id: '5',
  }),
  jobInfoFromObject({
    state: JobState.JOB_RUNNING,
    createdAt: {seconds: 1614126189, nanos: 0},
    id: '6',
  }),
  jobInfoFromObject({
    state: JobState.JOB_STARTING,
    createdAt: {seconds: 1614126189, nanos: 0},
    id: '7',
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
