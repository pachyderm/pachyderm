import {DatumFilter, NodeState, OriginKind} from '@graphqlTypes';
import isEmpty from 'lodash/isEmpty';
import isEqual from 'lodash/isEqual';
import omit from 'lodash/omit';
import {useCallback, useMemo} from 'react';
import {useHistory, useLocation} from 'react-router';

import {DagDirection} from '@dash-frontend/lib/types';

interface selectableResource {
  selectedPipelines?: string[];
  selectedRepos?: string[];
  selectedJobs?: string[];
}
export interface UrlState extends selectableResource {
  dagDirection?: DagDirection;
  sidebarWidth?: number;
  skipCenterOnSelect?: boolean;
  prevPath?: string;
  tutorialId?: string;
  sortBy?: string;
  globalIdFilter?: string;
  datumFilters?: DatumFilter[];
  jobStatus?: NodeState[];
  jobId?: string[];
  pipelineStep?: string[];
  pipelineState?: NodeState[];
  commitType?: OriginKind[];
  branchName?: string[];
  commitId?: string[];
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

  const getNewViewState = useCallback(
    (newState: Partial<UrlState>, path?: string) => {
      const searchParams = new URLSearchParams();
      const pathname = window.location.pathname;
      const updatedPath = path || pathname;

      searchParams.set('view', btoa(JSON.stringify(newState)));

      return browserHistory.push(`${updatedPath}?${searchParams}`);
    },
    [browserHistory],
  );

  const updateViewState = useCallback(
    (newState: Partial<UrlState>, path?: string) => {
      const searchParams = new URLSearchParams(window.location.search);
      const decodedViewState = getViewStateFromSearchParams(searchParams);

      const updatedState = Object.entries(newState).reduce(
        (accViewState, [key, newValue]) => {
          if (
            newValue === '' ||
            newValue === undefined ||
            newValue === null ||
            isEmpty(newValue)
          ) {
            return omit(accViewState, key);
          }
          return {...accViewState, [key]: newValue};
        },
        decodedViewState,
      );
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

  const toggleSelection = (param: keyof selectableResource, value: string) => {
    if (!viewState[param]) {
      return updateViewState({
        [param]: [value],
      });
    }
    if (viewState[param] && Array.isArray(viewState[param])) {
      return viewState[param]?.includes(value)
        ? updateViewState({
            [param]: viewState[param]?.filter((v) => v !== value),
          })
        : updateViewState({[param]: [...(viewState[param] || []), value]});
    }
  };

  const clearViewState = useCallback(() => {
    return browserHistory.push(window.location.pathname);
  }, [browserHistory]);

  return {
    viewState,
    getNewViewState,
    updateViewState,
    clearViewState,
    toggleSelection,
    getUpdatedSearchParams,
  };
};

export default useUrlQueryState;
