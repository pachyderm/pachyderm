import {useCallback} from 'react';
import {useLocation} from 'react-router';

import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';

import useUrlQueryState from './useUrlQueryState';

const useFileBrowserNavigation = () => {
  const {getUpdatedSearchParams, searchParams} = useUrlQueryState();
  const {pathname} = useLocation();

  const getPathToFileBrowser = useCallback(
    (args: Parameters<typeof fileBrowserRoute>[0]) => {
      return `${fileBrowserRoute(args, false)}?${getUpdatedSearchParams(
        {
          prevPath: pathname,
        },
        true,
      )}`;
    },
    [pathname, getUpdatedSearchParams],
  );

  const getPathFromFileBrowser = useCallback(
    (backupPath: string) => {
      return `${searchParams.prevPath || backupPath}?${getUpdatedSearchParams({
        prevPath: undefined,
      })}`;
    },
    [getUpdatedSearchParams, searchParams],
  );

  return {
    getPathToFileBrowser,
    getPathFromFileBrowser,
  };
};

export default useFileBrowserNavigation;
