import {JobInfo} from '@pachyderm/node-pachyderm';

import jobs from './jobs';
import {JOB_SETS} from './loadLimits';

const tutorial = {
  '23b9af7d5d4343219bc8e02ff44cd55a': [jobs['1'][0], jobs['1'][1]],
  '33b9af7d5d4343219bc8e02ff44cd55a': [jobs['1'][2]],
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
  '3': tutorial,
  '4': customerTeam,
  '5': {},
  '6': {},
  '9': getLoadJobSets(JOB_SETS),
  default: tutorial,
};

export default jobSets;
