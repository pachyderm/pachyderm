import {JobQuery, DatumFilter} from '@graphqlTypes';
import {
  formatDistanceStrict,
  formatDistanceToNowStrict,
  formatDuration,
  fromUnixTime,
} from 'date-fns';
import {useCallback, useMemo} from 'react';

import {useJob} from '@dash-frontend/hooks/useJob';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {timestampToSeconds} from '@dash-frontend/lib/timestampToSeconds';
import {Input} from '@dash-frontend/lib/types';
import {logsViewerDatumRoute} from '@dash-frontend/views/Project/utils/routes';

type infoPanelProps = {
  pipelineJob?: JobQuery['job'];
  pipelineLoading: boolean;
};

const useInfoPanel = ({pipelineJob, pipelineLoading}: infoPanelProps) => {
  const {jobId, projectId, pipelineId} = useUrlState();
  const {getUpdatedSearchParams} = useUrlQueryState();

  const {job: specifiedJob, loading: jobLoading} = useJob(
    {
      id: jobId,
      pipelineName: pipelineId,
      projectId,
    },
    {skip: pipelineLoading || !!pipelineJob},
  );
  const job = pipelineJob || specifiedJob;

  const started = useMemo(() => {
    return job?.startedAt
      ? formatDistanceToNowStrict(fromUnixTime(job?.startedAt), {
          addSuffix: true,
        })
      : 'N/A';
  }, [job]);

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
        label: 'Success',
        filter: DatumFilter.SUCCESS,
      },
      {
        value: job?.dataSkipped,
        label: 'Skipped',
        filter: DatumFilter.SKIPPED,
      },
      {
        value: job?.dataFailed,
        label: 'Failed',
        filter: DatumFilter.FAILED,
      },
      {
        value: job?.dataRecovered,
        label: 'Recovered',
        filter: DatumFilter.RECOVERED,
      },
    ];
  }, [job]);

  let logsDatumRoute = '';
  if ((job?.id || jobId) && pipelineId) {
    logsDatumRoute = logsViewerDatumRoute(
      {
        projectId,
        jobId: job?.id || jobId,
        pipelineId: pipelineId,
      },
      false,
    );
  }

  const addLogsQueryParams = useCallback(
    (path: string, filter: DatumFilter) => {
      return `${path}?${getUpdatedSearchParams({
        datumFilters: [filter],
      })}`;
    },
    [getUpdatedSearchParams],
  );

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
        bytes: undefined,
      },
      {
        duration: formatDuration({seconds: jobDetails.downloadTime}),
        bytes: jobDetails?.stats?.downloadBytes,
        label: 'Download',
      },
      {
        duration: formatDuration({seconds: jobDetails.processTime}),
        label: 'Processing',
        bytes: undefined,
      },
      {
        duration: formatDuration({seconds: jobDetails.uploadTime}),
        bytes: jobDetails?.stats?.uploadBytes,
        label: 'Upload',
      },
    ];
  }, [jobDetails, job?.finishedAt, job?.startedAt]);

  const getInputRepos = useCallback((input: Input) => {
    const inputs: {projectId: string; name: string}[] = [];
    if (input?.pfs?.repo) {
      inputs.push({projectId: input.pfs.project, name: input.pfs.repo});
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
    loading: jobLoading,
    started,
    datumMetrics,
    runtimeMetrics,
    inputs,
    jobConfig,
    logsDatumRoute,
    addLogsQueryParams,
  };
};

export default useInfoPanel;
