import {
  JobOverviewFragment,
  JobSetFieldsFragment,
  JobState,
} from '@graphqlTypes';
import cloneDeep from 'lodash/cloneDeep';
import countBy from 'lodash/countBy';
import {useCallback, useMemo} from 'react';

import {JobFilters} from '@dash-frontend/components/JobList/hooks/useJobFilters';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';

const useJobListStatusFilter = (
  jobs: (JobOverviewFragment | JobSetFieldsFragment)[],
  selectedFilters: JobFilters,
) => {
  const {updateViewState} = useUrlQueryState();

  const stateCounts = useMemo(() => countBy(jobs, (job) => job.state), [jobs]);

  const onChipClick = useCallback(
    (job?: JobState) => {
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
        updateViewState({jobFilters: updatedJobFilters});
      }
    },
    [updateViewState, selectedFilters],
  );

  return {stateCounts, onChipClick};
};
export default useJobListStatusFilter;
