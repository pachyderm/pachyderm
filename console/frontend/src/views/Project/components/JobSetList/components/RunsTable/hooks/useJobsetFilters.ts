import intersection from 'lodash/intersection';
import {useEffect, useMemo} from 'react';
import {useForm} from 'react-hook-form';

import {restJobStateToNodeState} from '@dash-frontend/api/utils/nodeStateMappers';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {getUnixSecondsFromISOString} from '@dash-frontend/lib/dateTime';
import {InternalJobSet, NodeState} from '@dash-frontend/lib/types';
import {useSort, numberComparator, SortableItem} from '@pachyderm/components';

type sortOptionsType = {
  [key: string]: SortableItem<InternalJobSet>;
};

const sortOptions: sortOptionsType = {
  'Created: Newest': {
    name: 'Created: Newest',
    reverse: true,
    func: numberComparator,
    accessor: (jobset: InternalJobSet) =>
      getUnixSecondsFromISOString(jobset?.created),
  },
  'Created: Oldest': {
    name: 'Created: Oldest',
    func: numberComparator,
    accessor: (jobset: InternalJobSet) =>
      getUnixSecondsFromISOString(jobset?.created),
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
  jobSets?: InternalJobSet[];
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
        const jobStates = jobSet.jobs.map((job) =>
          restJobStateToNodeState(job.state),
        );

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
    if (
      searchParams.sortBy &&
      sortOptions[searchParams.sortBy] &&
      comparatorName !== searchParams.sortBy
    ) {
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
