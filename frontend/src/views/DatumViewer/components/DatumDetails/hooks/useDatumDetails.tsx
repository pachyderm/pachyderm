import {DatumState, Timestamp} from '@graphqlTypes';
import {formatDuration} from 'date-fns';

import useDatum from '@dash-frontend/hooks/useDatum';
import {useJob} from '@dash-frontend/hooks/useJob';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {timestampToSeconds} from '@dash-frontend/lib/timestampToSeconds';

export const _deriveDuration = (times: Array<Timestamp | null | undefined>) => {
  const derivedDurationTimestamp = times.reduce<{
    seconds: number;
    nanos: number;
  }>(
    (acc, curr) => ({
      seconds: acc?.seconds + (curr?.seconds ?? 0),
      nanos: acc?.nanos + (curr?.nanos ?? 0),
    }),
    {seconds: 0, nanos: 0},
  );

  const derivedDuration = formatDuration({
    seconds: timestampToSeconds(derivedDurationTimestamp),
  });

  return derivedDuration;
};

export const _processDuration = (timestamp?: Timestamp | null) => {
  return (
    formatDuration({
      seconds: timestampToSeconds(timestamp || {}),
    }) ||
    (timestamp && '< 0.01 seconds') ||
    'N/A'
  );
};

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
      duration: _processDuration(datum?.downloadTimestamp),
      bytes: datum?.downloadBytes,
    },
    {
      label: 'Processing',
      duration: _processDuration(datum?.processTimestamp),
      bytes: undefined,
    },
    {
      label: 'Upload',
      duration: _processDuration(datum?.uploadTimestamp),
      bytes: datum?.uploadBytes,
    },
  ];

  const duration = _deriveDuration([
    datum?.downloadTimestamp,
    datum?.processTimestamp,
    datum?.uploadTimestamp,
  ]);

  const inputSpec = {input: JSON.parse(job?.inputString || '{}')};

  return {
    datumId,
    datum,
    loading: loadingJob || loadingDatum,
    started,
    duration,
    runtimeMetrics,
    inputSpec,
    skippedWithPreviousJob,
    previousJobId: datum?.jobId,
  };
};
