import {Job, JobState} from '@graphqlTypes';

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

const transform: Job['transform'] = {
  image: 'pachyderm/opencv',
  cmdList: ['python3', '/edges.py'],
};

const createdAt = 1622467490;
const finishedAt = 1622468200;

const pipelineJobs: Job[] = [
  {
    id: '0',
    pipelineName: 'split',
    state: JobState.JOB_FAILURE,
    inputString,
    inputBranch,
    transform,
    createdAt,
    finishedAt,
  },
  {
    id: '1',
    pipelineName: 'model',
    state: JobState.JOB_SUCCESS,
    inputString,
    inputBranch,
    transform,
    createdAt,
    finishedAt,
  },
  {
    id: '2',
    pipelineName: 'test',
    state: JobState.JOB_SUCCESS,
    inputString,
    inputBranch,
    transform,
    createdAt,
    finishedAt,
  },
  {
    id: '3',
    pipelineName: 'detect',
    state: JobState.JOB_SUCCESS,
    inputString,
    inputBranch,
    transform,
    createdAt,
    finishedAt,
  },
  {
    id: '4',
    pipelineName: 'select',
    state: JobState.JOB_SUCCESS,
    inputString,
    inputBranch,
    transform,
    createdAt,
    finishedAt,
  },
];

export default pipelineJobs;
