import {JobState, JobInfo} from '@pachyderm/node-pachyderm';
import {jobInfoFromObject} from '@pachyderm/node-pachyderm/dist/builders/pps';

import {JOBS} from './loadLimits';

const tutorial = [
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    createdAt: {seconds: 1616533099, nanos: 100},
    startedAt: {seconds: 1616533100, nanos: 100},
    finishedAt: {seconds: 1616533103, nanos: 100},
    job: {id: '23b9af7d5d4343219bc8e02ff44cd55a', pipeline: {name: 'montage'}},
    input: {
      crossList: [
        {pfs: {repo: 'edges', name: 'edges', branch: 'master'}},
        {pfs: {repo: 'images', name: 'images', branch: 'master'}},
      ],
    },
    outputCommit: {
      id: '23b9af7d5d4343219bc8e02ff44cd55a',
      branch: {
        name: 'master',
        repo: {
          name: 'montage',
        },
      },
    },
    dataFailed: 0,
    dataTotal: 100,
  }),
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    createdAt: {seconds: 1614126189, nanos: 100},
    startedAt: {seconds: 1614126190, nanos: 100},
    finishedAt: {seconds: 1614126193, nanos: 100},
    job: {id: '23b9af7d5d4343219bc8e02ff44cd55a', pipeline: {name: 'edges'}},
    input: {
      pfs: {repo: 'images', name: 'images', branch: 'master'},
    },
    outputCommit: {
      id: '23b9af7d5d4343219bc8e02ff44cd55a',
      branch: {
        name: 'master',
        repo: {
          name: 'edges',
        },
      },
    },
    dataFailed: 0,
    dataTotal: 100,
  }),
  jobInfoFromObject({
    state: JobState.JOB_FAILURE,
    reason:
      'datum 64b95f0fe1a787b6c26ec7ede800be6f2b97616f3224592d91cbfe1cfccd00a1 failed',
    createdAt: {seconds: 1614126189, nanos: 100},
    startedAt: {seconds: 1614126191, nanos: 100},
    finishedAt: {seconds: 1614126194, nanos: 100},
    job: {id: '33b9af7d5d4343219bc8e02ff44cd55a', pipeline: {name: 'montage'}},
    input: {
      crossList: [
        {pfs: {repo: 'edges', name: 'edges', branch: 'master'}},
        {pfs: {repo: 'images', name: 'images', branch: 'master'}},
      ],
    },
    pipelineVersion: 1,
    dataTotal: 5,
    dataFailed: 4,
    datumTries: 3,
    salt: 'd5631d7df40d4b1195bc46f1f146d6a5',
    stats: {
      downloadTime: {
        nanos: 269391100,
        seconds: 10,
      },
      processTime: {
        seconds: 20,
        nanos: 531186700,
      },
      uploadTime: {
        seconds: 30,
        nanos: 231186700,
      },
      downloadBytes: 2896,
    },
  }),
];

