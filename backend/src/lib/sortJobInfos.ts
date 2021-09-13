import {JobInfo} from '@pachyderm/node-pachyderm';
import sortBy from 'lodash/sortBy';

import timestampToNanos from '@dash-backend/lib/timestampToNanos';

const sortJobInfos = (jobInfos: JobInfo.AsObject[]) => {
  return sortBy(jobInfos, [
    (jobInfo) => (jobInfo.started ? timestampToNanos(jobInfo.started) : 0),
    (jobInfo) => jobInfo.job?.pipeline?.name || '',
  ]);
};

export default sortJobInfos;
