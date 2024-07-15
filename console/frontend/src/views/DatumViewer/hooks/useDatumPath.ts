import {useHistory} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  logsViewerDatumRoute,
  logsViewerJobRoute,
} from '@dash-frontend/views/Project/utils/routes';

const useDatumPath = () => {
  const {projectId, jobId, pipelineId, datumId} = useUrlState();
  const browserHistory = useHistory();

  const updateSelectedJob = (selectedJobId: string) => {
    browserHistory.push(
      logsViewerJobRoute({
        projectId,
        jobId: selectedJobId,
        pipelineId,
      }),
    );
  };

  const updateSelectedDatum = (
    selectedJobId: string,
    selectedDatumId?: string,
  ) => {
    browserHistory.push(
      logsViewerDatumRoute({
        projectId,
        jobId: selectedJobId,
        pipelineId,
        datumId: selectedDatumId,
      }),
    );
  };

  return {
    urlJobId: jobId,
    updateSelectedJob,
    currentDatumId: datumId,
    updateSelectedDatum,
  };
};

export default useDatumPath;