const cron = [
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    createdAt: {seconds: 1616533099, nanos: 100},
    startedAt: {seconds: 1616533100, nanos: 100},
    finishedAt: {seconds: 1616533103, nanos: 100},
    job: {
      id: '290989c8a294ce1064041f0caa405c85',
      pipeline: {name: 'processor'},
    },
    input: {
      pfs: {repo: 'cron', name: 'cron', branch: 'master'},
    },
    outputCommit: {
      id: '290989c8a294ce1064041f0caa405c85',
      branch: {
        name: 'master',
        repo: {
          name: 'processor',
        },
      },
    },
    dataFailed: 0,
    dataTotal: 100,
  }),
  jobInfoFromObject({
    state: JobState.JOB_KILLED,
    createdAt: {seconds: 1616533099, nanos: 100},
    startedAt: {seconds: 1616533100, nanos: 100},
    finishedAt: {seconds: 1616533103, nanos: 100},
    job: {
      id: '249a1835a00b64422e30a0fdcb32deaf',
      pipeline: {name: 'processor'},
    },
    input: {
      pfs: {repo: 'cron', name: 'cron', branch: 'master'},
    },
    outputCommit: {
      id: '249a1835a00b64422e30a0fdcb32deaf',
      branch: {
        name: 'master',
        repo: {
          name: 'processor',
        },
      },
    },
    dataFailed: 0,
    dataTotal: 100,
  }),
  jobInfoFromObject({
    state: JobState.JOB_FAILURE,
    createdAt: {seconds: 1616533099, nanos: 100},
    startedAt: {seconds: 1616533100, nanos: 100},
    finishedAt: {seconds: 1616533103, nanos: 100},
    job: {
      id: 'abdf311864379b0cedd95932628935a0',
      pipeline: {name: 'processor'},
    },
    input: {
      pfs: {repo: 'cron', name: 'cron', branch: 'master'},
    },
    outputCommit: {
      id: 'abdf311864379b0cedd95932628935a0',
      branch: {
        name: 'master',
        repo: {
          name: 'processor',
        },
      },
    },
    dataFailed: 100,
    dataTotal: 100,
  }),
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    createdAt: {seconds: 1616533099, nanos: 100},
    startedAt: {seconds: 1616533100, nanos: 100},
    finishedAt: {seconds: 1616533103, nanos: 100},
    job: {
      id: 'a7811954e2828d76b4642ac214f2a0e6',
      pipeline: {name: 'processor'},
    },
    input: {
      pfs: {repo: 'cron', name: 'cron', branch: 'master'},
    },
    outputCommit: {
      id: 'a7811954e2828d76b4642ac214f2a0e6',
      branch: {
        name: 'master',
        repo: {
          name: 'processor',
        },
      },
    },
    dataFailed: 0,
    dataTotal: 100,
  }),
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    createdAt: {seconds: 1616533099, nanos: 100},
    startedAt: {seconds: 1616533100, nanos: 100},
    finishedAt: {seconds: 1616533103, nanos: 100},
    job: {
      id: '24fcfa133462bfcf3bbecfdc43614349',
      pipeline: {name: 'processor'},
    },
    input: {
      pfs: {repo: 'cron', name: 'cron', branch: 'master'},
    },
    outputCommit: {
      id: '24fcfa133462bfcf3bbecfdc43614349',
      branch: {
        name: 'master',
        repo: {
          name: 'processor',
        },
      },
    },
    dataFailed: 0,
    dataTotal: 100,
  }),
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    createdAt: {seconds: 1616533099, nanos: 100},
    startedAt: {seconds: 1616533100, nanos: 100},
    finishedAt: {seconds: 1616533103, nanos: 100},
    job: {
      id: '85c09e20958ac73f8005b37815f747a9',
      pipeline: {name: 'processor'},
    },
    input: {
      pfs: {repo: 'cron', name: 'cron', branch: 'master'},
    },
    outputCommit: {
      id: '85c09e20958ac73f8005b37815f747a9',
      branch: {
        name: 'master',
        repo: {
          name: 'processor',
        },
      },
    },
    dataFailed: 0,
    dataTotal: 100,
  }),
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    createdAt: {seconds: 1616533099, nanos: 100},
    startedAt: {seconds: 1616533100, nanos: 100},
    finishedAt: {seconds: 1616533103, nanos: 100},
    job: {
      id: '7be17147600af973b162ad795e09ac80',
      pipeline: {name: 'processor'},
    },
    input: {
      pfs: {repo: 'cron', name: 'cron', branch: 'master'},
    },
    outputCommit: {
      id: '7be17147600af973b162ad795e09ac80',
      branch: {
        name: 'master',
        repo: {
          name: 'processor',
        },
      },
    },
    dataFailed: 0,
    dataTotal: 100,
  }),
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    createdAt: {seconds: 1616533099, nanos: 100},
    startedAt: {seconds: 1616533100, nanos: 100},
    finishedAt: {seconds: 1616533103, nanos: 100},
    job: {
      id: '6dd9d64968e97d35821ce84fd03c8fef',
      pipeline: {name: 'processor'},
    },
    input: {
      pfs: {repo: 'cron', name: 'cron', branch: 'master'},
    },
    outputCommit: {
      id: '6dd9d64968e97d35821ce84fd03c8fef',
      branch: {
        name: 'master',
        repo: {
          name: 'processor',
        },
      },
    },
    dataFailed: 0,
    dataTotal: 100,
  }),
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    createdAt: {seconds: 1616533099, nanos: 100},
    startedAt: {seconds: 1616533100, nanos: 100},
    finishedAt: {seconds: 1616533103, nanos: 100},
    job: {
      id: '2ca0773cbc32b569b27450e4f13adf65',
      pipeline: {name: 'processor'},
    },
    input: {
      pfs: {repo: 'cron', name: 'cron', branch: 'master'},
    },
    outputCommit: {
      id: '2ca0773cbc32b569b27450e4f13adf65',
      branch: {
        name: 'master',
        repo: {
          name: 'processor',
        },
      },
    },
    dataFailed: 0,
    dataTotal: 100,
  }),
];

