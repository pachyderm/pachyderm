import {useEffect, useMemo} from 'react';
import {useForm} from 'react-hook-form';

import {JobInfo} from '@dash-frontend/api/pps';
import {restJobStateToNodeState} from '@dash-frontend/api/utils/nodeStateMappers';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {getUnixSecondsFromISOString} from '@dash-frontend/lib/dateTime';
import {NodeState} from '@dash-frontend/lib/types';
import {useSort, numberComparator, SortableItem} from '@pachyderm/components';

type sortOptionsType = {
  [key: string]: SortableItem<JobInfo>;
};

const sortOptions: sortOptionsType = {
  'Created: Newest': {
    name: 'Created: Newest',
    reverse: true,
    func: numberComparator,
    accessor: (job: JobInfo) => getUnixSecondsFromISOString(job?.created),
  },
  'Created: Oldest': {
    name: 'Created: Oldest',
    func: numberComparator,
    accessor: (job: JobInfo) => getUnixSecondsFromISOString(job?.created),
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
  pipelineVersions: string[];
};

type useJobsFiltersProps = {
  jobs?: JobInfo[];
};

const useJobsListFilters = ({jobs = []}: useJobsFiltersProps) => {
  const {searchParams, getNewSearchParamsAndGo} = useUrlQueryState();
  const formCtx = useForm<FormValues>({
    mode: 'onChange',
    defaultValues: {
      sortBy: 'Created: Newest',
      jobStatus: [],
      jobIds: [],
      pipelineSteps: [],
      pipelineVersions: [],
    },
  });

  const {watch} = formCtx;
  const sortFilter = watch('sortBy');
  const jobStatusFilters = watch('jobStatus');
  const jobIdFilters = watch('jobIds');
  const pipelineStepsFilters = watch('pipelineSteps');
  const pipelineVersionsFilters = watch('pipelineVersions');

  useEffect(() => {
    const {selectedPipelines, selectedJobs} = searchParams;
    getNewSearchParamsAndGo({
      sortBy: sortFilter,
      jobStatus: jobStatusFilters,
      jobId: jobIdFilters,
      pipelineStep: pipelineStepsFilters,
      pipelineVersion: pipelineVersionsFilters,
      selectedPipelines,
      selectedJobs,
    });
  }, [
    getNewSearchParamsAndGo,
    jobIdFilters,
    jobStatusFilters,
    pipelineStepsFilters,
    pipelineVersionsFilters,
    searchParams,
    sortFilter,
  ]);

  const multiselectFilters = [
    {
      label: 'ID',
      name: 'jobIds',
      noun: 'job ID',
      formatLabel: (val: string) => `${val.slice(0, 6)}...`,
      values: [...new Set(jobs?.map((job) => job?.job?.id || ''))],
    },
    {
      label: 'Pipeline',
      name: 'pipelineSteps',
      noun: 'step',
      values: [...new Set(jobs?.map((job) => job?.job?.pipeline?.name || ''))],
    },
    {
      label: 'Pipeline Version',
      name: 'pipelineVersions',
      noun: 'version',
      values: [...new Set(jobs?.map((job) => job?.pipelineVersion || ''))],
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
    const pipelineVersions = pipelineVersionsFilters.map((version) => ({
      field: 'pipelineVersions',
      name: version,
      value: version,
    }));
    return [...jobStatuses, ...jobIds, ...pipelineSteps, ...pipelineVersions];
  }, [
    jobIdFilters,
    jobStatusFilters,
    pipelineStepsFilters,
    pipelineVersionsFilters,
  ]);

  const staticFilterKeys = [sortFilter];
  const filteredJobs = useMemo(
    () =>
      jobs?.filter((job) => {
        return (
          (!searchParams.jobStatus ||
            searchParams.jobStatus.includes(
              restJobStateToNodeState(job.state),
            )) &&
          (!searchParams.jobId ||
            searchParams.jobId.includes(job?.job?.id || '')) &&
          (!searchParams.pipelineStep ||
            searchParams.pipelineStep.includes(
              job?.job?.pipeline?.name || '',
            )) &&
          (!searchParams.pipelineVersion ||
            searchParams.pipelineVersion.includes(job?.pipelineVersion || ''))
        );
      }),
    [
      jobs,
      searchParams.jobStatus,
      searchParams.jobId,
      searchParams.pipelineStep,
      searchParams.pipelineVersion,
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
    sortedJobs,
    jobStatusFilters,
    clearableFiltersMap,
    staticFilterKeys,
    multiselectFilters,
  };
};

export default useJobsListFilters;
