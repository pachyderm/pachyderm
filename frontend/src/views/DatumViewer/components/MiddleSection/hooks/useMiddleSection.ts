import {DatumState} from '@graphqlTypes';
import {useMemo} from 'react';

import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import useDatum from '@dash-frontend/hooks/useDatum';
import {useJob} from '@dash-frontend/hooks/useJob';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useMiddleSection = () => {
  const {projectId, jobId, pipelineId, datumId} = useUrlState();
  const {pipeline, pipelineType, isServiceOrSpout} = useCurrentPipeline();

  const {job, loading: loadingJob} = useJob(
    {
      id: jobId,
      pipelineName: pipelineId,
      projectId,
    },
    {skip: !pipelineType || isServiceOrSpout},
  );

  const {datum, loading: loadingDatum} = useDatum(
    {
      id: datumId,
      jobId,
      pipelineId,
      projectId,
    },
    {skip: datumId === ''},
  );

  const currentJobId = jobId || job?.id;

  const {headerText, headerValue} = useMemo(() => {
    if (isServiceOrSpout) {
      return {
        headerText: 'Pipeline logs for',
        headerValue: pipelineId,
      };
    } else if (datumId) {
      return {headerText: 'Datum Logs for', headerValue: datumId};
    } else {
      return {
        headerText: 'Job Logs for',
        headerValue: currentJobId,
      };
    }
  }, [datumId, isServiceOrSpout, currentJobId, pipelineId]);

  const startTime = isServiceOrSpout ? pipeline?.createdAt : job?.createdAt;

  const isSkippedDatum = datum?.state === DatumState.SKIPPED;

  return {
    jobId: currentJobId,
    job,
    headerText,
    headerValue,
    startTime,
    loading: loadingJob || loadingDatum,
    isSkippedDatum,
    isServiceOrSpout,
  };
};

export default useMiddleSection;
