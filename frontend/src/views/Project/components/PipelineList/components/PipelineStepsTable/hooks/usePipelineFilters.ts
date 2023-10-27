import {Pipeline, NodeState, PipelinesQuery} from '@graphqlTypes';
import {useEffect, useMemo} from 'react';
import {useForm} from 'react-hook-form';

import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {
  useSort,
  stringComparator,
  numberComparator,
  SortableItem,
} from '@pachyderm/components';

type sortOptionsType = {
  [key: string]: SortableItem<Pipeline | null>;
};

const sortOptions: sortOptionsType = {
  'Created: Newest': {
    name: 'Created: Newest',
    reverse: true,
    func: numberComparator,
    accessor: (pipeline: Pipeline | null) => pipeline?.createdAt || 0,
  },
  'Created: Oldest': {
    name: 'Created: Oldest',
    func: numberComparator,
    accessor: (pipeline: Pipeline | null) => pipeline?.createdAt || 0,
  },
  'Alphabetical: A-Z': {
    name: 'Alphabetical: A-Z',
    func: stringComparator,
    accessor: (pipeline: Pipeline | null) => pipeline?.name || '',
  },
  'Alphabetical: Z-A': {
    name: 'Alphabetical: Z-A',
    reverse: true,
    func: stringComparator,
    accessor: (pipeline: Pipeline | null) => pipeline?.name || '',
  },
  'Job Status': {
    name: 'Job Status',
    reverse: true,
    func: stringComparator,
    accessor: (pipeline: Pipeline | null) => pipeline?.lastJobState || '',
  },
};

const jobStateOptions = [
  {name: 'Success', value: NodeState.SUCCESS},
  {name: 'Running', value: NodeState.RUNNING},
  {name: 'Error', value: NodeState.ERROR},
];

const pipelineStateOptions = [
  {name: 'Idle', value: NodeState.IDLE},
  {name: 'Paused', value: NodeState.PAUSED},
  {name: 'Busy', value: NodeState.BUSY},
  {name: 'Error', value: NodeState.ERROR},
];

export const pipelineFilters = [
  {
    label: 'Sort By',
    name: 'sortBy',
    type: 'radio',
    options: Object.entries(sortOptions).map(([, option]) => ({
      name: option.name,
      value: option.name,
    })),
  },
  {
    label: 'Pipeline state',
    name: 'pipelineStatus',
    type: 'checkbox',
    options: pipelineStateOptions,
  },
  {
    label: 'Last job status',
    name: 'jobStatus',
    type: 'checkbox',
    options: jobStateOptions,
  },
];

type FormValues = {
  sortBy: string;
  jobStatus: NodeState[];
  pipelineStatus: NodeState[];
};

type usePipelineFiltersProps = {
  pipelines?: PipelinesQuery['pipelines'];
};

const usePipelineFilters = ({pipelines = []}: usePipelineFiltersProps) => {
  const {searchParams, updateSearchParamsAndGo} = useUrlQueryState();
  const formCtx = useForm<FormValues>({
    mode: 'onChange',
    defaultValues: {
      sortBy: 'Created: Newest',
      jobStatus: [],
      pipelineStatus: [],
    },
  });

  const {watch} = formCtx;
  const sortFilter = watch('sortBy');
  const jobStatusFilter = watch('jobStatus');
  const pipelineStatusFilter = watch('pipelineStatus');

  useEffect(() => {
    updateSearchParamsAndGo({
      sortBy: sortFilter,
      jobStatus: jobStatusFilter,
      pipelineState: pipelineStatusFilter,
    });
  }, [
    jobStatusFilter,
    pipelineStatusFilter,
    sortFilter,
    updateSearchParamsAndGo,
  ]);

  const jobStatusFiltersMap = useMemo(
    () =>
      jobStateOptions
        .filter((statusEntry) => jobStatusFilter.includes(statusEntry.value))
        .map((entry) => ({
          field: 'jobStatus',
          name: entry.name,
          value: entry.value.toString(),
        })),
    [jobStatusFilter],
  );
  const pipelineStatusFiltersMap = useMemo(
    () =>
      pipelineStateOptions
        .filter((statusEntry) =>
          pipelineStatusFilter.includes(statusEntry.value),
        )
        .map((entry) => ({
          field: 'pipelineStatus',
          name: entry.name,
          value: entry.value.toString(),
        })),
    [pipelineStatusFilter],
  );

  const clearableFiltersMap = useMemo(
    () => [...jobStatusFiltersMap, ...pipelineStatusFiltersMap],
    [jobStatusFiltersMap, pipelineStatusFiltersMap],
  );

  const staticFilterKeys = [sortFilter];

  const filteredPipelines = useMemo(
    () =>
      pipelines?.filter((pipeline) => {
        let included = Boolean(pipeline);
        if (!pipeline?.state) {
          return false;
        }
        if (
          searchParams.pipelineState &&
          searchParams.pipelineState.length > 0
        ) {
          included = searchParams.pipelineState.includes(pipeline.nodeState);
        }
        if (searchParams.jobStatus && searchParams.jobStatus.length > 0) {
          included =
            !!pipeline?.lastJobNodeState &&
            searchParams.jobStatus.includes(pipeline.lastJobNodeState);
        }
        return included;
      }),
    [pipelines, searchParams.jobStatus, searchParams.pipelineState],
  );

  const {
    sortedData: sortedPipelines,
    setComparator,
    comparatorName,
  } = useSort({
    data: filteredPipelines,
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
    sortedPipelines,
    clearableFiltersMap,
    staticFilterKeys,
  };
};

export default usePipelineFilters;
