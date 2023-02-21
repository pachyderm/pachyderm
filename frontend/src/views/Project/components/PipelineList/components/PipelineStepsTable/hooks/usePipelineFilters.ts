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
  const {viewState, updateViewState, clearViewState} = useUrlQueryState();
  const formCtx = useForm<FormValues>({
    mode: 'onChange',
    defaultValues: {
      sortBy: 'Created: Newest',
      jobStatus: [],
      pipelineStatus: [],
    },
  });

  const {watch, reset} = formCtx;
  const sortFilter = watch('sortBy');
  const jobStatusFilter = watch('jobStatus');
  const pipelineStatusFilter = watch('pipelineStatus');

  useEffect(() => {
    clearViewState();
    reset();
  }, [clearViewState, reset]);

  useEffect(() => {
    updateViewState({
      sortBy: sortFilter,
      jobStatus: jobStatusFilter,
      pipelineState: pipelineStatusFilter,
    });
  }, [jobStatusFilter, pipelineStatusFilter, sortFilter, updateViewState]);

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
        if (viewState.pipelineState && viewState.pipelineState.length > 0) {
          included = viewState.pipelineState.includes(pipeline.nodeState);
        }
        if (viewState.jobStatus && viewState.jobStatus.length > 0) {
          included =
            !!pipeline?.lastJobNodeState &&
            viewState.jobStatus.includes(pipeline.lastJobNodeState);
        }
        return included;
      }),
    [pipelines, viewState.jobStatus, viewState.pipelineState],
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
    if (viewState.sortBy && comparatorName !== viewState.sortBy) {
      setComparator(sortOptions[viewState.sortBy]);
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
