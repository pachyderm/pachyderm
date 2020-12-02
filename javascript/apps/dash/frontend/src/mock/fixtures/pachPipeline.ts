import {
  JobState,
  Pipeline,
  PipelineInputType,
  PipelineState,
} from 'lib/graphqlTypes';

import {pachRepos} from './pachRepo';

export type PachPipelineFixtures = {
  [pachId: string]: Pipeline[];
};

export const pachPipelines: PachPipelineFixtures = {
  tutorial: [
    {
      id: 'montage',
      name: 'montage',
      version: 1,
      createdAt: 1606844888,
      state: PipelineState.Running,
      stopped: false,
      recentError: null,
      numOfJobsStarting: 0,
      numOfJobsRunning: 0,
      numOfJobsFailing: 0,
      numOfJobsSucceeding: 0,
      numOfJobsKilled: 0,
      numOfJobsMerging: 0,
      numOfJobsEgressing: 0,
      lastJobState: JobState.Success,
      inputs: [
        {
          id: 'edges',
          type: PipelineInputType.Pfs,
          joinedWith: [],
          groupedWith: [],
          crossedWith: [],
          unionedWith: [],
          pfsInput: {
            name: 'edges',
            repo: pachRepos['tutorial'][1],
          },
        },
        {
          id: 'images',
          type: PipelineInputType.Pfs,
          joinedWith: [],
          groupedWith: [],
          crossedWith: [],
          unionedWith: [],
          pfsInput: {
            name: 'images',
            repo: pachRepos['tutorial'][2],
          },
        },
      ],
      description: 'Creates a montage of images.',
    },
    {
      id: 'edges',
      name: 'edges',
      version: 1,
      createdAt: 1606844888,
      state: PipelineState.Running,
      stopped: false,
      recentError: null,
      numOfJobsStarting: 0,
      numOfJobsRunning: 0,
      numOfJobsFailing: 0,
      numOfJobsSucceeding: 0,
      numOfJobsKilled: 0,
      numOfJobsMerging: 0,
      numOfJobsEgressing: 0,
      lastJobState: JobState.Success,
      inputs: [
        {
          id: 'images',
          type: PipelineInputType.Pfs,
          joinedWith: [],
          groupedWith: [],
          crossedWith: [],
          unionedWith: [],
          pfsInput: {
            name: 'images',
            repo: pachRepos['tutorial'][2],
          },
        },
      ],
      description: 'Detects the edges of objects in the image.',
    },
  ],
};
