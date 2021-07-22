import {useMemo} from 'react';

import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {
  JobOverviewFragment,
  JobSetFieldsFragment,
  JobState,
} from '@graphqlTypes';

export type JobFilters = {
  [key in JobState]?: boolean;
};

const convertToObject = (JobStateList: JobState[]) => {
  return Object.values(JobStateList).reduce<JobFilters>((result, jobState) => {
    result[jobState] = true;
    return result;
  }, {});
};

interface UseJobFiltersArgs {
  jobs: (JobOverviewFragment | JobSetFieldsFragment)[];
}

const useJobFilters = ({jobs}: UseJobFiltersArgs) => {
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
    filteredJobs,
    noFiltersSelected,
    selectedFilters,
  };
};

export default useJobFilters;
