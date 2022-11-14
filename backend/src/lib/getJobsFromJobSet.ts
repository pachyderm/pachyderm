import compact from 'lodash/compact';

import {JobInfo} from '@dash-backend/proto';

import {Context} from './types';

interface getJobsFromJobSetArgs {
  jobSet: JobInfo.AsObject[];
  pachClient: Context['pachClient'];
  projectId: string;
}

const getJobsFromJobSet = async ({
  jobSet,
  pachClient,
  projectId,
}: getJobsFromJobSetArgs) => {
  const jobPromises = jobSet.map(({job}) => {
    if (job?.id && job?.pipeline?.name) {
      return pachClient
        .pps()
        .inspectJob({
          id: job.id,
          projectId,
          pipelineName: job.pipeline.name,
        })
        .catch(() => null);
    }
    return Promise.resolve(null);
  });

  const jobs = await Promise.all(jobPromises);

  return compact(jobs);
};

export default getJobsFromJobSet;
