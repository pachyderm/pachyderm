import {NodeState, JobSetsQuery} from '@graphqlTypes';
import intersection from 'lodash/intersection';
import {useEffect, useMemo} from 'react';
import {useForm} from 'react-hook-form';

import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {useSort, numberComparator, SortableItem} from '@pachyderm/components';

type sortOptionsType = {
  [key: string]: SortableItem<JobSetsQuery['jobSets'][number]>;
};

const sortOptions: sortOptionsType = {
  'Created: Newest': {
    name: 'Created: Newest',
    reverse: true,
    func: numberComparator,
    accessor: (jobset: JobSetsQuery['jobSets'][number]) =>
      jobset?.createdAt || 0,
  },
  'Created: Oldest': {
    name: 'Created: Oldest',
    func: numberComparator,
    accessor: (jobset: JobSetsQuery['jobSets'][number]) =>
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
  jobSets?: JobSetsQuery['jobSets'];
};

const useJobSetFilters = ({jobSets = []}: useJobSetsFiltersProps) => {
  const {viewState, updateViewState, clearViewState} = useUrlQueryState();
  const formCtx = useForm<FormValues>({
    mode: 'onChange',
    defaultValues: {
      sortBy: 'Created: Newest',
      jobsetStatus: [],
    },
  });

  const {watch, reset} = formCtx;
  const sortFilter = watch('sortBy');
  const jobsetStatusFilters = watch('jobsetStatus');

  useEffect(() => {
    clearViewState();
    reset();
  }, [clearViewState, reset]);

  useEffect(() => {
    updateViewState({
      sortBy: sortFilter,
      jobStatus: jobsetStatusFilters,
    });
  }, [jobsetStatusFilters, sortFilter, updateViewState]);

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

        if (viewState.jobStatus && viewState.jobStatus.length > 0) {
          included = intersection(viewState.jobStatus, jobStates).length > 0;
        }

        return included;
      }),
    [jobSets, viewState.jobStatus],
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
    if (viewState.sortBy && comparatorName !== viewState.sortBy) {
      setComparator(sortOptions[viewState.sortBy]);
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
