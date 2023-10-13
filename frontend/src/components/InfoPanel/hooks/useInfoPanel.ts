import {DatumFilter} from '@graphqlTypes';
import {useCallback, useMemo} from 'react';

import {useJob} from '@dash-frontend/hooks/useJob';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  getStandardDate,
  formatDurationFromSeconds,
} from '@dash-frontend/lib/dateTime';
import {Input} from '@dash-frontend/lib/types';
import {logsViewerDatumRoute} from '@dash-frontend/views/Project/utils/routes';

const useInfoPanel = () => {
  const {jobId, projectId, pipelineId} = useUrlState();
  const {getUpdatedSearchParams} = useUrlQueryState();

  const {job, loading: jobLoading} = useJob({
    id: jobId,
    pipelineName: pipelineId,
    projectId,
  });

  const started = useMemo(() => {
    return job?.startedAt ? getStandardDate(job?.startedAt) : 'N/A';
  }, [job]);

  const created = useMemo(() => {
    return job?.createdAt ? getStandardDate(job?.createdAt) : 'N/A';
  }, [job]);

  const totalRuntime = useMemo(() => {
    return job?.finishedAt && job?.startedAt
      ? formatDurationFromSeconds(job.finishedAt - job.startedAt)
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
      downloadTime: details?.stats?.downloadTime?.seconds,
      processTime: details?.stats?.processTime?.seconds,
      uploadTime: details?.stats?.uploadTime?.seconds,
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
            ? formatDurationFromSeconds(
                job.finishedAt -
                  (job.startedAt +
                    jobDetails.downloadTime +
                    jobDetails.processTime +
                    jobDetails.uploadTime),
              )
            : 'N/A',
        label: 'Setup',
        bytes: undefined,
      },
      {
        duration: formatDurationFromSeconds(jobDetails.downloadTime),
        bytes: jobDetails?.stats?.downloadBytes,
        label: 'Download',
      },
      {
        duration: formatDurationFromSeconds(jobDetails.processTime),
        label: 'Processing',
        bytes: undefined,
      },
      {
        duration: formatDurationFromSeconds(jobDetails.uploadTime),
        bytes: jobDetails?.stats?.uploadBytes,
        label: 'Upload',
      },
    ];
  }, [jobDetails, job?.finishedAt, job?.startedAt]);

  const getInputRepos = useCallback((input: Input) => {
    const inputs: {projectId: string; name: string}[] = [];
    if (input.pfs?.repo) {
      inputs.push({projectId: input.pfs.project, name: input.pfs.repo});
    }
    input.join.forEach((i) => {
      inputs.push(...getInputRepos(i));
    });
    input.group.forEach((i) => {
      inputs.push(...getInputRepos(i));
    });
    input.cross.forEach((i) => {
      inputs.push(...getInputRepos(i));
    });
    input.union.forEach((i) => {
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
    totalRuntime,
    jobDetails,
    jobLoading,
    started,
    created,
    datumMetrics,
    runtimeMetrics,
    inputs,
    jobConfig,
    logsDatumRoute,
    addLogsQueryParams,
  };
};

export default useInfoPanel;
