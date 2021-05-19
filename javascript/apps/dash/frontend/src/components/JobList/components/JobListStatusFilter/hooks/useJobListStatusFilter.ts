import cloneDeep from 'lodash/cloneDeep';
import countBy from 'lodash/countBy';
import {useMemo} from 'react';
import {useHistory} from 'react-router';

import {JobFilters} from '@dash-frontend/components/JobList/hooks/useJobList';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {GetJobsQuery, JobState} from '@graphqlTypes';

const useJobListStatusFilter = (
  jobs: GetJobsQuery['jobs'],
  selectedFilters: JobFilters,
) => {
  const {setUrlFromViewState} = useUrlQueryState();
  const browserHistory = useHistory();

  const stateCounts = useMemo(() => countBy(jobs, (job) => job.state), [jobs]);

  const onChipClick = (job?: JobState) => {
    if (job) {
      const updatedFilterList = cloneDeep(selectedFilters);
      updatedFilterList[job] = !updatedFilterList[job];

      const updatedJobFilters = Object.entries(updatedFilterList).reduce<
        JobState[]
      >((result, [key, selected]) => {
        if (selected) {
          result.push(key as JobState);
        }
        return result;
      }, []);
      browserHistory.push(setUrlFromViewState({jobFilters: updatedJobFilters}));
    }
  };

  return {stateCounts, onChipClick};
};
export default useJobListStatusFilter;
