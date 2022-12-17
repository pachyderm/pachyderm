import {DatumState} from '@graphqlTypes';

import useDatum from '@dash-frontend/hooks/useDatum';
import {useJob} from '@dash-frontend/hooks/useJob';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useMiddleSection = () => {
  const {projectId, jobId, pipelineId, datumId} = useUrlState();

  const {job, loading: loadingJob} = useJob({
    id: jobId,
    pipelineName: pipelineId,
    projectId,
  });

  const {datum, loading: loadingDatum} = useDatum(
    {
      id: datumId,
      jobId,
      pipelineId,
      projectId,
    },
    {skip: datumId === ''},
  );

  const headerText = datumId ? 'Datum Logs for' : 'Job Logs for';
  const startTime = job?.createdAt || undefined;
  const isSkippedDatum = datum?.state === DatumState.SKIPPED;

  return {
    jobId,
    datumId,
    headerText,
    startTime,
    loading: loadingJob || loadingDatum,
    isSkippedDatum,
  };
};

export default useMiddleSection;
