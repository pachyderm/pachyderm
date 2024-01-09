import parse from 'parse-duration';

import {useDatum} from '@dash-frontend/hooks/useDatum';
import {useJob} from '@dash-frontend/hooks/useJob';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {formatDurationFromSeconds} from '@dash-frontend/lib/dateTime';

export const useDatumDetails = () => {
  const {datumId, jobId, pipelineId, projectId} = useUrlState();

  const {datum, loading: loadingDatum} = useDatum({
    datum: {
      id: datumId,
      job: {
        id: jobId,
        pipeline: {
          name: pipelineId,
          project: {
            name: projectId,
          },
        },
      },
    },
  });

  const {job, loading: loadingJob} = useJob({
    id: jobId,
    pipelineName: pipelineId,
    projectId,
  });

  const started = undefined; // TODO: Datum Start time is not yet available from pachyderm.

  const downloadSeconds = parse(datum?.stats?.downloadTime ?? '', 's');
  const uploadSeconds = parse(datum?.stats?.uploadTime ?? '', 's');
  const processSeconds = parse(datum?.stats?.processTime ?? '', 's');
  const runtimeMetrics = [
    {
      label: 'Download',
      duration: formatDurationFromSeconds(downloadSeconds),
      bytes: Number(datum?.stats?.downloadBytes),
    },
    {
      label: 'Upload',
      duration: formatDurationFromSeconds(uploadSeconds),
      bytes: Number(datum?.stats?.uploadBytes),
    },
    {
      label: 'Processing',
      duration: formatDurationFromSeconds(processSeconds),
      bytes: undefined,
    },
  ];

  const cumulativeTime =
    downloadSeconds || uploadSeconds || processSeconds
      ? formatDurationFromSeconds(
          (downloadSeconds || 0) + (uploadSeconds || 0) + (processSeconds || 0),
        )
      : 'N/A';

  const inputSpec = {input: job?.details?.input};

  return {
    datumId,
    datum,
    loading: loadingJob || loadingDatum,
    started,
    cumulativeTime,
    runtimeMetrics,
    inputSpec,
  };
};
