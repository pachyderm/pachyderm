import {useMemo} from 'react';

import {usePipelineJobs} from '@dash-frontend/hooks/usePipelineJobs';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {PipelineJobsQueryArgs, PipelineJobState} from '@graphqlTypes';

export type PipelineJobFilters = {
  [key in PipelineJobState]?: boolean;
};

const convertToObject = (JobStateList: PipelineJobState[]) => {
  return Object.values(JobStateList).reduce<PipelineJobFilters>(
    (result, jobState) => {
      result[jobState] = true;
      return result;
    },
    {},
  );
};

const useJobList = ({projectId, pipelineId}: PipelineJobsQueryArgs) => {
  const {pipelineJobs, loading} = usePipelineJobs({projectId, pipelineId});
  const {viewState} = useUrlQueryState();

  const selectedFilters: PipelineJobFilters = useMemo(() => {
    if (viewState && viewState.pipelineJobFilters) {
      return convertToObject(viewState.pipelineJobFilters);
    }
    return convertToObject(Object.values(PipelineJobState));
  }, [viewState]);

  const filteredPipelineJobs = useMemo(() => {
    return pipelineJobs.filter((job) => {
      return selectedFilters[job.state];
    });
  }, [pipelineJobs, selectedFilters]);

  return {
    pipelineJobs,
    filteredPipelineJobs,
    selectedFilters,
    loading,
  };
};

export default useJobList;
