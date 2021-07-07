import {JobOverviewFragment, JobSetFieldsFragment} from '@graphqlTypes';

function isPipelineJob(
  job: JobOverviewFragment | JobSetFieldsFragment,
): job is JobOverviewFragment {
  return job.__typename === 'Job';
}

export default isPipelineJob;
