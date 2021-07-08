import countBy from 'lodash/countBy';
import sum from 'lodash/sum';
import values from 'lodash/values';
import {useCallback, useMemo} from 'react';

import {useJobSets} from '@dash-frontend/hooks/useJobSets';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {jobsRoute} from '@dash-frontend/views/Project/utils/routes';
import {JobState} from '@graphqlTypes';

import {useSearch} from './useSearch';

export const useDefaultDropdown = () => {
  const {projectId} = useUrlState();
  const {jobSets} = useJobSets({projectId});
  const {setUrlFromViewState} = useUrlQueryState();
  const {closeDropdown, setSearchValue} = useSearch();

  const stateCounts = useMemo(
    () => countBy(jobSets, (job) => job.state),
    [jobSets],
  );
  const allJobs = useMemo(() => sum(values(stateCounts)), [stateCounts]);

  const handleJobChipClick = useCallback(
    (value?: JobState) => {
      value
        ? setUrlFromViewState(
            {jobFilters: [value]},
            jobsRoute({projectId}, false),
          )
        : setUrlFromViewState(
            {
              jobFilters: Object.values(JobState),
            },
            jobsRoute({projectId}, false),
          );

      closeDropdown();
    },
    [closeDropdown, projectId, setUrlFromViewState],
  );

  const handleHistoryChipClick = useCallback(
    (value?: string) => {
      if (value) {
        setSearchValue(value);
      }
    },
    [setSearchValue],
  );

  return {stateCounts, allJobs, handleHistoryChipClick, handleJobChipClick};
};
