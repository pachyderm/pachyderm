import {
  formatDuration,
  formatDistanceStrict,
  fromUnixTime,
  formatDistanceToNowStrict,
} from 'date-fns';
import {useState, useMemo, useCallback} from 'react';

import {useJob} from '@dash-frontend/hooks/useJob';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {Input} from '@dash-frontend/lib/types';

const NANOS = 1e9;

const timestampToSeconds = (timestamp?: {seconds?: number; nanos?: number}) => {
  if (!timestamp) return 0;
  if (!timestamp.nanos) return timestamp.seconds || 0;
  return +Number((timestamp.seconds || 0) + timestamp.nanos / NANOS).toFixed(2);
};

const useInfoPanel = () => {
  const {jobId, projectId, pipelineId} = useUrlState();
  const {viewState} = useUrlQueryState();

  const [runtimeDetailsOpen, setRuntimeDetailsOpen] = useState(false);
  const [runtimeDetailsClosing, setRuntimeDetailsClosing] = useState(false);

  const toggleRunTimeDetailsOpen = () => {
    if (runtimeDetailsOpen) {
      setRuntimeDetailsClosing(true);
      // slide out animation
      setTimeout(() => {
        setRuntimeDetailsOpen(false);
        setRuntimeDetailsClosing(false);
      }, 300);
    } else {
      setRuntimeDetailsOpen(true);
    }
  };

  const {job, loading} = useJob({
    id: viewState.globalIdFilter || jobId,
    pipelineName: pipelineId,
    projectId,
  });

  const started = useMemo(() => {
    return job?.startedAt
      ? formatDistanceToNowStrict(fromUnixTime(job?.startedAt), {
          addSuffix: true,
        })
      : 'N/A';
  }, [job?.startedAt]);

  const duration = useMemo(() => {
    return job?.finishedAt && job?.startedAt
      ? formatDistanceStrict(
          fromUnixTime(job.startedAt),
          fromUnixTime(job.finishedAt),
        )
      : 'N/A';
  }, [job?.startedAt, job?.finishedAt]);

  const jobDetails = useMemo(() => {
    const details = JSON.parse(job?.jsonDetails || '{}');
    return {
      ...details,
      dataTotal: details?.dataTotal,
      datumTries: details?.datumTries,
      pipelineVersion: details?.pipelineVersion,
      salt: details?.salt,
      downloadBytes: details?.stats?.downloadBytes,
      uploadBytes: details?.stats?.uploadBytes,
      downloadTime: timestampToSeconds(details?.stats?.downloadTime),
      processTime: timestampToSeconds(details?.stats?.processTime),
      uploadTime: timestampToSeconds(details?.stats?.uploadTime),
    };
  }, [job?.jsonDetails]);

  const datumMetrics = useMemo(() => {
    return [
      {
        value: job?.dataProcessed,
        label: 'Processed',
      },
      {
        value: job?.dataSkipped,
        label: 'Skipped',
      },
      {
        value: job?.dataFailed,
        label: 'Failed',
      },
      {
        value: job?.dataRecovered,
        label: 'Recovered',
      },
    ];
  }, [job]);

  const runtimeMetrics = useMemo(() => {
    return [
      {
        duration:
          job?.finishedAt && job?.startedAt
            ? formatDistanceStrict(
                fromUnixTime(
                  job.startedAt +
                    jobDetails.downloadTime +
                    jobDetails.processTime +
                    jobDetails.uploadTime,
                ),
                fromUnixTime(job.finishedAt),
              )
            : 'N/A',
        label: 'Setup',
      },
      {
        duration: formatDuration({seconds: jobDetails.downloadTime}),
        bytes: jobDetails?.stats?.downloadBytes,
        label: 'Download',
      },
      {
        duration: formatDuration({seconds: jobDetails.processTime}),
        label: 'Processing',
      },
      {
        duration: formatDuration({seconds: jobDetails.uploadTime}),
        bytes: jobDetails?.stats?.uploadBytes,
        label: 'Upload',
      },
    ];
  }, [jobDetails, job?.finishedAt, job?.startedAt]);

  const getInputRepos = useCallback((input: Input) => {
    const inputs = [];
    if (input?.pfs?.repo) {
      inputs.push(input.pfs.repo);
    }
    input.joinList?.forEach((i) => {
      inputs.push(...getInputRepos(i));
    });
    input.groupList?.forEach((i) => {
      inputs.push(...getInputRepos(i));
    });
    input.crossList?.forEach((i) => {
      inputs.push(...getInputRepos(i));
    });
    input.unionList?.forEach((i) => {
      inputs.push(...getInputRepos(i));
    });

    return inputs;
  }, []);

  const inputs = useMemo(() => {
    if (job?.inputString) {
      const input = JSON.parse(job?.inputString);
      return getInputRepos(input);
    }
    return [];
  }, [job, getInputRepos]);

  const jobConfig = useMemo(() => {
    if (!job) return null;
    else {
      const transform = {...job.transform};
      delete transform.__typename;
      return {
        input: JSON.parse(job.inputString || '{}'),
        transform: JSON.parse(job.transformString || '{}'),
        details: JSON.parse(job.jsonDetails),
      };
    }
  }, [job]);

  return {
    job,
    duration,
    jobDetails,
    loading,
    started,
    runtimeDetailsOpen,
    runtimeDetailsClosing,
    toggleRunTimeDetailsOpen,
    datumMetrics,
    runtimeMetrics,
    inputs,
    jobConfig,
  };
};

export default useInfoPanel;
