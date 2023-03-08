import {NodeState, JobsQuery} from '@graphqlTypes';
import {useEffect, useMemo} from 'react';
import {useForm} from 'react-hook-form';

import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {useSort, numberComparator, SortableItem} from '@pachyderm/components';

type sortOptionsType = {
  [key: string]: SortableItem<JobsQuery['jobs'][number]>;
};

const sortOptions: sortOptionsType = {
  'Created: Newest': {
    name: 'Created: Newest',
    reverse: true,
    func: numberComparator,
    accessor: (job: JobsQuery['jobs'][number]) => job?.createdAt || 0,
  },
  'Created: Oldest': {
    name: 'Created: Oldest',
    func: numberComparator,
    accessor: (job: JobsQuery['jobs'][number]) => job?.createdAt || 0,
  },
};

const jobStatusOptions = [
  {name: 'Failed', value: NodeState.ERROR},
  {name: 'Running', value: NodeState.RUNNING},
  {name: 'Success', value: NodeState.SUCCESS},
];

export const jobsFilters = [
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
    name: 'jobStatus',
    type: 'checkbox',
    options: jobStatusOptions,
  },
];

type FormValues = {
  sortBy: string;
  jobStatus: NodeState[];
  jobIds: string[];
  pipelineSteps: string[];
};

type useJobsFiltersProps = {
  jobs?: JobsQuery['jobs'];
};

const useJobsListFilters = ({jobs = []}: useJobsFiltersProps) => {
  const {searchParams, updateSearchParamsAndGo, getNewSearchParamsAndGo} =
    useUrlQueryState();
  const formCtx = useForm<FormValues>({
    mode: 'onChange',
    defaultValues: {
      sortBy: 'Created: Newest',
      jobStatus: [],
      jobIds: [],
      pipelineSteps: [],
    },
  });

  const {watch, reset} = formCtx;
  const sortFilter = watch('sortBy');
  const jobStatusFilters = watch('jobStatus');
  const jobIdFilters = watch('jobIds');
  const pipelineStepsFilters = watch('pipelineSteps');

  useEffect(() => {
    const {selectedPipelines, selectedJobs} = searchParams;
    reset();
    getNewSearchParamsAndGo({
      selectedPipelines,
      selectedJobs,
    });
    // We want to clear the form and viewstate on a fresh render,
    // but NOT when searchParams changes
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [getNewSearchParamsAndGo, reset, updateSearchParamsAndGo]);

  useEffect(() => {
    updateSearchParamsAndGo({
      sortBy: sortFilter,
      jobStatus: jobStatusFilters,
      jobId: jobIdFilters,
      pipelineStep: pipelineStepsFilters,
    });
  }, [
    jobIdFilters,
    jobStatusFilters,
    pipelineStepsFilters,
    sortFilter,
    updateSearchParamsAndGo,
  ]);

  const multiselectFilters = [
    {
      label: 'ID',
      name: 'jobIds',
      noun: 'job ID',
      formatLabel: (val: string) => `${val.slice(0, 6)}...`,
      values: [...new Set(jobs?.map((job) => job.id))],
    },
    {
      label: 'Pipeline',
      name: 'pipelineSteps',
      noun: 'step',
      values: [...new Set(jobs?.map((job) => job.pipelineName))],
    },
  ];

  const clearableFiltersMap = useMemo(() => {
    const jobStatuses = jobStatusOptions
      .filter((statusEntry) => jobStatusFilters.includes(statusEntry.value))
      .map((entry) => ({
        field: 'jobStatus',
        name: entry.name,
        value: entry.value.toString(),
      }));
    const jobIds = jobIdFilters.map((id) => ({
      field: 'jobIds',
      name: `${id.slice(0, 6)}...`,
      value: id,
    }));
    const pipelineSteps = pipelineStepsFilters.map((step) => ({
      field: 'pipelineSteps',
      name: step,
      value: step,
    }));
    return [...jobStatuses, ...jobIds, ...pipelineSteps];
  }, [jobIdFilters, jobStatusFilters, pipelineStepsFilters]);

  const staticFilterKeys = [sortFilter];
  const filteredJobs = useMemo(
    () =>
      jobs?.filter((job) => {
        return (
          (!searchParams.jobStatus ||
            searchParams.jobStatus.includes(job.nodeState)) &&
          (!searchParams.jobId || searchParams.jobId.includes(job.id)) &&
          (!searchParams.pipelineStep ||
            searchParams.pipelineStep.includes(job.pipelineName))
        );
      }),
    [
      jobs,
      searchParams.jobStatus,
      searchParams.jobId,
      searchParams.pipelineStep,
    ],
  );

  const {
    sortedData: sortedJobs,
    setComparator,
    comparatorName,
  } = useSort({
    data: filteredJobs,
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
    sortedJobs,
    jobStatusFilters,
    clearableFiltersMap,
    staticFilterKeys,
    multiselectFilters,
  };
};

export default useJobsListFilters;
