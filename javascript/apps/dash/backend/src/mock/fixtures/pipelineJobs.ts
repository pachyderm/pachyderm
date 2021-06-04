import {
  PipelineJobInfo,
  PipelineJobState,
} from '@pachyderm/proto/pb/pps/pps_pb';

import {pipelineJobInfoFromObject} from '@dash-backend/grpc/builders/pps';

const tutorial = [
  pipelineJobInfoFromObject({
    state: PipelineJobState.JOB_SUCCESS,
    createdAt: {seconds: 1616533099, nanos: 0},
    pipelineJob: {id: '23b9af7d5d4343219bc8e02ff44cd55a'},
    pipeline: {name: 'montage'},
  }),
  pipelineJobInfoFromObject({
    state: PipelineJobState.JOB_SUCCESS,
    createdAt: {seconds: 1614126189, nanos: 0},
    pipelineJob: {id: '23b9af7d5d4343219bc8e02ff4acd33a'},
    pipeline: {name: 'edges'},
  }),
];

const customerTeam = [
  pipelineJobInfoFromObject({
    state: PipelineJobState.JOB_FAILURE,
    createdAt: {seconds: 1614136189, nanos: 0},
    pipelineJob: {id: '3'},
    pipeline: {name: 'likelihoods'},
  }),
  pipelineJobInfoFromObject({
    state: PipelineJobState.JOB_EGRESSING,
    createdAt: {seconds: 1614136189, nanos: 0},
    pipelineJob: {id: '4'},
    pipeline: {name: 'models'},
  }),
  pipelineJobInfoFromObject({
    state: PipelineJobState.JOB_KILLED,
    createdAt: {seconds: 1614136189, nanos: 0},
    pipelineJob: {id: '5'},
    pipeline: {name: 'joint_call'},
  }),
  pipelineJobInfoFromObject({
    state: PipelineJobState.JOB_RUNNING,
    createdAt: {seconds: 1614136189, nanos: 0},
    pipelineJob: {id: '6'},
    pipeline: {name: 'split'},
  }),
  pipelineJobInfoFromObject({
    state: PipelineJobState.JOB_STARTING,
    createdAt: {seconds: 1614136189, nanos: 0},
    pipelineJob: {id: '7'},
    pipeline: {name: 'test'},
  }),
];

const pipelineJobs: {[projectId: string]: PipelineJobInfo[]} = {
  '1': tutorial,
  '2': customerTeam,
  '3': tutorial,
  '4': customerTeam,
  '5': [],
  '6': [],
  default: [...tutorial, ...customerTeam],
};

export default pipelineJobs;
