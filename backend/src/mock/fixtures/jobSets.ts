import {JobInfo} from '@dash-backend/proto';

import jobs from './jobs';
import {JOB_SETS} from './loadLimits';

const tutorial = {
  '23b9af7d5d4343219bc8e02ff44cd55a': [
    jobs['Solar-Panel-Data-Sorting'][0],
    jobs['Solar-Panel-Data-Sorting'][1],
  ],
  '33b9af7d5d4343219bc8e02ff44cd55a': [jobs['Solar-Panel-Data-Sorting'][2]],
  '7798fhje5d4343219bc8e02ff4acd33a': [jobs['Solar-Panel-Data-Sorting'][3]],
  o90du4js5d4343219bc8e02ff4acd33a: [jobs['Solar-Panel-Data-Sorting'][4]],
};

const cron = {
  '290989c8a294ce1064041f0caa405c85': [
    jobs['Solar-Power-Data-Logger-Team-Collab'][0],
  ],
  '249a1835a00b64422e30a0fdcb32deaf': [
    jobs['Solar-Power-Data-Logger-Team-Collab'][1],
  ],
  abdf311864379b0cedd95932628935a0: [
    jobs['Solar-Power-Data-Logger-Team-Collab'][2],
  ],
  a7811954e2828d76b4642ac214f2a0e6: [
    jobs['Solar-Power-Data-Logger-Team-Collab'][3],
  ],
  '24fcfa133462bfcf3bbecfdc43614349': [
    jobs['Solar-Power-Data-Logger-Team-Collab'][4],
  ],
  '85c09e20958ac73f8005b37815f747a9': [
    jobs['Solar-Power-Data-Logger-Team-Collab'][5],
  ],
  '7be17147600af973b162ad795e09ac80': [
    jobs['Solar-Power-Data-Logger-Team-Collab'][6],
  ],
  '6dd9d64968e97d35821ce84fd03c8fef': [
    jobs['Solar-Power-Data-Logger-Team-Collab'][7],
  ],
  '2ca0773cbc32b569b27450e4f13adf65': [
    jobs['Solar-Power-Data-Logger-Team-Collab'][8],
  ],
};

const customerTeam = {
  '23b9af7d5d4343219bc8e02ff4acd33a': jobs['Data-Cleaning-Process'],
};

const multiProjectPipelineA = {
  '23b9af7d5d4343219bc8e02ff44cd55a': jobs['Multi-Project-Pipeline-A'],
};

const getLoadJobSets = (count: number) => {
  return [...new Array(count).keys()].reduce(
    (jobSets: Record<string, JobInfo[]>, index) => {
      jobSets[`0-${index}`] = jobs['Load-Project'];
      return jobSets;
    },
    {},
  );
};

const jobSets: {[projectId: string]: {[id: string]: JobInfo[]}} = {
  'Solar-Panel-Data-Sorting': tutorial,
  'Data-Cleaning-Process': customerTeam,
  'Solar-Power-Data-Logger-Team-Collab': cron,
  'Solar-Price-Prediction-Modal': customerTeam,
  'Egress-Examples': {},
  'Empty-Project': {},
  'Load-Project': getLoadJobSets(JOB_SETS),
  'Multi-Project-Pipeline-A': multiProjectPipelineA,

  default: tutorial,
};

export default jobSets;
