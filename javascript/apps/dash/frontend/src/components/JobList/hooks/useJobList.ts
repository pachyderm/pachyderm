import {useEffect, useMemo} from 'react';
import {useForm} from 'react-hook-form';

import {useJobs} from '@dash-frontend/hooks/useJobs';
import {GetJobsQueryVariables, JobState} from '@graphqlTypes';

const defaultValues = Object.values(JobState).reduce<{[key: string]: unknown}>(
  (result, state) => {
    result[state] = true;
    return result;
  },
  {},
);

const useJobList = ({projectId, pipelineId}: GetJobsQueryVariables['args']) => {
  const {jobs, loading} = useJobs({projectId, pipelineId});
  const formCtx = useForm({
    defaultValues,
  });

  const {reset, watch} = formCtx;
  const formValues = watch();

  const filteredJobs = useMemo(() => {
    const activeStates = Object.entries(formValues).reduce<string[]>(
      (result, pair) => {
        const [label, value] = pair;

        if (value) {
          result.push(label);
        }

        return result;
      },
      [],
    );

    return jobs.filter((job) => {
      return activeStates.includes(String(job.state));
    });
  }, [formValues, jobs]);

  useEffect(() => {
    reset(defaultValues);
  }, [reset]);

  return {
    jobs,
    filteredJobs,
    loading,
    formCtx,
  };
};

export default useJobList;
