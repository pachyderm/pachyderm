import {JobState, JobInfo} from '@dash-backend/proto';
import {jobInfoFromObject} from '@dash-backend/proto/builders/pps';

import {JOBS} from './loadLimits';

const solarPanelDataSorting = [
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    created: {seconds: 1616533099, nanos: 100},
    started: {seconds: 1616533100, nanos: 100},
    finished: {seconds: 1616533103, nanos: 100},
    job: {id: '23b9af7d5d4343219bc8e02ff44cd55a', pipeline: {name: 'montage'}},
    input: {
      crossList: [
        {
          pfs: {
            project: 'Solar-Panel-Data-Sorting',
            repo: 'edges',
            name: 'edges',
            branch: 'master',
          },
        },
        {
          pfs: {
            project: 'Solar-Panel-Data-Sorting',
            repo: 'images',
            name: 'images',
            branch: 'master',
          },
        },
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
    dataFailed: 1,
    dataProcessed: 2,
    dataSkipped: 1,
    dataTotal: 4,
  }),
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    created: {seconds: 1614126189, nanos: 100},
    started: {seconds: 1614126190, nanos: 100},
    finished: {seconds: 1614126193, nanos: 100},
    job: {id: '23b9af7d5d4343219bc8e02ff44cd55a', pipeline: {name: 'edges'}},
    input: {
      pfs: {
        project: 'Solar-Panel-Data-Sorting',
        repo: 'images',
        name: 'images',
        branch: 'master',
      },
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
    created: {seconds: 1614126190, nanos: 100},
    started: {seconds: 1614126191, nanos: 100},
    finished: {seconds: 1614126194, nanos: 100},
    job: {
      id: '33b9af7d5d4343219bc8e02ff44cd55a',
      pipeline: {name: 'montage', project: {name: 'Solar-Panel-Data-Sorting'}},
    },
    input: {
      crossList: [
        {
          pfs: {
            repo: 'edges',
            name: 'edges',
            branch: 'master',
            project: 'Solar-Panel-Data-Sorting',
          },
        },
        {
          pfs: {
            repo: 'images',
            name: 'images',
            branch: 'master',
            project: 'Solar-Panel-Data-Sorting',
          },
        },
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
        nanos: 589999999,
      },
      uploadTime: {
        seconds: 30,
        nanos: 231186700,
      },
      downloadBytes: 2896,
    },
  }),

  jobInfoFromObject({
    state: JobState.JOB_FINISHING,
    created: {seconds: 1614125000, nanos: 100},
    started: {seconds: 1614125000, nanos: 100},
    job: {
      id: '7798fhje5d4343219bc8e02ff4acd33a',
      pipeline: {name: 'montage', project: {name: 'Solar-Panel-Data-Sorting'}},
    },
  }),
  jobInfoFromObject({
    state: JobState.JOB_KILLED,
    created: {seconds: 1614123000, nanos: 100},
    started: {seconds: 1614123000, nanos: 100},
    job: {
      id: 'o90du4js5d4343219bc8e02ff4acd33a',
      pipeline: {name: 'montage', project: {name: 'Solar-Panel-Data-Sorting'}},
    },
  }),
];

