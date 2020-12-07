import {JobState, Pipeline, PipelineState} from 'lib/graphqlTypes';

import {pachInputs} from './pachInput';

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
      inputs: [pachInputs.tutorial[0], pachInputs.tutorial[1]],
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
      inputs: [pachInputs.tutorial[1]],
      description: 'Detects the edges of objects in the image.',
    },
  ],
};
