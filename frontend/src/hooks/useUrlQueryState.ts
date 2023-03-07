import {DatumFilter, NodeState, OriginKind} from '@graphqlTypes';
import isEmpty from 'lodash/isEmpty';
import {useCallback, useMemo} from 'react';
import {useHistory, useLocation} from 'react-router';

import {DagDirection} from '@dash-frontend/lib/types';

interface SelectableResource {
  selectedPipelines?: string[] | string;
  selectedRepos?: string[] | string;
  selectedJobs?: string[] | string;
}

interface ListParams extends SelectableResource {
  datumFilters?: DatumFilter[];
  jobStatus?: NodeState[];
  jobId?: string[];
  pipelineStep?: string[];
  pipelineState?: NodeState[];
  commitType?: OriginKind[];
  branchName?: string[];
  commitId?: string[];
}
export interface UrlState extends ListParams {
  dagDirection?: DagDirection;
  sidebarWidth?: number;
  skipCenterOnSelect?: boolean;
  prevPath?: string;
  tutorialId?: string;
  sortBy?: string;
  globalIdFilter?: string;
}
export interface ViewStateLists {
  selectedPipelines?: string[];
  selectedRepos?: string[];
  selectedJobs?: string[];
  datumFilters?: string[];
  jobStatus?: string[];
  jobId?: string[];
  pipelineStep?: string[];
  pipelineState?: string[];
  commitType?: string[];
  branchName?: string[];
  commitId?: string[];
}
export interface ViewStateValues {
  dagDirection?: string;
  sidebarWidth?: string;
  skipCenterOnSelect?: string;
  prevPath?: string;
  tutorialId?: string;
  sortBy?: string;
  globalIdFilter?: string;
}
export interface ViewState extends ViewStateValues, ViewStateLists {}

const LIST_TYPES: (keyof ListParams)[] = [
  'selectedPipelines',
  'selectedRepos',
  'selectedJobs',
  'datumFilters',
  'jobStatus',
  'jobId',
  'pipelineStep',
  'pipelineState',
  'commitType',
  'branchName',
  'commitId',
];

const useUrlQueryState = () => {
  const {search} = useLocation();
  const browserHistory = useHistory();

  const viewState = useMemo(() => {
    const searchParams = new URLSearchParams(search);
    const viewState: ViewState = {};

    for (const [key, value] of searchParams.entries()) {
      if (LIST_TYPES.includes(key as keyof ListParams)) {
        // use ListParams types as string[]
        viewState[key as keyof ViewStateLists] = value.split(',');
      } else {
        // use all other values as string
        viewState[key as keyof ViewStateValues] = value;
      }
    }
    return viewState;
  }, [search]);

  const getUpdatedSearchParams = useCallback(
    (newState: UrlState, clearParams?: boolean) => {
      const searchParams = new URLSearchParams(
        clearParams ? undefined : window.location.search,
      );

      Object.entries(newState).forEach(([key, value]) => {
        // delete any search params that come back as falsy or empty values
        // to trim the search string. E.g. a filter becoming an empty array
        if (
          value === '' ||
          value === undefined ||
          value === null ||
          ((typeof value === 'object' || Array.isArray(value)) &&
            isEmpty(value))
        ) {
          searchParams.delete(key);
        } else {
          searchParams.set(key, value);
        }
      });
      return searchParams;
    },
    [],
  );

  const getNewViewState = useCallback(
    (newState: Partial<UrlState>, path?: string) => {
      const pathname = window.location.pathname;
      const updatedPath = path || pathname;

      const searchParams = getUpdatedSearchParams(newState, true);

      return browserHistory.push(`${updatedPath}?${searchParams}`);
    },
    [browserHistory, getUpdatedSearchParams],
  );

  const updateViewState = useCallback(
    (newState: Partial<UrlState>, path?: string) => {
      const searchParams = getUpdatedSearchParams(newState);

      const pathname = window.location.pathname;
      const updatedPath = path || pathname;

      // De-duplicating unecessary updates to the history API.
      if (
        searchParams.toString() !== window.location.search ||
        pathname !== updatedPath
      ) {
        return browserHistory.push(`${updatedPath}?${searchParams}`);
      }
    },
    [browserHistory, getUpdatedSearchParams],
  );

  // params from SelectableResource are used to select subsets of a resource
  // like selectedPipelines=A,B. This function either adds a new selection
  // to the array or removes an existing one. (used in Table Views)
  const toggleSelection = useCallback(
    (param: keyof SelectableResource, value: string) => {
      const selectableResourceArray = viewState[param];
      if (!selectableResourceArray) {
        return updateViewState({
          [param]: [value],
        });
      }
      if (selectableResourceArray && Array.isArray(selectableResourceArray)) {
        if (selectableResourceArray.includes(value)) {
          const filteredSelections = selectableResourceArray.filter(
            (v: string) => v !== value,
          );
          updateViewState({
            [param]: filteredSelections,
          });
        } else {
          updateViewState({
            [param]: [...selectableResourceArray, value],
          });
        }
      }
    },
    [updateViewState, viewState],
  );

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
