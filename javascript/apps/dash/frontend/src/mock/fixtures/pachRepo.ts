import {Repo} from '@graphqlTypes';

export type PachRepoFixtures = {
  [pachId: string]: Repo[];
};

export const pachRepos: PachRepoFixtures = {
  tutorial: [
    {
      name: 'montage',
      createdAt: 1606844888,
      sizeInBytes: 1000,
      description: 'A montage',
      isPipelineOutput: true,
    },
    {
      name: 'edges',
      createdAt: 1606844888,
      sizeInBytes: 1000,
      description: 'Some edges',
      isPipelineOutput: true,
    },
    {
      name: 'images',
      createdAt: 1606844888,
      sizeInBytes: 1000,
      description: 'A bunch of images',
      isPipelineOutput: false,
    },
  ],
};
