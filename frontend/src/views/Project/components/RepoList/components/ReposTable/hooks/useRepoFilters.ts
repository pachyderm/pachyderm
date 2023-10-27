import {ReposWithCommitQuery} from '@graphqlTypes';
import {useEffect} from 'react';
import {useForm} from 'react-hook-form';

import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {
  useSort,
  numberComparator,
  stringComparator,
  SortableItem,
} from '@pachyderm/components';

type sortOptionsType = {
  [key: string]: SortableItem<ReposWithCommitQuery['repos'][number] | null>;
};

const sortOptions: sortOptionsType = {
  'Created: Newest': {
    name: 'Created: Newest',
    reverse: true,
    func: numberComparator,
    accessor: (repo: ReposWithCommitQuery['repos'][number]) =>
      repo?.createdAt || 0,
  },
  'Created: Oldest': {
    name: 'Created: Oldest',
    func: numberComparator,
    accessor: (repo: ReposWithCommitQuery['repos'][number]) =>
      repo?.createdAt || 0,
  },
  'Alphabetical: A-Z': {
    name: 'Alphabetical: A-Z',
    func: stringComparator,
    accessor: (repo: ReposWithCommitQuery['repos'][number]) => repo?.id || '',
  },
  'Alphabetical: Z-A': {
    name: 'Alphabetical: Z-A',
    reverse: true,
    func: stringComparator,
    accessor: (repo: ReposWithCommitQuery['repos'][number]) => repo?.id || '',
  },
  Size: {
    name: 'Size',
    func: numberComparator,
    accessor: (repo: ReposWithCommitQuery['repos'][number]) =>
      repo?.sizeBytes || 0,
  },
};

export const repoFilters = [
  {
    label: 'Sort By',
    name: 'sortBy',
    type: 'radio',
    options: Object.entries(sortOptions).map(([, option]) => ({
      name: option.name,
      value: option.name,
    })),
  },
];

type FormValues = {
  sortBy: string;
};

type useRepoFiltersProps = {
  repos?: ReposWithCommitQuery['repos'];
};

const useRepoFilters = ({repos = []}: useRepoFiltersProps) => {
  const {searchParams, updateSearchParamsAndGo} = useUrlQueryState();
  const formCtx = useForm<FormValues>({
    mode: 'onChange',
    defaultValues: {
      sortBy: 'Created: Newest',
    },
  });

  const {watch} = formCtx;
  const sortFilter = watch('sortBy');

  useEffect(() => {
    updateSearchParamsAndGo({
      sortBy: sortFilter,
    });
  }, [sortFilter, updateSearchParamsAndGo]);

  const staticFilterKeys = [sortFilter];

  const {
    sortedData: sortedRepos,
    setComparator,
    comparatorName,
  } = useSort({
    data: repos,
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
    sortedRepos,
    staticFilterKeys,
  };
};

export default useRepoFilters;
