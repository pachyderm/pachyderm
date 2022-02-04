import {useCallback} from 'react';
import {useLocation} from 'react-router';

import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';

import useUrlQueryState from './useUrlQueryState';

const useFileBrowserNavigation = () => {
  const {getUpdatedSearchParams, viewState} = useUrlQueryState();
  const {pathname} = useLocation();

  const getPathToFileBrowser = useCallback(
    (args: Parameters<typeof fileBrowserRoute>[0]) => {
      return `${fileBrowserRoute(args, false)}?${getUpdatedSearchParams({
        prevFileBrowserPath: pathname,
      })}`;
    },
    [pathname, getUpdatedSearchParams],
  );

  const getPathFromFileBrowser = useCallback(
    (path: string) => {
      return `${viewState.prevFileBrowserPath || path}?${getUpdatedSearchParams(
        {
          prevFileBrowserPath: undefined,
        },
      )}`;
    },
    [getUpdatedSearchParams, viewState],
  );

  return {
    getPathToFileBrowser,
    getPathFromFileBrowser,
  };
};

export default useFileBrowserNavigation;
