import isEmpty from 'lodash/isEmpty';
import {useCallback, useMemo} from 'react';
import {useHistory, useLocation} from 'react-router';

import {OriginKind} from '@dash-frontend/api/pfs';
import {DatumState} from '@dash-frontend/api/pps';
import {DatumFilter, NodeState} from '@dash-frontend/lib/types';

interface SelectableResource {
  selectedPipelines?: string[] | string;
  selectedRepos?: string[] | string;
  selectedJobs?: string[] | string;
}

interface ListParams extends SelectableResource {
  datumFilters?:
    | DatumFilter[]
    | Array<
        | DatumState.FAILED
        | DatumState.RECOVERED
        | DatumState.SKIPPED
        | DatumState.SUCCESS
      >;
  jobStatus?: NodeState[];
  jobId?: string[];
  pipelineStep?: string[];
  pipelineVersion?: string[];
  pipelineState?: NodeState[];
  commitType?: OriginKind[];
  branchName?: string[];
  commitId?: string[];
}
export interface UrlState extends ListParams {
  prevPath?: string;
  sortBy?: string;
  globalIdFilter?: string;
  branchId?: string;
}
export interface ViewStateLists {
  selectedPipelines?: string[];
  selectedRepos?: string[];
  selectedJobs?: string[];
  datumFilters?: string[];
  jobStatus?: string[];
  jobId?: string[];
  pipelineStep?: string[];
  pipelineVersion?: string[];
  pipelineState?: string[];
  commitType?: string[];
  branchName?: string[];
  commitId?: string[];
}
export interface ViewStateValues {
  prevPath?: string;
  sortBy?: string;
  globalIdFilter?: string;
  branchId?: string;
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
  'pipelineVersion',
  'commitType',
  'branchName',
  'commitId',
];

const useUrlQueryState = () => {
  const {search} = useLocation();
  const browserHistory = useHistory();

  const searchParams = useMemo(() => {
    const searchParamsIn = new URLSearchParams(search);
    const viewStateObject: ViewState = {};

    for (const [key, value] of searchParamsIn.entries()) {
      if (LIST_TYPES.includes(key as keyof ListParams)) {
        // use ListParams types as string[]
        viewStateObject[key as keyof ViewStateLists] = value.split(',');
      } else {
        // use all other values as string
        viewStateObject[key as keyof ViewStateValues] = value;
      }
    }
    return viewStateObject;
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

  const getNewSearchParamsAndGo = useCallback(
    (newState: Partial<UrlState>, path?: string) => {
      const pathname = window.location.pathname;
      const updatedPath = path || pathname;

      const searchParams = getUpdatedSearchParams(newState, true);

      return browserHistory.replace(`${updatedPath}?${searchParams}`);
    },
    [browserHistory, getUpdatedSearchParams],
  );

  const updateSearchParamsAndGo = useCallback(
    (newState: Partial<UrlState>, path?: string) => {
      const searchParams = getUpdatedSearchParams(newState);

      const pathname = window.location.pathname;
      const updatedPath = path || pathname;

      // De-duplicating unecessary updates to the history API.
      if (
        searchParams.toString() !== window.location.search ||
        pathname !== updatedPath
      ) {
        return browserHistory.replace(`${updatedPath}?${searchParams}`);
      }
    },
    [browserHistory, getUpdatedSearchParams],
  );

  // params from SelectableResource are used to select subsets of a resource
  // like selectedPipelines=A,B. This function either adds a new selection
  // to the array or removes an existing one. (used in Table Views)
  const toggleSearchParamsListEntry = useCallback(
    (param: keyof SelectableResource, value: string) => {
      const selectableResourceArray = searchParams[param];
      if (!selectableResourceArray) {
        return updateSearchParamsAndGo({
          [param]: [value],
        });
      }
      if (selectableResourceArray && Array.isArray(selectableResourceArray)) {
        if (selectableResourceArray.includes(value)) {
          const filteredSelections = selectableResourceArray.filter(
            (v: string) => v !== value,
          );
          updateSearchParamsAndGo({
            [param]: filteredSelections,
          });
        } else {
          updateSearchParamsAndGo({
            [param]: [...selectableResourceArray, value],
          });
        }
      }
    },
    [updateSearchParamsAndGo, searchParams],
  );

  const clearSearchParamsAndGo = useCallback(() => {
    return browserHistory.replace(window.location.pathname);
  }, [browserHistory]);

  return {
    searchParams,
    getUpdatedSearchParams,
    getNewSearchParamsAndGo,
    updateSearchParamsAndGo,
    toggleSearchParamsListEntry,
    clearSearchParamsAndGo,
  };
};

export default useUrlQueryState;
