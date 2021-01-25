import {Job, JobState} from '@graphqlTypes';

import {pachInputs} from './pachInput';
import {pachPipelines} from './pachPipeline';
import {pachRepos} from './pachRepo';

export type PachJobFixtures = {
  [pachId: string]: Job[];
};

const tutorialJobs: Job[] = [
  {
    id: '0',
    pipeline: pachPipelines.tutorial[1],
    parentJobId: null,
    startedAt: 1606844888,
    finishedAt: 1606844888,
    state: JobState.Success,
    reason: 'Job completed',
    outputRepo: pachRepos.tutorial[1],
    input: pachInputs.tutorial[0],
  },
  {
    id: '2',
    pipeline: pachPipelines.tutorial[0],
    parentJobId: null,
    startedAt: 1606844888,
    finishedAt: 1606844888,
    state: JobState.Success,
    reason: null,
    outputRepo: pachRepos.tutorial[0],
    input: pachInputs.tutorial[0],
  },
];

tutorialJobs.push({
  id: '1',
  pipeline: pachPipelines.tutorial[0],
  parentJobId: tutorialJobs[0].id,
  startedAt: 1606844888,
  finishedAt: null,
  state: JobState.Running,
  reason: null,
  outputRepo: pachRepos.tutorial[0],

  input: pachInputs.tutorial[1],
});

export const pachJobs: PachJobFixtures = {
  tutorial: tutorialJobs,
};
