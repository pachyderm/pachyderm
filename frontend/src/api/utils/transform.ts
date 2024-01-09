import sortBy from 'lodash/sortBy';

import {JobInfo, JobState} from '@dash-frontend/generated/proto/pps/pps.pb';
import {
  getISOStringFromUnix,
  getUnixSecondsFromISOString,
} from '@dash-frontend/lib/dateTime';
import {InternalJobSet} from '@dash-frontend/lib/types';

// TODO: We should ultimately handle most everything in this file on the Pachyderm side.

export const getAggregateJobState = (jobs: JobInfo[]) => {
  for (let i = 0; i < jobs.length; i++) {
    if (jobs[i].state !== JobState.JOB_SUCCESS) {
      return jobs[i].state;
    }
  }

  return JobState.JOB_SUCCESS;
};

export const sortJobInfos = (jobInfos: JobInfo[]) => {
  return sortBy(jobInfos, [
    (jobInfo) =>
      jobInfo.started ? getUnixSecondsFromISOString(jobInfo.started) : 0,
    (jobInfo) => jobInfo.job?.pipeline?.name || '',
  ]);
};

export const jobInfosToJobSet = (jobInfos: JobInfo[]): InternalJobSet => {
  const jobs = sortJobInfos(jobInfos);
  const job = jobs[0];
  const created =
    jobs.length > 0 ? jobs.find((job) => job.created)?.created : undefined;
  const started =
    jobs.length > 0 ? jobs.find((job) => job.started)?.started : undefined;
  const finishTimes = jobs
    .map((job) => getUnixSecondsFromISOString(job.finished))
    .filter((finished) => !isNaN(finished));
  const finished =
    finishTimes.length > 0
      ? getISOStringFromUnix(Math.max(...finishTimes))
      : undefined;
  const inProgress = Boolean(jobs.find((job) => !job.finished));

  return {
    job: job.job,
    created,
    started,
    finished,
    inProgress,
    state: getAggregateJobState(jobs),
    jobs,
  };
};

export const jobsMapToJobSet = (jobsMap: Map<string, JobInfo[]>) => {
  const data = [];

  for (const [_jobId, jobs] of jobsMap) {
    data.push(jobInfosToJobSet(jobs));
  }

  return data;
};
