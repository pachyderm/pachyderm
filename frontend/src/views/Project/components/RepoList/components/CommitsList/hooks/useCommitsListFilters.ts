import {Commit, OriginKind} from '@graphqlTypes';
import {useEffect, useMemo} from 'react';
import {useForm} from 'react-hook-form';

import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {useSort, numberComparator, SortableItem} from '@pachyderm/components';

type sortOptionsType = {
  [key: string]: SortableItem<Commit>;
};

const sortOptions: sortOptionsType = {
  'Created: Newest': {
    name: 'Created: Newest',
    reverse: true,
    func: numberComparator,
    accessor: (commit: Commit) => commit?.started || 0,
  },
  'Created: Oldest': {
    name: 'Created: Oldest',
    func: numberComparator,
    accessor: (commit: Commit) => commit?.started || 0,
  },
};

const commitTypeOptions = [
  {name: 'User', value: OriginKind.USER},
  {name: 'Alias', value: OriginKind.ALIAS},
  {name: 'Auto', value: OriginKind.AUTO},
];

export const commitsFilters = [
  {
    label: 'Sort By',
    name: 'sortBy',
    type: 'radio',
    options: Object.entries(sortOptions).map(([_key, option]) => ({
      name: option.name,
      value: option.name,
    })),
  },
  {
    label: 'Filter commit type',
    name: 'commitType',
    type: 'checkbox',
    options: commitTypeOptions,
  },
];

type FormValues = {
  sortBy: string;
  commitType: OriginKind[];
  branchNames: string[];
  commitIds: string[];
};

type useCommitsListFiltersProps = {
  commits?: Commit[];
};

const useCommitsListFilters = ({commits = []}: useCommitsListFiltersProps) => {
  const {viewState, updateViewState, getNewViewState} = useUrlQueryState();
  const formCtx = useForm<FormValues>({
    mode: 'onChange',
    defaultValues: {
      sortBy: 'Created: Newest',
      commitType: [],
      branchNames: [],
      commitIds: [],
    },
  });

  const {watch, reset} = formCtx;
  const sortFilter = watch('sortBy');
  const commitTypes = watch('commitType');
  const branchNames = watch('branchNames');
  const commitIds = watch('commitIds');

  useEffect(() => {
    const {selectedRepos} = viewState;
    reset();
    getNewViewState({
      selectedRepos,
    });
    // We want to clear viewstate and form state on fresh renders
    // but NOT on changes to viewState
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [getNewViewState, reset, updateViewState]);

  useEffect(() => {
    updateViewState({
      sortBy: sortFilter,
      commitType: commitTypes,
      branchName: branchNames,
      commitId: commitIds,
    });
  }, [branchNames, commitIds, commitTypes, sortFilter, updateViewState]);

  const multiselectFilters = [
    {
      label: 'Branch',
      name: 'branchNames',
      noun: 'branch name',
      values: [
        ...new Set(commits?.map((commit) => commit.branch?.name || 'default')),
      ],
    },
    {
      label: 'ID',
      name: 'commitIds',
      noun: 'commit ID',
      formatLabel: (val: string) => `${val.slice(0, 6)}...`,
      values: [...new Set(commits?.map((commit) => commit.id))],
    },
  ];

  const clearableFiltersMap = useMemo(() => {
    const commitTypesMap = commitTypeOptions
      .filter((typeEntry) => commitTypes.includes(typeEntry.value))
      .map((entry) => ({
        field: 'commitType',
        name: entry.name,
        value: entry.value.toString(),
      }));
    const branchNamesMap = branchNames.map((branch) => ({
      field: 'branchNames',
      name: branch,
      value: branch,
    }));
    const commitIdsMap = commitIds.map((id) => ({
      field: 'commitIds',
      name: `${id.slice(0, 6)}...`,
      value: id,
    }));
    return [...commitTypesMap, ...branchNamesMap, ...commitIdsMap];
  }, [branchNames, commitIds, commitTypes]);

  const staticFilterKeys = [sortFilter];
  const filteredCommits = useMemo(
    () =>
      commits?.filter((commit) => {
        let included = true;

        if (viewState.commitType && viewState.commitType.length > 0) {
          included =
            !!commit.originKind &&
            viewState.commitType.includes(commit.originKind);
        }

        if (viewState.branchName && viewState.branchName.length > 0) {
          included =
            !!commit.branch?.name &&
            viewState.branchName.includes(commit.branch.name);
        }

        if (viewState.commitId && viewState.commitId.length > 0) {
          included = viewState.commitId.includes(commit.id);
        }

        return included;
      }),
    [commits, viewState.commitType, viewState.branchName, viewState.commitId],
  );

  const {
    sortedData: sortedCommits,
    setComparator,
    comparatorName,
  } = useSort({
    data: filteredCommits,
    initialSort: sortOptions['Created: Newest'],
    initialDirection: -1,
  });

  useEffect(() => {
    if (viewState.sortBy && comparatorName !== viewState.sortBy) {
      setComparator(sortOptions[viewState.sortBy]);
    }
  });

  return {
    formCtx,
    sortedCommits,
    commitsFilters,
    clearableFiltersMap,
    staticFilterKeys,
    multiselectFilters,
  };
};

export default useCommitsListFilters;
