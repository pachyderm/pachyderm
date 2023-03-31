import {JobsQuery} from '@graphqlTypes';
import {useEffect, useMemo} from 'react';
import {useForm} from 'react-hook-form';

import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';

type FormValues = {
  jobIds: string[];
  pipelineSteps: string[];
};

type useRuntimesChartFiltersProps = {
  jobs?: JobsQuery['jobs']['items'];
};

const useRuntimesChartFilters = ({jobs = []}: useRuntimesChartFiltersProps) => {
  const {searchParams, updateSearchParamsAndGo, getNewSearchParamsAndGo} =
    useUrlQueryState();
  const formCtx = useForm<FormValues>({
    mode: 'onChange',
    defaultValues: {
      jobIds: [],
      pipelineSteps: [],
    },
  });

  const {watch, reset} = formCtx;
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
    // but NOT when viewState changes
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [getNewSearchParamsAndGo, reset]);

  useEffect(() => {
    updateSearchParamsAndGo({
      jobId: jobIdFilters,
      pipelineStep: pipelineStepsFilters,
    });
  }, [jobIdFilters, pipelineStepsFilters, updateSearchParamsAndGo]);

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
          (!searchParams.jobId || searchParams.jobId.includes(job.id)) &&
          (!searchParams.pipelineStep ||
            searchParams.pipelineStep.includes(job.pipelineName))
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
