import {commitInfoFromObject} from '@pachyderm/node-pachyderm/dist/builders/pfs';
import {CommitInfo} from '@pachyderm/proto/pb/pfs/pfs_pb';

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

const commits: {[projectId: string]: CommitInfo[]} = {
  '1': tutorial,
  '2': customerTeam,
  default: [...tutorial],
  '7': [],
};

export default commits;
