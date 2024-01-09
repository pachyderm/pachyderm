import {useMemo} from 'react';

import {DatumState} from '@dash-frontend/api/pps';
import {useCurrentPipeline} from '@dash-frontend/hooks/useCurrentPipeline';
import {useDatum} from '@dash-frontend/hooks/useDatum';
import {useJob} from '@dash-frontend/hooks/useJob';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useMiddleSection = () => {
  const {projectId, jobId, pipelineId, datumId} = useUrlState();
  const {
    pipeline,
    pipelineType,
    isServiceOrSpout,
    loading: pipelineLoading,
  } = useCurrentPipeline();

  const {job, loading: loadingJob} = useJob(
    {
      id: jobId,
      pipelineName: pipelineId,
      projectId,
    },
    !!pipelineType && !isServiceOrSpout,
  );

  const {datum, loading: loadingDatum} = useDatum(
    {
      datum: {
        id: datumId,
        job: {
          id: jobId,
          pipeline: {
            name: pipelineId,
            project: {name: projectId},
          },
        },
      },
    },
    datumId !== '',
  );

  const currentJobId = jobId || job?.job?.id;

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

  const startTime = isServiceOrSpout
    ? pipeline?.details?.createdAt
    : job?.created;

  const isSkippedDatum = datum?.state === DatumState.SKIPPED;

  return {
    jobId: currentJobId,
    headerText,
    headerValue,
    startTime,
    loading: loadingJob || loadingDatum || pipelineLoading,
    isSkippedDatum,
    isServiceOrSpout,
  };
};

export default useMiddleSection;
