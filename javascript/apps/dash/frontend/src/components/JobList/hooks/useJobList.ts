import {useMemo} from 'react';

import {useJobs} from '@dash-frontend/hooks/useJobs';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {JobsQueryArgs, JobState} from '@graphqlTypes';

export type JobFilters = {
  [key in JobState]?: boolean;
};

const convertToObject = (JobStateList: JobState[]) => {
  return Object.values(JobStateList).reduce<JobFilters>((result, jobState) => {
    result[jobState] = true;
    return result;
  }, {});
};

const useJobList = ({projectId, pipelineId}: JobsQueryArgs) => {
  const {jobs, loading} = useJobs({projectId, pipelineId});
  const {viewState} = useUrlQueryState();

  const selectedFilters: JobFilters = useMemo(() => {
    if (viewState && viewState.jobFilters) {
      return convertToObject(viewState.jobFilters);
    }
    return convertToObject(Object.values(JobState));
  }, [viewState]);

  const filteredJobs = useMemo(() => {
    return jobs.filter((job) => {
      return selectedFilters[job.state];
    });
  }, [jobs, selectedFilters]);

  const noFiltersSelected = filteredJobs?.length === 0 && jobs?.length !== 0;

  return {
    jobs,
    filteredJobs,
    noFiltersSelected,
    selectedFilters,
    loading,
  };
};

export default useJobList;
