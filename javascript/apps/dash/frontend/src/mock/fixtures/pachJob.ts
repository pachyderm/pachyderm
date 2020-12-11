import {Job, JobState} from 'lib/graphqlTypes';

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
    input: {
      name: '0-0',
      repo: pachRepos.tutorial[2],
      commit: pachRepos.tutorial[2].branches[0].commits[0],
      branchName: pachRepos.tutorial[2].branches[0].name,
    },
    outputRepo: pachRepos.tutorial[1],
    outputBranchName: pachRepos.tutorial[1].branches[0].name,
    outputCommit: pachRepos.tutoria[1].branches[0].commits[0],
  },
  {
    id: '2',
    pipeline: pachPipelines.tutorial[0],
    parentJobId: null,
    startedAt: 1606844888,
    finishedAt: 1606844888,
    state: JobState.Success,
    reason: null,
    input: {
      name: '2-0',
      repo: pachRepos.tutorial[2],
      commit: pachRepos.tutorial[2].branches[0].commits[0],
      branchName: pachRepos.tutorial[2].branches[0].name,
    },
    outputRepo: pachRepos.tutorial[0],
    outputBranchName: pachRepos.tutorial[0].branches[0].name,
    outputCommit: pachRepos.tutoria[0].branches[0].commits[0],
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
  input: {
    name: '1-0',
    repo: pachRepos.tutorial[1],
    commit: pachRepos.tutorial[1].branches[0].commits[0],
    branchName: pachRepos.tutorial[1].branches[0].name,
  },
  outputRepo: pachRepos.tutorial[0],
  outputBranchName: pachRepos.tutorial[0].branches[0].name,
});

export const pachJobs: PachJobFixtures = {
  tutorial: tutorialJobs,
};
