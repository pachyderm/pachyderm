import {useCallback} from 'react';
import {useLocation} from 'react-router';

import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';

import useUrlQueryState from './useUrlQueryState';

const useFileBrowserNavigation = () => {
  const {getUpdatedSearchParams, searchParams} = useUrlQueryState();
  const {pathname, search} = useLocation();

  const getPathToFileBrowser = useCallback(
    (
      args: Parameters<typeof fileBrowserRoute>[0],
      clearSearchParams = false,
    ) => {
      return `${fileBrowserRoute(args, false)}?${getUpdatedSearchParams(
        {
          prevPath: `${pathname}${search}`,
        },
        clearSearchParams,
      )}`;
    },
    [getUpdatedSearchParams, pathname, search],
  );

  const getPathFromFileBrowser = useCallback(
    (backupPath: string) => {
      return searchParams.prevPath || backupPath;
    },
    [searchParams],
  );

  return {
    getPathToFileBrowser,
    getPathFromFileBrowser,
  };
};

export default useFileBrowserNavigation;
