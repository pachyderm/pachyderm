import {FileType, OriginKind, Repo} from 'lib/graphqlTypes';

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
      branches: [
        {
          name: 'master',
          commits: [],
        },
      ],
    },
    {
      name: 'edges',
      createdAt: 1606844888,
      sizeInBytes: 1000,
      description: 'Some edges',
      isPipelineOutput: true,
      branches: [],
    },
    {
      name: 'images',
      createdAt: 1606844888,
      sizeInBytes: 1000,
      description: 'A bunch of images',
      isPipelineOutput: false,
      branches: [
        {
          name: 'master',
          commits: [
            {
              id: '296d1570c1204864a9f5a0d66dafec45',
              origin: OriginKind.User,
              parentCommitId: '9fdb1436f7dc477eb23fe5d56b4094d8',
              description: 'Second commit.',
              childCommitIds: [],
              started: 1606936556,
              finished: 1606936558,
              sizeInBytes: 2654026,
              files: [
                {
                  path: '/liberty.jpg',
                  fileType: FileType.File,
                  sizeInBytes: 58644,
                  committedAt: 1607092339
                },
                {
                  path: '/test',
                  fileType: FileType.Dir,
                  sizeInBytes: 58644,
                  committedAt: 1607100702
                },
              ],
            },
            {
              id: '9fdb1436f7dc477eb23fe5d56b4094d8',
              origin: OriginKind.User,
              description: 'First commit.',
              childCommitIds: ['296d1570c1204864a9f5a0d66dafec45'],
              started: 1606925585,
              finished: 1606925587,
              sizeInBytes: 1327013,
              files: [],
            },
          ],
        },
        {
          name: 'v1',
          commits: [],
        },
      ],
    },
  ],
};
