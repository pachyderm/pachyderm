import {JobInfo} from '@pachyderm/node-pachyderm';

import jobs from './jobs';
import {JOB_SETS} from './loadLimits';

const tutorial = {
  '23b9af7d5d4343219bc8e02ff44cd55a': [jobs['1'][0], jobs['1'][1]],
  '33b9af7d5d4343219bc8e02ff44cd55a': [jobs['1'][2]],
};

const cron = {
  '290989c8a294ce1064041f0caa405c85': [jobs['3'][0]],
  '249a1835a00b64422e30a0fdcb32deaf': [jobs['3'][1]],
  abdf311864379b0cedd95932628935a0: [jobs['3'][2]],
  a7811954e2828d76b4642ac214f2a0e6: [jobs['3'][3]],
  '24fcfa133462bfcf3bbecfdc43614349': [jobs['3'][4]],
  '85c09e20958ac73f8005b37815f747a9': [jobs['3'][5]],
  '7be17147600af973b162ad795e09ac80': [jobs['3'][6]],
  '6dd9d64968e97d35821ce84fd03c8fef': [jobs['3'][7]],
  '2ca0773cbc32b569b27450e4f13adf65': [jobs['3'][8]],
};

const customerTeam = {
  '23b9af7d5d4343219bc8e02ff4acd33a': jobs['2'],
};

const getLoadJobSets = (count: number) => {
  return [...new Array(count).keys()].reduce(
    (jobSets: Record<string, JobInfo[]>, index) => {
      jobSets[`0-${index}`] = jobs['9'];
      return jobSets;
    },
    {},
  );
};

const jobSets: {[projectId: string]: {[id: string]: JobInfo[]}} = {
  '1': tutorial,
  '2': customerTeam,
  '3': cron,
  '4': customerTeam,
  '5': {},
  '6': {},
  '9': getLoadJobSets(JOB_SETS),
  default: tutorial,
};

export default jobSets;
