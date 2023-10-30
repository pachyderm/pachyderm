import {DatumFilter} from '@graphqlTypes';
import {useCallback} from 'react';
import {useLocation} from 'react-router';

import {
  logsViewerJobRoute,
  logsViewerDatumRoute,
  logsViewerLatestRoute,
} from '@dash-frontend/views/Project/utils/routes';

import useUrlQueryState from './useUrlQueryState';

const useLogsNavigation = () => {
  const {getUpdatedSearchParams, searchParams} = useUrlQueryState();
  const {pathname, search} = useLocation();

  const getPathToJobLogs = useCallback(
    (
      args: Parameters<typeof logsViewerJobRoute>[0],
      clearSearchParams = false,
    ) => {
      return `${logsViewerJobRoute(args, false)}?${getUpdatedSearchParams(
        {
          prevPath: `${pathname}${search}`,
        },
        clearSearchParams,
      )}`;
    },
    [getUpdatedSearchParams, pathname, search],
  );

  const getPathToDatumLogs = useCallback(
    (
      args: Parameters<typeof logsViewerDatumRoute>[0],
      datumFilters: DatumFilter[],
      clearSearchParams = false,
    ) => {
      return `${logsViewerDatumRoute(args, false)}?${getUpdatedSearchParams(
        {
          prevPath: `${pathname}${search}`,
          datumFilters,
        },
        clearSearchParams,
      )}`;
    },
    [getUpdatedSearchParams, pathname, search],
  );

  const getPathToLatestJobLogs = useCallback(
    (
      args: Parameters<typeof logsViewerLatestRoute>[0],
      clearSearchParams = false,
    ) => {
      return `${logsViewerLatestRoute(args, false)}?${getUpdatedSearchParams(
        {
          prevPath: `${pathname}${search}`,
        },
        clearSearchParams,
      )}`;
    },
    [getUpdatedSearchParams, pathname, search],
  );

  const getPathFromLogs = useCallback(
    (backupPath: string) => {
      return searchParams.prevPath || backupPath;
    },
    [searchParams],
  );

  return {
    getPathToJobLogs,
    getPathToDatumLogs,
    getPathToLatestJobLogs,
    getPathFromLogs,
  };
};

export default useLogsNavigation;
