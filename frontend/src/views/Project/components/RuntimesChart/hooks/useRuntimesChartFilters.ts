import {useEffect, useMemo} from 'react';
import {useForm} from 'react-hook-form';

import {JobInfo} from '@dash-frontend/api/pps';
import {MultiselectFilter} from '@dash-frontend/components/TableView/components/TableViewFilters/components/DropdownFilter/DropdownFilter';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';

type FormValues = {
  jobIds: string[];
  pipelineSteps: string[];
};

type useRuntimesChartFiltersProps = {
  jobs?: JobInfo[];
};

const useRuntimesChartFilters = ({jobs = []}: useRuntimesChartFiltersProps) => {
  const {searchParams, getNewSearchParamsAndGo} = useUrlQueryState();
  const formCtx = useForm<FormValues>({
    mode: 'onChange',
    defaultValues: {
      jobIds: [],
      pipelineSteps: [],
    },
  });

  const {watch} = formCtx;
  const jobIdFilters = watch('jobIds');
  const pipelineStepsFilters = watch('pipelineSteps');

  useEffect(() => {
    const {selectedPipelines, selectedJobs} = searchParams;
    getNewSearchParamsAndGo({
      jobId: jobIdFilters,
      pipelineStep: pipelineStepsFilters,
      selectedPipelines,
      selectedJobs,
    });
  }, [
    jobIdFilters,
    pipelineStepsFilters,
    getNewSearchParamsAndGo,
    searchParams,
  ]);

  const multiselectFilters: MultiselectFilter[] = [
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
  ];

  const clearableFiltersMap = useMemo(() => {
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
    return [...jobIds, ...pipelineSteps];
  }, [jobIdFilters, pipelineStepsFilters]);

  const filteredJobs = useMemo(
    () =>
      jobs?.filter((job) => {
        return (
          (!searchParams.jobId ||
            searchParams.jobId.includes(job?.job?.id || '')) &&
          (!searchParams.pipelineStep ||
            searchParams.pipelineStep.includes(job?.job?.pipeline?.name || ''))
        );
      }),
    [jobs, searchParams.jobId, searchParams.pipelineStep],
  );

  return {
    formCtx,
    filteredJobs,
    clearableFiltersMap,
    multiselectFilters,
  };
};

export default useRuntimesChartFilters;
