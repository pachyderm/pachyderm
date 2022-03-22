import {JobState} from '@graphqlTypes';
import isEqual from 'lodash/isEqual';
import {useCallback, useMemo} from 'react';
import {useHistory, useLocation} from 'react-router';

import {DagDirection} from '@dash-frontend/lib/types';

export interface UrlState {
  dagDirection?: DagDirection;
  sidebarWidth?: number;
  skipCenterOnSelect?: boolean;
  jobFilters?: JobState[];
  prevFileBrowserPath?: string;
  tutorialId?: string;
  globalIdFilter?: string;
}

const getViewStateFromSearchParams = (searchParams: URLSearchParams) => {
  const encodedViewState = searchParams.get('view');

  let decodedViewState: UrlState = {};

  if (encodedViewState) {
    try {
      decodedViewState = JSON.parse(atob(encodedViewState));
    } catch (e) {
      decodedViewState = {};
    }
  }

  return decodedViewState;
};

const useUrlQueryState = () => {
  const {search} = useLocation();
  const browserHistory = useHistory();

  const searchParams = useMemo(() => {
    return new URLSearchParams(search);
  }, [search]);

  const viewState = useMemo(() => {
    return getViewStateFromSearchParams(searchParams);
  }, [searchParams]);

  const getUpdatedSearchParams = useCallback(
    (newState: Partial<UrlState>) => {
      const newSearchParams = new URLSearchParams(search);

      const updatedState: UrlState = {
        ...viewState,
        ...newState,
      };

      newSearchParams.set('view', btoa(JSON.stringify(updatedState)));

      return newSearchParams;
    },
    [search, viewState],
  );

  const updateViewState = useCallback(
    (newState: Partial<UrlState>, path?: string) => {
      const searchParams = new URLSearchParams(window.location.search);
      const decodedViewState = getViewStateFromSearchParams(searchParams);

      const updatedState = {...decodedViewState, ...newState};
      const pathname = window.location.pathname;
      const updatedPath = path || pathname;

      // De-duplicating unecessary updates to the history API.
      if (
        !isEqual(decodedViewState, updatedState) ||
        pathname !== updatedPath
      ) {
        searchParams.set('view', btoa(JSON.stringify(updatedState)));

        return browserHistory.push(`${updatedPath}?${searchParams}`);
      }
    },
    [browserHistory],
  );

  return {
    viewState,
    updateViewState,
    getUpdatedSearchParams,
  };
};

export default useUrlQueryState;