const customerTeam = [
  jobInfoFromObject({
    state: JobState.JOB_FAILURE,
    createdAt: {seconds: 1614136189, nanos: 100},
    startedAt: {seconds: 1614136190, nanos: 100},
    finishedAt: {seconds: 1614136193, nanos: 100},
    job: {
      id: '23b9af7d5d4343219bc8e02ff4acd33a',
      pipeline: {name: 'likelihoods'},
    },
    reason: 'inputs failed: images',
    dataFailed: 100,
    dataTotal: 100,
  }),
  jobInfoFromObject({
    state: JobState.JOB_EGRESSING,
    createdAt: {seconds: 1614136189, nanos: 100},
    startedAt: {seconds: 1614136191, nanos: 100},
    job: {id: '23b9af7d5d4343219bc8e02ff4acd33a', pipeline: {name: 'models'}},
    outputCommit: {
      id: '23b9af7d5d4343219bc8e02ff4acd33a',
      branch: {
        name: 'master',
        repo: {
          name: 'models',
        },
      },
    },
    dataFailed: 0,
    dataTotal: 100,
  }),
  jobInfoFromObject({
    state: JobState.JOB_KILLED,
    createdAt: {seconds: 1614136189, nanos: 100},
    startedAt: {seconds: 1614136192, nanos: 100},
    job: {
      id: '23b9af7d5d4343219bc8e02ff4acd33a',
      pipeline: {name: 'joint_call'},
    },
    reason:
      'datum 64b95f0fe1a787b6c26ec7ede800be6f2b97616f3224592d91cbfe1cfccd00a1 failed',
    dataFailed: 0,
    dataTotal: 0,
  }),
  jobInfoFromObject({
    state: JobState.JOB_RUNNING,
    createdAt: {seconds: 1614136189, nanos: 100},
    startedAt: {seconds: 1614136193, nanos: 100},
    job: {id: '23b9af7d5d4343219bc8e02ff4acd33a', pipeline: {name: 'split'}},
    dataFailed: 0,
    dataTotal: 100,
  }),
  jobInfoFromObject({
    state: JobState.JOB_STARTING,
    createdAt: {seconds: 1614136189, nanos: 100},
    startedAt: {seconds: 1614136194, nanos: 100},
    job: {id: '23b9af7d5d4343219bc8e02ff4acd33a', pipeline: {name: 'test'}},
    dataFailed: 0,
    dataTotal: 100,
  }),
];

const getLoadJobs = (jobCount: number) => {
  const jobStates = Object.values(JobState);
  const now = Math.floor(new Date().getTime() / 1000);
  return [...new Array(jobCount).keys()].map((jobIndex) => {
    return jobInfoFromObject({
      state: jobStates[
        Math.floor(Math.random() * jobStates.length)
      ] as JobState,
      createdAt: {seconds: now - jobIndex * 100, nanos: jobIndex * 100},
      startedAt: {seconds: now - jobIndex * 100, nanos: jobIndex * 100},
      job: {
        id: `0-${jobIndex}`,
        pipeline: {name: `load-pipeline-${jobIndex}`},
      },
      dataFailed: Math.floor(Math.random() * 100),
      dataTotal: Math.floor(Math.random() * 1000),
    });
  });
};

const jobs: {[projectId: string]: JobInfo[]} = {
  '1': tutorial,
  '2': customerTeam,
  '3': cron,
  '4': customerTeam,
  '5': [],
  '6': [],
  '7': [],
  '8': tutorial,
  '9': getLoadJobs(JOBS),
  default: [...tutorial, ...customerTeam],
};

export default jobs;
