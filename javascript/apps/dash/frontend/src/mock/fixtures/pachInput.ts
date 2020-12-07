import {Input, InputType} from 'lib/graphqlTypes';

import {pachRepos} from './pachRepo';

export type PachInputFixtures = {
  [pachId: string]: Input[];
};

export const pachInputs: PachInputFixtures = {
  tutorial: [
    {
      id: 'edges',
      type: InputType.Pfs,
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
      type: InputType.Pfs,
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
};
