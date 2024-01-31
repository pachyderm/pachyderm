import {useCallback, useMemo} from 'react';

import {restJobStateToNodeState} from '@dash-frontend/api/utils/nodeStateMappers';
import {jobInfosToJobSet} from '@dash-frontend/api/utils/transform';
import {useJobSet} from '@dash-frontend/hooks/useJobSet';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {calculateJobTotalRuntime} from '@dash-frontend/lib/dateTime';
import {NodeState} from '@dash-frontend/lib/types';

import {jobsRoute} from '../../../utils/routes';

export const useGlobalIDStatusBar = () => {
  const {projectId} = useUrlState();

  const {
    searchParams: {globalIdFilter},
    getNewSearchParamsAndGo,
    getUpdatedSearchParams,
  } = useUrlQueryState();
  const {jobSet: jobSets} = useJobSet(globalIdFilter, Boolean(globalIdFilter));

  const jobSet = useMemo(() => {
    if (jobSets?.length && jobSets?.[0]?.job?.id === globalIdFilter) {
      return jobSets[0];
    }
    return undefined;
  }, [globalIdFilter, jobSets]);

  const closeStatusBar = useCallback(() => {
    getNewSearchParamsAndGo({
      globalIdFilter: undefined,
    });
  }, [getNewSearchParamsAndGo]);

  const getPathToJobsTable = useCallback(
    (tabId: 'runtimes' | 'subjobs') => {
      return `${jobsRoute(
        {
          projectId: projectId || '',
          tabId: tabId,
        },
        false,
      )}?${getUpdatedSearchParams(
        {
          selectedJobs: globalIdFilter,
        },
        true,
      )}`;
    },
    [getUpdatedSearchParams, globalIdFilter, projectId],
  );

  const showErrorBorder =
    restJobStateToNodeState(jobSet?.state) === NodeState.ERROR;
  const totalRuntime = calculateJobTotalRuntime(jobSet, 'Still Processing');
  const internalJobSets = jobSets ? jobInfosToJobSet(jobSets) : undefined;

  const totalSubJobs = `${internalJobSets?.jobs.length} Subjob${
    (internalJobSets?.jobs.length || 0) > 1 ? 's' : ''
  } Total`;

  return {
    globalIdFilter,
    getPathToJobsTable,
    internalJobSets,
    totalRuntime,
    totalSubJobs,
    closeStatusBar,
    showErrorBorder,
  };
};
