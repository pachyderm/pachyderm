import {PipelineJob, PipelineJobState} from '@graphqlTypes';

const inputString = JSON.stringify(
  {
    pfs: {
      name: 'images',
      repo: 'images',
      branch: 'master',
      commit: '1233456',
      glob: '/*',
    },
  },
  null,
  2,
);

const inputBranch = 'master';

const transform: PipelineJob['transform'] = {
  image: 'pachyderm/opencv',
  cmdList: ['python3', '/edges.py'],
};

const createdAt = 1622467490;
const finishedAt = 1622468200;

const pipelineJobs: PipelineJob[] = [
  {
    id: '0',
    pipelineName: 'split',
    state: PipelineJobState.JOB_FAILURE,
    inputString,
    inputBranch,
    transform,
    createdAt,
    finishedAt,
  },
  {
    id: '1',
    pipelineName: 'model',
    state: PipelineJobState.JOB_SUCCESS,
    inputString,
    inputBranch,
    transform,
    createdAt,
    finishedAt,
  },
  {
    id: '2',
    pipelineName: 'test',
    state: PipelineJobState.JOB_SUCCESS,
    inputString,
    inputBranch,
    transform,
    createdAt,
    finishedAt,
  },
  {
    id: '3',
    pipelineName: 'detect',
    state: PipelineJobState.JOB_SUCCESS,
    inputString,
    inputBranch,
    transform,
    createdAt,
    finishedAt,
  },
  {
    id: '4',
    pipelineName: 'select',
    state: PipelineJobState.JOB_SUCCESS,
    inputString,
    inputBranch,
    transform,
    createdAt,
    finishedAt,
  },
];

export default pipelineJobs;
