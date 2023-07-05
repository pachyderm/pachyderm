import {DatumState} from '@graphqlTypes';
import {useMemo} from 'react';

import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import useDatum from '@dash-frontend/hooks/useDatum';
import {useJob} from '@dash-frontend/hooks/useJob';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useMiddleSection = () => {
  const {projectId, jobId, pipelineId, datumId} = useUrlState();

  const {pipeline, pipelineType, isSpout} = useCurrentPipeline();

  const {job, loading: loadingJob} = useJob(
    {
      id: jobId,
      pipelineName: pipelineId,
      projectId,
    },
    {skip: !pipelineType || isSpout},
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

  const {headerText, headerValue} = useMemo(() => {
    if (isSpout) {
      return {
        headerText: 'Pipeline logs for',
        headerValue: pipelineId,
      };
    } else if (datumId) {
      return {headerText: 'Datum Logs for', headerValue: datumId};
    } else {
      return {
        headerText: 'Job Logs for',
        headerValue: jobId || job?.id,
      };
    }
  }, [datumId, isSpout, job?.id, jobId, pipelineId]);

  const startTime = isSpout ? pipeline?.createdAt : job?.createdAt;

  const isSkippedDatum = datum?.state === DatumState.SKIPPED;

  return {
    jobId,
    job,
    headerText,
    headerValue,
    startTime,
    loading: loadingJob || loadingDatum,
    isSkippedDatum,
    isSpout,
  };
};

export default useMiddleSection;