const cron = [
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    created: {seconds: 1616533099, nanos: 100},
    started: {seconds: 1616533100, nanos: 100},
    finished: {seconds: 1616533103, nanos: 100},
    job: {
      id: '290989c8a294ce1064041f0caa405c85',
      pipeline: {
        name: 'processor',
        project: {name: 'Solar-Power-Data-Logger-Team-Collab'},
      },
    },
    input: {
      pfs: {
        project: 'Solar-Power-Data-Logger-Team-Collab',
        repo: 'cron',
        name: 'cron',
        branch: 'master',
      },
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
    created: {seconds: 1616533099, nanos: 100},
    started: {seconds: 1616533100, nanos: 100},
    finished: {seconds: 1616533103, nanos: 100},
    job: {
      id: '249a1835a00b64422e30a0fdcb32deaf',
      pipeline: {
        name: 'processor',
        project: {name: 'Solar-Power-Data-Logger-Team-Collab'},
      },
    },
    input: {
      pfs: {
        project: 'Solar-Power-Data-Logger-Team-Collab',
        repo: 'cron',
        name: 'cron',
        branch: 'master',
      },
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
    created: {seconds: 1616533099, nanos: 100},
    started: {seconds: 1616533100, nanos: 100},
    finished: {seconds: 1616533103, nanos: 100},
    job: {
      id: 'abdf311864379b0cedd95932628935a0',
      pipeline: {
        name: 'processor',
        project: {name: 'Solar-Power-Data-Logger-Team-Collab'},
      },
    },
    input: {
      pfs: {
        project: 'Solar-Power-Data-Logger-Team-Collab',
        repo: 'cron',
        name: 'cron',
        branch: 'master',
      },
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
    created: {seconds: 1616533099, nanos: 100},
    started: {seconds: 1616533100, nanos: 100},
    finished: {seconds: 1616533103, nanos: 100},
    job: {
      id: 'a7811954e2828d76b4642ac214f2a0e6',
      pipeline: {
        name: 'processor',
        project: {name: 'Solar-Power-Data-Logger-Team-Collab'},
      },
    },
    input: {
      pfs: {
        project: 'Solar-Power-Data-Logger-Team-Collab',
        repo: 'cron',
        name: 'cron',
        branch: 'master',
      },
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
    created: {seconds: 1616533099, nanos: 100},
    started: {seconds: 1616533100, nanos: 100},
    finished: {seconds: 1616533103, nanos: 100},
    job: {
      id: '24fcfa133462bfcf3bbecfdc43614349',
      pipeline: {
        name: 'processor',
        project: {name: 'Solar-Power-Data-Logger-Team-Collab'},
      },
    },
    input: {
      pfs: {
        project: 'Solar-Power-Data-Logger-Team-Collab',
        repo: 'cron',
        name: 'cron',
        branch: 'master',
      },
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
    created: {seconds: 1616533099, nanos: 100},
    started: {seconds: 1616533100, nanos: 100},
    finished: {seconds: 1616533103, nanos: 100},
    job: {
      id: '85c09e20958ac73f8005b37815f747a9',
      pipeline: {
        name: 'processor',
        project: {name: 'Solar-Power-Data-Logger-Team-Collab'},
      },
    },
    input: {
      pfs: {
        project: 'Solar-Power-Data-Logger-Team-Collab',
        repo: 'cron',
        name: 'cron',
        branch: 'master',
      },
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
    created: {seconds: 1616533099, nanos: 100},
    started: {seconds: 1616533100, nanos: 100},
    finished: {seconds: 1616533103, nanos: 100},
    job: {
      id: '7be17147600af973b162ad795e09ac80',
      pipeline: {
        name: 'processor',
        project: {name: 'Solar-Power-Data-Logger-Team-Collab'},
      },
    },
    input: {
      pfs: {
        project: 'Solar-Power-Data-Logger-Team-Collab',
        repo: 'cron',
        name: 'cron',
        branch: 'master',
      },
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
    created: {seconds: 1616533099, nanos: 100},
    started: {seconds: 1616533100, nanos: 100},
    finished: {seconds: 1616533103, nanos: 100},
    job: {
      id: '6dd9d64968e97d35821ce84fd03c8fef',
      pipeline: {
        name: 'processor',
        project: {name: 'Solar-Power-Data-Logger-Team-Collab'},
      },
    },
    input: {
      pfs: {
        project: 'Solar-Power-Data-Logger-Team-Collab',
        repo: 'cron',
        name: 'cron',
        branch: 'master',
      },
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
    created: {seconds: 1616533099, nanos: 100},
    started: {seconds: 1616533100, nanos: 100},
    finished: {seconds: 1616533103, nanos: 100},
    job: {
      id: '2ca0773cbc32b569b27450e4f13adf65',
      pipeline: {
        name: 'processor',
        project: {name: 'Solar-Power-Data-Logger-Team-Collab'},
      },
    },
    input: {
      pfs: {
        project: 'Solar-Power-Data-Logger-Team-Collab',
        repo: 'cron',
        name: 'cron',
        branch: 'master',
      },
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

const dataCleaningProcess = [
  jobInfoFromObject({
    state: JobState.JOB_FAILURE,
    created: {seconds: 1614136189, nanos: 100},
    started: {seconds: 1614136190, nanos: 100},
    finished: {seconds: 1614136193, nanos: 100},
    job: {
      id: '23b9af7d5d4343219bc8e02ff4acd33a',
      pipeline: {
        name: 'likelihoods',
        project: {name: 'Data-Cleaning-Process'},
      },
    },
    reason: 'inputs failed: images',
    dataFailed: 100,
    dataTotal: 100,
  }),
  jobInfoFromObject({
    state: JobState.JOB_EGRESSING,
    created: {seconds: 1614136189, nanos: 100},
    started: {seconds: 1614136191, nanos: 100},
    job: {
      id: '23b9af7d5d4343219bc8e02ff4acd33a',
      pipeline: {name: 'models', project: {name: 'Data-Cleaning-Process'}},
    },
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
    created: {seconds: 1614136189, nanos: 100},
    started: {seconds: 1614136192, nanos: 100},
    job: {
      id: '23b9af7d5d4343219bc8e02ff4acd33a',
      pipeline: {name: 'joint_call', project: {name: 'Data-Cleaning-Process'}},
    },
    reason:
      'datum 64b95f0fe1a787b6c26ec7ede800be6f2b97616f3224592d91cbfe1cfccd00a1 failed',
    dataFailed: 0,
    dataTotal: 0,
  }),
  jobInfoFromObject({
    state: JobState.JOB_RUNNING,
    created: {seconds: 1614136189, nanos: 100},
    started: {seconds: 1614136193, nanos: 100},
    job: {
      id: '23b9af7d5d4343219bc8e02ff4acd33a',
      pipeline: {name: 'split', project: {name: 'Data-Cleaning-Process'}},
    },
    dataFailed: 0,
    dataTotal: 100,
  }),
  jobInfoFromObject({
    state: JobState.JOB_STARTING,
    created: {seconds: 1614136189, nanos: 100},
    started: {seconds: 1614136194, nanos: 100},
    job: {
      id: '23b9af7d5d4343219bc8e02ff4acd33a',
      pipeline: {name: 'test', project: {name: 'Data-Cleaning-Process'}},
    },
    dataFailed: 0,
    dataTotal: 100,
  }),
];

const multiProjectPipelineA = [
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    created: {seconds: 1616533099, nanos: 100},
    started: {seconds: 1616533100, nanos: 100},
    finished: {seconds: 1616533103, nanos: 100},
    job: {
      id: '23b9af7d5d4343219bc8e02ff44cd55a',
      pipeline: {name: 'Node_2', project: {name: 'Multi-Project-Pipeline-A'}},
    },
    input: {
      pfs: {
        project: 'Multi-Project-Pipeline-B',
        repo: 'Node_1',
        name: 'Node_1',
        branch: 'master',
      },
    },
    outputCommit: {
      id: '23b9af7d5d4343219bc8e02ff44cd55a',
      branch: {
        name: 'master',
        repo: {
          name: 'Node_2',
        },
      },
    },
    dataFailed: 1,
    dataProcessed: 2,
    dataSkipped: 1,
    dataTotal: 4,
  }),
];

const multiProjectPipelineB = [
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    created: {seconds: 1616533099, nanos: 100},
    started: {seconds: 1616533100, nanos: 100},
    finished: {seconds: 1616533103, nanos: 100},
    job: {id: '23b9af7d5d4343219bc8e02ff44cd55a', pipeline: {name: 'Node_2'}},
    input: {
      pfs: {
        project: 'Multi-Project-Pipeline-A',
        repo: 'Node_1',
        name: 'Node_1',
        branch: 'master',
      },
    },
    outputCommit: {
      id: '23b9af7d5d4343219bc8e02ff44cd55a',
      branch: {
        name: 'master',
        repo: {
          name: 'Node_2',
        },
      },
    },
    dataFailed: 1,
    dataProcessed: 2,
    dataSkipped: 1,
    dataTotal: 4,
  }),
];

const pipelinesProject = [
  jobInfoFromObject({
    state: JobState.JOB_RUNNING,
    created: {seconds: 1616533098, nanos: 100},
    started: {seconds: 1616533098, nanos: 100},
    job: {
      id: '5940382d5d4343219bc8e02ff44cd55a',
      pipeline: {name: 'service-pipeline'},
    },
    input: {
      pfs: {
        project: 'Pipelines-Project',
        repo: 'service-pipeline-input',
        name: 'service-pipeline-input',
        branch: 'master',
      },
    },
  }),
  jobInfoFromObject({
    state: JobState.JOB_SUCCESS,
    created: {seconds: 1616533099, nanos: 100},
    started: {seconds: 1616533100, nanos: 100},
    finished: {seconds: 1616533103, nanos: 100},
    job: {
      id: '23b9af7d5d4343219bc8e02ff44cd55a',
      pipeline: {name: 'service-pipeline'},
    },
    input: {
      pfs: {
        project: 'Pipelines-Project',
        repo: 'service-pipeline-input',
        name: 'service-pipeline-input',
        branch: 'master',
      },
    },
    outputCommit: {
      id: '23b9af7d5d4343219bc8e02ff44cd55a',
      branch: {
        name: 'master',
        repo: {
          name: 'service-pipeline',
        },
      },
    },
    dataFailed: 1,
    dataProcessed: 2,
    dataSkipped: 1,
    dataTotal: 4,
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
      created: {seconds: now - jobIndex * 100, nanos: jobIndex * 100},
      started: {seconds: now - jobIndex * 100, nanos: jobIndex * 100},
      job: {
        id: `0-${jobIndex}`,
        pipeline: {
          name: `load-pipeline-${jobIndex}`,
          project: {name: 'Load-Project'},
        },
      },
      dataFailed: Math.floor(Math.random() * 100),
      dataTotal: Math.floor(Math.random() * 1000),
    });
  });
};

const jobs: {[projectId: string]: JobInfo[]} = {
  'Solar-Panel-Data-Sorting': solarPanelDataSorting,
  'Data-Cleaning-Process': dataCleaningProcess,
  'Solar-Power-Data-Logger-Team-Collab': cron,
  'Solar-Price-Prediction-Modal': dataCleaningProcess,
  'Egress-Examples': [],
  'Empty-Project': [],
  'Trait-Discovery': [],
  'OpenCV-Tutorial': solarPanelDataSorting,
  'Load-Project': getLoadJobs(JOBS),
  default: [...solarPanelDataSorting, ...dataCleaningProcess],
  'Multi-Project-Pipeline-A': multiProjectPipelineA,
  'Multi-Project-Pipeline-B': multiProjectPipelineB,
  'Pipelines-Project': pipelinesProject,
};

export default jobs;
