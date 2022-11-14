import {CommitInfo, OriginKind} from '@dash-backend/proto';
import {commitInfoFromObject} from '@dash-backend/proto/builders/pfs';

import {COMMITS} from './loadLimits';
import repos from './repos';

const tutorial = [
  commitInfoFromObject({
    commit: {
      id: '9d5daa0918ac4c43a476b86e3bb5e88e',
      branch: {
        name: 'master',
        repo: {name: repos['3'][0].getRepo()?.getName() || ''},
      },
    },
    description: 'added shinra hq building specs',
    sizeBytes: 543211,
    started: {
      seconds: 1614136189,
      nanos: 0,
    },
    finished: {
      seconds: 1614136191,
      nanos: 0,
    },
    originKind: OriginKind.USER,
  }),
  commitInfoFromObject({
    commit: {
      id: '0918ac4c43a476b86e3bb5e88e9d5daa',
      branch: {
        name: 'master',
        repo: {name: repos['3'][0].getRepo()?.getName() || ''},
      },
    },
    description: 'added fire materia',
    sizeBytes: 34371,
    started: {
      seconds: 1614136289,
      nanos: 0,
    },
    finished: {
      seconds: 1614136291,
      nanos: 0,
    },
    originKind: OriginKind.USER,
  }),
  commitInfoFromObject({
    commit: {
      id: '0918ac9d5daa76b86e3bb5e88e4c43a4',
      branch: {
        name: 'master',
        repo: {name: repos['3'][0].getRepo()?.getName() || ''},
      },
    },
    description: 'added mako',
    sizeBytes: 44276,
    started: {
      seconds: 1614136389,
      nanos: 0,
    },
    finished: {
      seconds: 1614136391,
      nanos: 0,
    },
  }),
  commitInfoFromObject({
    commit: {
      id: '0918ac9d5daa76b86e3bb5e88e4c43a5',
      branch: {
        name: 'master',
        repo: {name: repos['3'][0].getRepo()?.getName() || ''},
      },
    },
    description: 'in progress',
    sizeBytes: 100,
    started: {
      seconds: 1614133389,
      nanos: 0,
    },
  }),
  commitInfoFromObject({
    commit: {
      id: '0218ac9d5daa76b86e3bb5e88e4c43a5',
      branch: {
        name: 'master',
        repo: {name: repos['3'][0].getRepo()?.getName() || ''},
      },
    },
    description: 'in progress 2',
    sizeBytes: 100,
    started: {
      seconds: 1614133389,
      nanos: 0,
    },
  }),
  commitInfoFromObject({
    commit: {
      id: '0518ac9d5daa76b86e3bb5e88e4c43a5',
      branch: {
        name: 'master',
        repo: {name: repos['3'][0].getRepo()?.getName() || ''},
      },
    },
    description: 'in progress 3',
    sizeBytes: 100,
    started: {
      seconds: 1614133389,
      nanos: 0,
    },
  }),
  commitInfoFromObject({
    commit: {
      id: 'f4e23cf347c342d98bd9015e4c3ad52a',
      branch: {
        name: 'master',
        repo: {name: repos['3'][1].getRepo()?.getName() || ''},
      },
    },
    description: 'done',
    sizeBytes: 100,
    started: {
      seconds: 1614133389,
      nanos: 0,
    },
    finished: {
      seconds: 1614136391,
      nanos: 0,
    },
  }),
];

