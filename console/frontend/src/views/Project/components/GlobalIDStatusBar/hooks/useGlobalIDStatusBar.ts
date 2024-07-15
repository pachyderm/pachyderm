import {useCallback} from 'react';

import {restJobStateToNodeState} from '@dash-frontend/api/utils/nodeStateMappers';
import {
  getAggregateJobState,
  jobInfosToJobSet,
} from '@dash-frontend/api/utils/transform';
import {useJobSet} from '@dash-frontend/hooks/useJobSet';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {calculateJobSetTotalRuntime} from '@dash-frontend/lib/dateTime';
import {NodeState} from '@dash-frontend/lib/types';

import {jobsRoute} from '../../../utils/routes';

export const useGlobalIDStatusBar = () => {
  const {projectId} = useUrlState();

  const {
    searchParams: {globalIdFilter},
    getNewSearchParamsAndGo,
    getUpdatedSearchParams,
  } = useUrlQueryState();
  const {jobSet} = useJobSet(globalIdFilter, Boolean(globalIdFilter));

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

  const showErrorBorder = jobSet
    ? restJobStateToNodeState(getAggregateJobState(jobSet)) === NodeState.ERROR
    : false;
  const internalJobSet = jobSet ? jobInfosToJobSet(jobSet) : undefined;
  const totalRuntime = calculateJobSetTotalRuntime(internalJobSet);
  const totalSubJobs = `${internalJobSet?.jobs.length} Subjob${
    (internalJobSet?.jobs.length || 0) > 1 ? 's' : ''
  } Total`;

  return {
    globalIdFilter,
    getPathToJobsTable,
    internalJobSet,
    totalRuntime,
    totalSubJobs,
    closeStatusBar,
    showErrorBorder,
  };
};
