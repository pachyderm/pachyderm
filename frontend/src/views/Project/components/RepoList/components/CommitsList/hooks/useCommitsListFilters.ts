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
  const {searchParams, updateSearchParamsAndGo, getNewSearchParamsAndGo} =
    useUrlQueryState();
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
    const {selectedRepos} = searchParams;
    reset();
    getNewSearchParamsAndGo({
      selectedRepos,
    });
    // We want to clear viewstate and form state on fresh renders
    // but NOT on changes to searchParams
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [getNewSearchParamsAndGo, reset, updateSearchParamsAndGo]);

  useEffect(() => {
    updateSearchParamsAndGo({
      sortBy: sortFilter,
      commitType: commitTypes,
      branchName: branchNames,
      commitId: commitIds,
    });
  }, [
    branchNames,
    commitIds,
    commitTypes,
    sortFilter,
    updateSearchParamsAndGo,
  ]);

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

        if (searchParams.commitType && searchParams.commitType.length > 0) {
          included =
            !!commit.originKind &&
            searchParams.commitType.includes(commit.originKind);
        }

        if (searchParams.branchName && searchParams.branchName.length > 0) {
          included =
            !!commit.branch?.name &&
            searchParams.branchName.includes(commit.branch.name);
        }

        if (searchParams.commitId && searchParams.commitId.length > 0) {
          included = searchParams.commitId.includes(commit.id);
        }

        return included;
      }),
    [
      commits,
      searchParams.commitType,
      searchParams.branchName,
      searchParams.commitId,
    ],
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
    if (searchParams.sortBy && comparatorName !== searchParams.sortBy) {
      setComparator(sortOptions[searchParams.sortBy]);
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
