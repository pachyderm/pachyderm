import {NodeState, JobSetsQuery} from '@graphqlTypes';
import intersection from 'lodash/intersection';
import {useEffect, useMemo} from 'react';
import {useForm} from 'react-hook-form';

import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {useSort, numberComparator, SortableItem} from '@pachyderm/components';

type sortOptionsType = {
  [key: string]: SortableItem<JobSetsQuery['jobSets']['items'][number]>;
};

const sortOptions: sortOptionsType = {
  'Created: Newest': {
    name: 'Created: Newest',
    reverse: true,
    func: numberComparator,
    accessor: (jobset: JobSetsQuery['jobSets']['items'][number]) =>
      jobset?.createdAt || 0,
  },
  'Created: Oldest': {
    name: 'Created: Oldest',
    func: numberComparator,
    accessor: (jobset: JobSetsQuery['jobSets']['items'][number]) =>
      jobset?.createdAt || 0,
  },
};

const jobsetStatusOptions = [
  {name: 'Failed', value: NodeState.ERROR},
  {name: 'Running', value: NodeState.RUNNING},
  {name: 'Success', value: NodeState.SUCCESS},
];

export const jobSetFilters = [
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
    label: 'Filter status',
    name: 'jobsetStatus',
    type: 'checkbox',
    options: jobsetStatusOptions,
  },
];

type FormValues = {
  sortBy: string;
  jobsetStatus: NodeState[];
};

type useJobSetsFiltersProps = {
  jobSets?: JobSetsQuery['jobSets']['items'];
};

const useJobSetFilters = ({jobSets = []}: useJobSetsFiltersProps) => {
  const {searchParams, updateSearchParamsAndGo} = useUrlQueryState();
  const formCtx = useForm<FormValues>({
    mode: 'onChange',
    defaultValues: {
      sortBy: 'Created: Newest',
      jobsetStatus: [],
    },
  });

  const {watch} = formCtx;
  const sortFilter = watch('sortBy');
  const jobsetStatusFilters = watch('jobsetStatus');

  useEffect(() => {
    updateSearchParamsAndGo({
      sortBy: sortFilter,
      jobStatus: jobsetStatusFilters,
    });
  }, [jobsetStatusFilters, sortFilter, updateSearchParamsAndGo]);

  const clearableFiltersMap = useMemo(
    () =>
      jobsetStatusOptions
        .filter((statusEntry) =>
          jobsetStatusFilters.includes(statusEntry.value),
        )
        .map((entry) => ({
          field: 'jobsetStatus',
          name: entry.name,
          value: entry.value.toString(),
        })),
    [jobsetStatusFilters],
  );

  const staticFilterKeys = [sortFilter];

  const filteredJobsets = useMemo(
    () =>
      jobSets?.filter((jobSet) => {
        let included = true;
        const jobStates = jobSet.jobs.map((job) => job.nodeState);

        if (!jobSet?.state) {
          return false;
        }

        if (searchParams.jobStatus && searchParams.jobStatus.length > 0) {
          included = intersection(searchParams.jobStatus, jobStates).length > 0;
        }

        return included;
      }),
    [jobSets, searchParams.jobStatus],
  );

  const {
    sortedData: sortedJobsets,
    setComparator,
    comparatorName,
  } = useSort({
    data: filteredJobsets,
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
    sortedJobsets,
    clearableFiltersMap,
    staticFilterKeys,
  };
};

export default useJobSetFilters;
