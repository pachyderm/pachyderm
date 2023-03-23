import {DatumState} from '@graphqlTypes';

import useDatum from '@dash-frontend/hooks/useDatum';
import {useJob} from '@dash-frontend/hooks/useJob';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {formatDurationFromSeconds} from '@dash-frontend/lib/dateTime';

export const useDatumDetails = () => {
  const {datumId, jobId, pipelineId, projectId} = useUrlState();

  const {datum, loading: loadingDatum} = useDatum({
    id: datumId,
    jobId,
    pipelineId,
    projectId,
  });

  const {job, loading: loadingJob} = useJob({
    id: jobId,
    pipelineName: pipelineId,
    projectId,
  });

  const hasPreviousJobId = !!(
    datum?.requestedJobId &&
    datum?.jobId &&
    datum?.requestedJobId !== datum?.jobId
  );

  const skippedWithPreviousJob =
    hasPreviousJobId && datum.state === DatumState.SKIPPED;

  const started = undefined; // TODO: Datum Start time is not yet available from pachyderm.

  const runtimeMetrics = [
    {
      label: 'Download',
      duration: formatDurationFromSeconds(datum?.downloadTimestamp?.seconds),
      bytes: datum?.downloadBytes,
    },
    {
      label: 'Processing',
      duration: formatDurationFromSeconds(datum?.processTimestamp?.seconds),
      bytes: undefined,
    },
    {
      label: 'Upload',
      duration: formatDurationFromSeconds(datum?.uploadTimestamp?.seconds),
      bytes: datum?.uploadBytes,
    },
  ];

  const totalRuntime =
    datum?.downloadTimestamp?.seconds ||
    datum?.processTimestamp?.seconds ||
    datum?.uploadTimestamp?.seconds
      ? formatDurationFromSeconds(
          (datum?.downloadTimestamp?.seconds || 0) +
            (datum?.processTimestamp?.seconds || 0) +
            (datum?.uploadTimestamp?.seconds || 0),
        )
      : 'N/A';

  const inputSpec = {input: JSON.parse(job?.inputString || '{}')};

  return {
    datumId,
    datum,
    loading: loadingJob || loadingDatum,
    started,
    totalRuntime,
    runtimeMetrics,
    inputSpec,
    skippedWithPreviousJob,
    previousJobId: datum?.jobId,
  };
};
