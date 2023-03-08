import {DatumFilter} from '@graphqlTypes';
import {useCallback} from 'react';
import {useLocation} from 'react-router';

import {
  logsViewerJobRoute,
  logsViewerDatumRoute,
} from '@dash-frontend/views/Project/utils/routes';

import useUrlQueryState from './useUrlQueryState';

const useLogsNavigation = () => {
  const {getUpdatedSearchParams, searchParams} = useUrlQueryState();
  const {pathname} = useLocation();

  const getPathToJobLogs = useCallback(
    (args: Parameters<typeof logsViewerJobRoute>[0]) => {
      return `${logsViewerJobRoute(args, false)}?${getUpdatedSearchParams({
        prevPath: pathname,
      })}`;
    },
    [pathname, getUpdatedSearchParams],
  );

  const getPathToDatumLogs = useCallback(
    (
      args: Parameters<typeof logsViewerDatumRoute>[0],
      datumFilters: DatumFilter[],
    ) => {
      return `${logsViewerDatumRoute(args, false)}?${getUpdatedSearchParams({
        prevPath: pathname,
        datumFilters,
      })}`;
    },
    [pathname, getUpdatedSearchParams],
  );

  const getPathFromLogs = useCallback(
    (path: string) => {
      return `${searchParams.prevPath || path}?${getUpdatedSearchParams({
        prevPath: undefined,
      })}`;
    },
    [getUpdatedSearchParams, searchParams],
  );

  return {
    getPathToJobLogs,
    getPathToDatumLogs,
    getPathFromLogs,
  };
};

export default useLogsNavigation;