const customerTeam = [
  commitInfoFromObject({
    commit: {
      id: '23b9af7d5d4343219bc8e02ff4acd33a',
      branch: {
        name: 'master',
        repo: {name: 'training'},
      },
    },
    sizeBytes: 44276,
    started: {
      seconds: 1614136389,
      nanos: 0,
    },
    finished: {
      seconds: 1614136391,
      nanos: 0,
    },
    originKind: OriginKind.USER,
  }),
  commitInfoFromObject({
    commit: {
      id: '23b9af7d5d4343219bc8e02ff4acd33b',
      branch: {
        name: 'test',
        repo: {name: 'training'},
      },
    },
    sizeBytes: 44276,
    started: {
      seconds: 1614136389,
      nanos: 0,
    },
    finished: {
      seconds: 1614136391,
      nanos: 0,
    },
    originKind: OriginKind.USER,
  }),
  commitInfoFromObject({
    commit: {
      id: '23b9af7d5d4343219bc8e02ff4acd33c',
      branch: {
        name: 'test',
        repo: {name: 'training'},
      },
    },
    sizeBytes: 44276,
    started: {
      seconds: 1614136389,
      nanos: 0,
    },
    finished: {
      seconds: 1614136391,
      nanos: 0,
    },
    originKind: OriginKind.USER,
  }),
  commitInfoFromObject({
    commit: {
      id: '23b9af7d5d4343219bc8e02ff4acd33a',
      branch: {
        name: 'master',
        repo: {name: 'likelihoods'},
      },
    },
    started: {
      seconds: 1614136389,
      nanos: 0,
    },
    finished: {
      seconds: 1614136391,
      nanos: 0,
    },
  }),
  commitInfoFromObject({
    commit: {
      id: '23b9af7d5d4343219bc8e02ff4acd33a',
      branch: {
        name: 'master',
        repo: {name: 'models'},
      },
    },
    sizeBytes: 44276,
    started: {
      seconds: 1614136389,
      nanos: 0,
    },
    finished: {
      seconds: 1614136391,
      nanos: 0,
    },
    originKind: OriginKind.USER,
  }),
  commitInfoFromObject({
    commit: {
      id: '23b9af7d5d4343219bc8e02ff4acd33a',
      branch: {
        name: 'master',
        repo: {name: 'joint_call'},
      },
    },
    started: {
      seconds: 1614136389,
      nanos: 0,
    },
    finished: {
      seconds: 1614136391,
      nanos: 0,
    },
    originKind: OriginKind.USER,
  }),
  commitInfoFromObject({
    commit: {
      id: '23b9af7d5d4343219bc8e02ff4acd33a',
      branch: {
        name: 'master',
        repo: {name: 'split'},
      },
    },
    started: {
      seconds: 1614136389,
      nanos: 0,
    },
  }),
  commitInfoFromObject({
    commit: {
      id: '23b9af7d5d4343219bc8e02ff4acd33a',
      branch: {
        name: 'master',
        repo: {name: 'test'},
      },
    },
    started: {
      seconds: 1614136389,
      nanos: 0,
    },
  }),
];

const nestedFolderCommits = [
  commitInfoFromObject({
    commit: {
      id: 'd350c8d08a644ed5b2ee98c035ab6b34',
      branch: {
        name: 'master',
        repo: {name: 'images'},
      },
    },
    started: {
      seconds: 1614136389,
      nanos: 0,
    },
  }),
];

const getLoadCommits = (repoCount: number, commitCount: number) => {
  const now = Math.floor(new Date().getTime() / 1000);
  return [...new Array(repoCount).keys()].reduce(
    (commits: CommitInfo[], repoIndex) => {
      const repoCommits = [...new Array(commitCount).keys()].map(
        (commitIndex) => {
          return commitInfoFromObject({
            commit: {
              id: `${repoIndex}-${commitIndex}`,
              branch: {
                name: 'master',
                repo: {name: `load-repo-${repoIndex}`},
              },
            },
            started: {
              seconds: now + commitIndex * 10,
              nanos: 0,
            },
          });
        },
      );
      commits.push(...repoCommits);
      return commits;
    },
    [],
  );
};

const commits: {[projectId: string]: CommitInfo[]} = {
  '1': tutorial,
  '2': customerTeam,
  '3': tutorial,
  '5': nestedFolderCommits,
  default: [...tutorial],
  '7': [],
  '6': [],
  '9': getLoadCommits(1, COMMITS),
};

export default commits;
