import {useMemo, useState} from 'react';
import {useForm} from 'react-hook-form';

import {JobState} from '@dash-frontend/api/pps';
import {useJobs} from '@dash-frontend/hooks/useJobs';
import useUrlState from '@dash-frontend/hooks/useUrlState';

export type DatumFilterFormValues = {
  jobs: string;
};

export const jobSortOrder = (state?: JobState) => {
  switch (state) {
    case JobState.JOB_SUCCESS:
    case JobState.JOB_CREATED:
      return 3;
    case JobState.JOB_EGRESSING:
    case JobState.JOB_RUNNING:
    case JobState.JOB_STARTING:
    case JobState.JOB_FINISHING:
      return 2;
    case JobState.JOB_FAILURE:
    case JobState.JOB_KILLED:
    case JobState.JOB_UNRUNNABLE:
      return 1;
    case JobState.JOB_STATE_UNKNOWN:
    default:
      return 4;
  }
};

const useLeftPanel = () => {
  const [isExpanded, setIsExpanded] = useState(false);

  const {projectId, pipelineId} = useUrlState();

  const formCtx = useForm<DatumFilterFormValues>({
    mode: 'onChange',
    defaultValues: {
      jobs: 'newest',
    },
  });

  const {jobs, loading} = useJobs({
    limit: 30,
    pipelineIds: [pipelineId],
    projectName: projectId,
  });

  const {watch} = formCtx;
  const jobsSort = watch('jobs');

  const sortedJobs = useMemo(
    () =>
      jobs && jobsSort === 'status'
        ? [...jobs].sort((a, b) => {
            const aState = jobSortOrder(a.state);
            const bState = jobSortOrder(b.state);
            return aState - bState;
          })
        : jobs,
    [jobs, jobsSort],
  );

  return {
    isExpanded,
    setIsExpanded,
    jobs: sortedJobs,
    loading,
    formCtx,
  };
};

export default useLeftPanel;
